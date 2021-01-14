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

func TestAccDataSourceB2BucketFilesSingleFile(t *testing.T) {
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
				Config: testAccDataSourceB2BucketFilesConfigSingleFile(bucketName, tempFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(dataSourceName, "file_versions.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_id", resourceName, "file_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_name", resourceName, "file_name"),
					resource.TestCheckResourceAttr(dataSourceName, "file_versions.0.action", "upload"),
				),
			},
		},
	})
}

func testAccDataSourceB2BucketFilesConfigSingleFile(bucketName string, tempFile string) string {
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

func TestAccDataSourceB2BucketFilesMultipleFilesWithoutVersions(t *testing.T) {
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
				Config: testAccDataSourceB2BucketFilesConfigMultipleFiles(bucketName, tempFile, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(dataSourceName, "file_versions.#", "2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_id", resource2Name, "file_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_name", resource2Name, "file_name"),
					resource.TestCheckResourceAttr(dataSourceName, "file_versions.0.action", "upload"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_info", resource2Name, "file_info"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.file_id", resource3Name, "file_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.file_name", resource3Name, "file_name"),
					resource.TestCheckResourceAttr(dataSourceName, "file_versions.1.action", "upload"),
				),
			},
		},
	})
}

func TestAccDataSourceB2BucketFilesMultipleFilesWithVersions(t *testing.T) {
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
				Config: testAccDataSourceB2BucketFilesConfigMultipleFiles(bucketName, tempFile, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(dataSourceName, "file_versions.#", "3"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_id", resource2Name, "file_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_name", resource2Name, "file_name"),
					resource.TestCheckResourceAttr(dataSourceName, "file_versions.0.action", "upload"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.0.file_info", resource2Name, "file_info"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.file_id", resource1Name, "file_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.1.file_name", resource1Name, "file_name"),
					resource.TestCheckResourceAttr(dataSourceName, "file_versions.1.action", "upload"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.2.file_id", resource3Name, "file_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_versions.2.file_name", resource3Name, "file_name"),
					resource.TestCheckResourceAttr(dataSourceName, "file_versions.2.action", "upload"),
				),
			},
		},
	})
}

func testAccDataSourceB2BucketFilesConfigMultipleFiles(bucketName string, tempFile string, showVersions string) string {
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
