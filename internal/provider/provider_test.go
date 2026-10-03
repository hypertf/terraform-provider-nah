package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hypertf/terraform-provider-nah/internal/client"
)

func TestSchema(t *testing.T) {
	server := providerserver.NewProtocol6(New("test")())()
	resp, err := server.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil || len(resp.Diagnostics) != 0 {
		t.Fatalf("schema: %v %v", err, resp.Diagnostics)
	}
	if len(resp.ResourceSchemas) != 14 || len(resp.DataSourceSchemas) != 14 {
		t.Fatal("missing registered schema")
	}
}

func TestConfigure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, `{"id":"configured","name":%q}`, r.Header.Get("Authorization"))
	}))
	defer server.Close()
	t.Setenv("NAH_TOKEN", "environment-token")
	t.Setenv("NAH_ENDPOINT", server.URL)
	for _, tc := range []struct {
		name            string
		endpoint, token types.String
		wantError       bool
	}{
		{"environment", types.StringNull(), types.StringNull(), false},
		{"explicit", types.StringValue(server.URL + "/"), types.StringValue("explicit-token"), false},
		{"unknown token", types.StringNull(), types.StringUnknown(), true},
		{"unknown endpoint", types.StringUnknown(), types.StringNull(), true},
		{"empty token overrides environment", types.StringNull(), types.StringValue(""), true},
		{"relative endpoint", types.StringValue("example.test"), types.StringNull(), true},
		{"credentials in endpoint", types.StringValue("https://user:secret@example.test"), types.StringNull(), true},
		{"query in endpoint", types.StringValue("https://example.test?token=secret"), types.StringNull(), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &NahProvider{}
			var schema fwprovider.SchemaResponse
			p.Schema(t.Context(), fwprovider.SchemaRequest{}, &schema)
			state := tfsdk.State{Schema: schema.Schema}
			if d := state.Set(t.Context(), NahProviderModel{tc.endpoint, tc.token}); d.HasError() {
				t.Fatal(d)
			}
			var resp fwprovider.ConfigureResponse
			p.Configure(t.Context(), fwprovider.ConfigureRequest{Config: tfsdk.Config{Schema: schema.Schema, Raw: state.Raw}}, &resp)
			if resp.Diagnostics.HasError() != tc.wantError {
				t.Fatalf("diagnostics: %v", resp.Diagnostics)
			}
			if !tc.wantError {
				c, ok := resp.ResourceData.(*client.Client)
				if !ok {
					t.Fatal("missing configured client")
				}
				project, err := c.GetProject(t.Context(), "test")
				if err != nil {
					t.Fatal(err)
				}
				want := "Bearer environment-token"
				if tc.name == "explicit" {
					want = "Bearer explicit-token"
				}
				if project.Name != want {
					t.Fatal("configuration/environment token precedence is incorrect")
				}
			}
		})
	}
}

func TestStringValidation(t *testing.T) {
	for _, tc := range []struct {
		v       apiString
		value   types.String
		invalid bool
	}{
		{slugValidator, types.StringValue("a-9"), false},
		{slugValidator, types.StringValue("9abc"), true},
		{slugValidator, types.StringValue("a/b"), true},
		{slugValidator, types.StringValue(strings.Repeat("a", 63)), false},
		{slugValidator, types.StringValue(strings.Repeat("a", 64)), true},
		{nameValidator, types.StringValue("Web_9-a"), false},
		{nameValidator, types.StringValue("has space"), true},
		{apiString{min: 1, max: 255}, types.StringValue(strings.Repeat("é", 127)), false},
		{apiString{min: 1, max: 255}, types.StringValue(strings.Repeat("é", 128)), true},
		{slugValidator, types.StringNull(), false},
		{slugValidator, types.StringUnknown(), false},
		{slugValidator, types.StringValue(""), true},
	} {
		var resp validator.StringResponse
		tc.v.ValidateString(t.Context(), validator.StringRequest{Path: path.Root("test"), ConfigValue: tc.value}, &resp)
		if resp.Diagnostics.HasError() != tc.invalid {
			t.Fatalf("%v: %v", tc.value, resp.Diagnostics)
		}
	}
}

func TestResourceDeletionAndFailures(t *testing.T) {
	for _, status := range []int{404, 401, 403, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) }))
			defer server.Close()
			for _, factory := range (&NahProvider{}).Resources(t.Context()) {
				r := factory()
				var schema resource.SchemaResponse
				r.Schema(t.Context(), resource.SchemaRequest{}, &schema)
				state := tfsdk.State{Schema: schema.Schema}
				state.RemoveResource(t.Context())
				for key := range schema.Schema.Attributes {
					value := any("test")
					switch key {
					case "cpu", "memory_mb", "size_gb", "port", "weight", "seed", "duration_seconds":
						value = int64(1)
					case "enabled", "healthy":
						value = true
					case "actions":
						value = []string{"*"}
					}
					if d := state.SetAttribute(t.Context(), path.Root(key), value); d.HasError() {
						t.Fatal(d)
					}
				}
				var configure resource.ConfigureResponse
				configurable, ok := r.(resource.ResourceWithConfigure)
				if !ok {
					t.Fatalf("%T is not configurable", r)
				}
				configurable.Configure(t.Context(), resource.ConfigureRequest{ProviderData: client.NewClient(server.URL, "test")}, &configure)
				read := resource.ReadResponse{State: state}
				r.Read(t.Context(), resource.ReadRequest{State: state}, &read)
				if status == 404 {
					if read.Diagnostics.HasError() || !read.State.Raw.IsNull() {
						t.Fatalf("%T: 404 did not remove state: %v", r, read.Diagnostics)
					}
				} else if !read.Diagnostics.HasError() || read.State.Raw.IsNull() {
					t.Fatalf("%T: %d must preserve state with error", r, status)
				}
				var deleted resource.DeleteResponse
				r.Delete(t.Context(), resource.DeleteRequest{State: state}, &deleted)
				if deleted.Diagnostics.HasError() != (status != 404) {
					t.Fatalf("%T delete: %v", r, deleted.Diagnostics)
				}
			}
		})
	}
}

func TestMalformedImports(t *testing.T) {
	for _, id := range []string{"", "project", "/id", "project/", "project/id/extra", "../id"} {
		var resp resource.ImportStateResponse
		importScoped(t.Context(), resource.ImportStateRequest{ID: id}, &resp, "project", "id")
		if !resp.Diagnostics.HasError() {
			t.Fatalf("accepted malformed import %q", id)
		}
	}
}

func TestDiskSizeReplacement(t *testing.T) {
	for _, tc := range []struct {
		state, plan int64
		replace     bool
	}{{20, 30, false}, {30, 20, true}, {20, 20, false}} {
		var resp planmodifier.Int64Response
		replaceOnShrink{}.PlanModifyInt64(t.Context(), planmodifier.Int64Request{
			StateValue: types.Int64Value(tc.state),
			PlanValue:  types.Int64Value(tc.plan),
		}, &resp)
		if resp.RequiresReplace != tc.replace {
			t.Fatalf("size %d -> %d: replace=%t", tc.state, tc.plan, resp.RequiresReplace)
		}
	}
}
