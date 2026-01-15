---
page_title: "nah_project Data Source - NahCloud"
subcategory: ""
description: |-
  Fetches information about a NahCloud project.
---

# nah_project (Data Source)

Fetches information about a NahCloud project.

## Example Usage

```terraform
data "nah_project" "example" {
  id = "existing-project-id"
}

output "project_name" {
  value = data.nah_project.example.name
}
```

## Schema

### Required

- `id` (String) The unique identifier of the project.

### Read-Only

- `name` (String) The name of the project.
- `created_at` (String) The timestamp when the project was created.
- `updated_at` (String) The timestamp when the project was last updated.
