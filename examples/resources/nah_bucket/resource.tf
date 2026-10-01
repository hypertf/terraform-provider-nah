resource "nah_project" "example" {
  slug = "my-project"
  name = "My project"
}

# Create a storage bucket
resource "nah_bucket" "assets" {
  project = nah_project.example.slug
  name    = "my-assets"
}

output "bucket_id" {
  value = nah_bucket.assets.id
}
