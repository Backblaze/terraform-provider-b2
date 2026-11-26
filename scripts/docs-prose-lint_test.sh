#!/usr/bin/env bash
set -euo pipefail
script=$(cd "$(dirname "$0")" && pwd)/docs-prose-lint.sh
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
count=0
case_run() {
  local name=$1 expected=$2 content=$3 setup=${4:-} mode=${5:---strict} pattern=${6:-}
  count=$((count + 1))
  local dir="$tmp/$count" result
  mkdir -p "$dir" && cd "$dir"
  git init -q
  git config user.name Test
  git config user.email test@example.invalid
  printf 'ok\n' > existing.md
  printf '[bad](missing.md)\n' > unchanged.md
  git add . && git commit -qm baseline
  local base; base=$(git rev-parse HEAD)
  printf '%s\n' "$content" > changed.md
  case "$setup" in
    file) printf 'ok\n' > 'a file.md' ;;
    paren) printf 'ok\n' > 'notes_(v2).md' ;;
    skip) mkdir docs; printf '[bad](missing.md)\n' > docs/bad.md; printf '[bad](missing.md)\n' > CHANGELOG.md ;;
    staged) git add changed.md ;;
    tmpl) mv changed.md changed.md.tmpl ;;
    escape) printf 'ok\n' > ../README.md ;;
    committed) git add changed.md; git commit -qm change ;;
  esac
  result=0
  "$script" --base "$base" "$mode" > "$dir/output" 2>&1 || result=$?
  if [ "$expected" = pass ] && [ "$result" -ne 0 ]; then cat "$dir/output"; echo "FAIL $name: expected success, got $result"; exit 1; fi
  if [ "$expected" = fail ] && [ "$result" -ne 1 ]; then cat "$dir/output"; echo "FAIL $name: expected exit 1, got $result"; exit 1; fi
  if [ -n "$pattern" ] && ! grep -Fq "$pattern" "$dir/output"; then cat "$dir/output"; echo "FAIL $name: missing $pattern"; exit 1; fi
  echo "ok   $name"
}
case_run 'plain document' pass '# Hello'
case_run 'broken relative link' fail '[bad](missing.md)' '' --strict 'missing.md'
case_run 'existing link' pass '[good](existing.md)'
case_run 'root-relative existing link' pass '[good](/existing.md)'
case_run 'root-relative broken link' fail '[bad](/missing.md)' '' --strict '/missing.md'
case_run 'percent-encoded destination' pass '[file](a%20file.md)' file
case_run 'same-file anchor' pass '[section](#section)'
case_run 'file anchor' pass '[section](existing.md#section)'
case_run 'external link' pass '[web](https://example.com)'
case_run 'defined reference' pass $'[link][ref]\n\n[ref]: existing.md'
case_run 'undefined reference' fail '[link][unknown]' '' --strict 'unknown'
case_run 'collapsed reference' pass $'[label][]\n[label]: existing.md'
case_run 'broken reference target' fail $'[link][ref]\n[ref]: missing.md' '' --strict 'missing.md'
case_run 'fenced code ignored' pass $'```\n[bad](missing.md)\n```'
# The backticks are literal Markdown.
# shellcheck disable=SC2016
case_run 'inline code ignored' pass '`[bad](missing.md)`'
case_run 'four-backtick fence' pass $'````\n```\n[bad](missing.md)\n````'
case_run 'tilde fence' pass $'~~~\n[bad](missing.md)\n~~~'
case_run 'indented marker does not open fence' fail $'    ```\n[bad](missing.md)' '' --strict 'missing.md'
case_run 'angle destination with space' pass '[file](<a file.md>)' file
case_run 'parenthesized destination' pass '[file](notes_(v2).md)' paren
case_run 'missing file with fragment' fail '[bad](missing.md#section)' '' --strict 'missing.md'
case_run 'unclosed front matter' fail $'---\ntitle: Example\n# No closing marker' '' --strict 'unclosed front'
case_run 'empty front matter value' pass $'---\ntitle: # pending\n---\n# Content'
case_run 'closed front matter' pass $'---\ntitle: Example # comment\n---\n# Content'
case_run 'skip generated docs and changelog' pass '# Good' skip
case_run 'unchanged broken document ignored' pass '# Good'
case_run 'staged document checked' fail '[bad](missing.md)' staged --strict 'missing.md'
case_run 'committed document checked' fail '[bad](missing.md)' committed --strict 'missing.md'
case_run 'warning mode succeeds' pass '[bad](missing.md)' '' --warn-only 'WARNING'
case_run 'checkboxes and bare text' pass $'- [ ] task\n- [x] done\n[ordinary]'
case_run 'reference label case and whitespace' pass $'[link][  MIXED   case ]\n[mixed case]: existing.md'
case_run 'empty destination' pass '[empty]()'
case_run 'closing fence length' fail $'````\n```\n````\n[bad](missing.md)' '' --strict 'missing.md'
case_run 'link escaping the repository is broken' fail '[x](../README.md)' escape --strict 'outside'
case_run 'backtick fence info string with a backtick is not a fence' fail $'```lang`bad\n[bad](missing.md)\n```' '' --strict 'missing.md'
case_run 'template broken link' fail '[bad](missing.md)' tmpl --strict 'missing.md'
case_run 'template undefined reference' fail '[x][undefined]' tmpl --strict 'undefined'
case_run 'template placeholder destination ignored' pass '[x]({{ .Path }})' tmpl
echo "PASS: $count cases"
