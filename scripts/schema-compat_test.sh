#!/usr/bin/env bash
# Test the schema comparer against small fixture schemas. It needs no provider build.
set -euo pipefail
cd "$(dirname "$0")/.."
dir=scripts/testdata/schema-compat
fail=0
# check NAME EXIT_CODE PATTERN: compare NAME.base.json with NAME.head.json.
check() {
  local name=$1 want=$2 pattern=$3 out rc=0
  out=$(python3 -I scripts/schema-compat-compare.py "$dir/$name.base.json" "$dir/$name.head.json" 2>&1) || rc=$?
  if [ "$rc" -ne "$want" ] || ! grep -Eq -- "$pattern" <<<"$out"; then
    echo "FAIL $name: exit $rc (want $want), pattern '$pattern'" >&2
    echo "$out" >&2
    fail=1
  else
    echo "PASS $name"
  fi
}
check removed-resource-breaking 1 'BREAKING.*b2_b.*removed'
check removed-resource-clean 0 'No differences'
check removed-datasource-breaking 1 'BREAKING.*data source b2_e.*removed'
check removed-datasource-clean 0 'NON-BREAKING.*data source b2_e.*added'
check removed-attribute-breaking 1 'BREAKING.*b2_a.*note.*removed'
check removed-attribute-clean 0 'NON-BREAKING.*b2_a.*note.*added'
check changed-type-breaking 1 'BREAKING.*b2_a.*n.*type changed'
check changed-type-clean 0 'No differences'
check optional-to-required-breaking 1 'BREAKING.*b2_a.*n.*optional.*required'
check optional-to-required-clean 0 'No differences|NON-BREAKING'
check new-required-breaking 1 'BREAKING.*b2_a.*n.*new required'
check new-required-clean 0 'NON-BREAKING'
check sensitive-dropped-breaking 1 'BREAKING.*b2_a.*k.*sensitive'
check sensitive-dropped-clean 0 'No differences|NON-BREAKING'
check deprecation-clean 0 'DEPRECATION.*b2_a.*n'
check block-attribute-breaking 1 'BREAKING.*b2_a.*rule\.x.*removed'
check block-attribute-clean 0 'NON-BREAKING.*rule\.y.*added'
check nesting-mode-breaking 1 '^BREAKING.*b2_a.*rule.*nesting mode changed'
check nesting-mode-clean 0 'No differences'
check min-items-breaking 1 '^BREAKING.*b2_a.*rule.*minimum item count increased'
check min-items-clean 0 'No differences'
check max-items-breaking 1 '^BREAKING.*b2_a.*rule.*maximum item count decreased'
check max-items-added-breaking 1 '^BREAKING.*b2_a.*rule.*maximum item count decreased'
check max-items-clean 0 'No differences'
check new-required-block-breaking 1 '^BREAKING.*b2_a.*rule.*new required block'
check new-optional-block-clean 0 '^NON-BREAKING.*b2_a.*rule.*block added'
check removed-block-breaking 1 '^BREAKING.*b2_a.*rule.*block removed'
# A crash is reported as an error (exit 2), never as a breaking change (exit 1).
check invalid-json 2 'ERROR'
exit "$fail"
