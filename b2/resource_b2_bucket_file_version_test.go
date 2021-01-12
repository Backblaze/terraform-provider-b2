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
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceB2BucketFileVersion(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_bucket_file_version.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	tempFile := createTempFile(t, "hello")

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2BucketFileVersionConfig(bucketName, tempFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(resourceName, "file_name", "temp.txt"),
					resource.TestCheckResourceAttr(resourceName, "source", tempFile),
					resource.TestCheckResourceAttr(resourceName, "content_md5", "5d41402abc4b2a76b9719d911017c592"),
					resource.TestCheckResourceAttr(resourceName, "content_sha1", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "text/plain"),
					resource.TestCheckResourceAttr(resourceName, "size", "5"),
					resource.TestCheckResourceAttr(resourceName, "file_info.#", "0"),
				),
			},
		},
	})
}

func testAccResourceB2BucketFileVersionConfig(bucketName string, tempFile string) string {
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
