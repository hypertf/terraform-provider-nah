---
page_title: "nah_object Resource - NahCloud"
subcategory: ""
description: |-
  Manages a NahCloud storage object within a bucket.
---

# nah_object (Resource)

Manages a NahCloud storage object within a bucket. Content is stored as base64-encoded string.

## Example Usage

```terraform
resource "nah_bucket" "assets" {
  name = "my-assets"
}

resource "nah_object" "config" {
  bucket_id = nah_bucket.assets.id
  path      = "config/settings.json"
  content   = base64encode(jsonencode({
    debug = true
    level = "info"
  }))
}
```

## Schema

### Required

- `bucket_id` (String) The ID of the bucket this object belongs to. Changing this forces a new resource.
- `path` (String) The path of the object within the bucket.
- `content` (String) The content of the object (base64-encoded).

### Read-Only

- `id` (String) The unique identifier of the object.

## Import

Objects can be imported using the object ID:

```shell
terraform import nah_object.example <object-id>
```
