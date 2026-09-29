#!/usr/bin/env bash
# Shared exact-revision Terraform CLI configuration for repository-owned tests.

sdkharness_tf_cli_config() {
  local destination="$1"
  local provider_dir="${SDKHARNESS_TERRAFORM_PROVIDER_DIR:-}"
  local encoded
  [ -n "$provider_dir" ] || return 1
  [ -x "$provider_dir/terraform-provider-b2" ] || return 1
  encoded="$(python3 -c 'import json, sys; print(json.dumps(sys.argv[1]))' "$provider_dir")" \
    || return 1
  cat >"$destination" <<EOF
provider_installation {
  dev_overrides {
    "Backblaze/b2" = $encoded
  }
  direct {}
}
EOF
}
