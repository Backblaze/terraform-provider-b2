#!/usr/bin/env bash
# Preserve the former action's all-files contract and -changelog opt-out.
set -euo pipefail
base=${1:?usage: changelog-check.sh BASE}
git rev-parse --verify "$base^{commit}" >/dev/null
files=$(git diff --name-only --diff-filter=AMR "$base"...HEAD)
if git cat-file -e "$base:.changie.yaml" 2>/dev/null && grep -qx 'CHANGELOG.md' <<< "$files"; then
  head=${PR_HEAD_REF:-$(git symbolic-ref --quiet --short HEAD || true)}
  [[ "$head" =~ ^release/v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]] || {
    echo 'Assemble CHANGELOG.md only on release/vX.Y.Z' >&2; exit 1;
  }
  bash "$(dirname "$0")/release-notes.sh" "${head#release/}" >/dev/null
  while IFS= read -r file; do
    case "$file" in CHANGELOG.md|.changes/*) ;; *) echo 'Release PRs may change only CHANGELOG.md and .changes/' >&2; exit 1 ;; esac
  done <<< "$files"
fi
if python3 -c 'import json,os,sys; sys.exit(0 if "-changelog" in json.loads(os.environ.get("PR_LABELS_JSON", "[]")) else 1)'; then
  echo 'Changelog presence waived by -changelog label'
  exit 0
fi
grep -Eq '^\.changes/(unreleased/[^/]+\.ya?ml|v[0-9][^/]*\.md)$' <<< "$files" || {
  echo 'Add a Changie fragment, or have a maintainer apply -changelog' >&2; exit 1;
}
echo 'Changelog presence: PASS'
