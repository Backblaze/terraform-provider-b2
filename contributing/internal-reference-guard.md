# Internal reference guard

The `Internal reference guard` workflow stops pull requests from naming internal hosts or internal tracker keys. This repository is public.

## What it checks

The workflow scans four things:

- The pull request title.
- The pull request body.
- Each commit message in the pull request.
- Each file the pull request adds or changes.

A tracker key is a configured prefix, a hyphen and digits, such as a key followed by `-12`. The guard ignores a key that is part of a longer word. An internal host is a configured host name. The host match ignores case. Public references such as `#231` and links to `github.com` pass.

The workflow runs `scripts/internal-reference-guard.sh`. It reads the pull request checkout, so it catches accidents and does not stop a deliberate change to the script.

## Run it locally

Set both lists, then scan a file, a title, a body or a commit message. The example below uses fake sample values.

```sh
export TRACKER_KEYS='EXAMPLE|SAMPLE'
export INTERNAL_HOSTS='internal\.example\.test'
scripts/internal-reference-guard.sh file README.md
scripts/internal-reference-guard.sh title 'A pull request title'
scripts/internal-reference-guard.sh commit "$(git rev-parse HEAD)" "$(git log -1 --format=%B)"
```

To scan every tracked file in the working tree, run this loop:

```sh
git ls-files -z | while IFS= read -r -d '' f; do scripts/internal-reference-guard.sh file "$f"; done
```

The script exits 0 when the text is clean, 1 when it finds a reference and 2 on a usage error.

Run the test script with `scripts/internal-reference-guard_test.sh`. It reads the lists from the workflow.

## Where maintainers set the lists

Maintainers edit `TRACKER_KEYS` and `INTERNAL_HOSTS` in the `env` block at the top of `.github/workflows/internal-reference-guard.yml`. The committed values are fake placeholders. Each value is a regular expression alternation, such as `EXAMPLE|SAMPLE`. Do not copy the real values into the script, the tests or documentation.
