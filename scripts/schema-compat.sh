#!/usr/bin/env bash
# Compare the provider schema at a base ref with the working tree, offline.
# usage: scripts/schema-compat.sh [--base REF] [--warn-only]
set -euo pipefail
cd "$(dirname "$0")/.."
root=$PWD
base=${BASE_REF:-origin/master}
warn_only=0
while [ "$#" -gt 0 ]; do
  case "$1" in
    --base) if [ "$#" -lt 2 ]; then echo 'ERROR: --base requires a ref' >&2; exit 2; fi; base=$2; shift 2 ;;
    --warn-only) warn_only=1; shift ;;
    *) echo "ERROR: unknown argument: $1" >&2; exit 2 ;;
  esac
done
tmp=$(mktemp -d)
base_tree="$tmp/base-tree"
# shellcheck disable=SC2329 # invoked through the EXIT trap
cleanup() {
  git -C "$root" worktree remove --force "$base_tree" 2>/dev/null || true
  rm -rf "$tmp"
}
trap cleanup EXIT
unset B2_APPLICATION_KEY B2_APPLICATION_KEY_ID TF_CLI_ARGS TF_CLI_ARGS_init TF_CLI_ARGS_providers

# dump_schema SRC_DIR OUT_JSON: build the provider in SRC_DIR and dump its schema through a dev override.
dump_schema() {
  local src=$1 out=$2 work
  work="$tmp/work-$(basename "$out" .json)"
  mkdir -p "$work/provider"
  # The Go build embeds this file; an empty stub is enough because no resource runs.
  test -f "$src/b2/py-terraform-provider-b2" || touch "$src/b2/py-terraform-provider-b2"
  (cd "$src" && go build -tags netgo -o "$work/provider/terraform-provider-b2" .)
  cat > "$work/terraform.rc" <<CONFIG
provider_installation {
  dev_overrides {
    "registry.terraform.io/backblaze/b2" = "$work/provider"
  }
}
CONFIG
  mkdir -p "$work/config"
  cat > "$work/config/main.tf" <<'CONFIG'
terraform {
  required_providers {
    b2 = {
      source = "Backblaze/b2"
    }
  }
}
CONFIG
  (cd "$work/config" && TF_CLI_CONFIG_FILE="$work/terraform.rc" terraform providers schema -json > "$out")
}

echo "Base ref: $base"
git worktree add --detach "$base_tree" "$base" >/dev/null
dump_schema "$base_tree" "$tmp/base.json"
dump_schema "$root" "$tmp/head.json"

rc=0
python3 -I "$root/scripts/schema-compat-compare.py" "$tmp/base.json" "$tmp/head.json" || rc=$?
if [ "$rc" -eq 1 ] && [ "$warn_only" -eq 1 ]; then
  echo 'Breaking changes found; exiting 0 because of --warn-only.'
  rc=0
fi
exit "$rc"
