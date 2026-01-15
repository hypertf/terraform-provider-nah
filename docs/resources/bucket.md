---
page_title: "nah_bucket Resource - NahCloud"
subcategory: ""
description: |-
  Manages a NahCloud storage bucket. Buckets are logical containers for objects.
---

# nah_bucket (Resource)

Manages a NahCloud storage bucket. Buckets are logical containers for objects.

## Example Usage

```terraform
resource "nah_bucket" "assets" {
  name = "my-assets"
}
```

## Schema

### Required

- `name` (String) The name of the bucket. Must be unique.

### Read-Only

- `id` (String) The unique identifier of the bucket.

## Import

Buckets can be imported using the bucket ID:

```shell
terraform import nah_bucket.example <bucket-id>
```
