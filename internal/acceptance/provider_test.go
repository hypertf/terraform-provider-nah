package acceptance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hypertf/nahcloud/api"
	"github.com/hypertf/nahcloud/domain"
	"github.com/hypertf/nahcloud/service"
	"github.com/hypertf/nahcloud/storage/sqlite"
	"github.com/hypertf/terraform-provider-nah/internal/client"
)

// TestAccCLI uses the built provider binary, a real CLI, the upstream HTTP router,
// and a disposable SQLite database by default. NAH_ACC_ENDPOINT explicitly opts
// into a remote deployment, with a fresh organization for each CLI.
func TestAccCLI(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("set TF_ACC=1 to run real CLI acceptance tests")
	}
	bin := t.TempDir()
	build := exec.CommandContext(t.Context(), "go", "build", "-o", filepath.Join(bin, "terraform-provider-nah"), "../..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build provider: %v\n%s", err, output)
	}
	clis := []string{"terraform", "tofu"}
	if cli := os.Getenv("NAH_TEST_CLI"); cli != "" {
		clis = []string{cli}
	}
	for _, cli := range clis {
		t.Run(filepath.Base(cli), func(t *testing.T) { testLifecycle(t, cli, bin) })
	}
}

func testLifecycle(t *testing.T, cli, bin string) {
	t.Helper()
	endpoint, org := testAPI(t)
	c := client.NewClient(endpoint, org.APIKey.Token)
	dir := t.TempDir()
	write := func(name, value string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("cli.tfrc", fmt.Sprintf("provider_installation {\n dev_overrides {\n \"registry.terraform.io/hypertf/nah\" = %q\n }\n direct {}\n}\n", bin))
	env := append(os.Environ(), "TF_CLI_CONFIG_FILE="+filepath.Join(dir, "cli.tfrc"), "TF_IN_AUTOMATION=1", "TF_INPUT=0", "NAH_ENDPOINT="+endpoint, "NAH_TOKEN="+org.APIKey.Token)
	run := func(want int, args ...string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, cli, args...)
		cmd.Dir, cmd.Env = dir, env
		output, err := cmd.CombinedOutput()
		code := 0
		if err != nil {
			if e, ok := err.(*exec.ExitError); ok {
				code = e.ExitCode()
			} else {
				t.Fatal(err)
			}
		}
		if code != want {
			t.Fatalf("%s %v: exit %d, want %d\n%s", cli, args, code, want, strings.ReplaceAll(string(output), org.APIKey.Token, "[REDACTED]"))
		}
		return output
	}
	t.Cleanup(func() {
		if !t.Failed() {
			return
		}
		// Destroy only this test's state, even if the failure left invalid HCL.
		write("main.tf", strings.Split(config(1), `resource "nah_project"`)[0])
		write("missing.tf", "")
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, cli, "destroy", "-auto-approve", "-no-color")
		cmd.Dir, cmd.Env = dir, env
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("cleanup failed: %v\n%s", err, strings.ReplaceAll(string(output), org.APIKey.Token, "[REDACTED]"))
		}
	})
	state := func() map[string]map[string]any {
		t.Helper()
		var s struct {
			Values struct {
				RootModule struct {
					Resources []struct {
						Address string
						Values  map[string]any
					}
				} `json:"root_module"`
			}
		}
		if err := json.Unmarshal(run(0, "show", "-json"), &s); err != nil {
			t.Fatal(err)
		}
		out := map[string]map[string]any{}
		for _, r := range s.Values.RootModule.Resources {
			out[r.Address] = r.Values
		}
		return out
	}
	id := func(s map[string]map[string]any, name string) string {
		t.Helper()
		value, ok := s["nah_"+name+".test"]["id"].(string)
		if !ok || value == "" {
			t.Fatalf("%s has no string ID", name)
		}
		return value
	}
	check := func(s map[string]map[string]any, address, field string, want any) {
		t.Helper()
		if s[address][field] != want {
			if field == "token" {
				t.Fatalf("%s.token differs from expected token", address)
			}
			t.Fatalf("%s.%s: got %v, want %v", address, field, s[address][field], want)
		}
	}
	write("main.tf", config(1))
	run(0, "validate", "-no-color")
	run(0, "apply", "-auto-approve", "-no-color")
	run(0, "plan", "-detailed-exitcode", "-no-color")
	initial := state()
	check(initial, "nah_instance.test", "cpu", float64(2))
	check(initial, "nah_instance.test", "memory_mb", float64(1536))
	check(initial, "data.nah_instance.test", "region", "eu-west-1")
	check(initial, "data.nah_project.test", "name", "Project 1")
	check(initial, "data.nah_bucket.test", "name", "assets")
	check(initial, "data.nah_object.test", "content", "b25l")
	check(initial, "data.nah_metadata.test", "value", "one")
	check(initial, "data.nah_organization.test", "slug", org.Slug)
	if token, ok := initial["nah_api_key.test"]["token"].(string); !ok || token == "" {
		t.Fatal("missing created API token")
	}

	// Mutable updates must retain IDs and converge to an empty plan.
	write("main.tf", config(2))
	run(0, "apply", "-auto-approve", "-no-color")
	run(0, "plan", "-detailed-exitcode", "-no-color")
	updated := state()
	for _, name := range []string{"project", "instance", "bucket", "object", "metadata", "api_key"} {
		check(updated, "nah_"+name+".test", "id", initial["nah_"+name+".test"]["id"])
	}
	check(updated, "nah_instance.test", "cpu", float64(3))
	check(updated, "nah_instance.test", "status", "stopped")
	check(updated, "nah_bucket.test", "name", "assets-renamed")
	check(updated, "nah_metadata.test", "value", "")
	check(updated, "nah_object.test", "path", "nested/new.json")
	check(updated, "nah_api_key.test", "token", initial["nah_api_key.test"]["token"])

	// Exercise every import with real CLI state removal, refresh, and an empty plan.
	imports := map[string]string{
		"project": "test-project", "instance": "test-project/" + id(updated, "instance"),
		"bucket":   "test-project/" + id(updated, "bucket"),
		"object":   "test-project/" + id(updated, "bucket") + "/" + id(updated, "object"),
		"metadata": id(updated, "metadata"), "api_key": id(updated, "api_key"),
	}
	for _, name := range []string{"project", "instance", "bucket", "object", "metadata", "api_key"} {
		run(0, "state", "rm", "nah_"+name+".test")
		run(0, "import", "-no-color", "nah_"+name+".test", imports[name])
	}
	run(0, "plan", "-detailed-exitcode", "-no-color")
	check(state(), "nah_api_key.test", "token", nil)

	// Remote changes are detected and repaired, not silently adopted.
	drift := "external"
	if _, err := c.UpdateProject(t.Context(), "test-project", drift); err != nil {
		t.Fatal(err)
	}
	if _, err := c.UpdateMetadata(t.Context(), imports["metadata"], &client.UpdateMetadataRequest{Value: &drift}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.UpdateBucket(t.Context(), "test-project", id(updated, "bucket"), "external-bucket"); err != nil {
		t.Fatal(err)
	}
	run(2, "plan", "-detailed-exitcode", "-no-color")
	run(0, "apply", "-auto-approve", "-no-color")
	check(state(), "nah_project.test", "name", "Project 2")
	check(state(), "nah_bucket.test", "name", "assets-renamed")
	check(state(), "nah_bucket.test", "id", id(updated, "bucket"))
	check(state(), "nah_object.test", "id", id(updated, "object"))

	// Immutable image changes replace instead of attempting an invalid PATCH.
	write("main.tf", strings.ReplaceAll(config(2), "ubuntu:24.04", "debian:12"))
	run(0, "apply", "-auto-approve", "-no-color")
	replaced := state()
	if replaced["nah_instance.test"]["id"] == updated["nah_instance.test"]["id"] {
		t.Fatal("image change did not replace instance")
	}
	check(replaced, "nah_instance.test", "image", "debian:12")

	// Deleted remote resources must be recreated by refresh/apply. Remove data
	// sources first: a missing data source should be an error, not state removal.
	resourceConfig := strings.Split(strings.ReplaceAll(config(2), "ubuntu:24.04", "debian:12"), "# Data sources")[0]
	write("main.tf", resourceConfig)
	run(0, "apply", "-auto-approve", "-no-color")
	if err := c.DeleteInstance(t.Context(), "test-project", id(replaced, "instance")); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteObject(t.Context(), "test-project", id(updated, "bucket"), id(updated, "object")); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteMetadata(t.Context(), imports["metadata"]); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteAPIKey(t.Context(), imports["api_key"]); err != nil {
		t.Fatal(err)
	}
	run(2, "plan", "-detailed-exitcode", "-no-color")
	run(0, "apply", "-auto-approve", "-no-color")
	run(0, "plan", "-detailed-exitcode", "-no-color")
	for _, name := range []string{"instance", "object", "metadata", "api_key"} {
		if state()["nah_"+name+".test"]["id"] == replaced["nah_"+name+".test"]["id"] {
			t.Fatalf("%s not recreated after deletion", name)
		}
	}
	// Parent deletion cascades through object state and still converges.
	recreated := state()
	if err := c.DeleteBucket(t.Context(), "test-project", id(recreated, "bucket")); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteInstance(t.Context(), "test-project", id(recreated, "instance")); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteProject(t.Context(), "test-project"); err != nil {
		t.Fatal(err)
	}
	run(0, "apply", "-auto-approve", "-no-color")
	run(0, "plan", "-detailed-exitcode", "-no-color")
	if state()["nah_project.test"]["id"] == recreated["nah_project.test"]["id"] {
		t.Fatal("project not recreated")
	}

	// Validation must reject values locally, before invalid requests reach the API.
	for _, pair := range [][2]string{{"cpu = 3", "cpu = 0"}, {"cpu = 3", "cpu = 65"}, {"memory_mb = 1536", "memory_mb = 524289"}, {`region = "eu-west-1"`, `region = "invalid"`}, {`status = "stopped"`, `status = "broken"`}} {
		write("main.tf", strings.ReplaceAll(resourceConfig, pair[0], pair[1]))
		run(1, "validate", "-no-color")
	}
	write("main.tf", resourceConfig)
	write("missing.tf", `data "nah_project" "missing" { slug = "missing" }`)
	output := run(1, "plan", "-no-color")
	if !strings.Contains(string(output), "404") {
		t.Fatal("missing data source did not report 404")
	}
	write("missing.tf", "")
	run(0, "destroy", "-auto-approve", "-no-color")
	if _, err := c.GetProject(t.Context(), "test-project"); !client.IsNotFound(err) {
		t.Fatalf("project survived destroy: %v", err)
	}
	if got := len(state()); got != 0 {
		t.Fatalf("destroy left %d resources", got)
	}
}

