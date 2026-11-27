## Summary

## Motivation

Why this change is needed. Link the issue: `Closes #123`.

## Changes

What changed. If this PR is stacked on another, name it and the merge order.

## Testing

Paste the tail of `make lint` and `make test`. For each behavior this PR adds or fixes, name the test that
fails when the change is reverted. List the acceptance tests (`make testacc`) you ran, or write NOT RUN.

## Risk

What could break for existing configurations and state: a new plan diff, a forced replacement
(`ForceNew`), a state upgrade, or a change in what is sent to B2. Name the riskiest part of the diff.

## How to revert

Whether reverting this commit is enough, or whether users who applied with this version would be left with
state or B2 resources the previous release cannot read.

## Checklist

- [ ] `CHANGELOG.md` has an `[Unreleased]` entry, or the PR has the `-changelog` label.
- [ ] `make docs` was run if a schema changed.
- [ ] New or changed behavior is pinned by a test that fails when the change is reverted.
- [ ] The description matches the final code.
- [ ] The person who directed the agent has read Testing, Risk and How to revert and signs off on them.
