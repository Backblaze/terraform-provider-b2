# Releasing the B2 provider

A release is a `vX.Y.Z` tag on `master`. Pushing the tag runs `.github/workflows/cd.yml`, which builds the
Python bindings for every platform, builds and signs the provider with goreleaser, creates the GitHub
release, and the Terraform Registry picks it up from there. The provider version comes from the tag
(`-X main.version` in `.goreleaser.yml`), so no source file holds a version string.

The changelog is assembled from `.changes/unreleased` entries with [changie](https://changie.dev).

## Versioning

Semantic versioning. Before 1.0, a change that needs users to edit their configuration or that forces
replacement of existing resources bumps the minor version. changie picks the bump from the entry kinds
(`Added`, `Changed`, `Deprecated`, `Removed` → minor; `Fixed`, `Security`, `Infrastructure` → patch); pass
an explicit version to `changie batch` to override it. `v1.0.0` is always a deliberate `changie batch v1.0.0`.

## Security and quality gate

A release is held while any solvable security or quality finding is open on the
[Security tab](https://github.com/Backblaze/terraform-provider-b2/security). `make release-check` runs the
gate and exits non-zero to hold:

| Source | Holds the release when | Clear it by |
| --- | --- | --- |
| Code scanning (CodeQL security and quality queries) | Any alert is open on the default branch | Fixing it, or a maintainer dismissing it with a written reason |
| Dependabot alerts (Go modules and `python-bindings` requirements) | An open alert has a patched version | Upgrading the dependency |
| `govulncheck` | A reachable known vulnerability is reported | Upgrading the module or Go toolchain |
| Repository security advisories | An advisory is in triage or draft that this release does not fix | Fixing it in this release (see below) |
| Secret scanning | Any alert is open | Revoking the secret, then closing the alert |

Alerts with no patched version yet do not hold the release; the gate lists them as notes. A section the
script cannot read (for example secret scanning without security access) is reported as UNVERIFIED and
must be checked by hand on the Security tab; it never passes silently.

Never open a public issue for a repository advisory or a code scanning security alert: track those in the
private advisory. A draft advisory fixed by the pending release stays private through tagging; list its id
in `RELEASE_FIXES_ADVISORIES=GHSA-xxxx-xxxx-xxxx make release-check` so the gate reports it instead of
holding, and publish the advisory right after the release is out (step 8).

## Steps

1. Make sure `master` is green and every PR meant for the release has merged. Install dependencies with
   `make deps`, then run `make release-check`; if it holds, fix or dismiss the findings first.
2. Branch from `master`: `git switch -c release/vX.Y.Z origin/master`.
3. Batch the entries and rebuild the changelog:

   ```bash
   go run github.com/miniscruff/changie@v1.24.0 batch auto --dry-run
   go run github.com/miniscruff/changie@v1.24.0 batch auto
   go run github.com/miniscruff/changie@v1.24.0 merge
   ```

4. Check the diff touches only `CHANGELOG.md` and `.changes/`. Edit `.changes/vX.Y.Z.md` for readability if
   needed and run `merge` again.
5. Run the acceptance tests against a real account and the local gate:

   ```bash
   export B2_TEST_APPLICATION_KEY_ID=your_app_key_id
   export B2_TEST_APPLICATION_KEY=your_app_key
   make testacc
   make ci
   ```

6. Commit as `Version X.Y.Z` and open a PR with the changelog section and the `make release-check` output
   as its body; leave out any advisory ids until they are published. A maintainer approves and CI is green.
7. After it merges, tag the release PR's merge commit, not the tip of `master`, which may already include
   later PRs:

   ```bash
   git switch master && git pull --ff-only
   sha=$(gh pr view <release-pr-number> --json mergeCommit --jq .mergeCommit.oid)
   git log -1 --oneline "$sha"   # must be the "Version X.Y.Z" commit
   git tag vX.Y.Z "$sha"
   git push origin vX.Y.Z
   ```

8. Wait for the CD workflow to finish, then check:
   - the [GitHub release](https://github.com/Backblaze/terraform-provider-b2/releases) exists with the
     binaries for every platform, the checksums and their signature;
   - the [Terraform Registry](https://registry.terraform.io/providers/Backblaze/b2/latest) shows vX.Y.Z;
   - `terraform init -upgrade` with `version = "X.Y.Z"` installs it.

   Then publish each advisory listed in `RELEASE_FIXES_ADVISORIES`, naming vX.Y.Z as the patched version.

## Never

- Release while `make release-check` holds, or with an UNVERIFIED section nobody checked by hand.
- Move or delete a published tag. The Registry has already indexed it; release a new patch version instead.
- Edit a released `.changes/vX.Y.Z.md` without a reason recorded in the PR.
