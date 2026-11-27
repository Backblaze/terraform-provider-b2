# Security Policy

## Supported versions

Please check the latest released version when reporting a security issue. Maintainers will assess reports
affecting older versions case by case. Fixes ship in a new release; see [`RELEASE.md`](RELEASE.md).

## Reporting a vulnerability

Please report suspected vulnerabilities privately through [Backblaze's Bugcrowd program](https://bugcrowd.com/engagements/backblaze),
as directed by the [Backblaze Trust Center](https://www.backblaze.com/company/trust-center). Do not open a
public GitHub issue for a security report.

Include the affected version, a description of the issue, steps to reproduce it, and any relevant impact.
Do not include credentials or other sensitive customer data.

Please allow time for the maintainers to investigate and coordinate a fix before disclosing the issue
publicly. The security team can coordinate follow-up and disclosure through the reporting channel. Fixes
ship as a release with a `Security` changelog entry.

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
- **Sensitive attributes**: `application_key`, SSE-C `secret_b64`, `hmac_sha256_signing_secret` and
  `b2_bucket_file_signed_url.signed_url` are marked `Sensitive`, so Terraform hides them in plan output; they are still stored in state.
- **Logs and errors**: `TF_LOG=DEBUG` logs provider input, output and bindings stderr after redacting
  values the schema marks `Sensitive`, the provider credentials, signed URLs and webhook custom header
  values (`b2/client.go`); errors returned to Terraform carry only the operation and the reported error.
  Any new secret must be marked `Sensitive` so the redaction covers it.
- **Tests and CI**: acceptance tests read `B2_TEST_APPLICATION_KEY_ID` and `B2_TEST_APPLICATION_KEY` from
  repository secrets in `ci.yml`; Dependabot and fork PRs do not receive them. The `.sdkharness` checks
  refuse any non-loopback endpoint and drop ambient `B2_*` variables.

## Best practices for users

- Use application keys scoped to the buckets and capabilities the configuration manages.
- Store Terraform state in an encrypted backend with restricted access; it contains keys and secrets.
- Pass credentials through environment variables or a secrets manager, not literals in `.tf` files.
- Rotate application keys regularly.
