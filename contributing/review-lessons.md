# Review lessons

What reviewers raised on merged PRs, grouped by theme, with the PRs where it came up. It comes from every
review, inline and conversation comment on the 73 non-Dependabot PRs merged up to #174 (83 comments on 21
PRs; Copilot has not reviewed any of them), plus fixes and follow-up commits where a review would have
caught the problem. A lesson that has come up on at least two PRs is also listed in
[AGENTS.md](../AGENTS.md#things-agents-get-wrong-here).

When a review teaches something new, add it here in the same PR, with the PR link.

## Compatibility for existing configurations

- A schema or docs change that drops, renames or re-types an argument is a breaking change; reviewers
  flag it ("Smells like non-backward-compatible change", #17). Keep deprecated arguments working
  (`bucket_id` kept on the v3 API, #132; v4 API with deprecated fields mapped, #119) and document the
  break under `Changed` with what users must do (#157).
- Helpers that translate deprecated fields to current ones come in pairs; give them symmetric names and
  APIs (`DeprecatedToCurrent` / `CurrentToDeprecated`) so the round trip is obvious (#119).

## Schema, plan and state

- Shared blocks in `b2/templates.go` serve both resources and data sources. Data source variants need
  `Computed: true` where the resource has `Optional`, or the data source never gets the value (#17); use
  the `If(ds, ...)` pattern already in that file.
- Put validators in `b2/validators.go`, not inline in a template (#17).
- Data sources that always produce a new value (signed URLs) re-read on every plan; keep their `id` stable
  so Terraform sees the same object (#31).
- Perpetual diffs and re-uploads found after release: `revision` drift (#146), `bucket_info` key case
  (#151), B2-added `file_info` keys on a `ForceNew` map (#153), lifecycle rule order (#154), the SSE
  default re-sent from state (#157). Each needed an acceptance step that re-plans the same configuration.

## Python bindings

- Raise `RuntimeError` for provider errors, like the rest of `provider_tool.py` (#17).
- Keep imports in alphabetical order, give every helper a one-line docstring, and move large helpers to
  their own module (#15).
- When code relies on b2sdk internals that change between minor versions, say so in a comment next to it,
  because it forces the b2sdk pin (#24).
- Pass whole objects, not leftover parameters from an earlier attempt (#125), and remove arguments that
  duplicate a default or a flag already passed (`-v`, #15).
- Pin b2sdk to a released version, never a git branch (#163, closed; replaced by #165).

## Release binaries and build environments

- The Python bindings are frozen with PyInstaller and, on Linux, made static with staticx so the provider
  runs on older glibc and on Alpine (the official Terraform images). Do not move the build to a newer
  runner or a different Python setup without checking the binary still runs there: a newer Ubuntu runner
  would have broken older systems (#15), `actions/setup-python` produces binaries staticx cannot make
  static, which is why Linux builds use `deadsnakes` (#36), and Alpine support needs CI coverage (#19).
- Bump build tools to the versions that fix a known problem rather than relying on a loose pin
  (`pyinstaller` for embedded files with staticx, #19).

## Dependencies

- `go.sum` is a checksum file, not a lock file: never hand-edit it. Upgrade with `go get`, then
  `go mod tidy`; lines for versions that are not used may stay (#66).
- Regenerate `docs/` with `make docs` after any schema change; CI fails on a mismatch (#83, #121).

## Changelog

- Entries start with a capital letter, like the SDK and CLI changelogs (#15).
- Compare links must name the right versions (#66). changie now writes them at release time.

## Tests and CI

- Fork PRs do not get the acceptance-test secrets, so `testacc` fails or is skipped there; a maintainer
  runs the tests on a branch in this repository (#12). A failure can also come from the test key lacking a
  capability rather than from the change (#91): check the error before changing code.
- Merge or rebase onto current `master` before re-running a failed CI run; the failure may already be
  fixed there (#12).
- State exactly what ran and what did not, and pin each fix with a test that fails on the previous head
  (#162).
- `.sdkharness` checks never build in, or write to, the checkout under test (#162).
