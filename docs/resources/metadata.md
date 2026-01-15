---
page_title: "nah_metadata Resource - NahCloud"
subcategory: ""
description: |-
  Manages NahCloud key-value metadata with path-based hierarchy.
---

# nah_metadata (Resource)

Manages NahCloud key-value metadata with path-based hierarchy.

## Example Usage

```terraform
resource "nah_metadata" "config" {
  path  = "/config/app/debug"
  value = "true"
}

resource "nah_metadata" "settings" {
  path  = "/config/app/max_connections"
  value = "100"
}
```

## Schema

### Required

- `path` (String) The path for the metadata entry (e.g., `/config/app/setting`).
- `value` (String) The value for the metadata entry.

### Read-Only

- `id` (String) The unique identifier of the metadata entry.

## Import

Metadata can be imported using the metadata ID:

```shell
terraform import nah_metadata.example <metadata-id>
```
