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

## Before the passes

1. Read the linked issue and, if present, `intent/<slug>/spec.md` and `plan.md`.
2. Check out the PR merged onto current `master` and run `make ci`. Say what you ran.
3. For each behavior the PR claims to fix or add, revert that line and confirm a test fails. Name any
   behavior no test pins, and whether its acceptance tests ran.

## Passes

Tag each finding with its pass.

- **Bugs**: logic errors, nil dereferences, wrong type assertions on `schema.ResourceData`, error paths
  that leave a resource half-created in B2 but missing from state.
- **Plan and state**: the same configuration plans empty after apply; `ForceNew` is reached only by
  changes the user made; omitted optional arguments send nothing; `Computed` values B2 owns are not
  diffed; import and read produce the same state as create; existing state keeps working (or a state
  upgrade exists).
- **Bindings**: Go input structs, `convert.go` and the Python handler agree on field names and shapes;
  validation lives in Go; b2sdk is called the way its docs describe, and pins are released versions.
- **Security**: new secrets are `Sensitive` and kept out of `tflog`, errors and warnings; credentials go to
  the bindings on stdin only; workflow changes keep secrets away from fork and Dependabot PRs.
- **Compatibility**: removed or renamed arguments, new `ForceNew`, changed defaults and changed `Computed`
  behavior each need a `Changed` or `Removed` entry that says what users must do.
- **Compliance**: the change matches the issue, `spec.md` and `plan.md`; the PR description matches the
  code; a `.changes/unreleased` entry exists and reads well for users; schema descriptions and `docs/`
  agree; `AGENTS.md` is updated when the change makes it stale.

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

Every finding gets an answer before merge: fix it, or resolve the thread with a one-line reason (false
positive, intentional, out of scope with an issue link). Do not resolve a thread silently.

## Format

Start with a count (`2 Important, 1 Nit, 0 Pre-existing`, or `No issues found`), then what you checked
and against what (issue number, acceptance tests or simulator, commands run). Then a table sorted by
severity: `Severity | File:Line | Pass: finding`. Do not report generated files (`docs/`, `CHANGELOG.md`)
or what CI already enforces (formatting, vet, lint, headers).
