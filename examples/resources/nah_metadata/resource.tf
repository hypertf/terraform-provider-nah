# Create metadata entries
resource "nah_metadata" "app_config" {
  path  = "/config/app/debug"
  value = "true"
}

resource "nah_metadata" "db_config" {
  path  = "/config/database/host"
  value = "localhost:5432"
}
