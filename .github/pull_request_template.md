Summary:

## Motivation

Why this change is needed. Link the issue (`Fixes #123`) and, if one exists, `intent/<slug>/`.

## Changes

What changed. If this PR is stacked on another, name it and the merge order.

## Testing

Paste the tail of `make ci`. For each behavior this PR adds or fixes, name the test that fails when the
change is reverted. List the acceptance tests (`make testacc`) or `.sdkharness` scenarios you ran, and any
you could not.

## Risk

What could break for existing configurations and state: a new plan diff, a forced replacement
(`ForceNew`), a state upgrade, or a change in what is sent to B2. Name the riskiest part of the diff.

## How to revert

Whether reverting this commit is enough, or whether users who applied with this version would be left with
state or B2 resources the previous release cannot read.

## Checklist

- [ ] `make ci` passes locally.
- [ ] A fresh-context review (`/code-review`, `/security-review` or a subagent) ran on the diff; findings are fixed or answered below.
- [ ] A `.changes/unreleased` entry exists (`make changelog-new`), or the PR is labelled `-changelog`.
- [ ] Schema changes are reflected in `docs/` (`make docs`) and `examples/`.
- [ ] New or changed behavior is pinned by a test that fails when the change is reverted.
- [ ] `AGENTS.md` is updated if this change makes it stale.
- [ ] The description matches the final code.
