//####################################################################
//
// File: b2/data_source_b2_bucket_test.go
//
// Copyright 2021 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceB2Bucket_basic(t *testing.T) {
	resourceName := "b2_bucket.test"
	dataSourceName := "data.b2_bucket.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceB2BucketConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "account_id", resourceName, "account_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_id", resourceName, "bucket_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_info", resourceName, "bucket_info"),
					resource.TestCheckResourceAttr(dataSourceName, "bucket_name", bucketName),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_name", resourceName, "bucket_name"),
					resource.TestCheckResourceAttr(dataSourceName, "bucket_type", "allPublic"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_type", resourceName, "bucket_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_rules", resourceName, "cors_rules"),
					resource.TestCheckResourceAttrPair(dataSourceName, "default_server_side_encryption", resourceName, "default_server_side_encryption"),
					resource.TestCheckResourceAttrPair(dataSourceName, "lifecycle_rules", resourceName, "lifecycle_rules"),
					resource.TestCheckResourceAttrPair(dataSourceName, "options", resourceName, "options"),
					resource.TestCheckResourceAttrPair(dataSourceName, "revision", resourceName, "revision"),
				),
			},
		},
	})
}

func TestAccDataSourceB2Bucket_all(t *testing.T) {
	resourceName := "b2_bucket.test"
	dataSourceName := "data.b2_bucket.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceB2BucketConfig_all(bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "account_id", resourceName, "account_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_id", resourceName, "bucket_id"),
					resource.TestCheckResourceAttr(dataSourceName, "bucket_info.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "bucket_info.description", "the bucket"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_info", resourceName, "bucket_info"),
					resource.TestCheckResourceAttr(dataSourceName, "bucket_name", bucketName),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_name", resourceName, "bucket_name"),
					resource.TestCheckResourceAttr(dataSourceName, "bucket_type", "allPrivate"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_type", resourceName, "bucket_type"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_rules.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_rules.0.cors_rule_name", "downloadFromAnyOrigin"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_rules.0.allowed_origins.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_rules.0.allowed_origins.0", "https"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_rules.0.allowed_operations.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_rules.0.allowed_operations.0", "b2_download_file_by_id"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_rules.0.allowed_operations.1", "b2_download_file_by_name"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_rules.0.expose_headers.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_rules.0.expose_headers.0", "x-bz-content-sha1"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_rules.0.allowed_headers.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_rules.0.allowed_headers.0", "range"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_rules.0.max_age_seconds", "3600"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_rules", resourceName, "cors_rules"),
					resource.TestCheckResourceAttr(dataSourceName, "default_server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "default_server_side_encryption.0.mode", "SSE-B2"),
					resource.TestCheckResourceAttr(dataSourceName, "default_server_side_encryption.0.algorithm", "AES256"),
					resource.TestCheckResourceAttrPair(dataSourceName, "default_server_side_encryption", resourceName, "default_server_side_encryption"),
					resource.TestCheckResourceAttr(dataSourceName, "lifecycle_rules.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "lifecycle_rules.0.file_name_prefix", ""),
					resource.TestCheckResourceAttr(dataSourceName, "lifecycle_rules.0.days_from_hiding_to_deleting", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "lifecycle_rules.0.days_from_uploading_to_hiding", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "lifecycle_rules", resourceName, "lifecycle_rules"),
					resource.TestCheckResourceAttrPair(dataSourceName, "options", resourceName, "options"),
					resource.TestCheckResourceAttrPair(dataSourceName, "revision", resourceName, "revision"),
				),
			},
		},
	})
}

func testAccDataSourceB2BucketConfig_basic(bucketName string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPublic"
}

data "b2_bucket" "test" {
  bucket_name = b2_bucket.test.bucket_name

  depends_on = [
    b2_bucket.test,
  ]
}
`, bucketName)
}

func testAccDataSourceB2BucketConfig_all(bucketName string) string {
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
  default_server_side_encryption {
    mode = "SSE-B2"
    algorithm = "AES256"
  }
  lifecycle_rules {
    file_name_prefix = ""
    days_from_hiding_to_deleting = 2
    days_from_uploading_to_hiding = 1
  }
}

data "b2_bucket" "test" {
  bucket_name = b2_bucket.test.bucket_name

  depends_on = [
    b2_bucket.test,
  ]
}
`, bucketName)
}
