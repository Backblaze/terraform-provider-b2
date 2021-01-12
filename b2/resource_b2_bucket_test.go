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
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceB2Bucket(t *testing.T) {
	resourceName := "b2_bucket.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp-")

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2BucketConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "bucket_name", bucketName),
					resource.TestCheckResourceAttr(resourceName, "bucket_type", "allPublic"),
					resource.TestCheckResourceAttr(resourceName, "bucket_info.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cors_rules.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rules.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0", "s3"),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
		},
	})
}

func testAccResourceB2BucketConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPublic"
}
`, bucketName)
}
