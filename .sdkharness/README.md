# SDK harness contracts

`tests.tsv` is the versioned entry-point contract used by the centralized SDK
quality harness. The provider repository owns the executable health,
conformance, and resilience assertions. The harness supplies a fresh local B2
simulator, fault controls, the provider built from this exact revision,
orchestration, evidence retention, and fleet reporting.

## What these checks can and cannot reach

- **Only the loopback simulator.** Every check refuses to run unless its simulator URL is a
  bare IPv4 loopback HTTP origin (`http://127.0.0.1:<port>`). There is no staging or
  production mode.
- **Only the simulator's fixed test credential** (`test-key-id` / `test-key`). Ambient
  `B2_*` variables are dropped before any check runs, so a real key cannot reach the provider.
  The health check keeps only the fixed pair and `B2_BUCKET_NAME` and unsets every other `B2_*`.
- **The exact-revision provider.** Terraform selects the supplied provider executable
  through `dev_overrides` (`SDKHARNESS_TERRAFORM_PROVIDER_DIR`), so a passing result
  cannot silently come from a previously published registry build. The CLI configuration
  allows no other installation method, and checks do **not** run `terraform init`:
  with `dev_overrides` nothing needs installing, and `init` would contact the provider
  registry.
- **No other network while asserting.** Proxy-aware clients are pointed at a dead proxy with
  loopback exempt, so a connection to anything but the simulator fails at once and is reported
  as a `FAIL` ("a non-loopback connection was attempted"), not as an amber "unreachable".
- **Building is a separate, earlier step and never touches the checkout.** The conformance and
  resilience dispatchers require `SDKHARNESS_TERRAFORM_PROVIDER_DIR` and never build. The health
  check uses it too; only when it is unset does the health check build the provider from a
  scratch **copy** of the checkout (`make build STATICX=0`), before any assertion. That build
  may download Go modules and Python packages, so it is not covered by the guard above. For a
  fully offline run, build the provider yourself first (below) and pass the directory.
- **What is `COULD-NOT-RUN` (`SKIP`), at every level.** Only a missing Terraform CLI or a missing
  provider build (no executable in `SDKHARNESS_TERRAFORM_PROVIDER_DIR`; for the health check also
  no `go`/`make` when it has to build) is amber, reason `missing-runtime`. Because a check only
  runs against the simulator URL it was given, a rejected credential, an unreachable or reset
  simulator, a timeout, a provider crash and a failed build are all a `FAIL`. Conformance,
  resilience and health apply the same policy, including a missing `terraform` in health.
  A missing check-owned tool (`python3`, `jq`, `curl`, a sha1 tool) is a `FAIL` (setup).
- A check that reports `COULD-NOT-RUN` must exit 0; the dispatcher reports a nonzero exit
  after such a verdict as a `FAIL`, never as amber evidence.

`tests/selftest` pins these properties and needs no simulator, network, Terraform or provider.

## Prerequisites

- Go as declared in `go.mod` (1.25.x), `make`, Terraform (any release that supports `dev_overrides`; the provider's own CI matrix
  (`.github/workflows/ci.yml`) tests 1.13.* and 1.14.*, and the central SDK harness workflows run
  these checks with 1.12.2), `python3`, `jq`.
- To build the provider: Python 3.10+ with `python-bindings/requirements.txt` and
  `requirements-dev.txt` installed (the provider embeds a PyInstaller-built b2sdk binding).
  On Linux under CI interpreters such as `actions/setup-python`, pass `STATICX=0`: the
  Makefile's `staticx` wrapper rejects those interpreters and only matters for shipping the
  binding, not for simulator checks.

Build the provider into its own directory first (this is the step that may use the network).
Do it in a scratch copy so the checkout stays untouched:

```bash
tmp="$(mktemp -d)" && cp -R . "$tmp/provider"
(cd "$tmp/provider" && make build STATICX=0)     # writes $tmp/provider/terraform-provider-b2
export SDKHARNESS_TERRAFORM_PROVIDER_DIR="$tmp/provider"
```

## Run one check by hand

Start a simulator (the standalone one from `backblaze-labs/b2-simulator`; the harness's
embedded `sdkharness/bin/simulator/serve.mjs` behaves the same). Resilience needs `--control`:

```bash
node bin/simulator/serve.mjs --control
# SIMULATOR-LISTENING http://127.0.0.1:<port>
# SIMULATOR-LISTENING https://127.0.0.1:<https-port>
# SIMULATOR-CONTROL   http://127.0.0.1:<control-port>
```

Export what the dispatchers read. (`b2-simulator`'s shell helper `bin/lib/simulator.sh` sets
the same values as `B2SIM_URL`, `B2SIM_CONTROL_URL` and so on; the harness maps them onto
these names.)

```bash
export SDKHARNESS_SIMULATOR_URL=http://127.0.0.1:<port>
export SDKHARNESS_SIMULATOR_CONTROL_URL=http://127.0.0.1:<control-port>   # resilience only
export SDKHARNESS_TERRAFORM_PROVIDER_DIR="$tmp/provider"   # from "Prerequisites" above
```

Then, from the repository root:

```bash
# conformance
SDKHARNESS_TEST_LEVEL=conformance SDKHARNESS_SCENARIO=bucket.crud \
  .sdkharness/tests/run-conformance

# resilience
SDKHARNESS_TEST_LEVEL=resilience SDKHARNESS_SCENARIO=upload.retry_503 \
  .sdkharness/tests/run-resilience

# customer health: the bucket must already exist in the simulator, with the fixed credential
SDKHARNESS_TEST_LEVEL=health SDKHARNESS_SCENARIO=golden-path \
  B2_APPLICATION_KEY_ID=test-key-id B2_APPLICATION_KEY=test-key B2_BUCKET_NAME=sdkharness-healthcheck \
  .sdkharness/tests/health-golden-path
```

Each prints one `SDKHARNESS_RESULT<TAB>level<TAB>scenario<TAB>PASS|FAIL|SKIP<TAB>detail`
line (plus diagnostics) and exits 0 for `PASS` and `SKIP`, 1 for `FAIL`. Create the health
bucket against a fresh simulator with an authorize call and `b2_create_bucket`
(`curl -u test-key-id:test-key "$SDKHARNESS_SIMULATOR_URL/b2api/v4/b2_authorize_account"`, then
POST `{"accountId": ..., "bucketName": "sdkharness-healthcheck", "bucketType": "allPrivate"}`
to `/b2api/v4/b2_create_bucket` with the returned `authorizationToken`).

The leaf files under `tests/conformance/` and `tests/resilience/` are not entry points: run
directly they refuse anything but the loopback simulator (the resilience leaves find their helper
module themselves, so a direct run prints a clean refusal rather than a Python import error), but the dispatchers above are the
supported way to run them.
