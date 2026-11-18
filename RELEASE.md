# Releasing the Terraform provider

The provider version comes from its Git tag through GoReleaser. Release artifacts keep
the Terraform Registry format: platform ZIPs, SHA256SUMS, and a detached GPG signature
of SHA256SUMS. GoReleaser's automatic changelog stays disabled.

## Changelog

Like [Backblaze/blazer](https://github.com/Backblaze/blazer/blob/main/RELEASE.md),
this repository uses Changie 1.24.0. Add a fragment with `make changelog-new`; do not
edit `CHANGELOG.md` directly. Existing history is in `.changes/vX.Y.Z.md`, and
`changie merge` assembles it with `.changes/header.tpl.md`.

The PR check requires a fragment for every change, matching the previous action's
default all-files scope and Dependabot exemption. A maintainer may apply `-changelog` to waive presence;
that does not waive the release-branch rule or generation checks. The previous
action also validated Markdown headings and package.json versions; the replacement
uses Changie's own assembly and versioned notes instead of Node package checks.

Before 1.0, use a minor bump for breaking changes and a patch for compatible fixes.
Choose major versions deliberately. `Infrastructure` is retained alongside the
Keep a Changelog kinds. Check the proposed version rather than relying solely on
`batch auto` when fragments describe a breaking change.

## Release milestone and preparation

1. Create a `vX.Y.Z` milestone. Put every intended public issue and PR in it. Track
   unpublished security work privately; never copy private findings into the milestone.
2. Confirm all intended changes have merged and review deferred items explicitly.
   Run `make release-health` locally. It checks release-tool tests, generated notes,
   Go unit tests and vet without publishing. Also run `make lint`, `make docs-lint`,
   and live acceptance tests with the documented test account before release.
3. Start from up-to-date public `origin/master`:

   ```sh
   git fetch origin
   git switch -c release/vX.Y.Z origin/master
   go run github.com/miniscruff/changie@v1.24.0 batch vX.Y.Z --dry-run
   go run github.com/miniscruff/changie@v1.24.0 batch vX.Y.Z
   go run github.com/miniscruff/changie@v1.24.0 merge
   ```

   If that version already has assembled candidate notes, update the untagged
   candidate using Changie's batch options instead of creating a duplicate version.
   Never rewrite notes for a published version. Review the release date.
   Update the `[Unreleased]` comparison in `.changes/header.tpl.md` to compare
   the new release tag with `HEAD`, then run `changie merge` again.
4. The release PR may change only `CHANGELOG.md` and `.changes/`. Commit as
   `chore(release): vX.Y.Z`, push the release branch, and open a PR to `master` with
   Context, Changes and Verification. Assign the milestone and obtain maintainer
   approval. CI enforces `release/vX.Y.Z` for assembled changelog changes; repository
   branch protection must require that check and human approval. Never push release
   commits directly to `master`.
5. After green CI and human merge, identify the release PR's exact merge commit:

   ```sh
   release_sha=$(gh pr view RELEASE_PR --json mergeCommit --jq .mergeCommit.oid)
   git fetch origin
   git show --stat "$release_sha"
   ```

   Confirm the milestone scope and assembled `.changes/vX.Y.Z.md` at that commit.
   Later commits on `master` are not automatically part of this release.

## Tagging and publication

An authorized maintainer creates and pushes the release tag on that exact commit.
Do not tag until the release checks and approvals are complete. A pushed `v*` tag
starts `.github/workflows/cd.yml`, which builds the five platform targets and
publishes the signed Registry artifacts. It uses the assembled Changie notes as
the GitHub Release body. Do not create a separate manual release first.

After publication, verify the GitHub Release, all five ZIPs, SHA256SUMS and its GPG
signature, and the Terraform Registry version. Confirm the signature against the
Registry's enrolled signing key. Close the milestone only when publication and
installation have been verified. Merging a release PR alone does not ship it.

Never move or replace a published tag or artifact. Fix a failed published release
with a new version. Keep release signing material out of logs and local fixtures.
