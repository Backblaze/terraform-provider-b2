#!/usr/bin/env bash
# Check the index structure, repository links, and provider documentation coverage.
set -euo pipefail
cd "$(dirname "$0")/.."

if [ ! -f llms.txt ]; then
  echo 'FAIL llms.txt: file not found'
  exit 1
fi
if ! head -n 1 llms.txt | grep -q '^# [^#]'; then
  echo 'FAIL llms.txt: first line must be an H1'
  exit 1
fi
if ! awk '/^## / {section=1; exit} /^> / {found=1} END {exit !(found && section)}' llms.txt; then
  echo 'FAIL llms.txt: summary must precede the first H2'
  exit 1
fi
repo='https://github.com/Backblaze/terraform-provider-b2'
prefix="$repo/blob/master/"
links=$(grep -Eo ']\([^)]+\)' llms.txt | sed -e 's/^](//' -e 's/)$//' || true)
if [ -z "$links" ]; then
  echo 'FAIL llms.txt: no links found'
  exit 1
fi
while IFS= read -r link; do
  lower=$(printf '%s' "$link" | tr '[:upper:]' '[:lower:]')
  case "$lower" in
    https://github.com/backblaze/terraform-provider-b2*)
      # Exact case and the blob/tree route that matches the target type.
      case "$link" in
        "$repo/blob/master/"*) path=${link#"$repo/blob/master/"}; want=file ;;
        "$repo/tree/master/"*) path=${link#"$repo/tree/master/"}; want=dir ;;
        *) echo "FAIL llms.txt: invalid repository ref: $link"; exit 1 ;;
      esac
      if [ ! -e "$path" ]; then
        echo "FAIL llms.txt: missing path $path"
        exit 1
      fi
      if [ "$want" = file ] && [ ! -f "$path" ]; then
        echo "FAIL llms.txt: $path is a directory; use /tree/master/ ($link)"
        exit 1
      fi
      if [ "$want" = dir ] && [ ! -d "$path" ]; then
        echo "FAIL llms.txt: $path is a file; use /blob/master/ ($link)"
        exit 1
      fi
      ;;
    http://*|https://*) ;;
    *) echo "FAIL llms.txt: use an absolute URL, not a relative link: $link"; exit 1 ;;
  esac
done <<< "$links"
echo 'ok   llms.txt: structure and repository links'

for kind in ResourcesMap DataSourcesMap; do
  case "$kind" in ResourcesMap) directory=resources ;; DataSourcesMap) directory=data-sources ;; esac
  names=$(awk -v kind="$kind" '$0 ~ kind ": map\\[string\\]\\*schema.Resource\\{" {inside=1; next} inside && /^[[:space:]]*},/ {exit} inside && /"b2_/ {line=$0; sub(/^[^"]*"b2_/, "", line); sub(/".*/, "", line); print line}' b2/provider.go)
  if [ -z "$names" ]; then
    echo "FAIL b2/provider.go: no $kind entries found"
    exit 1
  fi
  while IFS= read -r name; do
    page="docs/$directory/$name.md"
    if [ ! -f "$page" ]; then
      echo "FAIL $kind: missing documentation page $page"
      exit 1
    fi
    if ! grep -Fq "${prefix}${page})" llms.txt; then
      echo "FAIL llms.txt: missing $kind entry for b2_$name ($page)"
      exit 1
    fi
  done <<< "$names"
done
echo 'ok   llms.txt: provider resources and data sources'
