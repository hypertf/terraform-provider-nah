terraform {
  required_providers {
    nah = { source = "registry.terraform.io/hypertf/nah" }
  }
}

provider "nah" {}

variable "config_version" { type = number }
variable "image" { type = string }

locals {
  updated = var.config_version == 2
}

resource "nah_project" "test" {
  slug = "test-project"
  name = "Project ${var.config_version}"
}
resource "nah_network" "test" {
  project = nah_project.test.slug
  name    = local.updated ? "primary-renamed" : "primary"
  region  = "eu-west-1"
}
resource "nah_subnet" "test" {
  project    = nah_project.test.slug
  network_id = nah_network.test.id
  name       = local.updated ? "primary-renamed" : "primary"
  cidr       = "10.42.1.0/24"
}
resource "nah_instance" "test" {
  project   = nah_project.test.slug
  subnet_id = nah_subnet.test.id
  name      = "web"
  region    = "eu-west-1"
  cpu       = var.config_version + 1
  memory_mb = 1536
  image     = var.image
  status    = local.updated ? "stopped" : "running"
}
resource "nah_bucket" "test" {
  project = nah_project.test.slug
  name    = local.updated ? "assets-renamed" : "assets"
}
resource "nah_object" "test" {
  project   = nah_project.test.slug
  bucket_id = nah_bucket.test.id
  path      = local.updated ? "nested/new.json" : "nested/original.json"
  content   = local.updated ? "dHdv" : "b25l"
}
resource "nah_metadata" "test" {
  path  = "/config/test"
  value = local.updated ? "" : "one"
}
resource "nah_api_key" "test" { name = "automation" }
resource "nah_disk" "test" {
  project = nah_project.test.slug
  name    = local.updated ? "primary-renamed" : "primary"
  size_gb = local.updated ? 30 : 20
  region  = "eu-west-1"
  type    = "ssd"
}
resource "nah_disk_attachment" "test" {
  project     = nah_project.test.slug
  disk_id     = nah_disk.test.id
  instance_id = nah_instance.test.id
  device      = "vdb"
}
resource "nah_policy" "test" {
  name        = local.updated ? "primary-renamed" : "primary"
  description = "Terraform E2E policy"
  effect      = "allow"
  actions     = ["network:read"]
}
resource "nah_policy_binding" "test" {
  policy_id      = nah_policy.test.id
  principal_type = "api_key"
  principal_id   = nah_api_key.test.id
  target_type    = "project"
  target_id      = nah_project.test.id
}
resource "nah_load_balancer" "test" {
  project           = nah_project.test.slug
  subnet_id         = nah_subnet.test.id
  name              = local.updated ? "primary-renamed" : "primary"
  protocol          = "http"
  port              = 80
  algorithm         = local.updated ? "least_connections" : "round_robin"
  health_check_path = "/health"
}
resource "nah_load_balancer_backend" "test" {
  project          = nah_project.test.slug
  load_balancer_id = nah_load_balancer.test.id
  instance_id      = nah_instance.test.id
  port             = 8080
  weight           = local.updated ? 50 : 100
  enabled          = true
}

# Data sources
data "nah_project" "test" { slug = nah_project.test.slug }
data "nah_instance" "test" {
  project = nah_project.test.slug
  id      = nah_instance.test.id
}
data "nah_bucket" "test" {
  project = nah_project.test.slug
  id      = nah_bucket.test.id
}
data "nah_object" "test" {
  project   = nah_project.test.slug
  bucket_id = nah_bucket.test.id
  id        = nah_object.test.id
}
data "nah_metadata" "test" { id = nah_metadata.test.id }
data "nah_organization" "test" {}
data "nah_network" "test" {
  project = nah_project.test.slug
  id      = nah_network.test.id
}
data "nah_subnet" "test" {
  project    = nah_project.test.slug
  network_id = nah_network.test.id
  id         = nah_subnet.test.id
}
data "nah_disk" "test" {
  project = nah_project.test.slug
  id      = nah_disk.test.id
}
data "nah_disk_attachment" "test" {
  project = nah_project.test.slug
  disk_id = nah_disk.test.id
  id      = nah_disk_attachment.test.id
}
data "nah_policy" "test" { id = nah_policy.test.id }
data "nah_policy_binding" "test" {
  policy_id = nah_policy.test.id
  id        = nah_policy_binding.test.id
}
data "nah_load_balancer" "test" {
  project = nah_project.test.slug
  id      = nah_load_balancer.test.id
}
data "nah_load_balancer_backend" "test" {
  project          = nah_project.test.slug
  load_balancer_id = nah_load_balancer.test.id
  id               = nah_load_balancer_backend.test.id
}
