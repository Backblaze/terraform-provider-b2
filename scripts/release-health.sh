#!/usr/bin/env bash
# Local, non-publishing release tooling and unit-test check.
set -euo pipefail
if [ "$#" -gt 1 ] || { [ "$#" -eq 1 ] && [ "$1" != --changelog-only ]; }; then
  echo 'usage: release-health.sh [--changelog-only]' >&2
  exit 2
fi
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
go run github.com/miniscruff/changie@v1.24.0 merge --dry-run > "$tmp"
diff -u CHANGELOG.md "$tmp"
if [ -d .changes/unreleased ] && find .changes/unreleased -maxdepth 1 -type f \( -name '*.yaml' -o -name '*.yml' \) -print -quit | grep -q .; then
  go run github.com/miniscruff/changie@v1.24.0 batch auto --dry-run >/dev/null
fi
if [ "${1:-}" = --changelog-only ]; then
  echo 'Changelog validation: PASS'
  exit 0
fi
env -u TF_ACC make test
go vet ./...
git diff --check
echo 'Release health: PASS (live acceptance tests and publication not run)'
