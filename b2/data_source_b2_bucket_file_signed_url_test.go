//####################################################################
//
// File: b2/data_source_b2_bucket_file_signed_url_test.go
//
// Copyright 2022 Backblaze Inc. All Rights Reserved.
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

func TestAccDataSourceB2BucketFileSignedUrl_singleFile(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_bucket_file_version.test"
	dataSourceName := "data.b2_bucket_file_signed_url.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	tempFile := createTempFile(t, "hello")
	defer os.Remove(tempFile)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceB2BucketFileSignedUrlConfig_singleFile(bucketName, tempFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_name", resourceName, "file_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "signed_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", dataSourceName, "signed_url"),
				),
			},
		},
	})
}

func testAccDataSourceB2BucketFileSignedUrlConfig_singleFile(bucketName string, tempFile string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPrivate"
}

resource "b2_bucket_file_version" "test" {
  bucket_id = b2_bucket.test.id
  file_name = "temp.txt"
  source = "%s"
}

data "b2_bucket_file_signed_url" "test" {
  bucket_id = b2_bucket_file_version.test.bucket_id
  file_name = b2_bucket_file_version.test.file_name
  duration = 3600
}
`, bucketName, tempFile)
}
