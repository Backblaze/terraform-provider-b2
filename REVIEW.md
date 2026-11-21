# Review instructions

For human reviewers and review agents. Approval always comes from a maintainer with write access who did
not author the change; an agent's findings never approve or block on their own, and the agent that wrote
a change never reviews it.

## What failure looks like here

The provider fails by changing users' infrastructure in ways their configuration did not ask for: a plan
that never goes empty (perpetual diff), a forced replacement that re-creates a bucket or re-uploads a file,
an update that re-sends stale state to B2, or state that the previous release cannot read. It also fails by
leaking a secret into logs, errors or plan output, and by shipping a release binary with a vulnerable
embedded dependency. A test that passes either way is also a failure: it claims coverage it does not give.

## Depth by risk

Match the depth to what a missed bug would cost. The change sets the depth, not the author.

| Depth | When | What it means |
| --- | --- | --- |
| Quick | Markdown outside `docs/`, `*_test.go` only, `.changes/`, `intent/`, `examples/` | Compliance pass; check claims against the diff |
| Standard | `b2/` schemas and CRUD, `python-bindings/b2_terraform/`, `.sdkharness/` | All passes, revert-to-fail check for each claimed fix |
| Deep | `b2/client.go`, `b2/bindings.go`, `b2/convert.go`, `python-bindings/requirements*.txt`, `.github/workflows/`, `.goreleaser.yml`, release PRs | All passes plus the Security pass line by line; a person always reviews |

## Before the passes

1. Read the linked issue and, if present, `intent/<slug>/spec.md` and `plan.md`.
2. Check out the PR merged onto current `master` and run `make ci`. Say what you ran.
3. For each behavior the PR claims to fix or add, revert that line and confirm a test fails. Name any
   behavior no test pins, and whether its acceptance tests ran.

## Passes

Tag each finding with its pass.

- **Intent**: the change solves the problem it states and fits the provider; for a bug fix, it fixes
  the root cause, not one symptom.
- **Bugs**: logic errors, nil dereferences, wrong type assertions on `schema.ResourceData`, error paths
  that leave a resource half-created in B2 but missing from state.
- **Plan and state**: the same configuration plans empty after apply; `ForceNew` is reached only by
  changes the user made; omitted optional arguments send nothing; `Computed` values B2 owns are not
  diffed; import and read produce the same state as create; existing state keeps working (or a state
  upgrade exists); data source variants of shared blocks are `Computed`.
- **Bindings**: Go input structs, `convert.go` and the Python handler agree on field names and shapes;
  validation lives in Go; b2sdk is called the way its docs describe, and pins are released versions.
- **Build**: changes to runners, Python setup, PyInstaller, staticx or goreleaser keep the release binary
  static and runnable on older glibc and Alpine.
- **Security**: new secrets are `Sensitive` and kept out of `tflog`, errors and warnings; credentials go to
  the bindings on stdin only; workflow changes keep secrets away from fork and Dependabot PRs.
- **Compatibility**: removed or renamed arguments, new `ForceNew`, changed defaults and changed `Computed`
  behavior each need a `Changed` or `Removed` entry that says what users must do.
- **Compliance**: the change matches the issue, `spec.md` and `plan.md`; the PR description matches the
  code; a `.changes/unreleased` entry exists and reads well for users; schema descriptions and `docs/`
  agree; `AGENTS.md` is updated when the change makes it stale.

Past findings and the PRs they came from are in
[contributing/review-lessons.md](contributing/review-lessons.md); check the diff against them.

## Severity

- **Important**: an unrequested plan diff or replacement, a wrong request to B2, lost or unreadable
  state, a leaked secret, or a claimed behavior with no test pinning it.
- **Nit**: style, naming, comments, description wording. At most five per review; summarize the rest as
  a count.
- **Pre-existing**: a real bug this PR did not introduce. Report it so it is not lost, but it does not
  block this PR; suggest an issue (or a private advisory for a security bug).

## Evidence

Every finding cites `file:line` and says what goes wrong, with the configuration or sequence that triggers
it. Do not infer behavior from a function's name; read it. If you could not verify something (no
acceptance tests, no simulator), say so.

## Converging

- Round 1 reviews the whole diff. From round 3 on, review only the commits since your last review.
- Name the class of bug, not one instance, so the author fixes every occurrence in one round.
- Never repeat a finding already posted on the PR, resolved or not.
- After three rounds with Important findings still open, stop and ask the author to reconsider the
  approach instead of posting more line comments.

## Answering findings

Every finding gets an answer before merge: fix it, or resolve the thread with a one-line reason. Start the
reply with one of these labels so the reviewer's noise can be measured (`scripts/review-metrics.sh`):

- `Fixed:` with the commit that fixes it.
- `Wrong:` the finding is a false positive; say why.
- `Intentional:` the behavior is deliberate; say why.
- `Didn't matter:` real but not worth changing; say why, or link an issue.

Do not resolve a thread silently. When the same finding comes back on more than one PR, fix it at the
source in the same PR: a rule in `AGENTS.md`, a check in `make ci`, a hook or a lint rule, and a line in
[contributing/review-lessons.md](contributing/review-lessons.md).

## Skip

Do not review or report on generated or vendored files: `docs/` (check only that it matches the schema),
`CHANGELOG.md` and released `.changes/v*.md`, `go.sum`, `tools/go.sum`, and the playbook skills in
`.claude/skills/`. Do not report what CI already enforces (formatting, vet, lint, license headers, docs
drift, changelog presence) or style-only cleanups.

## For the human approver

An AI finding never approves and never fills a human slot. Before approving, read the evidence: the
`make ci` output, the tests named as pinning the change, the acceptance tests run, and the reviewer's
findings and their answers. If you cannot say what those checks would miss, review the code properly
instead of approving. New commits after your approval need another look.

## Format

Start with a count (`2 Important, 1 Nit, 0 Pre-existing`, or `No issues found`), then what you checked
and against what (issue number, acceptance tests or simulator, commands run). Then a table sorted by
severity: `Severity | File:Line | Pass: finding`. When nothing needs fixing, say `No issues found` and stop.
