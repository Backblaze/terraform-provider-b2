#!/usr/bin/env bash
# Import only the release key, in an isolated keyring retained until GoReleaser finishes.
set +x
set -euo pipefail
umask 077

: "${GPG_PRIVATE_KEY:?GPG_PRIVATE_KEY is required}"
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"
: "${GITHUB_ENV:?GITHUB_ENV is required}"
: "${GITHUB_OUTPUT:?GITHUB_OUTPUT is required}"

export GNUPGHOME
GNUPGHOME=$(mktemp -d "$RUNNER_TEMP/provider-signing.XXXXXX")
# Register cleanup before attempting the import, including failed imports.
printf 'GNUPGHOME=%s\n' "$GNUPGHOME" >> "$GITHUB_ENV"
printf 'allow-preset-passphrase\ndefault-cache-ttl 7200\nmax-cache-ttl 7200\n' > "$GNUPGHOME/gpg-agent.conf"

fail() {
  printf '%s\n' "$1" >&2
  exit 1
}

if [[ "$GPG_PRIVATE_KEY" == *'-----BEGIN PGP PRIVATE KEY BLOCK-----'* ]]; then
  printf '%s' "$GPG_PRIVATE_KEY" > "$GNUPGHOME/key"
elif ! printf '%s' "$GPG_PRIVATE_KEY" | base64 --decode > "$GNUPGHOME/key" 2>/dev/null; then
  fail 'Could not import release signing key'
fi
if ! gpg --batch --import "$GNUPGHOME/key" > "$GNUPGHOME/import.log" 2>&1; then
  fail 'Could not import release signing key'
fi
rm -f "$GNUPGHOME/key" "$GNUPGHOME/import.log"
unset GPG_PRIVATE_KEY

gpg --batch --with-colons --with-keygrip --list-secret-keys > "$GNUPGHOME/keys" 2>/dev/null
fingerprints=$(awk -F: '$1 == "sec" { primary = 1; next } primary && $1 == "fpr" { print $10; primary = 0 }' "$GNUPGHOME/keys")
if [[ ! "$fingerprints" =~ ^[A-F0-9]{40,64}$ ]]; then
  fail 'Expected exactly one primary release signing key'
fi
fingerprint=$fingerprints
printf '%s:6:\n' "$fingerprint" | gpg --batch --import-ownertrust > /dev/null 2>&1

# GoReleaser uses plain `gpg --batch --local-user ... --detach-sign`.
# Seed the agent for the primary key and signing subkeys, as the former Action did,
# so passphrase-protected keys need no pinentry or changes to .goreleaser.yml.
if [[ -n "${PASSPHRASE:-}" ]]; then
  preset="$(gpgconf --list-dirs libexecdir)/gpg-preset-passphrase"
  [[ -x "$preset" ]] || fail 'gpg-preset-passphrase is unavailable'
  gpgconf --launch gpg-agent
  while IFS= read -r keygrip; do
    if ! printf '%s' "$PASSPHRASE" | "$preset" --preset "$keygrip" > /dev/null 2>&1; then
      fail 'Could not unlock release signing key'
    fi
  done < <(awk -F: '$1 == "grp" {print $10}' "$GNUPGHOME/keys")
fi
unset PASSPHRASE
rm -f "$GNUPGHOME/keys"

printf 'GPG_FINGERPRINT=%s\n' "$fingerprint" >> "$GITHUB_ENV"
printf 'fingerprint=%s\n' "$fingerprint" >> "$GITHUB_OUTPUT"
printf 'Imported release signing fingerprint: %s\n' "$fingerprint"
