resource "nah_project" "example" {
  slug = "my-project"
  name = "My project"
}

# Create a bucket first
resource "nah_bucket" "assets" {
  project = nah_project.example.slug
  name    = "my-assets"
}

# Create an object in the bucket
resource "nah_object" "config" {
  project   = nah_project.example.slug
  bucket_id = nah_bucket.assets.id
  path      = "config/settings.json"
  content = base64encode(jsonencode({
    debug = true
    level = "info"
  }))
}

output "object_id" {
  value = nah_object.config.id
}
