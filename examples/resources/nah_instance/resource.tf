# Create a project first
resource "nah_project" "example" {
  slug = "my-project"
  name = "my-project"
}

# Create a compute instance
resource "nah_instance" "web" {
  project   = nah_project.example.slug
  region    = "us-east-1"
  name      = "web-server"
  cpu       = 2
  memory_mb = 1024
  image     = "ubuntu:22.04"
  status    = "running"
}

output "instance_id" {
  value = nah_instance.web.id
}
