---
page_title: "NahCloud Provider"
subcategory: ""
description: |-
  The NahCloud provider allows you to manage resources in NahCloud, a fake cloud API for testing Terraform tooling.
---

# NahCloud Provider

The NahCloud provider allows you to manage resources in NahCloud, a fake cloud API for testing Terraform tooling without provisioning real infrastructure.

## Features

- **Projects** - Top-level containers for organizing resources
- **Instances** - Compute resources with CPU, memory, image, and status
- **Metadata** - Key-value storage with path-based hierarchy
- **Buckets** - Storage containers for objects
- **Objects** - Blob storage within buckets

## Example Usage

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
```

## Authentication

The provider supports optional bearer token authentication. You can configure it via:

1. Provider configuration block:
   ```terraform
   provider "nah" {
     token = "your-token"
   }
   ```

2. Environment variable:
   ```bash
   export NAH_TOKEN="your-token"
   ```

## Schema

### Optional

- `endpoint` (String) The NahCloud API endpoint. Defaults to `https://nahcloud.com`. Can also be set via `NAH_ENDPOINT` environment variable.
- `token` (String, Sensitive) The NahCloud API token for authentication. Can also be set via `NAH_TOKEN` environment variable.
