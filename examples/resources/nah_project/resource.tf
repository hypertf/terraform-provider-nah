# Create a NahCloud project
resource "nah_project" "example" {
  slug = "my-project"
  name = "my-project"
}

output "project_id" {
  value = nah_project.example.id
}
