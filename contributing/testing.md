# Testing the provider

A green run proves only the behavior it exercised. Run commands from the repository root and report failures, skips, and checks not run.

## Test levels

| Level | Where | Run it with | Proves | Does not prove | Needs credentials |
| --- | --- | --- | --- | --- | --- |
| Go unit | `b2/*_test.go` | `make test` | Provider logic, schema and regression behavior exercised by the tests. It excludes `TestAcc` tests. | Live B2 behavior or the packaged bindings when that fixture is skipped. | No |
| Known defects | `b2/*_knowndefect_test.go` | `make test-known-defects` | Whether the specified expected behavior holds; uses the `knowndefect` tag and race detector. | That other behavior or live B2 works. | No |
| Python bindings | `python-bindings/tests/` | `B2_SIMULATOR_DIR=<b2-simulator checkout> python -m pytest python-bindings/tests` | Bindings behavior exercised by pytest, including tests against a local b2-simulator. | Behavior of a packaged executable or real B2. | Fixed simulator test credentials only; never live credentials |
| Harness contract | `b2/sdkharness_contract_test.go`, `.sdkharness/` | `go test ./b2 -run '^TestSDKHarnessContract$' -count=1` and `.sdkharness/tests/selftest` | Contract rows, entry points and harness safety and verdict handling. | Actual simulator scenario results or live B2. | No |
| Harness simulator | `.sdkharness/tests/`, `.sdkharness/tests.tsv` | See [harness instructions](../.sdkharness/README.md) for conformance and resilience dispatchers. | The selected scenario against the loopback simulator and, for resilience, its fault controls. | Real B2 behavior or behavior the simulator does not model. | Fixed simulator test credentials only; never live credentials |
| Acceptance | `b2/*_test.go` (`TestAcc*`) | `make testacc` | The selected end-to-end provider operations against real B2. | Other account settings or all future B2 behavior. | Yes: `B2_TEST_APPLICATION_KEY_ID` and `B2_TEST_APPLICATION_KEY` |

Go commands need Go and make. Python bindings tests need Python, pytest and the bindings requirements, plus Node.js and `B2_SIMULATOR_DIR` set to a checkout of [backblaze-labs/b2-simulator](https://github.com/backblaze-labs/b2-simulator) (v0.6.0 or later); the tests that use it fail with a clear error when it is not set. `TestAccResourceB2BucketDefaultRetentionNoneRoundTrip` also needs `B2_SIMULATOR_DIR`, because its helper (`b2/testdata/retention_simulator.py`) starts that simulator and uses its virtual clock to delete files under compliance retention. Some Go tests use Terraform and skip when it is unavailable; the packaged-bindings cancellation test needs its fixture. Harness simulator scenarios need a running loopback B2 simulator, a provider build, Terraform, Python and the tools listed in the harness README; resilience also needs the simulator control URL. Harness selftest does not need a simulator. Acceptance tests need Terraform, Python, PyInstaller, the bindings dependencies and live B2 credentials. They create real B2 resources and may cost money; use disposable resources and check cleanup. On Linux, building the packaged bindings may also need staticx. CI skips acceptance steps when its credentials are missing, so green CI alone does not establish that they ran.

## Known-defect tests

Files with the `knowndefect` build tag contain tests of expected behavior for defects that may still exist. `make test` excludes them because they are intended to fail until fixed. `make test-known-defects` reports each test separately and can fail while some tagged tests pass. When fixing a defect, move its test to an untagged file in the regular suite and remove the `TestKnownDefect_` prefix; verify the test fails before the fix and passes afterward. Repository-policy checks are text heuristics and may need adjustment when the implementation changes.

## Choosing checks for a PR

- For a schema or diff change, use Go unit tests and a Terraform plan or state test when behavior depends on Terraform's schema processing. Check that a regression test fails for the intended reason before the fix.
- For Python bindings changes, run the bindings tests and relevant provider tests; use harness scenarios when their simulator models the affected interaction.
- For changes that depend on how B2 answers, use an appropriate live acceptance test as well as local checks. Simulator and fake responses do not establish real service behavior. If credentials or a required capability are unavailable, report the live check as not run and explain the gap.
