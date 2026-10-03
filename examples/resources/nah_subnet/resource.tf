resource "nah_subnet" "app" {
  project    = nah_project.app.slug
  network_id = nah_network.app.id
  name       = "app_a"
  cidr       = "10.42.1.0/24"
}
