# Security Policy

Security-response owner: Backblaze Security. Repository maintainers triage reports and ship the fix.

## Supported Versions

The latest minor release line (currently `0.14.x`) is supported. Earlier releases are not patched in
place; upgrade to the newest `v0.x` release in the [Terraform Registry](https://registry.terraform.io/providers/Backblaze/b2/latest)
to receive fixes. See [`RELEASE.md`](RELEASE.md).

| Version | Supported |
|---|---|
| Latest minor (`0.14.x`) | Yes |
| Older releases | No |

## Reporting a Vulnerability

**Do not file a public GitHub issue for a security vulnerability.**

Instead, report it privately via GitHub Security Advisories:

1. Go to the [Security tab](https://github.com/Backblaze/terraform-provider-b2/security/advisories/new) of this repository
2. Click **Report a vulnerability**
3. Provide a clear description, the affected provider version, a minimal configuration, reproduction steps, and impact assessment

You can also email **security@backblaze.com** if you cannot use GitHub Security Advisories.

## What to expect

- Acknowledgement within 3 business days
- Initial assessment within 7 business days
- Coordinated disclosure once a fix is available, shipped as a patch release with a `Security` changelog
  entry and a GitHub security advisory

## Scope

In scope:

- Secrets exposed by the provider: application keys, SSE-C keys (`secret_b64`), event notification
  signing secrets (`hmac_sha256_signing_secret`) or authorization tokens written to logs, error messages,
  plan output or state fields not marked `Sensitive`
- The embedded Python bindings (`python-bindings/`): how the Go provider passes credentials to them, and
  vulnerable dependencies bundled into the binary
- Transport flaws: TLS verification bypassed, or requests sent to an endpoint other than the configured one
- Release integrity: build, signing and publishing in `.github/workflows/cd.yml` and `.goreleaser.yml`

Out of scope:

- Vulnerabilities in the Backblaze B2 service itself — report those to Backblaze directly
- Secrets that a configuration stores in Terraform state by design (state must be protected by its backend)
- Misuse of credentials by the user, such as committing a master application key to a repository
- Issues requiring root or local access on the machine running Terraform

## Security-relevant areas

- **Credentials**: `application_key_id` and `application_key` come from the provider block or the
  `B2_APPLICATION_KEY_ID` and `B2_APPLICATION_KEY` environment variables. The Go provider passes them to
  the embedded Python bindings on stdin, never on the command line.
- **Sensitive attributes**: `application_key`, SSE-C `secret_b64` and `hmac_sha256_signing_secret` are
  marked `Sensitive`, so Terraform hides them in plan output; they are still stored in state.
- **Logs**: `TF_LOG=DEBUG` logs provider input and output. Output is sanitized against the schema's
  `Sensitive` flags before logging; any new secret must be marked `Sensitive` and kept out of logs.
- **Tests and CI**: acceptance tests read `B2_TEST_APPLICATION_KEY_ID` and `B2_TEST_APPLICATION_KEY` from
  repository secrets in `ci.yml`; Dependabot and fork PRs do not receive them. The `.sdkharness` checks
  refuse any non-loopback endpoint and drop ambient `B2_*` variables.

## Best practices for users

- Use application keys scoped to the buckets and capabilities the configuration manages.
- Store Terraform state in an encrypted backend with restricted access; it contains keys and secrets.
- Pass credentials through environment variables or a secrets manager, not literals in `.tf` files.
- Rotate application keys regularly.
