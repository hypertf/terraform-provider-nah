resource "nah_network" "app" {
  project = nah_project.app.slug
  name    = "app"
  region  = "eu-west-1"
}
