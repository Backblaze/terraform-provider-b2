# terraform-provider-b2

Terraform provider for Backblaze B2 (`registry.terraform.io/Backblaze/b2`). The Go provider (Terraform
Plugin SDK v2) defines schemas and diffs; every B2 call is made by embedded Python bindings built on
[b2-sdk-python](https://github.com/Backblaze/b2-sdk-python). This file is the map for coding agents and new
contributors; `CLAUDE.md` points here. Keep it under about 150 lines and link out for detail.

## Related projects

- **[b2-mcp](https://github.com/backblaze-labs/b2-mcp)** — a Model Context Protocol server for B2.

## Commands

- Setup: `make deps` installs the Python bindings dependencies into the active Python environment (use a
  virtualenv) and the Go and dev tools. Go and Terraform must already be installed. It also runs `go mod tidy`,
  which can change `go.mod` and `go.sum`.
- Local gate before every push: `make lint`, `make test` and `make deps-check`, plus `make docs-lint` when a
  schema or documentation input changed. Paste the tail of their output in the PR. CI also builds a
  packaged-bindings fixture and runs `python -m pytest python-bindings/tests`; run that pytest suite
  yourself when you change the bindings.
- `make lint` runs license-header checks, `golangci-lint`, `go vet` and Python `ruff`.
- Unit tests, no credentials: `make test`. Acceptance tests against real B2 (create resources, cost money):
  `B2_TEST_APPLICATION_KEY_ID=... B2_TEST_APPLICATION_KEY=... make testacc TESTARGS='-run TestAccX'`.
- Go version: `go.mod` pins `toolchain go1.26.11`, the version CI uses.
- Format: `make format` (`golangci-lint fmt`, `terraform fmt` on `examples/`, `ruff format`).
- Docs: `make docs` regenerates `docs/` from schemas, `templates/` and `examples/`. Never hand-edit `docs/`.
- Dependency check: `make deps-check`. Vulnerabilities: `make vulncheck`.
- `make docs-prose-lint`: checks the Markdown you changed against the prose rules, in warning-only mode.
- `make examples-check`: validates `examples/` against a locally built provider, with no credentials.
- `make schema-compat`: compares the provider schema with the base ref (default origin/master) and reports breaking changes.
- Guides: [writing style](contributing/writing-style.md), [testing](contributing/testing.md),
  [schema compatibility](contributing/schema-compat.md) and
  [internal reference guard](contributing/internal-reference-guard.md).
- Changelog: edit the `[Unreleased]` section of `CHANGELOG.md` by hand (see Conventions).
- Releases: see [README.release.md](README.release.md).

## Architecture

| Path | Role |
| --- | --- |
| `main.go` | Plugin entry point; the version comes from the release tag via goreleaser `ldflags`. |
| `b2/provider.go`, `b2/resource_*.go`, `b2/data_source_*.go` | Schemas, CRUD functions, `CustomizeDiff` and diff suppression. |
| `b2/templates.go`, `b2/validators.go`, `b2/utils.go` | Shared schema blocks (resource and data source variants), validators, helpers. |
| `b2/json_types.go`, `b2/convert.go` | Typed input and output passed to the Python bindings as JSON. |
| `b2/client.go`, `b2/bindings.go` | Runs the embedded bindings binary (`go:embed b2/py-terraform-provider-b2`) with JSON on stdin. |
| `python-bindings/b2_terraform/provider_tool.py` | One handler per resource and data source operation; calls b2sdk. Built with PyInstaller. |
| `python-bindings/requirements*.txt` | Pinned b2sdk, certifi and build tools, bundled into every release binary. |
| `templates/`, `examples/`, `docs/` | tfplugindocs input and its generated output (published on the Registry). |
| `tools/` | Separate module pinning tfplugindocs and other dev tools. |
| `.sdkharness/` | Contract checks run by the central SDK harness against the B2 simulator on loopback. See its README. |
| `scripts/` | Repository checks and helpers, such as `scripts/check-headers.py`. |
| `contributing/` | Contributor guides on testing, writing style, schema compatibility and the internal reference guard; kept out of `docs/`. |

A change to what is sent to B2 usually touches both sides: the Go input struct and schema, and the Python
handler that maps it to b2sdk. `make test` touches neither the bindings nor B2; only `make testacc` does.

## Conventions

- Every `.go` and `.py` file starts with the Backblaze license header naming its own path
  (`scripts/check-headers.py`, run by `make lint`).
- Schema `Description` text is the user documentation. Take wording from the b2sdk docstrings or B2 docs.
- Keep the Python bindings thin: map input to b2sdk calls and shape output. Validation, defaults and diff
  logic live in Go.
- Pin b2sdk and other Python dependencies to released versions, never a git branch.
- A breaking change (a removed or renamed argument, a new `ForceNew`, a changed default) needs a `Changed`
  or `Removed` entry that says what users must do before upgrading.
- Changelog: add a bullet under `[Unreleased]` in `CHANGELOG.md`, in the matching `###` section (`Added`,
  `Changed`, `Deprecated`, `Removed`, `Fixed`, `Security` or `Infrastructure` for toolchain, dependency, CI
  and test-harness changes). The changelog check in `.github/workflows/ci.yml`
  fails a PR that changes other files without a `CHANGELOG.md` change. Markdown-only, test-only and docs-only PRs take the `-changelog` label instead.
- Comments are short and explain non-obvious decisions.

## Things agents get wrong here

1. **Missing changelog entry.** Every change outside Markdown and tests needs a `CHANGELOG.md` entry or
   the `-changelog` label, including CI and `.sdkharness` work (`Infrastructure`).
2. **Hand-edited `docs/`.** Change the schema `Description`, `templates/` or `examples/`, then `make docs`.
3. **Perpetual diffs.** After apply, the same configuration must plan empty. Fix the cause (B2 reorders
   lists, lowercases keys, adds computed keys) with a diff suppress or `CustomizeDiff`, and pin it with an
   acceptance step that re-applies the same config with `PlanOnly: true`.
4. **Accidental replacement.** A key or value B2 adds to a `ForceNew` attribute must never force a
   re-create (for `b2_bucket_file_version`, a re-upload). Check every `ForceNew` the change reaches.
5. **Sending stored state the user did not set.** When an optional argument is omitted, send nothing and
   let B2 keep or default it; do not re-send the last value read into state.
6. **Silent conversion.** Warn at plan time when B2 will change a value (key case, reserved file info
   keys); leave validation that B2 already does to B2.
7. **Tests that do not pin the fix.** Revert the change and the test must fail. Say which test and paste
   the failure. Pure helpers get unit tests that `make test` runs without credentials.
8. **Claimed coverage that did not run.** CI runs acceptance tests only with repository secrets, and never
   runs the `.sdkharness` simulator scenarios (only the `TestSDKHarnessContract` unit test runs in `make test`). List what you ran (`make testacc` filter, harness scenarios) and what you did not.
   A skipped check is NOT RUN, never PASS.
9. **Secrets.** New secret attributes are `Sensitive`; never put a key, SSE-C secret or signing secret in
   a log line, error, warning or example. Check `tflog` calls on both the input and output paths.
10. **Platform assumptions.** Tests and scripts run on Linux, macOS and Windows runners: no POSIX-only
    checks without a skip, LF line endings for harness files (`.gitattributes`).
11. **`.sdkharness` safety.** Checks stay loopback-only and fail closed: no real endpoints, no ambient
    `B2_*` credentials, no `terraform init`, no build inside the checkout.
12. **PR text that disagrees with code.** Re-read the description after the last commit.
13. **Breaking existing configurations.** Keep deprecated arguments working and map them to the current
    ones; a removed, renamed or re-typed argument needs a `Changed` entry with upgrade steps.
14. **Resource and data source variants.** Shared blocks in `b2/templates.go` serve both; the data source
    side needs `Computed: true` (`If(ds, ...)`) or it never gets the value.
15. **Bindings style.** In `python-bindings/b2_terraform/provider_tool.py`, raise `RuntimeError` for failed operations (the file uses `ValueError` for invalid input), keep imports sorted, give helpers a
    one-line docstring, drop leftover parameters, and comment code that depends on b2sdk internals.
16. **Binary portability.** The bindings are frozen with PyInstaller and made static with staticx (Linux
    uses `deadsnakes`, not `setup-python`) so the provider runs on older glibc and Alpine. Do not change
    runners, Python setup or build tools without checking the binary still runs there.

## Workflow

- Bug fixes: commit the failing regression test first, then the fix, without editing the test.
- Before opening a PR, review your own diff in a fresh context (a subagent or session that did not write
  the code) and fix or answer what it finds. Say in the PR what it reported.
- Review: [REVIEW.md](REVIEW.md). Every finding gets fixed or resolved with a reason. Agents never approve
  or merge their own PRs, and a PR is not merged by its author without another maintainer's review.
- The person directing an agent owns its PR: they count as one of its reviewers and fix what it breaks.
- Keep PRs small and reversible; the blast radius sets the review depth ([REVIEW.md](REVIEW.md#depth-by-risk)).
  If the direction looks wrong, say so early in the thread; closing a PR is cheap.
- Security reports: [SECURITY.md](SECURITY.md).

## Verifying your work

Run `make lint` and `make test` and paste the tail of their output. If a test fails, fix the code, not the
test. For changes to resources, data sources or the bindings, run the matching acceptance tests with
`make testacc` and say which; if you have no test credentials, say the acceptance tests were NOT RUN.
