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
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceB2BucketFileVersion_basic(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_bucket_file_version.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	tempFile := createTempFileString(t, "hello")
	defer func() { _ = os.Remove(tempFile) }()

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
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.mode", "SSE-B2"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.algorithm", "AES256"),
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
	tempFile := createTempFileString(t, "hello")
	defer func() { _ = os.Remove(tempFile) }()

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
	tempFile := createTempFileString(t, "hello")
	defer func() { _ = os.Remove(tempFile) }()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2BucketFileVersionConfig_basic(bucketName, tempFile),
			},
			{
				Config: testAccResourceB2BucketFileVersionConfig_sse_c(bucketName, tempFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "action", "upload"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(resourceName, "content_md5", "5d41402abc4b2a76b9719d911017c592"),
					resource.TestCheckResourceAttr(resourceName, "content_sha1", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "octet/stream"),
					resource.TestCheckResourceAttr(resourceName, "file_info.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "file_info.description", "the file"),
					resource.TestCheckResourceAttr(resourceName, "file_info.sse_c_key_id", "test_id"),
					resource.TestCheckResourceAttr(resourceName, "file_name", "temp.bin"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.mode", "SSE-C"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.algorithm", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "source", tempFile),
					resource.TestCheckResourceAttr(resourceName, "size", "5"),
					resource.TestMatchResourceAttr(resourceName, "upload_timestamp", regexp.MustCompile("^[0-9]{13}$")),
				),
			},
		},
	})
}

func TestAccResourceB2BucketFileVersion_largeFile(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_bucket_file_version.test"
	var fileSize int64 = 105 * 1000 * 1000 // 105MB file is uploaded as a large file

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	tempFile := createTempFileTruncate(t, fileSize)
	defer func() { _ = os.Remove(tempFile) }()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2BucketFileVersionConfig_basic(bucketName, tempFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "action", "upload"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(resourceName, "content_md5", ""),      // empty for large files
					resource.TestCheckResourceAttr(resourceName, "content_sha1", "none"), // "none" for large files
					resource.TestCheckResourceAttr(resourceName, "content_type", "text/plain"),
					resource.TestCheckResourceAttr(resourceName, "file_info.%", "1"),
					resource.TestMatchResourceAttr(resourceName, "file_info.large_file_sha1", regexp.MustCompile("^[a-z0-9]{40}$")),
					resource.TestCheckResourceAttr(resourceName, "file_name", "temp.txt"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.mode", "SSE-B2"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.algorithm", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "size", strconv.Itoa(int(fileSize))),
					resource.TestCheckResourceAttr(resourceName, "source", tempFile),
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
	tempFile := createTempFileString(t, "hello")
	defer func() { _ = os.Remove(tempFile) }()

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
					resource.TestCheckResourceAttr(resourceName, "file_info.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "file_info.description", "the file"),
					resource.TestCheckResourceAttr(resourceName, "file_info.sse_c_key_id", "test_id"),
					resource.TestCheckResourceAttr(resourceName, "file_name", "temp.bin"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.mode", "SSE-C"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.algorithm", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "source", tempFile),
					resource.TestCheckResourceAttr(resourceName, "size", "5"),
					resource.TestMatchResourceAttr(resourceName, "upload_timestamp", regexp.MustCompile("^[0-9]{13}$")),
				),
			},
			{
				// The sse_c_key_id added by B2 must not be planned as a removal of a file info key
				Config:   testAccResourceB2BucketFileVersionConfig_sse_c(bucketName, tempFile),
				PlanOnly: true,
			},
		},
	})
}

func TestAccResourceB2BucketFileVersion_fileInfoKeyCase(t *testing.T) {
	resourceName := "b2_bucket_file_version.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	tempFile := createTempFileString(t, "hello")
	defer func() { _ = os.Remove(tempFile) }()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// B2 stores the key in lower case, which must not show up as a permanent diff
				Config: testAccResourceB2BucketFileVersionConfig_fileInfo(bucketName, tempFile,
					`ManagedBy = "Terraform"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "file_info.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "file_info.managedby", "Terraform"),
				),
			},
			{
				// The key stored in lower case must not be planned as a new file version
				Config: testAccResourceB2BucketFileVersionConfig_fileInfo(bucketName, tempFile,
					`ManagedBy = "Terraform"`),
				PlanOnly: true,
			},
			{
				// Changing the key case only is not a change for B2
				Config: testAccResourceB2BucketFileVersionConfig_fileInfo(bucketName, tempFile,
					`managedby = "Terraform"`),
				PlanOnly: true,
			},
			{
				// Adding a key requires a new file version, as B2 stores file info on upload only
				Config: testAccResourceB2BucketFileVersionConfig_fileInfo(bucketName, tempFile,
					`ManagedBy = "Terraform"
    description = "the file"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "file_info.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "file_info.managedby", "Terraform"),
					resource.TestCheckResourceAttr(resourceName, "file_info.description", "the file"),
				),
			},
			{
				// Removing a key requires a new file version too, so it must not be suppressed
				Config: testAccResourceB2BucketFileVersionConfig_fileInfo(bucketName, tempFile,
					`ManagedBy = "Terraform"`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
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

func testAccResourceB2BucketFileVersionConfig_fileInfo(bucketName string, tempFile string, fileInfo string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPublic"
}

resource "b2_bucket_file_version" "test" {
  bucket_id = b2_bucket.test.id
  file_name = "temp.txt"
  source = "%s"
  file_info = {
    %s
  }
}
`, bucketName, tempFile, fileInfo)
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
	  secret_b64 = "notarealkey11111111111111111111111111111111="
      key_id = "test_id"
    }
  }
}
`, bucketName, tempFile)
}

// When the file version no longer exists the bindings return None, which run() turns into an object
// that holds only _sha1. That must remove the resource from state instead of failing every refresh.
func TestResourceB2BucketFileVersionReadRemovesDeletedFileFromState(t *testing.T) {
	exec := stubBindings(t, `printf '{"_sha1":"da39a3ee5e6b4b0d3255bfef95601890afd80709"}'`)
	client := stubClient(t, exec)

	res := resourceB2BucketFileVersion()
	d := res.TestResourceData()
	d.SetId("4_zdeleted_out_of_band")

	if diags := resourceB2BucketFileVersionRead(context.Background(), d, client); diags.HasError() {
		t.Fatal(diags)
	}
	if d.Id() != "" {
		t.Errorf("a file version that no longer exists must be dropped from state, but the ID is still %q", d.Id())
	}
}
