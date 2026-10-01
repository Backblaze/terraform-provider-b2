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
}
EOF
}

# A bare IPv4 loopback HTTP origin (http://127.0.0.1:<port>), nothing else.
sdkharness_valid_loopback_origin() {
  python3 -c '
import sys
from urllib.parse import urlsplit
url = urlsplit(sys.argv[1])
valid = (url.scheme == "http" and url.hostname == "127.0.0.1" and url.port is not None
         and url.username is None and url.password is None and url.path in ("", "/")
         and not url.query and not url.fragment)
raise SystemExit(0 if valid else 1)
' "$1"
}

# Drop every ambient B2_* variable: the check sets the simulator's fixed credential itself,
# so a real key can never reach the provider.
sdkharness_drop_ambient_b2_env() {
  local name
  for name in $(compgen -e); do
    case "$name" in B2_*) unset "$name" ;; esac
  done
}

# Assertions may talk to the loopback simulator only. Point every proxy-aware client at a dead
# proxy and exempt loopback: a connection to anything else fails at once with a "proxyconnect" /
# ProxyError, which sdkharness_tf_trouble_is_proxy_refusal recognises and reports as a FAIL
# instead of an amber "unreachable". (Terraform's `init` would contact the provider registry even
# with dev_overrides, so checks do not run it: dev_overrides needs no installation.)
sdkharness_tf_network_guard() {
  export HTTP_PROXY=http://127.0.0.1:1 HTTPS_PROXY=http://127.0.0.1:1 ALL_PROXY=http://127.0.0.1:1
  export http_proxy=http://127.0.0.1:1 https_proxy=http://127.0.0.1:1 all_proxy=http://127.0.0.1:1
  export NO_PROXY=127.0.0.1,localhost,::1 no_proxy=127.0.0.1,localhost,::1
  export CHECKPOINT_DISABLE=1
}

# $1: lower-cased terraform output. True when a non-loopback connection was refused by the guard.
sdkharness_tf_trouble_is_proxy_refusal() {
  case "$1" in *proxyconnect*|*"cannot connect to proxy"*|*proxyerror*) return 0 ;; esac
  return 1
}
