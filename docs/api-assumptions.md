# Frozen cloud graph API contract

The provider assumes stable NahCloud routes, opaque lowercase 128-bit hex IDs,
snake_case JSON, and ownership derived from the authenticated organization and
route parents. `{project}` is a project slug. Every URL ancestor is validated;
wrong project, parent, or tenant scope returns 404.

| Terraform object | Canonical collection/item route | Mutable fields | Replacement fields | Import ID |
|---|---|---|---|---|
| `nah_network` | `/v1/projects/{project}/networks[/{network_id}]` | `name` | `project`, `region` | `project/network_id` |
| `nah_subnet` | `/v1/projects/{project}/networks/{network_id}/subnets[/{subnet_id}]` | `name` | `project`, `network_id`, `cidr` | `project/network_id/subnet_id` |
| `nah_disk` | `/v1/projects/{project}/disks[/{disk_id}]` | `name`, size expansion | `project`, `region`, `type`, size shrink | `project/disk_id` |
| `nah_disk_attachment` | `/v1/projects/{project}/disks/{disk_id}/attachments[/{attachment_id}]` | none | all configured fields | `project/disk_id/attachment_id` |
| `nah_policy` | `/v1/policies[/{policy_id}]` | `name`, `description`, `effect`, `actions` | none | `policy_id` |
| `nah_policy_binding` | `/v1/policies/{policy_id}/bindings[/{binding_id}]` | none | all configured fields | `policy_id/binding_id` |
| `nah_load_balancer` | `/v1/projects/{project}/load-balancers[/{load_balancer_id}]` | `name`, `algorithm`, `health_check_path` | `project`, `subnet_id`, `protocol`, `port` | `project/load_balancer_id` |
| `nah_load_balancer_backend` | `/v1/projects/{project}/load-balancers/{load_balancer_id}/backends[/{backend_id}]` | `port`, `weight`, `enabled` | `project`, `load_balancer_id`, `instance_id` | `project/load_balancer_id/backend_id` |

The existing `nah_instance` accepts optional `subnet_id`. Adding, changing, or
removing it replaces the instance. The subnet must belong to the same project,
and its network region must match the instance region.

Networks require one of the provider's five regions. Subnets require canonical
private IPv4 CIDRs with prefix /16 through /28. Disks require `region`, type
`standard` or `ssd`, and 1–16384 GiB; expansion PATCHes in place while a
state-aware plan modifier replaces on shrink. Attachments are immutable, require
same-project/same-region disk and instance, and use a `vd[b-z]` device.

Policies are organization-scoped evaluation data. `effect` is `allow` or `deny`;
`actions` has 1–32 exact operation names or `*`. Bindings allow principals
`organization` or `api_key` and targets `organization`, `project`, `network`,
`subnet`, `instance`, `disk`, `load_balancer`, or `bucket`. Policy evaluation is
POST-only at `/v1/policy-evaluations`, returns `allow`, `deny`, or
`not_applicable` plus matching binding IDs, and does not authorize ordinary CRUD.

Load balancers bind to a subnet and derive their region. Protocol is `http` or
`tcp`; algorithm is `round_robin` or `least_connections`; HTTP health paths begin
with `/` and TCP paths are empty. Backends must reference an instance in the same
project and subnet. `healthy` is computed as enabled plus a running instance.

Create/read/update/delete statuses are 201/200/200/204. Resource read 404 removes
state, delete 404 succeeds, and data-source 404 is an error. Writes are never
retried and no idempotency key is sent.

Fault rules (`/v1/fault-rules`) are intentionally API-only. The provider exposes
no fault resource or data source, preventing fault configuration from disrupting
the requests Terraform would need to repair that configuration.
