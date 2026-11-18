#!/usr/bin/env bash
# Validate generated notes on ordinary PRs and already-batched release PRs alike.
set -euo pipefail
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
go run github.com/miniscruff/changie@v1.24.0 merge --dry-run > "$tmp"
diff -u CHANGELOG.md "$tmp"
if [ -d .changes/unreleased ] && find .changes/unreleased -maxdepth 1 -type f \( -name '*.yaml' -o -name '*.yml' \) -print -quit | grep -q .; then
  go run github.com/miniscruff/changie@v1.24.0 batch auto --dry-run >/dev/null
fi
