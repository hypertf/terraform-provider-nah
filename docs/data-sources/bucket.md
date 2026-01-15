---
page_title: "nah_bucket Data Source - NahCloud"
subcategory: ""
description: |-
  Fetches information about a NahCloud storage bucket.
---

# nah_bucket (Data Source)

Fetches information about a NahCloud storage bucket.

## Example Usage

```terraform
data "nah_bucket" "assets" {
  id = "existing-bucket-id"
}

output "bucket_name" {
  value = data.nah_bucket.assets.name
}
```

## Schema

### Required

- `id` (String) The unique identifier of the bucket.

### Read-Only

- `name` (String) The name of the bucket.
- `created_at` (String) The timestamp when the bucket was created.
- `updated_at` (String) The timestamp when the bucket was last updated.
