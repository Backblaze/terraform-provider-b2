//####################################################################
//
// File: b2/data_source_b2_bucket_files_test.go
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
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceB2BucketFiles_noFiles(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	dataSourceName := "data.b2_bucket_files.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	tempFile := createTempFile(t, "hello")
	defer os.Remove(tempFile)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceB2BucketFilesConfig_noFiles(bucketName, tempFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(dataSourceName, "file_versions.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "folder_name", "non_existing_folder"),
				),
			},
		},
	})
}

func TestAccDataSourceB2BucketFiles_singleFile(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_bucket_file_version.test"
	dataSourceName := "data.b2_bucket_files.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	tempFile := createTempFile(t, "hello")
	defer os.Remove(tempFile)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceB2BucketFilesConfig_singleFile(bucketName, tempFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(dataSourceName, "file_versions.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.action", resourceName, "action"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.content_md5", resourceName, "content_md5"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.content_sha1", resourceName, "content_sha1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.action", resourceName, "action"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_id", resourceName, "file_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_info", resourceName, "file_info"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_name", resourceName, "file_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.size", resourceName, "size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.upload_timestamp", resourceName, "upload_timestamp"),
					resource.TestCheckResourceAttr(dataSourceName, "folder_name", ""),
				),
			},
		},
	})
}

func TestAccDataSourceB2BucketFiles_multipleFilesWithoutVersions(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resource2Name := "b2_bucket_file_version.test2"
	resource3Name := "b2_bucket_file_version.test3"
	dataSourceName := "data.b2_bucket_files.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	tempFile := createTempFile(t, "hello")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceB2BucketFilesConfig_multipleFiles(bucketName, tempFile, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(dataSourceName, "file_versions.#", "2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.action", resource2Name, "action"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.content_md5", resource2Name, "content_md5"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.content_sha1", resource2Name, "content_sha1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.content_type", resource2Name, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.action", resource2Name, "action"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_id", resource2Name, "file_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_info", resource2Name, "file_info"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_name", resource2Name, "file_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.size", resource2Name, "size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.upload_timestamp", resource2Name, "upload_timestamp"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.action", resource3Name, "action"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.content_md5", resource3Name, "content_md5"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.content_sha1", resource3Name, "content_sha1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.content_type", resource3Name, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.action", resource3Name, "action"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.file_id", resource3Name, "file_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.file_info", resource3Name, "file_info"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.file_name", resource3Name, "file_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.size", resource3Name, "size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.upload_timestamp", resource3Name, "upload_timestamp"),
					resource.TestCheckResourceAttr(dataSourceName, "folder_name", ""),
				),
			},
		},
	})
}

func TestAccDataSourceB2BucketFiles_multipleFilesWithVersions(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resource1Name := "b2_bucket_file_version.test1"
	resource2Name := "b2_bucket_file_version.test2"
	resource3Name := "b2_bucket_file_version.test3"
	dataSourceName := "data.b2_bucket_files.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	tempFile := createTempFile(t, "hello")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceB2BucketFilesConfig_multipleFiles(bucketName, tempFile, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(dataSourceName, "file_versions.#", "3"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.action", resource2Name, "action"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.content_md5", resource2Name, "content_md5"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.content_sha1", resource2Name, "content_sha1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.content_type", resource2Name, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.action", resource2Name, "action"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_id", resource2Name, "file_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_info", resource2Name, "file_info"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_name", resource2Name, "file_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.size", resource2Name, "size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.upload_timestamp", resource2Name, "upload_timestamp"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.action", resource1Name, "action"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.content_md5", resource1Name, "content_md5"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.content_sha1", resource1Name, "content_sha1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.content_type", resource1Name, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.action", resource1Name, "action"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.file_id", resource1Name, "file_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.file_info", resource1Name, "file_info"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.file_name", resource1Name, "file_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.size", resource1Name, "size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.upload_timestamp", resource1Name, "upload_timestamp"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.2.action", resource3Name, "action"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.2.content_md5", resource3Name, "content_md5"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.2.content_sha1", resource3Name, "content_sha1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.2.content_type", resource3Name, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.2.action", resource3Name, "action"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.2.file_id", resource3Name, "file_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.2.file_info", resource3Name, "file_info"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.2.file_name", resource3Name, "file_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.2.size", resource3Name, "size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.2.upload_timestamp", resource3Name, "upload_timestamp"),
					resource.TestCheckResourceAttr(dataSourceName, "folder_name", ""),
				),
			},
		},
	})
}

func testAccDataSourceB2BucketFilesConfig_noFiles(bucketName string,  tempFile string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPublic"
}

resource "b2_bucket_file_version" "test" {
  bucket_id = b2_bucket.test.id
  file_name = "existing_folder/temp.txt"
  source = "%s"
}

data "b2_bucket_files" "test" {
  bucket_id = b2_bucket_file_version.test.bucket_id
  folder_name = "non_existing_folder"
}
`, bucketName, tempFile)
}

func testAccDataSourceB2BucketFilesConfig_singleFile(bucketName string, tempFile string) string {
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

data "b2_bucket_files" "test" {
  bucket_id = b2_bucket_file_version.test.bucket_id
}
`, bucketName, tempFile)
}

func testAccDataSourceB2BucketFilesConfig_multipleFiles(bucketName string, tempFile string, showVersions string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPublic"
}

resource "b2_bucket_file_version" "test1" {
  bucket_id = b2_bucket.test.id
  file_name = "temp1.txt"
  source = "%s"
}

resource "b2_bucket_file_version" "test2" {
  bucket_id = b2_bucket_file_version.test1.bucket_id
  file_name = b2_bucket_file_version.test1.file_name
  source = b2_bucket_file_version.test1.source
  file_info = {
    description = "second version"
  }

  depends_on = [
    b2_bucket_file_version.test1,
  ]
}

resource "b2_bucket_file_version" "test3" {
  bucket_id = b2_bucket_file_version.test2.bucket_id
  file_name = "temp2.txt"
  source = b2_bucket_file_version.test2.source

   depends_on = [
    b2_bucket_file_version.test2,
  ]
}

data "b2_bucket_files" "test" {
  bucket_id = b2_bucket_file_version.test3.bucket_id
  show_versions = %s
}
`, bucketName, tempFile, showVersions)
}
