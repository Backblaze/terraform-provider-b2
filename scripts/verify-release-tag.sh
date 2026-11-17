#!/usr/bin/env bash
# Public keys only. GitHub account control is the trust anchor; see RELEASE.md.
set -euo pipefail
enforce=${ENFORCE_SIGNED_TAGS:-false}
case "$enforce" in true|false) ;; *) echo 'ENFORCE_SIGNED_TAGS must be true or false' >&2; exit 1 ;; esac
tag=${1:?usage: verify-release-tag.sh vX.Y.Z}
verify() (
  [[ "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]] || exit 1
  test "$(git cat-file -t "$tag")" = tag || exit 1
  scratch=$(mktemp -d)
  trap 'gpgconf --homedir "$scratch/gpg" --kill all >/dev/null 2>&1 || true; rm -rf "$scratch"' EXIT
  mkdir -m 700 "$scratch/gpg"
  export GNUPGHOME="$scratch/gpg"
  : > "$scratch/allowed-signers"
  count=0
  while read -r user rest; do
    case "$user" in ''|\#*) continue ;; esac
    [[ "$user" =~ ^[A-Za-z0-9-]+$ ]] || exit 1
    curl --fail --silent --show-error --location --proto '=https' --proto-redir '=https' --tlsv1.2 --max-time 30 --retry 2 \
      "https://github.com/$user.gpg" > "$scratch/public.gpg" 2>/dev/null || exit 1
    curl --fail --silent --show-error --location --proto '=https' --proto-redir '=https' --tlsv1.2 --max-time 30 --retry 2 \
      "https://github.com/$user.keys" > "$scratch/public.keys" 2>/dev/null || exit 1
    if [ ! -s "$scratch/public.gpg" ] && [ ! -s "$scratch/public.keys" ]; then exit 1; fi
    if [ -s "$scratch/public.gpg" ]; then
      gpg --batch --quiet --import "$scratch/public.gpg" >/dev/null 2>&1 || exit 1
    fi
    while read -r kind key _comment; do
      case "$kind" in ssh-*|ecdsa-*) printf '%s %s %s\n' "$user" "$kind" "$key" >> "$scratch/allowed-signers" ;; '') ;; *) exit 1 ;; esac
    done < "$scratch/public.keys"
    count=$((count + 1))
  done < .github/signing-maintainers
  test "$count" -gt 0 || exit 1
  git -c gpg.program=gpg verify-tag "$tag" >/dev/null 2>&1 && exit 0
  git -c gpg.ssh.allowedSignersFile="$scratch/allowed-signers" verify-tag "$tag" >/dev/null 2>&1
)
if verify; then
  echo 'Release tag signature: PASS'
elif [ "$enforce" = true ]; then
  echo 'Release tag signature: HOLD (signature or maintainer enrollment could not be verified)' >&2
  exit 1
else
  echo '::warning::Release tag signature not verified; ENFORCE_SIGNED_TAGS is off. Complete maintainer enrollment before enabling enforcement.'
fi
