resource "nah_load_balancer" "app" {
  project           = nah_project.app.slug
  subnet_id         = nah_subnet.app.id
  name              = "app"
  protocol          = "http"
  port              = 80
  algorithm         = "round_robin"
  health_check_path = "/health"
}
