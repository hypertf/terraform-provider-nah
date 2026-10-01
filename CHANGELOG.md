## 0.1.0 (Unreleased)

FEATURES:

- Organization-scoped authentication and correctly scoped project, instance, bucket, object, and metadata routes.
- API key resource and current-organization data source.
- Validated configuration, scoped imports, deletion drift handling, and immutable-field replacement.
- Real Terraform/OpenTofu binary acceptance tests against a pinned NahCloud API and disposable SQLite.

BREAKING CHANGES:

- API token is required via configuration or `NAH_TOKEN`.
- Projects require a slug; nested resources and data sources require `project` (slug), and instances require `region`.
- Project data sources use `slug`; instance `project_id` is computed. Re-import prototype state using the documented scoped identifiers.
