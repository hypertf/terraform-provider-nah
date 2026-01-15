---
page_title: "nah_instance Resource - NahCloud"
subcategory: ""
description: |-
  Manages a NahCloud compute instance.
---

# nah_instance (Resource)

Manages a NahCloud compute instance.

## Example Usage

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
```

## Schema

### Required

- `project_id` (String) The ID of the project this instance belongs to. Changing this forces a new resource.
- `name` (String) The name of the instance.
- `image` (String) The image to use for the instance.

### Optional

- `cpu` (Number) The number of CPUs for the instance. Defaults to `1`.
- `memory_mb` (Number) The amount of memory in MB for the instance. Defaults to `512`.
- `status` (String) The status of the instance. Valid values: `running`, `stopped`. Defaults to `running`.

### Read-Only

- `id` (String) The unique identifier of the instance.

## Import

Instances can be imported using the instance ID:

```shell
terraform import nah_instance.example <instance-id>
```
