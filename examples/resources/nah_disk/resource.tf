resource "nah_disk" "data" {
  project = nah_project.app.slug
  name    = "data"
  size_gb = 100
  region  = "eu-west-1"
  type    = "ssd"
}
