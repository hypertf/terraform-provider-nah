---
page_title: "nah_metadata Data Source - NahCloud"
subcategory: ""
description: |-
  Fetches information about a NahCloud metadata entry.
---

# nah_metadata (Data Source)

Fetches information about a NahCloud metadata entry.

## Example Usage

```terraform
data "nah_metadata" "config" {
  id = "existing-metadata-id"
}

output "config_value" {
  value = data.nah_metadata.config.value
}
```

## Schema

### Required

- `id` (String) The unique identifier of the metadata entry.

### Read-Only

- `path` (String) The path for the metadata entry.
- `value` (String) The value for the metadata entry.
- `created_at` (String) The timestamp when the metadata was created.
- `updated_at` (String) The timestamp when the metadata was last updated.
