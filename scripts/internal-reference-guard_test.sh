#!/usr/bin/env bash
# Tests for internal-reference-guard.sh. Uses only fake sample keys and hosts.
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
guard="$root/scripts/internal-reference-guard.sh"
workflow=${GUARD_TEST_WORKFLOW:-$root/.github/workflows/internal-reference-guard.yml}
TRACKER_KEYS=$(sed -n "s/^  TRACKER_KEYS: '\(.*\)'$/\1/p" "$workflow")
INTERNAL_HOSTS=$(sed -n "s/^  INTERNAL_HOSTS: '\(.*\)'$/\1/p" "$workflow")
if [[ -z $TRACKER_KEYS || -z $INTERNAL_HOSTS ]]; then
  echo 'FAIL: workflow TRACKER_KEYS and INTERNAL_HOSTS must both be present' >&2
  exit 1
fi
export TRACKER_KEYS INTERNAL_HOSTS
# Split literals so this file does not match the guard when it scans itself.
key='EXAM''PLE-12'
host='internal.''example.''test'
check() {
  local expected=$1; shift
  local status=0
  "$guard" "$@" > "$tmp/out" 2>&1 || status=$?
  if [[ $expected == fail && $status != 1 ]] || [[ $expected == pass && $status != 0 ]]; then
    echo "FAIL: expected $expected for $* (exit $status)"; cat "$tmp/out"; exit 1
  fi
}
# Each mode fails on a key and on a host, and passes on clean text.
for source in title body commit file; do
  for sample in "$key" "see https://$host/wiki" "fix_${key%12}4521" "${host}_x" "SAM""PLE-7" "INTERNAL.Exam""ple.TEST"; do
    case "$source" in
      commit) check fail commit abc "$sample" ;;
      file) printf '%s\n' "$sample" > "$tmp/case"; check fail file "$tmp/case" ;;
      *) check fail "$source" "$sample" ;;
    esac
  done
  # False positives: public numbers and URLs, look-alike words, lower case.
  for sample in 'clean text' 'Fixes #231' 'PR #274 and issue #175' \
    'https://github.com/Backblaze/terraform-provider-b2/issues/231' \
    'https://registry.terraform.io/providers/Backblaze/b2/latest' \
    'EXAMPLES-12' 'NOEXAMPLE-12' 'EXAMPLEX-12' 'EXAMPLE-abc' 'example-12' 'sample-7' \
    'notinternal.example.testing' 'internalXexampleXtest'; do
    case "$source" in
      commit) check pass commit abc "$sample" ;;
      file) printf '%s\n' "$sample" > "$tmp/case"; check pass file "$tmp/case" ;;
      *) check pass "$source" "$sample" ;;
    esac
  done
done
# A host configured with upper-case letters still matches text in any case.
for configured in 'Internal\.Exam''ple\.Test' 'INTERNAL\.EXAM''PLE\.TEST'; do
  for text in "see https://$host/x" "INTERNAL.EXAM""PLE.TEST"; do
    INTERNAL_HOSTS=$configured check fail body "$text"
  done
done
# Key at the start and end of a line, and on a later line.
printf '%s\n' "$key start" > "$tmp/line"; check fail file "$tmp/line"
printf '%s\n' "end $key" > "$tmp/line"; check fail file "$tmp/line"
printf '%s\n' 'clean' "later $key" > "$tmp/line"; check fail file "$tmp/line"
# A missing file is an error (exit 2), not a pass.
status=0
"$guard" file "$tmp/does-not-exist" > "$tmp/out" 2>&1 || status=$?
[[ $status == 2 ]] || { echo "FAIL: missing file exit $status, want 2"; exit 1; }
# Missing configuration is an error (exit 2).
status=0
env -u TRACKER_KEYS "$guard" title x > "$tmp/out" 2>&1 || status=$?
[[ $status == 2 ]] || { echo "FAIL: unset TRACKER_KEYS exit $status, want 2"; exit 1; }
# A binary-looking file with a key still fails and one without passes.
printf 'abc\0def %s\0\n' "$key" > "$tmp/bin"; check fail file "$tmp/bin"
printf 'abc\0def\0\n' > "$tmp/bin"; check pass file "$tmp/bin"
# The repository's own guard files must not trip the guard.
cd "$root"
check pass file scripts/internal-reference-guard.sh
check pass file scripts/internal-reference-guard_test.sh
check pass file .github/workflows/internal-reference-guard.yml
check pass file contributing/internal-reference-guard.md
echo 'internal-reference-guard tests PASS'
