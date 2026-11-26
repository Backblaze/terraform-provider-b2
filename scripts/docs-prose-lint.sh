#!/usr/bin/env bash
# Check changed hand-written Markdown. Generated documentation is excluded.
set -euo pipefail
base=${BASE_REF:-origin/master}
mode=strict
while [ "$#" -gt 0 ]; do
  case "$1" in
    --base) if [ "$#" -lt 2 ]; then echo 'ERROR: --base requires a ref' >&2; exit 2; fi; base=$2; shift 2 ;;
    --strict) mode=strict; shift ;;
    --warn-only) mode=warn; shift ;;
    *) echo "ERROR: unknown argument: $1" >&2; exit 2 ;;
  esac
done
root=$(git rev-parse --show-toplevel)
cd "$root"
rootp=$(pwd -P)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
if ! ancestor=$(git merge-base "$base" HEAD); then
  echo "ERROR: cannot find merge base for $base" >&2
  exit 2
fi
git diff --name-only -z --diff-filter=ACMR "$ancestor" -- '*.md' '*.md.tmpl' > "$tmp/files"
git ls-files --others --exclude-standard -z -- '*.md' '*.md.tmpl' >> "$tmp/files"
errors=0
report() {
  if [ "$mode" = warn ]; then echo "WARNING: $*"; else echo "ERROR: $*"; fi
  errors=$((errors + 1))
}
while IFS= read -r -d '' file; do
  case "$file" in docs/*|CHANGELOG.md|.git/*) continue ;; esac
  [ -f "$file" ] || continue
  if [ "$(head -n 1 "$file")" = '---' ] && ! awk 'NR > 1 && /^---[[:space:]]*$/ {found=1; exit} END {exit !found}' "$file"; then
    report "$file: unclosed front matter"
  fi
  : > "$tmp/undefined"
  awk -v missing="$tmp/undefined" '
    function norm(x) {gsub(/[[:space:]]+/, " ", x); sub(/^ /, "", x); sub(/ $/, "", x); return tolower(x)}
    /^ ? ? ?(```|~~~)/ {
      marker=$0; sub(/^[[:space:]]*/, "", marker)
      char=substr(marker,1,1); run=marker
      if (char == "`") sub(/[^`].*$/, "", run); else sub(/[^~].*$/, "", run)
      info=substr(marker,length(run)+1)
      if (!fence && char == "`" && index(info, "`")) {
        # Not a fence: a backtick fence info string cannot contain a backtick. Treat as text.
      } else {
        if (!fence) {fence=char; length_open=length(run)}
        else if (char == fence && length(run) >= length_open && info ~ /^[[:space:]]*$/) fence=""
        next
      }
    }
    fence {next}
    { line=$0; gsub(/`[^`]*`/, "", line)
      if (line ~ /^[[:space:]]*\[[^]]+\]:[[:space:]]*/) {
        label=line; sub(/^[[:space:]]*\[/, "", label); sub(/\].*$/, "", label)
        defs[norm(label)]=1
        sub(/^[[:space:]]*\[[^]]+\]:[[:space:]]*/, "", line)
        print line; next
      }
      scan=line
      while (match(scan, /\[[^][]+\]\[[^][]*\]/)) {
        ref=substr(scan,RSTART,RLENGTH); label=ref
        sub(/^\[[^]]+\]\[/, "", label); sub(/\]$/, "", label)
        if (label=="") {label=ref; sub(/^\[/, "", label); sub(/\].*$/, "", label)}
        refs[norm(label)]=1; scan=substr(scan,RSTART+RLENGTH)
      }
      while (match(line, /\]\(/)) {
        rest=substr(line,RSTART+2); depth=1; angle=0; end=0
        for (i=1; i<=length(rest); i++) {
          c=substr(rest,i,1)
          if (i==1 && c=="<") angle=1
          else if (angle && c==">") angle=0
          else if (!angle && c=="(") depth++
          else if (!angle && c==")") {depth--; if (depth==0) {end=i; break}}
        }
        if (!end) break
        print substr(rest,1,end-1); line=substr(rest,end+1)
      }
    }
    END {for (label in refs) if (!(label in defs)) print label > missing}
  ' "$file" > "$tmp/links"
  while IFS= read -r label; do
    [ -n "$label" ] && report "$file: undefined reference $label"
  done < "$tmp/undefined"
  while IFS= read -r target; do
    case "$target" in
      '<'*'>'*) target=${target#<}; target=${target%%>*} ;;
      *) target=${target%%[[:space:]]*} ;;
    esac
    target=${target%%\#*}; target=${target%%\?*}
    case "$target" in ''|*://*|mailto:*|tel:*|*'{{'*) continue ;; esac
    encoded=$(printf '%s' "$target" | sed -E 's/%([0-9A-Fa-f]{2})/\\x\1/g')
    target=$(printf '%b' "$encoded")
    case "$target" in /*) resolved="$root$target" ;; *) resolved="$(dirname "$file")/$target" ;; esac
    dir=$(cd "$(dirname "$resolved")" 2>/dev/null && pwd -P) || dir=
    case "$dir/" in
      "$rootp"/*) ;;
      *) if [ -n "$dir" ]; then report "$file: link points outside the repository $target"; else report "$file: broken link $target"; fi; continue ;;
    esac
    if [ ! -e "$resolved" ]; then report "$file: broken link $target"; fi
  done < "$tmp/links"
done < "$tmp/files"
printf 'docs prose lint: %s problem(s)\n' "$errors"
if [ "$mode" = strict ] && [ "$errors" -gt 0 ]; then exit 1; fi
