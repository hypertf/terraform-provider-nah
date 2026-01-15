# Terraform Provider for NahCloud

A Terraform provider for [NahCloud](https://github.com/hypertf/nahcloud-server), a fake cloud API for testing Terraform tooling without provisioning real infrastructure.

## Features

- **Projects** - Top-level containers for organizing resources
- **Instances** - Compute resources with CPU, memory, image, and status
- **Metadata** - Key-value storage with path-based hierarchy
- **Buckets** - Storage containers for objects
- **Objects** - Blob storage within buckets

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building)
- [NahCloud Server](https://github.com/hypertf/nahcloud-server) running locally or remotely

## Quick Start

### 1. Start NahCloud Server

```bash
cd nahcloud-server
make run-server
```

### 2. Configure the Provider

```terraform
terraform {
  required_providers {
    nah = {
      source = "hypertf/nah"
    }
  }
}

provider "nah" {
  endpoint = "https://nahcloud.com"
  # token = "optional-auth-token"
}
```

### 3. Create Resources

```terraform
resource "nah_project" "example" {
  name = "my-project"
}

resource "nah_instance" "web" {
  project_id = nah_project.example.id
  name       = "web-server"
  cpu        = 2
  memory_mb  = 1024
  image      = "ubuntu:22.04"
  status     = "running"
}

resource "nah_bucket" "assets" {
  name = "my-assets"
}

resource "nah_object" "config" {
  bucket_id = nah_bucket.assets.id
  path      = "config/settings.json"
  content   = base64encode(jsonencode({
    debug = true
  }))
}
```

## Configuration

### Provider Arguments

| Argument | Description | Default | Environment Variable |
|----------|-------------|---------|---------------------|
| `endpoint` | NahCloud API endpoint | `https://nahcloud.com` | `NAH_ENDPOINT` |
| `token` | Authentication token (optional) | - | `NAH_TOKEN` |

## Resources

- `nah_project` - Manages projects
- `nah_instance` - Manages compute instances
- `nah_metadata` - Manages key-value metadata
- `nah_bucket` - Manages storage buckets
- `nah_object` - Manages objects within buckets

## Data Sources

- `nah_project` - Fetches project information
- `nah_instance` - Fetches instance information
- `nah_metadata` - Fetches metadata information
- `nah_bucket` - Fetches bucket information
- `nah_object` - Fetches object information

## Building the Provider

```bash
git clone https://github.com/nicolas/terraform-provider-nah
cd terraform-provider-nah
go install
```

## Local Development

Create a `~/.terraformrc` file for dev overrides:

```hcl
provider_installation {
  dev_overrides {
    "hypertf/nah" = "/path/to/your/GOPATH/bin"
  }
  direct {}
}
```

## Running Tests

```bash
# Unit tests
make test

# Acceptance tests (requires running NahCloud server)
make testacc
```

## License

MPL-2.0
