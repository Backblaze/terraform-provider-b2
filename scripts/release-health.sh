#!/usr/bin/env bash
# Local, non-publishing release tooling and unit-test check.
set -euo pipefail
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
go run github.com/miniscruff/changie@v1.24.0 merge --dry-run > "$tmp"
diff -u CHANGELOG.md "$tmp"
go run github.com/miniscruff/changie@v1.24.0 batch auto --dry-run >/dev/null
env -u TF_ACC make test
go vet ./...
git diff --check
echo 'Release health: PASS (live acceptance tests and publication not run)'
