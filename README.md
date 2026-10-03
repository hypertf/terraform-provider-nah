# Terraform / OpenTofu Provider for NahCloud

Manage [NahCloud](https://github.com/hypertf/nahcloud) projects, compute,
networking, storage, evaluation policies, load balancing, metadata, and API keys
with Terraform or OpenTofu. NahCloud is a
simulated cloud: these resources do not provision real infrastructure.

## Quick start

Create an organization through NahCloud and save its API token securely. Export
`NAH_TOKEN` in your shell or CI secret store; do not commit it to HCL. The token
selects the organization. Anonymous provisioning is deliberately unsupported.

```hcl
terraform {
  required_providers {
    nah = { source = "hypertf/nah" }
  }
}

provider "nah" {
  endpoint = "https://nahcloud.com" # optional; also NAH_ENDPOINT
}

resource "nah_project" "app" {
  slug = "my-app"
  name = "My application"
}

resource "nah_instance" "web" {
  project   = nah_project.app.slug
  name      = "web"
  region    = "eu-west-1"
  image     = "ubuntu:24.04"
  cpu       = 2
  memory_mb = 1024
}

resource "nah_bucket" "assets" {
  project = nah_project.app.slug
  name    = "assets"
}

resource "nah_object" "config" {
  project   = nah_project.app.slug
  bucket_id = nah_bucket.assets.id
  path      = "config/settings.json"
  content   = base64encode(jsonencode({ debug = false }))
}
```

Use `terraform init`, `terraform plan`, and `terraform apply` (or `tofu` in place
of `terraform`). See [generated reference documentation](docs/index.md) and
[examples](examples/README.md). Building from this checkout is supported before
a registry release; no release tag is required for development.

## Resource behavior

* `project` always means the **project slug**, not its generated `id`.
* Project slugs, instance project/region/image, and object parent scope require
  replacement. Instance name/CPU/memory/status, project and bucket names, metadata path/value,
  and object path/content update in place.
* Buckets and objects use stable-ID API routes. Renaming a bucket, including
  outside Terraform, preserves the bucket and its objects and is detected on refresh.
* A resource missing remotely is removed from state during refresh and recreated
  by the next apply. Missing data sources produce errors. Authentication and
  server failures preserve state and report diagnostics; they are not deletions.
* Projects must be empty before deletion. Bucket deletion cascades to objects,
  including objects not managed by Terraform. Use `lifecycle.prevent_destroy`
  when that would be unsafe.
* API key names require replacement. Tokens are returned only on creation;
  importing a key leaves `token = null`. Do not manage the token used to
  authenticate this provider. Use a separate bootstrap/administrative key.
* API tokens, metadata values, and object contents are sensitive. Terraform state
  still contains these values in plaintext: use an encrypted, access-controlled
  backend and restrict state access.
* Policies and bindings are organization-scoped evaluation data; they do not
  authorize ordinary CRUD. Fault rules are intentionally not Terraform-managed.
* HTTP requests have a 30-second timeout, honor cancellation, reuse connections,
  and never follow redirects. Writes are not automatically retried: the API has
  no idempotency keys, and retrying a lost create response could create duplicates.

## Imports

Imports use `/`-separated scope and identity, not logical metadata/object paths:

```sh
terraform import nah_project.app my-app
terraform import nah_instance.web my-app/INSTANCE_ID
terraform import nah_bucket.assets my-app/BUCKET_ID
terraform import nah_object.config my-app/BUCKET_ID/OBJECT_ID
terraform import nah_metadata.config METADATA_ID
terraform import nah_api_key.automation KEY_ID
terraform import nah_network.app my-app/NETWORK_ID
terraform import nah_subnet.app my-app/NETWORK_ID/SUBNET_ID
terraform import nah_disk.data my-app/DISK_ID
terraform import nah_disk_attachment.data my-app/DISK_ID/ATTACHMENT_ID
terraform import nah_policy.read POLICY_ID
terraform import nah_policy_binding.read POLICY_ID/BINDING_ID
terraform import nah_load_balancer.app my-app/LOAD_BALANCER_ID
terraform import nah_load_balancer_backend.app my-app/LOAD_BALANCER_ID/BACKEND_ID
```

Organization creation/update/deletion is not represented as a managed resource:
the current API does not expose a complete organization lifecycle. Use
`data "nah_organization" "current" {}` to inspect the authenticated organization.
The HTTP state backend belongs in Terraform's `backend "http"` configuration,
not in a provider resource.

## Migrating from the initial prototype

The prototype targeted nonexistent unscoped routes. This is a breaking schema
correction: add project `slug`; replace instance `project_id` configuration with
`project = nah_project.example.slug`; add instance `region`; add `project` to
buckets, objects, and their data sources. Project data sources take `slug`, not
`id`. `project_id` remains a computed instance attribute. Export `NAH_TOKEN`.
Back up existing state before migration. If resources were imported under the
old schema, remove their old state entries and re-import with the scoped forms
above rather than applying a destructive replacement plan blindly.

## Development and verification

Requires Go 1.25.5+, a C compiler for SQLite acceptance tests, Terraform and/or
OpenTofu. CI tests real CLI executions against the pinned upstream NahCloud
module in `go.mod` plus a local cloud-graph contract server, with a fresh SQLite
database per test. The exact assumed routes and fields are documented in
[`docs/api-assumptions.md`](docs/api-assumptions.md). No network account, shared
server, token secret, or production access is required.

```sh
go build ./...
go test -race ./...
make testacc                         # both terraform and tofu on PATH
NAH_TEST_CLI=tofu make testacc        # one CLI, or an absolute binary path
make generate                       # regenerate registry docs and format examples
make lint
```

Acceptance tests build and launch the actual provider binary and verify create,
empty plans, updates retaining IDs, all imports, data sources, drift repair,
replacement, remote deletion/recreation, and destroy. Unit tests cover HTTP
contracts, errors, cancellation, secret redaction, validation, and state safety.

Acceptance is intentionally local-only and ignores remote endpoint settings. It
combines the pinned upstream router for existing resources with an in-memory
contract server for the assumed cloud-graph routes, so it cannot mutate a shared
or production deployment.

For manual development, run `go install .` and point a CLI configuration file at
the resulting binary directory:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/hypertf/nah" = "/absolute/path/to/go/bin"
  }
  direct {}
}
```

Set `TF_CLI_CONFIG_FILE` to that file and run plan/apply directly; development
overrides do not require provider installation through `init`.

## License

MPL-2.0
