# Intent: review and merge process scaled by risk
Author: Gonzalo Peña-Castellanos (with Claude). Status: draft, for the repository maintainers.
Source: [How the Claude Code team does code review](https://claude.dev/blog/solving-code-review-at-anthropic/)
(claude.dev), the AI-native SDLC playbook (PR review loop) and the matching proposal in Backblaze/blazer.

## Problem

Most provider PRs are now written with coding agents, and review has not kept up with them:

- From #138 to #174 no PR has a recorded review or inline comment; authors merged their own work. Since
  2026-01-01, 21 of 23 merged PRs had no review from anyone but the author, and the median human
  approvals per PR is 0 (`make review-metrics SINCE=2026-01-01`). Over the repo's history, 35 of 74 merged
  PRs had no review; no PR has ever been reviewed by Copilot.
- `master` has no branch protection (the API returns 404) and no `CODEOWNERS`, so nothing requires a
  second person before merge, and pushes go straight to `master` (the old release steps said so).
- Lessons that should have been review findings came back as follow-up commits instead: missing changelog
  entries (#164), perpetual diffs found after release (#146, #151, #153, #154), and a git-branch b2sdk pin
  that had to be closed (#163).
- Every PR gets the same attention whatever its risk; a typo fix and a change to how credentials reach the
  Python bindings are treated alike.

The adoption PR (#176) covers what needs no settings change:

- `AGENTS.md` (short, links out, recurring findings) and `CLAUDE.md`; an edit hook for formatting and
  license headers.
- `REVIEW.md`: what failure looks like here, review depth by risk (Quick, Standard, Deep), an Intent pass
  (does it solve the stated problem, root cause for fixes), Important / Nit / Pre-existing severities with a
  nit cap, `file:line` evidence, a count-first summary, paths to skip, convergence rules (diff-only
  re-review from round 3, no repeated findings, rethink after three rounds), answer labels for every
  finding (`Fixed`, `Wrong`, `Intentional`, `Didn't matter`), and what a human approver must read.
- `.github/workflows/ai-review.yml`: Claude code review following `REVIEW.md` plus a separate security
  review on every ready PR, posting findings only, never approvals. It is inert until the
  `ANTHROPIC_API_KEY` secret is added.
- PR template with Testing, Risk and How to revert, signed off by the person who directed the agent; a
  fresh-context self-review before a PR opens.
- `contributing/review-lessons.md` (every past review comment, by theme) and
  `contributing/review-evals.md` (seed eval set from bugs that shipped).
- `scripts/review-metrics.sh` / `make review-metrics` for the measures below.
- Release candidates (`vX.Y.Z-rc.N`, goreleaser pre-release) for High-risk releases.

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
5. **Automated AI review and security review on every PR**, with open findings blocking merge through
   required conversation resolution. Add the `ANTHROPIC_API_KEY` secret to turn on `ai-review.yml`, run
   it advisory for a few weeks while tuning `REVIEW.md`, then require it. Decide whether to use the
   managed Code Review service instead (review depth set per repo), and whether Copilot review stays.
6. **Always-human paths.** CODEOWNERS entries for `.github/workflows/`, `.goreleaser.yml`,
   `b2/client.go`, `b2/bindings.go` and `python-bindings/requirements*.txt`, with "require review from
   code owners" and "dismiss stale approvals when new commits are pushed" in branch protection.
7. **Merge queue**, so checks rerun on `master` plus the change and two PRs that pass alone cannot combine
   into a broken `master`.
8. **A tier check.** Once the tiers settle, a small workflow (or classifier) labels each PR with its tier
   from the paths it touches, so the tier is set by the change and not by the author.
9. **Past bugs become reviewer evals.** Build a set from bugs that shipped or were caught late, confirmed
   by a maintainer, for example #146 (`revision` drift), #151 (`bucket_info` key case), #153 (`file_info`
   keys forcing re-upload), #154 (lifecycle rule order), #157 (re-sent SSE default). Score the reviewer
   against the set whenever `REVIEW.md`, `AGENTS.md` or the model changes. The seed set is in
   `contributing/review-evals.md`; each case needs a maintainer to confirm it. Also run the reviewer on
   known-clean PRs to count harmful comments.
10. **Measure it.** `make review-metrics`: human approvals per PR (median, p95), reviews per PR, hours to
    first human review, PRs merged without review, reverts, and how findings were answered (noise).
    Reverts and post-release fix PRs are the guardrail; tighten a tier if they rise.
11. **Later.** A scheduled review agent over existing code to find old bugs, and a circuit breaker that
    raises approval requirements automatically when reverts or post-release fixes go up.

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
- Which AI reviewer runs in CI (`ai-review.yml` with an API key, the managed Code Review service, Copilot),
  and who owns the key and its cost?
- Where should the eval set live and run: this repo, or the central SDK harness?
