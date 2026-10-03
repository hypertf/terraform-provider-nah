resource "nah_policy_binding" "automation" {
  policy_id      = nah_policy.read_network.id
  principal_type = "api_key"
  principal_id   = nah_api_key.automation.id
  target_type    = "project"
  target_id      = nah_project.app.id
}
