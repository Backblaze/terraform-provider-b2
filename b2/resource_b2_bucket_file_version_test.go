//####################################################################
//
// File: b2/resource_b2_bucket_file_version_test.go
//
// Copyright 2021 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceB2BucketFileVersion_basic(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_bucket_file_version.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	tempFile := createTempFile(t, "hello")
	defer os.Remove(tempFile)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2BucketFileVersionConfig_basic(bucketName, tempFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "action", "upload"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(resourceName, "content_md5", "5d41402abc4b2a76b9719d911017c592"),
					resource.TestCheckResourceAttr(resourceName, "content_sha1", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "text/plain"),
					resource.TestCheckResourceAttr(resourceName, "file_info.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "file_name", "temp.txt"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.algorithm", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "5"),
					resource.TestCheckResourceAttr(resourceName, "source", tempFile),
					resource.TestMatchResourceAttr(resourceName, "upload_timestamp", regexp.MustCompile("^[0-9]{13}$")),
				),
			},
		},
	})
}

func TestAccResourceB2BucketFileVersion_all(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_bucket_file_version.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	tempFile := createTempFile(t, "hello")
	defer os.Remove(tempFile)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2BucketFileVersionConfig_all(bucketName, tempFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "action", "upload"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(resourceName, "content_md5", "5d41402abc4b2a76b9719d911017c592"),
					resource.TestCheckResourceAttr(resourceName, "content_sha1", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "octet/stream"),
					resource.TestCheckResourceAttr(resourceName, "file_info.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "file_info.description", "the file"),
					resource.TestCheckResourceAttr(resourceName, "file_name", "temp.bin"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.mode", "SSE-B2"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.algorithm", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "source", tempFile),
					resource.TestCheckResourceAttr(resourceName, "size", "5"),
					resource.TestMatchResourceAttr(resourceName, "upload_timestamp", regexp.MustCompile("^[0-9]{13}$")),
				),
			},
		},
	})
}

func TestAccResourceB2BucketFileVersion_forceNew(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_bucket_file_version.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	tempFile := createTempFile(t, "hello")
	defer os.Remove(tempFile)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2BucketFileVersionConfig_basic(bucketName, tempFile),
			},
			{
				Config: testAccResourceB2BucketFileVersionConfig_all(bucketName, tempFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "action", "upload"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(resourceName, "content_md5", "5d41402abc4b2a76b9719d911017c592"),
					resource.TestCheckResourceAttr(resourceName, "content_sha1", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "octet/stream"),
					resource.TestCheckResourceAttr(resourceName, "file_info.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "file_info.description", "the file"),
					resource.TestCheckResourceAttr(resourceName, "file_name", "temp.bin"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.mode", "SSE-B2"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.algorithm", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "source", tempFile),
					resource.TestCheckResourceAttr(resourceName, "size", "5"),
					resource.TestMatchResourceAttr(resourceName, "upload_timestamp", regexp.MustCompile("^[0-9]{13}$")),
				),
			},
		},
	})
}

func TestAccResourceB2BucketFileVersion_sse_c(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_bucket_file_version.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	tempFile := createTempFile(t, "hello")
	defer os.Remove(tempFile)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2BucketFileVersionConfig_sse_c(bucketName, tempFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "action", "upload"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(resourceName, "content_md5", "5d41402abc4b2a76b9719d911017c592"),
					resource.TestCheckResourceAttr(resourceName, "content_sha1", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "octet/stream"),
					resource.TestCheckResourceAttr(resourceName, "file_info.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "file_info.description", "the file"),
					resource.TestCheckResourceAttr(resourceName, "file_name", "temp.bin"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.mode", "SSE-C"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.algorithm", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key.0.key_id", "test_id"),
					resource.TestCheckResourceAttr(resourceName, "source", tempFile),
					resource.TestCheckResourceAttr(resourceName, "size", "5"),
					resource.TestMatchResourceAttr(resourceName, "upload_timestamp", regexp.MustCompile("^[0-9]{13}$")),
				),
			},
		},
	})
}

func testAccResourceB2BucketFileVersionConfig_basic(bucketName string, tempFile string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPublic"
}

resource "b2_bucket_file_version" "test" {
  bucket_id = b2_bucket.test.id
  file_name = "temp.txt"
  source = "%s"
}
`, bucketName, tempFile)
}

func testAccResourceB2BucketFileVersionConfig_all(bucketName string, tempFile string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPublic"
}

resource "b2_bucket_file_version" "test" {
  bucket_id = b2_bucket.test.id
  file_name = "temp.bin"
  source = "%s"
  content_type = "octet/stream"
  file_info = {
    description = "the file"
  }
  server_side_encryption {
    mode = "SSE-B2"
    algorithm = "AES256"
  }
}
`, bucketName, tempFile)
}

func testAccResourceB2BucketFileVersionConfig_sse_c(bucketName string, tempFile string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPublic"
}

resource "b2_bucket_file_version" "test" {
  bucket_id = b2_bucket.test.id
  file_name = "temp.bin"
  source = "%s"
  content_type = "octet/stream"
  file_info = {
    description = "the file"
  }
  server_side_encryption {
    mode = "SSE-C"
    algorithm = "AES256"
	key {
	  secret_b64 = "CL31zAM5wJs8OZ6ORPakyLRYBqLDA8Z+AZ5tEBxqCKw=\n"
      key_id = "test_id"
    }
  }
}
`, bucketName, tempFile)
}
