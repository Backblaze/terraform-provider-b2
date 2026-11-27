#!/usr/bin/env bash
# PostToolUse: reject an edit that leaves a Go file unformatted or a Go/Python file without the license
# header that `make lint` requires; exit 2 hands the message back to Claude.
set -euo pipefail
command -v jq > /dev/null || { echo "check-edit hook: jq not found, skipping (make lint still checks)." >&2; exit 0; }
file=$(jq -r '.tool_input.file_path // empty')
[ -f "$file" ] || exit 0
case "$file" in
  *.go | *.py) ;;
  *) exit 0 ;;
esac
rel=${file#"${CLAUDE_PROJECT_DIR:-$PWD}/"}
case "$rel" in
  python-bindings/build/* | python-bindings/dist/*) exit 0 ;;
esac
msg=""
if ! head -9 "$file" | grep -q 'All Rights Reserved' || ! head -9 "$file" | grep -qF "File: $rel"; then
  msg="$rel is missing the license header (copy it from a sibling file, with 'File: $rel'; see scripts/check-headers.py)."
fi
if [ "${file##*.}" = go ] && command -v gofmt > /dev/null; then
  if ! out=$(gofmt -l "$file" 2>&1) || [ -n "$out" ]; then
    msg="${msg:+$msg }gofmt: $rel is not formatted. Run 'gofmt -w $rel'."
  fi
fi
[ -z "$msg" ] && exit 0
echo "$msg" >&2
exit 2
