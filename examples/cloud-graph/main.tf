resource "nah_project" "app" {
  slug = "graph-example"
  name = "Cloud graph example"
}

resource "nah_network" "app" {
  project = nah_project.app.slug
  name    = "app"
  region  = "eu-west-1"
}

resource "nah_subnet" "app" {
  project    = nah_project.app.slug
  network_id = nah_network.app.id
  name       = "app_a"
  cidr       = "10.42.1.0/24"
}

resource "nah_instance" "app" {
  project   = nah_project.app.slug
  subnet_id = nah_subnet.app.id
  name      = "app"
  region    = "eu-west-1"
  image     = "ubuntu:24.04"
  cpu       = 2
  memory_mb = 2048
}

resource "nah_disk" "data" {
  project = nah_project.app.slug
  name    = "data"
  size_gb = 100
  region  = "eu-west-1"
  type    = "ssd"
}

resource "nah_disk_attachment" "data" {
  project     = nah_project.app.slug
  disk_id     = nah_disk.data.id
  instance_id = nah_instance.app.id
  device      = "vdb"
}

resource "nah_policy" "read_network" {
  name        = "read_network"
  description = "Allow network reads during policy evaluation"
  effect      = "allow"
  actions     = ["network:read"]
}

resource "nah_api_key" "automation" {
  name = "graph-example-automation"
}

resource "nah_policy_binding" "automation" {
  policy_id      = nah_policy.read_network.id
  principal_type = "api_key"
  principal_id   = nah_api_key.automation.id
  target_type    = "project"
  target_id      = nah_project.app.id
}

resource "nah_load_balancer" "app" {
  project           = nah_project.app.slug
  subnet_id         = nah_subnet.app.id
  name              = "app"
  protocol          = "http"
  port              = 80
  algorithm         = "round_robin"
  health_check_path = "/health"
}

resource "nah_load_balancer_backend" "app" {
  project          = nah_project.app.slug
  load_balancer_id = nah_load_balancer.app.id
  instance_id      = nah_instance.app.id
  port             = 8080
  weight           = 100
  enabled          = true
}
