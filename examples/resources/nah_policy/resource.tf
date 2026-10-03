resource "nah_policy" "read_network" {
  name        = "read_network"
  description = "Allow network reads during policy evaluation"
  effect      = "allow"
  actions     = ["network:read"]
}
