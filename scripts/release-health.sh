#!/usr/bin/env bash
# Local, non-publishing release tooling and unit-test check.
set -euo pipefail
bash "$(dirname "$0")/changelog-validate.sh"
env -u TF_ACC make test
go vet ./...
git diff --check
echo 'Release health: PASS (live acceptance tests and publication not run)'
