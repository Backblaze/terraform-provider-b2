# Testing the provider

A green run proves only the behavior it exercised. Run commands from the repository root and report failures, skips, and checks not run.

## Test levels

| Level | Where | Run it with | Proves | Does not prove | Needs credentials |
| --- | --- | --- | --- | --- | --- |
| Go unit | `b2/*_test.go` | `make test` | Provider logic, schema and regression behavior exercised by the tests. It excludes `TestAcc` tests. | Live B2 behavior or the packaged bindings when that fixture is skipped. | No |
| Known defects | `b2/*_knowndefect_test.go` | `make test-known-defects` | Whether the specified expected behavior holds; uses the `knowndefect` tag and race detector. | That other behavior or live B2 works. | No |
| Python bindings | `python-bindings/tests/` | `python -m pytest python-bindings/tests` | Bindings behavior exercised by pytest, including tests using the b2sdk simulator. | Behavior of a packaged executable or real B2. | No |
| Harness contract | `b2/sdkharness_contract_test.go`, `.sdkharness/` | `go test ./b2 -run '^TestSDKHarnessContract$' -count=1` and `.sdkharness/tests/selftest` | Contract rows, entry points and harness safety and verdict handling. | Actual simulator scenario results or live B2. | No |
| Harness simulator | `.sdkharness/tests/`, `.sdkharness/tests.tsv` | See [harness instructions](../.sdkharness/README.md) for conformance and resilience dispatchers. | The selected scenario against the loopback simulator and, for resilience, its fault controls. | Real B2 behavior or behavior the simulator does not model. | Fixed simulator test credentials only; never live credentials |
| Acceptance | `b2/*_test.go` (`TestAcc*`) | `make testacc` | The selected end-to-end provider operations against the B2 endpoint in `B2_ENDPOINT` (real B2 by default). | Real B2 behavior when run against the simulator; other account settings or all future B2 behavior. | Yes: `B2_TEST_APPLICATION_KEY_ID` and `B2_TEST_APPLICATION_KEY` (the simulator accepts its fixed test credentials) |

Go commands need Go and make. Python bindings tests need Python, pytest and the bindings requirements. Some Go tests use Terraform and skip when it is unavailable; the packaged-bindings cancellation test needs its fixture. Harness simulator scenarios need a running loopback B2 simulator, a provider build, Terraform, Python and the tools listed in the harness README; resilience also needs the simulator control URL. Harness selftest does not need a simulator. Acceptance tests need Terraform, Python, PyInstaller, the bindings dependencies and live B2 credentials. They create real B2 resources and may cost money; use disposable resources and check cleanup. On Linux, building the packaged bindings may also need staticx. In pull-request and release CI, acceptance tests run against the B2 Simulator commit in `SIMULATOR_REVISION` (`ci.yml`), one job per platform. The `acceptance` job fails when any of them fails, is cancelled or is skipped, and pull requests from forks fail at once because they cannot fetch the private Simulator. Acceptance tests against real B2 run in `live-b2.yml`, weekly and on manual dispatch, and fail when the credentials are missing; a failed scheduled run opens or updates an issue. A release is published only after `live-b2.yml` has passed on the tagged commit. A green pull request does not establish real B2 behavior.

## Known-defect tests

Files with the `knowndefect` build tag contain tests of expected behavior for defects that may still exist. `make test` excludes them because they are intended to fail until fixed. `make test-known-defects` reports each test separately and can fail while some tagged tests pass. When fixing a defect, move its test to an untagged file in the regular suite and remove the `TestKnownDefect_` prefix; verify the test fails before the fix and passes afterward. Repository-policy checks are text heuristics and may need adjustment when the implementation changes.

## Choosing checks for a PR

- For a schema or diff change, use Go unit tests and a Terraform plan or state test when behavior depends on Terraform's schema processing. Check that a regression test fails for the intended reason before the fix.
- For Python bindings changes, run the bindings tests and relevant provider tests; use harness scenarios when their simulator models the affected interaction.
- For changes that depend on how B2 answers, use an appropriate live acceptance test as well as local checks. Simulator and fake responses do not establish real service behavior. If credentials or a required capability are unavailable, report the live check as not run and explain the gap.
