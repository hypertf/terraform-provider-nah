data "nah_organization" "current" {}

output "organization_slug" {
  value = data.nah_organization.current.slug
}
