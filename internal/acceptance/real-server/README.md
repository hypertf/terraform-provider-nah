# Pinned real-server NahCloud provider E2E

This maintained manual acceptance suite runs the provider against a pinned,
real NahCloud server. `run-matrix.sh` makes the clone/build/service commands
and the Terraform/OpenTofu lifecycle reproducible in one invocation.

It runs only against `127.0.0.1:18080`, using the real NahCloud `cmd/server`
binary and a disposable file-backed SQLite database. It does not use the
provider's `httptest`/in-memory contract server and does not contact production.

## Pinned inputs

- Provider: the current checkout's exact `HEAD` (override with `PROVIDER_SHA`)
- NahCloud: `9131ee4a4da84b8d6ea2faa0fc0f44a2de15bb00`
- Terraform: `1.16.4`
- OpenTofu: `1.13.0`

Required commands: `amp`, `git`, `go`, `curl`, `jq`, `/usr/bin/time`,
`terraform`, and `tofu`.

## Run

```bash
chmod +x run-matrix.sh lifecycle.sh
./run-matrix.sh
```

The runner always stops the supervised orb service. By default it also removes
the temporary clones, SQLite database, state, logs, and mode-0600 token files.
Set `KEEP_E2E_ARTIFACTS=1` to retain the temporary run directory for diagnosis:

```bash
KEEP_E2E_ARTIFACTS=1 ./run-matrix.sh
```

To reproduce another pushed provider revision without changing the checkout:

```bash
PROVIDER_SHA=<full-commit-sha> ./run-matrix.sh
```

No `terraform init`/`tofu init` is needed: `lifecycle.sh` writes a CLI config
using `provider_installation.dev_overrides` pointed at the exact locally built
provider binary.

## Matrix

For each CLI, in a separate organization:

1. Validate, create 14 resources, and resolve 14 data sources.
2. Assert a stable detailed-exitcode plan (`0`).
3. Apply mutable updates and assert another stable plan.
4. Remove and import all 14 resource types, then assert a stable plan.
5. Apply external drift through the real API; require plan exit `2`, repair,
   and assert stability.
6. Change the immutable instance image and assert that its ID changes.
7. Install a deterministic one-shot project-read fault (`seed=42`,
   `max_triggers=1`, HTTP 503); require plan exit `1`, then recovery plan `0`.
8. Reset and delete that exact fault rule.
9. Require 409 restrict responses for an in-use disk, subnet, and project.
10. Remove data sources from configuration before deleting backing resources.
11. Delete a bucket, verify its object cascades to 404, require plan exit `2`,
    recreate, and assert stability.
12. Delete attachment and policy binding before their parent instance/API key,
    then independently delete instance, object, metadata, and API key; require
    plan exit `2`, recreate, and assert stability.
13. Restore all data sources and assert stability.
14. Destroy, require empty state, and require project lookup 404.

## Harness-only fixes incorporated

These fixes were made while iterating on the harness; none changed either
pinned source clone:

1. `version` was renamed to `config_version` because Terraform reserves the
   former as an input variable name.
2. The Bash `api()` locals were split across two declarations so `out` does
   not expand `label` before assignment under `set -u`.
3. Fault reset and fault deletion are separate operations. Resetting only
   counters caused the rule to trigger again; the final harness resets and then
   deletes the exact rule ID.
4. Data sources are removed before remote backing-resource deletion because a
   missing data-source target is correctly a hard read error, not state removal.
5. Disk attachment and policy binding are deleted before instance and API key;
   deleting the parents first correctly cascades the children and makes later
   child deletes return 404.
6. The bootstrap token is persisted only as a mode-0600 temporary file so an
   interrupted run can be diagnosed or manually destroyed; deleting the run
   directory and SQLite database guarantees cleanup even after early failure.

The successful observed totals were 11.49 seconds for Terraform and 11.72
seconds for OpenTofu, excluding builds. Provider build was 16.01 seconds and
server build was 53.62 seconds on the original orb.
