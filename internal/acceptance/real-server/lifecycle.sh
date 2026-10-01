#!/usr/bin/env bash
set -euo pipefail

cli=$1
root=$2
endpoint=http://127.0.0.1:18080
dir="$root/$cli"
log="$root/logs/$cli.log"
timings="$root/logs/$cli.timings"
mkdir -p "$dir"
cp "$root/main.tf" "$dir/main.tf"
cat >"$dir/cli.tfrc" <<EOF
provider_installation {
  dev_overrides { "registry.terraform.io/hypertf/nah" = "$root/bin" }
  direct {}
}
EOF
cat >"$dir/terraform.tfvars" <<EOF
config_version = 1
image = "ubuntu:24.04"
EOF

slug="e2e-$cli-$(date +%s%N)"
org_json=$(curl -fsS -X POST "$endpoint/v1/orgs" -H 'Content-Type: application/json' \
  --data "{\"slug\":\"$slug\",\"name\":\"Pinned $cli E2E\"}")
token=$(jq -er '.api_key.token' <<<"$org_json")
org_id=$(jq -er '.id' <<<"$org_json")
umask 077
printf '%s' "$token" >"$dir/.bootstrap-token"
printf 'organization=%s org_id=%s\n' "$slug" "$org_id" >>"$log"
export TF_CLI_CONFIG_FILE="$dir/cli.tfrc" TF_IN_AUTOMATION=1 TF_INPUT=0
export NAH_ENDPOINT="$endpoint" NAH_TOKEN="$token"

run() {
  local expected=$1 label=$2; shift 2
  local start end rc
  start=$(date +%s%N)
  set +e
  (cd "$dir" && "$cli" "$@") >>"$log" 2>&1
  rc=$?
  set -e
  end=$(date +%s%N)
  awk -v l="$label" -v n="$((end-start))" 'BEGIN {printf "%s %.3f\n", l, n/1000000000}' >>"$timings"
  printf '%s exit=%d expected=%d\n' "$label" "$rc" "$expected" >>"$log"
  if [[ $rc -ne $expected ]]; then
    echo "$label: exit $rc, expected $expected" >&2
    tail -80 "$log" >&2
    return 1
  fi
}

state_json() { (cd "$dir" && "$cli" show -json); }
id() { state_json | jq -er --arg a "nah_$1.test" '.values.root_module.resources[] | select(.address==$a) | .values.id'; }
api() {
  local expected=$1 label=$2 method=$3 path=$4 body=${5-}
  local out="$root/logs/$cli-$label.json" code
  if [[ -n $body ]]; then
    code=$(curl -sS -o "$out" -w '%{http_code}' -X "$method" "$endpoint$path" -H "Authorization: Bearer $token" -H 'Content-Type: application/json' --data "$body")
  else
    code=$(curl -sS -o "$out" -w '%{http_code}' -X "$method" "$endpoint$path" -H "Authorization: Bearer $token")
  fi
  printf 'api %s %s %s status=%s expected=%s payload=%s\n' "$label" "$method" "$path" "$code" "$expected" "$(jq -c . "$out" 2>/dev/null || cat "$out")" >>"$log"
  [[ $code == "$expected" ]]
}

run 0 validate validate -no-color
run 0 create apply -auto-approve -no-color
run 0 stable_initial plan -detailed-exitcode -no-color
initial_instance=$(id instance)

cat >"$dir/terraform.tfvars" <<EOF
config_version = 2
image = "ubuntu:24.04"
EOF
run 0 update apply -auto-approve -no-color
run 0 stable_updated plan -detailed-exitcode -no-color

declare -A ids
for name in project instance bucket object metadata api_key network subnet disk disk_attachment policy policy_binding load_balancer load_balancer_backend; do ids[$name]=$(id "$name"); done
declare -A imports=(
  [project]="test-project"
  [instance]="test-project/${ids[instance]}"
  [bucket]="test-project/${ids[bucket]}"
  [object]="test-project/${ids[bucket]}/${ids[object]}"
  [metadata]="${ids[metadata]}"
  [api_key]="${ids[api_key]}"
  [network]="test-project/${ids[network]}"
  [subnet]="test-project/${ids[network]}/${ids[subnet]}"
  [disk]="test-project/${ids[disk]}"
  [disk_attachment]="test-project/${ids[disk]}/${ids[disk_attachment]}"
  [policy]="${ids[policy]}"
  [policy_binding]="${ids[policy]}/${ids[policy_binding]}"
  [load_balancer]="test-project/${ids[load_balancer]}"
  [load_balancer_backend]="test-project/${ids[load_balancer]}/${ids[load_balancer_backend]}"
)
for name in project instance bucket object metadata api_key network subnet disk disk_attachment policy policy_binding load_balancer load_balancer_backend; do
  run 0 "state_rm_$name" state rm "nah_$name.test"
  run 0 "import_$name" import -no-color "nah_$name.test" "${imports[$name]}"
