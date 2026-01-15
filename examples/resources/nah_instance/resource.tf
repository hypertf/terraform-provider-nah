# Create a project first
resource "nah_project" "example" {
  name = "my-project"
}

# Create a compute instance
resource "nah_instance" "web" {
  project_id = nah_project.example.id
  name       = "web-server"
  cpu        = 2
  memory_mb  = 1024
  image      = "ubuntu:22.04"
  status     = "running"
}

output "instance_id" {
  value = nah_instance.web.id
}
