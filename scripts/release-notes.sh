#!/usr/bin/env bash
# Read only Changie's assembled notes for an exact release tag.
set -euo pipefail
tag=${1:?usage: release-notes.sh vX.Y.Z}
[[ "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]] || { echo 'Invalid release tag' >&2; exit 1; }
file=".changes/$tag.md"
test -s "$file" || { echo 'Missing assembled release notes' >&2; exit 1; }
IFS= read -r heading < "$file"
[[ "$heading" == "## [${tag#v}] - "* ]] || { echo 'Release notes version does not match tag' >&2; exit 1; }
awk 'NR > 1 { print }' "$file"
