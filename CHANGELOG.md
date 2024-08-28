# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
* Fixed `allowed_operations` stability issue

### Infrastructure
* Replace removed macOS 11 Big Sur in favour of macOS 12 Monterey in CI/CD

## [0.8.12] - 2024-06-20

### Infrastructure
* Fixed goreleaser v2 configuration

## [0.8.11] - 2024-06-20

### Infrastructure
* Fixed continuous delivery issue caused by goreleaser update

## [0.8.10] - 2024-06-20

### Changed
* Upgraded b2sdk from v1.18.0 to v2.4.1
* Upgraded pyinstaller from v6.3.0 to v6.8.0

## [0.8.9] - 2024-01-05

### Changed
* Upgraded github.com/cloudflare/circl from v1.3.3 to v1.3.7

### Fixed
* Fixed bucket update when is_file_lock_enabled is not set

## [0.8.8] - 2023-12-28

### Changed
* Upgraded pyinstaller to v6.3.0

## [0.8.7] - 2023-12-27

### Changed
* Upgraded github.com/hashicorp/terraform-plugin-sdk to v2.31.0

## [0.8.6] - 2023-12-22

### Fixed
* Fix arm64 builds

## [0.8.5] - 2023-11-24

### Changed
* Upgraded go to 1.20 and github.com/hashicorp/terraform-plugin-sdk to v2.30.0

### Infrastructure
* Disable changelog verification for dependabot PRs
* Upgrade macOS version in CI/CD
* Upgrade Terraform version for ACC tests
* Run ACC tests for all supported Terraform versions
* Do not use deprecated `::set-output` GitHub Actions command in favor of `GITHUB_OUTPUT` env

### Fixed
* Reconcile missing Application Key caused by the resource drift
* Fix reconciliation of missing Bucket caused by the resource drift
* Fix bucket cleanup after failed creation

## [0.8.4] - 2023-03-13

### Infrastructure
* Upgraded terraform-plugin-docs 0.5.1 -> 0.13.0
* Upgraded golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd -> 0.7.0

## [0.8.3] - 2023-02-20

### Infrastructure
* Upgraded golang.org/x/net 0.5.0 -> 0.7.0

## [0.8.2] - 2023-02-17

### Infrastructure
* Upgraded goutils 1.1.0 -> 1.1.1 and aws to 1.33.0
* Ensured that changelog validation only happens on pull requests

## [0.8.1] - 2022-06-24

### Changed
* Upgraded github.com/hashicorp/terraform-plugin-sdk/ to v2.17.0 and github.com/hashicorp/go-getter to v1.6.2

### Fixed
* Fixed golangcli-lint breaking on Github

## [0.8.0] - 2022-03-27

### Added
* Added importer for b2_bucket and b2_application_key resources
* Added signed URL as data source to allow downloading files from private bucket during provisioning without storing an API key

### Changed
* Upgraded go to 1.18 and github.com/hashicorp/terraform-plugin-sdk/ to v2.12.0
* Upgraded b2sdk to 1.14.1, which allowed using improved API calls for listing files and making Python parts simpler
* Upgraded PyInstaller to 4.10, which should help resolve some issues with running on Apple M1 silicon

## [0.7.1] - 2021-10-14

### Changed
* When a bucket that is in state no longer exists, warning is logged and the bucket is removed from state

## [0.7.0] - 2021-09-24

### Fixed
* Fix for static linking of Python bindings for Linux (CD uses python container)

## [0.6.1] - 2021-09-01

### Changed
* When deleting bucket that bucket no longer exists, error is silently ignored
* Terraform 1.0.0 or later required

## [0.6.0] - 2021-07-06

### Added
* Support SSE-C encryption mode for files
* Initial (experimental) support for Alpine Linux
* Support for Apple M1 (arm64) architecture on Mac OS (for main plugin binary, Python bindings will still use Rosetta)

## [0.5.0] - 2021-05-31

### Added
* Support isFileLockEnabled for buckets
* Support defaultRetention for buckets

### Fixed
* Fix acceptance tests breaking when new response fields are added to the API

### Changed
* Upgraded b2sdk version to 1.8.0

## [0.4.0] - 2021-04-08

### Added
* Show S3-compatible API URL in `b2_account_info` data source

### Fixed
* Upgrade b2sdk version - fix for server response change regarding SSE

## [0.3.0] - 2021-03-27

### Added
* Added `b2_account_info` data source
* Add support for SSE-B2 server-side encryption mode

### Changed
* Better handling sensitive data in Terraform logs
* Upgrade b2sdk version `>=1.4.0,<2.0.0`

## [0.2.1] - 2021-02-11

### Changed
* Upgrade b2sdk version

### Fixed
* Append Terraform versions to the User-Agent
* Fix the defaults for lifecycle rules #4

## [0.2.0] - 2021-01-22

### Added
* Added `b2_bucket` data source
* Added `b2_bucket_file` data source
* Added `b2_bucket_files` data source
* Added `b2_application_key` resource
* Added `b2_bucket` resource
* Added `b2_bucket_file_version` resource

### Changed
* Extended `b2` provider
* Extended `b2_application_key` data source
* Improved python bindings

## [0.1.0] - 2020-11-30

### Added
* Implementation of PoC (simple `b2_application_key` data source)

[Unreleased]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.8.12...HEAD
[0.8.11]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.8.11...v0.8.12
[0.8.11]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.8.10...v0.8.11
[0.8.10]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.8.9...v0.8.10
[0.8.9]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.8.8...v0.8.9
[0.8.8]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.8.7...v0.8.8
[0.8.7]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.8.6...v0.8.7
[0.8.6]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.8.5...v0.8.6
[0.8.5]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.8.4...v0.8.5
[0.8.4]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.8.3...v0.8.4
[0.8.3]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.8.2...v0.8.3
[0.8.2]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.8.1...v0.8.2
[0.8.1]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.8.0...v0.8.1
[0.8.0]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.7.1...v0.8.0
[0.7.1]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.6.1...v0.7.0
[0.6.1]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/Backblaze/terraform-provider-b2/compare/240851d...v0.1.0
