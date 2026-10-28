# Intent: review and merge process scaled by risk
Author: Gonzalo Peña-Castellanos (with Claude). Status: draft, for the repository maintainers.
Source: the AI-native SDLC playbook (PR review loop) and the matching proposal in Backblaze/blazer.

## Problem

Most provider PRs are now written with coding agents, and review has not kept up with them:

- From #138 to #174 no PR has a recorded review or inline comment; authors merged their own work. Only six
  PRs in #120–#174 were reviewed at all.
- `master` has no branch protection (the API returns 404) and no `CODEOWNERS`, so nothing requires a
  second person before merge, and pushes go straight to `master` (the old release steps said so).
- Lessons that should have been review findings came back as follow-up commits instead: missing changelog
  entries (#164), perpetual diffs found after release (#146, #151, #153, #154), and a git-branch b2sdk pin
  that had to be closed (#163).
- Every PR gets the same attention whatever its risk; a typo fix and a change to how credentials reach the
  Python bindings are treated alike.

The adoption PR covers the parts that need no settings change: `AGENTS.md`, `REVIEW.md` severities and
convergence rules, Risk and How-to-revert in the PR template, a fresh-context self-review before a PR
opens, the changie `changelog` check and CodeQL.

## Proposed outcome

Human review time goes where the risk is, with no rise in reverts or post-release fixes:

1. **A second person on every non-trivial merge.** Add `.github/CODEOWNERS` and branch protection on
   `master` requiring one approval from someone other than the author, and the `lint`, `test`, `build`
   and `changelog` checks.
2. **Risk tiers set by the change, not the author.** A first proposal to accept or change:

   | Tier | Paths | Human approvals |
   | --- | --- | --- |
   | Low | `*.md` outside `docs/`, `*_test.go` only, `.changes/`, `intent/`, `examples/` | None once CI and automated review are green, for authors on the trusted list |
   | Medium | `b2/` schemas and CRUD, `python-bindings/b2_terraform/`, `.sdkharness/` | One maintainer |
   | High | `b2/client.go`, `b2/bindings.go` (credentials and the bindings boundary), `python-bindings/requirements*.txt`, `.github/workflows/`, `.goreleaser.yml`, `.changie.yaml`, release PRs | One maintainer always; two for workflow and release changes |

   Authors outside the trusted list always need a person. A bot's approval never fills a human slot.
3. **Approvals cover one commit.** New commits pushed after an approval need re-approval for Medium and
   High tiers.
4. **Every review thread resolved before merge** (GitHub "require conversation resolution").
5. **Automated AI review and security review on every PR**, with open Important findings blocking merge
   once tuned. Start advisory, tune `REVIEW.md` for a few weeks, then decide on blocking.
6. **Past bugs become reviewer evals.** Build a set from bugs that shipped or were caught late, confirmed
   by a maintainer, for example #146 (`revision` drift), #151 (`bucket_info` key case), #153 (`file_info`
   keys forcing re-upload), #154 (lifecycle rule order), #157 (re-sent SSE default). Score the reviewer
   against the set whenever `REVIEW.md`, `AGENTS.md` or the model changes.
7. **Measure it.** Human approvals per PR, review rounds to pass, time a PR waits on a person; reverts and
   post-release fix PRs as the guardrail.

## Affected users and systems

- Contributors and agents opening PRs; the maintainers who review them.
- Repo settings: branch protection or rulesets on `master`, conversation resolution, CODEOWNERS.
- Workflows: a new review workflow, possibly a risk-tier check.
- Provider users only indirectly: fewer perpetual diffs and replacement surprises in releases.

## Constraints

- Settings changes need a repo admin; this intent proposes them, it does not apply them.
- Separation of duties: the agent that wrote a change never approves it, and the person directing it
  counts as one reviewer.
- Workflow changes keep the Dependabot guard (`github.actor` and `pull_request.user.login`).
- Acceptance tests need repository secrets and real B2 resources; required checks must not depend on
  them for fork PRs.
- Start strict and loosen with evidence; any tier can be tightened if reverts rise.

## Open questions

- Who are the code owners, and who is on the trusted-author list for the Low tier?
- Do the tier paths above match where the risk is? Which paths always need two people?
- Which checks become required on `master`, including `changelog` and CodeQL?
- Which AI reviewer runs in CI (managed Code Review, `claude-code-action`, Copilot), and who pays for it?
- Where should the eval set live and run: this repo, or the central SDK harness?
