//####################################################################
//
// File: b2/resource_b2_bucket_test.go
//
// Copyright 2021 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceB2Bucket_basic(t *testing.T) {
	resourceName := "b2_bucket.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2BucketConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "account_id", regexp.MustCompile("^[a-zA-Z0-9]{12}$")),
					resource.TestMatchResourceAttr(resourceName, "bucket_id", regexp.MustCompile("^[a-zA-Z0-9]{24}$")),
					resource.TestCheckResourceAttr(resourceName, "bucket_info.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", bucketName),
					resource.TestCheckResourceAttr(resourceName, "bucket_type", "allPublic"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "file_lock_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "file_lock_configuration.is_file_lock_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "file_lock_configuration.mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "default_server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_server_side_encryption.0.mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "default_server_side_encryption.0.algorithm", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0", "s3"),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
		},
	})
}

func TestAccResourceB2Bucket_all(t *testing.T) {
	resourceName := "b2_bucket.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2BucketConfig_all(bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "account_id", regexp.MustCompile("^[a-zA-Z0-9]{12}$")),
					resource.TestMatchResourceAttr(resourceName, "bucket_id", regexp.MustCompile("^[a-zA-Z0-9]{24}$")),
					resource.TestCheckResourceAttr(resourceName, "bucket_info.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "bucket_info.description", "the bucket"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", bucketName),
					resource.TestCheckResourceAttr(resourceName, "bucket_type", "allPrivate"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.cors_rule_name", "downloadFromAnyOrigin"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.allowed_origins.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.allowed_origins.0", "https"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.allowed_operations.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.allowed_operations.0", "b2_download_file_by_id"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.allowed_operations.1", "b2_download_file_by_name"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.expose_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.expose_headers.0", "x-bz-content-sha1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.allowed_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.allowed_headers.0", "range"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.max_age_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "file_lock_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "file_lock_configuration.is_file_lock_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "file_lock_configuration.default_retention.mode", "governance"),
					resource.TestCheckResourceAttr(resourceName, "file_lock_configuration.default_retention.period.duration", "7"),
					resource.TestCheckResourceAttr(resourceName, "file_lock_configuration.default_retention.period.unit", "days"),
					resource.TestCheckResourceAttr(resourceName, "default_server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_server_side_encryption.0.mode", "SSE-B2"),
					resource.TestCheckResourceAttr(resourceName, "default_server_side_encryption.0.algorithm", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.0.file_name_prefix", "c/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.0.days_from_hiding_to_deleting", "2"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.0.days_from_uploading_to_hiding", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0", "s3"),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
		},
	})
}

func TestAccResourceB2Bucket_lifecycleRulesDefaults(t *testing.T) {
	resourceName := "b2_bucket.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2BucketConfig_lifecycleRulesDefaults(bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "account_id", regexp.MustCompile("^[a-zA-Z0-9]{12}$")),
					resource.TestCheckResourceAttr(resourceName, "bucket_info.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", bucketName),
					resource.TestCheckResourceAttr(resourceName, "bucket_type", "allPublic"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.0.file_name_prefix", "a/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.0.days_from_hiding_to_deleting", "2"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.0.days_from_uploading_to_hiding", "0"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.1.file_name_prefix", "b/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.1.days_from_hiding_to_deleting", "0"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.1.days_from_uploading_to_hiding", "2"),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0", "s3"),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
		},
	})
}

func TestAccResourceB2Bucket_update(t *testing.T) {
	resourceName := "b2_bucket.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2BucketConfig_basic(bucketName),
			},
			{
				Config: testAccResourceB2BucketConfig_all(bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "account_id", regexp.MustCompile("^[a-zA-Z0-9]{12}$")),
					resource.TestMatchResourceAttr(resourceName, "bucket_id", regexp.MustCompile("^[a-zA-Z0-9]{24}$")),
					resource.TestCheckResourceAttr(resourceName, "bucket_info.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "bucket_info.description", "the bucket"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", bucketName),
					resource.TestCheckResourceAttr(resourceName, "bucket_type", "allPrivate"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.cors_rule_name", "downloadFromAnyOrigin"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.allowed_origins.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.allowed_origins.0", "https"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.allowed_operations.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.allowed_operations.0", "b2_download_file_by_id"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.allowed_operations.1", "b2_download_file_by_name"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.expose_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.expose_headers.0", "x-bz-content-sha1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.allowed_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.allowed_headers.0", "range"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.0.max_age_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "default_server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_server_side_encryption.0.mode", "SSE-B2"),
					resource.TestCheckResourceAttr(resourceName, "default_server_side_encryption.0.algorithm", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.0.file_name_prefix", "c/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.0.days_from_hiding_to_deleting", "2"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.0.days_from_uploading_to_hiding", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0", "s3"),
					resource.TestCheckResourceAttr(resourceName, "revision", "3"),
				),
			},
			{
				Config: testAccResourceB2BucketConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "account_id", regexp.MustCompile("^[a-zA-Z0-9]{12}$")),
					resource.TestMatchResourceAttr(resourceName, "bucket_id", regexp.MustCompile("^[a-zA-Z0-9]{24}$")),
					resource.TestCheckResourceAttr(resourceName, "bucket_info.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", bucketName),
					resource.TestCheckResourceAttr(resourceName, "bucket_type", "allPublic"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_server_side_encryption.0.mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "default_server_side_encryption.0.algorithm", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0", "s3"),
					resource.TestCheckResourceAttr(resourceName, "revision", "4"),
				),
			},
		},
	})
}

func testAccResourceB2BucketConfig_basic(bucketName string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPublic"
}
`, bucketName)
}

func testAccResourceB2BucketConfig_lifecycleRulesDefaults(bucketName string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPublic"

  lifecycle_rules {
    file_name_prefix = "a/"
    days_from_hiding_to_deleting = 2
  }
  lifecycle_rules {
    file_name_prefix = "b/"
    days_from_uploading_to_hiding = 2
  }
}
`, bucketName)
}

func testAccResourceB2BucketConfig_all(bucketName string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPrivate"
  bucket_info = {
    description = "the bucket"
  }
  cors_rules {
    cors_rule_name = "downloadFromAnyOrigin"
    allowed_origins = [
      "https"
    ]
    allowed_operations = [
      "b2_download_file_by_id",
      "b2_download_file_by_name"
    ]
    expose_headers = ["x-bz-content-sha1"]
    allowed_headers = ["range"]
    max_age_seconds = 3600
  }
  file_lock_configuration {
	is_file_lock_enabled = true
	default_retention {
	  mode = "governance"
	  period {
		duration = 7
		unit = "days"
	  }
	}
  }
  default_server_side_encryption {
    mode = "SSE-B2"
    algorithm = "AES256"
  }
  lifecycle_rules {
    file_name_prefix = "c/"
    days_from_hiding_to_deleting = 2
    days_from_uploading_to_hiding = 1
  }
}
`, bucketName)
}

