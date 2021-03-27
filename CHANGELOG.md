# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/Backblaze/terraform-provider-b2/compare/240851d...v0.1.0
