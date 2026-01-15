# Create a storage bucket
resource "nah_bucket" "assets" {
  name = "my-assets"
}

output "bucket_id" {
  value = nah_bucket.assets.id
}
