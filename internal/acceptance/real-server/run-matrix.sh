#!/usr/bin/env bash
set -euo pipefail

fixture_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
provider_repo=$(git -C "$fixture_dir" rev-parse --show-toplevel)
provider_sha=${PROVIDER_SHA:-$(git -C "$provider_repo" rev-parse HEAD)}
server_sha=9131ee4a4da84b8d6ea2faa0fc0f44a2de15bb00
service_name=nah-e2e
root=$(mktemp -d /tmp/nah-real-server-e2e.XXXXXX)
service_started=0

cleanup() {
  if (( service_started )); then
    amp orb service stop "$service_name" || true
  fi
  if [[ ${KEEP_E2E_ARTIFACTS:-0} == 1 ]]; then
    printf 'Preserved run directory: %s\n' "$root"
  else
    rm -rf "$root"
  fi
}
trap cleanup EXIT

terraform version | head -2
tofu version | head -2
[[ $(terraform version -json | jq -r .terraform_version) == 1.16.4 ]]
[[ $(tofu version -json | jq -r .terraform_version) == 1.13.0 ]]

mkdir -p "$root/bin" "$root/logs" "$root/server-data"
cp "$fixture_dir/main.tf" "$fixture_dir/lifecycle.sh" "$root/"
chmod +x "$root/lifecycle.sh"

git clone --quiet --no-checkout https://github.com/hypertf/terraform-provider-nah.git "$root/provider"
git -C "$root/provider" fetch --quiet origin "$provider_sha"
git -C "$root/provider" checkout --quiet --detach "$provider_sha"
test "$(git -C "$root/provider" rev-parse HEAD)" = "$provider_sha"

git clone --quiet --no-checkout https://github.com/hypertf/nahcloud.git "$root/server"
git -C "$root/server" fetch --quiet origin "$server_sha"
git -C "$root/server" checkout --quiet --detach "$server_sha"
test "$(git -C "$root/server" rev-parse HEAD)" = "$server_sha"

(
  cd "$root/provider"
  /usr/bin/time -f 'provider_build_seconds=%e' -o "$root/logs/provider-build.time" \
    go build -o "$root/bin/terraform-provider-nah" .
)
(
  cd "$root/server"
  /usr/bin/time -f 'server_build_seconds=%e' -o "$root/logs/server-build.time" \
    go build -ldflags "-X main.Version=$server_sha" \
      -o "$root/bin/nahcloud-server" ./cmd/server
)

amp orb service start "$service_name" --command \
  "'$root/bin/nahcloud-server' --addr 127.0.0.1:18080 --sqlite-dsn '$root/server-data/nahcloud.sqlite'"
service_started=1

for _ in $(seq 1 100); do
  if curl -fsS http://127.0.0.1:18080/buildz >"$root/logs/buildz.json"; then
    break
  fi
  sleep 0.1
done
jq -e --arg sha "$server_sha" '.version == $sha' "$root/logs/buildz.json" >/dev/null

/usr/bin/time -f 'terraform_total_seconds=%e' -o "$root/logs/terraform-total.time" \
  "$root/lifecycle.sh" terraform "$root"
/usr/bin/time -f 'tofu_total_seconds=%e' -o "$root/logs/tofu-total.time" \
  "$root/lifecycle.sh" tofu "$root"

grep -F 'RESULT PASS cli=terraform' "$root/logs/terraform.log"
grep -F 'RESULT PASS cli=tofu' "$root/logs/tofu.log"
test "$(TF_CLI_CONFIG_FILE="$root/terraform/cli.tfrc" terraform -chdir="$root/terraform" state list | wc -l)" -eq 0
test "$(TF_CLI_CONFIG_FILE="$root/tofu/cli.tfrc" tofu -chdir="$root/tofu" state list | wc -l)" -eq 0
git -C "$root/provider" diff --quiet
git -C "$root/server" diff --quiet

cat "$root/logs/provider-build.time" "$root/logs/server-build.time" \
  "$root/logs/terraform-total.time" "$root/logs/tofu-total.time"
