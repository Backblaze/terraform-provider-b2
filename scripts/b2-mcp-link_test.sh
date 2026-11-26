#!/usr/bin/env bash
set -eu
cd "$(dirname "$0")/.."
expected='- **[b2-mcp](https://github.com/backblaze-labs/b2-mcp)** — a Model Context Protocol server for B2.'
for file in README.md llms.txt AGENTS.md; do
  if [ "$file" != README.md ] && [ ! -f "$file" ]; then
    echo "SKIP $file (absent)"
    continue
  fi
  if ! grep -Fqx -- "$expected" "$file"; then
    echo "FAIL $file: expected b2-mcp link missing" >&2
    exit 1
  fi
  echo "PASS $file"
done
