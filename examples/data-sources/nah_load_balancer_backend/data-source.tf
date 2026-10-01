data "nah_load_balancer_backend" "app" {
  project          = "my-app"
  load_balancer_id = "LOAD_BALANCER_ID"
  id               = "BACKEND_ID"
}