// Remote tests are opt-in and create an isolated organization per CLI. The API
// lacks organization deletion, so cleanup revokes its bootstrap key instead.
func testAPI(t *testing.T) (string, *domain.OrganizationWithAPIKey) {
	t.Helper()
	endpoint := strings.TrimRight(os.Getenv("NAH_ACC_ENDPOINT"), "/")
	remote := endpoint != ""
	if !remote {
		db, err := sqlite.NewDB(filepath.Join(t.TempDir(), "nah.sqlite"))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := db.Close(); err != nil {
				t.Error(err)
			}
		})
		svc := service.NewService(sqlite.NewOrganizationRepository(db), sqlite.NewAPIKeyRepository(db), sqlite.NewSessionRepository(db), sqlite.NewProjectRepository(db), sqlite.NewInstanceRepository(db), sqlite.NewMetadataRepository(db), sqlite.NewBucketRepository(db), sqlite.NewObjectRepository(db))
		server := httptest.NewServer(api.SetupRouter(api.NewHandler(svc), svc, "acceptance"))
		t.Cleanup(server.Close)
		endpoint = server.URL
	}
	slug := fmt.Sprintf("provider-acc-%d", time.Now().UnixNano())
	body, err := json.Marshal(domain.CreateOrganizationRequest{Slug: slug, Name: "Provider acceptance test"})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequestWithContext(t.Context(), "POST", endpoint+"/v1/orgs", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create test organization: HTTP %d", resp.StatusCode)
	}
	var org domain.OrganizationWithAPIKey
	if err := json.NewDecoder(resp.Body).Decode(&org); err != nil {
		t.Fatal(err)
	}
	if org.APIKey.Token == "" {
		t.Fatal("test organization has no API key")
	}
	if remote {
		t.Logf("isolated remote test organization: %s", slug)
		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := client.NewClient(endpoint, org.APIKey.Token).DeleteAPIKey(ctx, org.APIKey.ID); err != nil {
				t.Errorf("revoke test bootstrap key: %v", err)
			}
		})
	}
	return endpoint, &org
}

