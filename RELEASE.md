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
   Install GoReleaser 2.20.2, GnuPG and OpenSSH, then run `make release-health`
   locally. It checks release-tool tests, generated notes,
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

   The imported v0.15.0 section is an untagged candidate. If the chosen version
   already has candidate notes, first generate the new entries with
   `changie batch vX.Y.Z --dry-run` into a temporary file. Review and incorporate
   those entries into `.changes/vX.Y.Z.md`, preserving its existing entries and
   comparison link, remove only the incorporated fragments, then run `changie merge`.
   Do not use `batch --force` to append: it overwrites the version file. Never
   rewrite notes for a published version. Review the release date.
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

An authorized maintainer creates and pushes a signed release tag on that exact commit:

```sh
git tag -s vX.Y.Z "$release_sha" -m "vX.Y.Z"
git verify-tag vX.Y.Z
# Push only after the approved release checks are complete:
git push origin vX.Y.Z
```

For SSH signing, configure `gpg.format=ssh`, `user.signingkey`, and a local
`gpg.ssh.allowedSignersFile` before these commands. GPG and SSH tag signatures are
supported. Tag-signing keys and the Registry checksum-signing key have separate roles.

Do not tag until the release checks and approvals are complete. A pushed `v*` tag
starts `.github/workflows/cd.yml`, which builds the five platform targets and
publishes the signed Registry artifacts. It uses the assembled Changie notes as
the GitHub Release body through `gh release create`; GoReleaser creates the archives
and checksum signature but does not publish. The workflow uploads all artifacts to
a draft, then publishes it. An upload failure leaves the draft for maintainer
inspection; do not overwrite an already published release on retry. Do not create
a separate manual release first.

After publication, verify the GitHub Release, all five ZIPs, SHA256SUMS and its GPG
signature, and the Terraform Registry version. Confirm the signature against the
Registry's enrolled signing key. Close the milestone only when publication and
installation have been verified. Merging a release PR alone does not ship it.

Never move or replace a published tag or artifact. Fix a failed published release
with a new version. Keep release signing material out of logs and local fixtures.

## Required repository setup and gates

- Require the changelog check and human approval on `master`. Restrict `v*` tag
  creation to release maintainers through a repository ruleset. Workflow code alone
  cannot prevent someone with workflow-write permissions from removing its gates.
- Use the version milestone on the release PR. The tag gate requires the tag commit
  to be on `master` and to equal the merge commit of a same-repository
  `release/vX.Y.Z` PR to `master` with the matching milestone. Do not tag a later tip.
- **Before enforcing signatures:** each account in `.github/signing-maintainers`
  (`goanpeca`, `ffumero2003`, `sophiecarreras`) must create a GPG or SSH signing key
  and register it on GitHub. Review the enrolled keys through an independent channel,
  then set repository variable `ENFORCE_SIGNED_TAGS=true`. Until then, missing keys,
  unsigned tags and failed verification produce a warning rather than block. This
  is a staged rollout, not an enforced signature policy yet.
- At verification time the gate downloads each account's `.gpg` and the public
  `/users/{username}/ssh_signing_keys` API over
  HTTPS into an isolated keyring/allowed-signers file and runs `git verify-tag`.
  It first checks `/users/{username}/gpg_keys`; an empty list permits SSH-only
  enrollment without fetching an absent GPG export. Unexpected or unreadable
  responses still hold verification.
  GitHub account/key control is the trust anchor, not a pinned fingerprint. The
  SSH key must be registered as a **Signing Key**, not just an authentication key.
  Allowed SSH keys are constrained to the `git` namespace and accept any tagger
  email; the enrolled key establishes authorization. Hardware-backed SSH signing
  keys are supported. SSH authentication keys are not enrolled for signing.
  Unavailable key sources fail verification; they never establish a valid signature.
- Configure `RELEASE_SECURITY_TOKEN` with read access to code-scanning, Dependabot,
  secret-scanning alerts and private repository advisories. Some advisory APIs may
  require an administrator-authorized token. Keep it limited to the gate job. The
  gate blocks on open medium/high/critical or unknown severity, including published
  advisories unless closed or withdrawn, all open secret
  alerts, or any unreadable source. It prints counts only. No findings are silently
  waived for lack of a patch. Review/dismiss findings through the approved private
  process; do not disclose advisory details in public release PRs. Run
  `make release-check` with `gh` authenticated for those read permissions; never
  put a token value in shell history.
- Keep `GPG_PRIVATE_KEY` and `PASSPHRASE` for the already-enrolled Registry signing key.
  Optionally set public variable `REGISTRY_GPG_FINGERPRINT` to its verified full
  fingerprint; a mismatch blocks import. The publishing job imports exactly one
  secret key in a mode-0700 temporary GnuPG home, sets owner trust, and passes its
  public fingerprint and a mode-0600 passphrase file to GoReleaser. An always-run
  step removes the keyring and passphrase before the GitHub release is created.
  Never reuse tag-verification fixtures or newly generated keys for Registry signing.
- The Registry flow remains GPG-signed SHA256SUMS; no OIDC token replaces that signature.
  The workflow requests no `id-token: write` permission because there is no OIDC
  exchange in this publishing path. Tag verification and security gates complete
  before any signing secrets are made available.

For a non-publishing packaging rehearsal, supply the five real platform binding
executables under `python-bindings/dist/`, then run
`goreleaser release --snapshot --clean --skip=publish -p 1`. The `-p 1` is required:
the existing Go embed hooks reuse `b2/py-terraform-provider-b2` between targets.
A smoke test with marked binding fixtures checks archive/checksum layout only and
cannot replace the real platform builds. Use a disposable test key and passphrase
file for snapshot signing, never the Registry key. To rehearse without any key,
add `--skip=publish,sign` instead. The offline signing tests also build all five
archives with marked binding fixtures, verify their checksums, and verify the
detached checksum signature with a disposable key. They do not publish anything.

## Action and platform maintenance

All actions are pinned to full commit SHAs with version comments. Review updates
before changing pins. The community GPG importer and release/changelog actions are
replaced by in-workflow GnuPG, `gh`, `awk`, and repository checks. Official GitHub
actions and the upstream HashiCorp/GoReleaser vendor actions remain; repository-level
allowlist settings still require administrator confirmation.

Linux CI uses the same official CPython bookworm image as CD, so staticx keeps
testing the shipping packaging path. macOS and Windows use `actions/setup-python`.
Do not disable staticx merely to make setup-python's Linux interpreter pass. The
existing macOS runner labels still require organization runner availability; fix
that independently rather than bypassing platform builds during a release.
