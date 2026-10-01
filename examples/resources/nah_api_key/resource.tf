resource "nah_api_key" "automation" {
  name = "automation"
}

# The token is sensitive and remains in Terraform state.
# Keep the provider's authentication key separate from managed keys.
output "automation_token" {
  value     = nah_api_key.automation.token
  sensitive = true
}