func config(version int) string {
	value, content, path, status := "one", "b25l", "nested/original.json", "running"
	bucket := "assets"
	if version == 2 {
		value, content, path, status = "", "dHdv", "nested/new.json", "stopped"
		bucket = "assets-renamed"
	}
	return fmt.Sprintf(`
terraform {
  required_providers {
    nah = { source = "registry.terraform.io/hypertf/nah" }
  }
}
provider "nah" {}
resource "nah_project" "test" {
  slug = "test-project"
  name = "Project %d"
}
resource "nah_instance" "test" {
  project = nah_project.test.slug
  name = "web"
  region = "eu-west-1"
  cpu = %d
  memory_mb = 1536
  image = "ubuntu:24.04"
  status = %q
}
resource "nah_bucket" "test" {
  project = nah_project.test.slug
  name = %q
}
resource "nah_object" "test" {
  project = nah_project.test.slug
  bucket_id = nah_bucket.test.id
  path = %q
  content = %q
}
resource "nah_metadata" "test" {
  path = "/config/test"
  value = %q
}
resource "nah_api_key" "test" { name = "automation" }
# Data sources
data "nah_project" "test" { slug = nah_project.test.slug }
data "nah_instance" "test" {
  project = nah_project.test.slug
  id = nah_instance.test.id
}
data "nah_bucket" "test" {
  project = nah_project.test.slug
  id = nah_bucket.test.id
}
data "nah_object" "test" {
  project = nah_project.test.slug
  bucket_id = nah_bucket.test.id
  id = nah_object.test.id
}
data "nah_metadata" "test" { id = nah_metadata.test.id }
data "nah_organization" "test" {}
`, version, version+1, status, bucket, path, content, value)
}
