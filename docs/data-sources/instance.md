---
page_title: "nah_instance Data Source - NahCloud"
subcategory: ""
description: |-
  Fetches information about a NahCloud compute instance.
---

# nah_instance (Data Source)

Fetches information about a NahCloud compute instance.

## Example Usage

```terraform
data "nah_instance" "example" {
  id = "existing-instance-id"
}

output "instance_status" {
  value = data.nah_instance.example.status
}
```

## Schema

### Required

- `id` (String) The unique identifier of the instance.

### Read-Only

- `project_id` (String) The ID of the project this instance belongs to.
- `name` (String) The name of the instance.
- `cpu` (Number) The number of CPUs for the instance.
- `memory_mb` (Number) The amount of memory in MB for the instance.
- `image` (String) The image used for the instance.
- `status` (String) The status of the instance.
- `created_at` (String) The timestamp when the instance was created.
- `updated_at` (String) The timestamp when the instance was last updated.
