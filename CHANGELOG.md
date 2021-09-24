# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.6.1...HEAD
[0.6.1]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/Backblaze/terraform-provider-b2/compare/240851d...v0.1.0
