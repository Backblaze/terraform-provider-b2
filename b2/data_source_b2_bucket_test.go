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

func TestAccDataSourceB2Bucket(t *testing.T) {
	resourceName := "b2_bucket.test"
	dataSourceName := "data.b2_bucket.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceB2BucketConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_id", resourceName, "bucket_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_type", resourceName, "bucket_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "account_id", resourceName, "account_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_info", resourceName, "bucket_info"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_rules", resourceName, "cors_rules"),
					resource.TestCheckResourceAttrPair(dataSourceName, "lifecycle_rules", resourceName, "lifecycle_rules"),
					resource.TestCheckResourceAttrPair(dataSourceName, "options", resourceName, "options"),
					resource.TestCheckResourceAttrPair(dataSourceName, "revision", resourceName, "revision"),
					resource.TestCheckResourceAttr(dataSourceName, "bucket_name", bucketName),
					resource.TestCheckResourceAttr(dataSourceName, "bucket_type", "allPublic"),
				),
			},
		},
	})
}

func testAccDataSourceB2BucketConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPublic"
}

data "b2_bucket" "test" {
  bucket_name = b2_bucket.test.bucket_name
}
`, bucketName)
}
