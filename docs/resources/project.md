---
page_title: "nah_project Resource - NahCloud"
subcategory: ""
description: |-
  Manages a NahCloud project. Projects are top-level containers for other resources.
---

# nah_project (Resource)

Manages a NahCloud project. Projects are top-level containers for other resources.

## Example Usage

```terraform
resource "nah_project" "example" {
  name = "my-project"
}
```

## Schema

### Required

- `name` (String) The name of the project.

### Read-Only

- `id` (String) The unique identifier of the project.

## Import

Projects can be imported using the project ID:

```shell
terraform import nah_project.example <project-id>
```
