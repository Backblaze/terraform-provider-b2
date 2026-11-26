#!/usr/bin/env bash
# Validate examples against a locally built provider without credentials or downloads.
set -euo pipefail
cd "$(dirname "$0")/.."
root=$PWD
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
platform="$(go env GOOS)_$(go env GOARCH)"
mkdir -p "$tmp/provider" "$tmp/mirror/registry.terraform.io/backblaze/b2/0.0.0/$platform"
test -f b2/py-terraform-provider-b2 || touch b2/py-terraform-provider-b2
go build -tags netgo -o "$tmp/provider/terraform-provider-b2" .
cp "$tmp/provider/terraform-provider-b2" "$tmp/mirror/registry.terraform.io/backblaze/b2/0.0.0/$platform/terraform-provider-b2_v0.0.0"
cat > "$tmp/terraform.rc" <<CONFIG
provider_installation {
  dev_overrides {
    "registry.terraform.io/backblaze/b2" = "$tmp/provider"
  }
  filesystem_mirror {
    path = "$tmp/mirror"
  }
}
CONFIG
export TF_CLI_CONFIG_FILE="$tmp/terraform.rc"
unset B2_APPLICATION_KEY B2_APPLICATION_KEY_ID TF_CLI_ARGS TF_CLI_ARGS_init TF_CLI_ARGS_validate
for dir in "$root"/examples/*/; do
  [ -d "$dir" ] || continue
  name=$(basename "$dir")
  echo "Checking examples/$name"
  mkdir -p "$tmp/examples/$name"
  cp -R "$dir". "$tmp/examples/$name/"
  (cd "$tmp/examples/$name" && terraform init -backend=false -input=false -no-color && terraform validate -no-color)
done
