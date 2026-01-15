---
page_title: "nah_object Data Source - NahCloud"
subcategory: ""
description: |-
  Fetches information about a NahCloud storage object.
---

# nah_object (Data Source)

Fetches information about a NahCloud storage object.

## Example Usage

```terraform
data "nah_object" "config" {
  id        = "existing-object-id"
  bucket_id = "existing-bucket-id"
}

output "object_content" {
  value = base64decode(data.nah_object.config.content)
}
```

## Schema

### Required

- `id` (String) The unique identifier of the object.
- `bucket_id` (String) The ID of the bucket this object belongs to.

### Read-Only

- `path` (String) The path of the object within the bucket.
- `content` (String) The content of the object (base64-encoded).
- `created_at` (String) The timestamp when the object was created.
- `updated_at` (String) The timestamp when the object was last updated.
