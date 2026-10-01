resource "nah_load_balancer_backend" "app" {
  project          = nah_project.app.slug
  load_balancer_id = nah_load_balancer.app.id
  instance_id      = nah_instance.app.id
  port             = 8080
  weight           = 100
  enabled          = true
}