done
run 0 stable_imports plan -detailed-exitcode -no-color

# External mutable drift, then repair.
api 200 drift_project PATCH /v1/projects/test-project '{"name":"external"}'
api 200 drift_metadata PATCH "/v1/metadata/${ids[metadata]}" '{"value":"external"}'
api 200 drift_bucket PATCH "/v1/projects/test-project/buckets-by-id/${ids[bucket]}" '{"name":"external-bucket"}'
api 200 drift_network PATCH "/v1/projects/test-project/networks/${ids[network]}" '{"name":"external-network"}'
api 200 drift_backend PATCH "/v1/projects/test-project/load-balancers/${ids[load_balancer]}/backends/${ids[load_balancer_backend]}" '{"weight":99}'
run 2 drift_plan plan -detailed-exitcode -no-color
run 0 drift_repair apply -auto-approve -no-color
run 0 stable_after_drift plan -detailed-exitcode -no-color

# Immutable image replacement.
old_instance=$(id instance)
cat >"$dir/terraform.tfvars" <<EOF
config_version = 2
image = "debian:12"
EOF
run 0 replacement apply -auto-approve -no-color
new_instance=$(id instance)
[[ $new_instance != "$old_instance" ]]
run 0 stable_after_replacement plan -detailed-exitcode -no-color

# Deterministic one-shot 503: first project refresh fails, second succeeds.
fault_body='{"name":"one-shot project read","operation":"projects.get","after_matches":0,"every_nth":1,"failure_percent":100,"seed":42,"max_triggers":1,"status_code":503}'
api 201 fault_create POST /v1/fault-rules "$fault_body"
fault_id=$(jq -er '.id' "$root/logs/$cli-fault_create.json")
run 1 deterministic_fault plan -no-color
run 0 fault_recovery plan -detailed-exitcode -no-color
api 200 fault_list GET /v1/fault-rules
api 200 fault_reset POST "/v1/fault-rules/$fault_id/reset"
api 204 fault_delete DELETE "/v1/fault-rules/$fault_id"

# Restrict semantics while dependencies exist.
api 409 restrict_disk DELETE "/v1/projects/test-project/disks/${ids[disk]}"
api 409 restrict_subnet DELETE "/v1/projects/test-project/networks/${ids[network]}/subnets/${ids[subnet]}"
api 409 restrict_project DELETE /v1/projects/test-project

# Missing backing resources are an expected hard error for data sources, so
# remove them before testing remote deletion and restore them afterward.
cp "$dir/main.tf" "$dir/main-with-data-sources.tf.disabled"
sed -i '/^# Data sources$/,$d' "$dir/main.tf"
run 0 remove_data_sources apply -auto-approve -no-color

# Cascade bucket->object, then Terraform observes/recreates both.
current_bucket=$(id bucket); current_object=$(id object)
api 204 cascade_bucket DELETE "/v1/projects/test-project/buckets-by-id/$current_bucket"
api 404 cascade_object_get GET "/v1/projects/test-project/buckets-by-id/$current_bucket/objects/$current_object"
run 2 cascade_plan plan -detailed-exitcode -no-color
run 0 cascade_recreate apply -auto-approve -no-color
run 0 stable_after_cascade plan -detailed-exitcode -no-color

# Independent remote deletes and convergence.
for name in instance metadata api_key disk_attachment policy_binding; do ids[$name]=$(id "$name"); done
ids[bucket]=$(id bucket); ids[object]=$(id object); ids[disk]=$(id disk); ids[policy]=$(id policy)
api 204 delete_attachment DELETE "/v1/projects/test-project/disks/${ids[disk]}/attachments/${ids[disk_attachment]}"
api 204 delete_binding DELETE "/v1/policies/${ids[policy]}/bindings/${ids[policy_binding]}"
api 204 delete_instance DELETE "/v1/projects/test-project/instances/${ids[instance]}"
api 204 delete_object DELETE "/v1/projects/test-project/buckets-by-id/${ids[bucket]}/objects/${ids[object]}"
api 204 delete_metadata DELETE "/v1/metadata/${ids[metadata]}"
api 204 delete_api_key DELETE "/v1/api-keys/${ids[api_key]}"
run 2 remote_delete_plan plan -detailed-exitcode -no-color
run 0 remote_delete_recreate apply -auto-approve -no-color
run 0 stable_final plan -detailed-exitcode -no-color

mv "$dir/main-with-data-sources.tf.disabled" "$dir/main.tf"
run 0 restore_data_sources apply -auto-approve -no-color
run 0 stable_with_data_sources plan -detailed-exitcode -no-color

run 0 destroy destroy -auto-approve -no-color
run 0 empty_state state list
api 404 project_absent GET /v1/projects/test-project
printf 'RESULT PASS cli=%s initial_instance=%s replacement_instance=%s\n' "$cli" "$initial_instance" "$new_instance" >>"$log"
