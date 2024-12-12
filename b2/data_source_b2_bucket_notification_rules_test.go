//####################################################################
//
// File: b2/data_source_b2_bucket_notification_rules_test.go
//
// Copyright 2024 Backblaze Inc. All Rights Reserved.
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

func TestAccDataSourceB2BucketNotificationRules_basic(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_bucket_notification_rules.test"
	dataSourceName := "data.b2_bucket_notification_rules.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	ruleName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceB2BucketNotificationRulesConfig_basic(bucketName, ruleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_id", resourceName, "bucket_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "notification_rules", resourceName, "notification_rules"),
				),
			},
		},
	})
}

func testAccDataSourceB2BucketNotificationRulesConfig_basic(bucketName string, ruleName string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPublic"
}

resource "b2_bucket_notification_rules" "test" {
  bucket_id = b2_bucket.test.id
  notification_rules {
    name        = "%s"
    event_types = ["b2:ObjectCreated:*"]
    target_configuration {
      target_type = "webhook"
      url         = "https://example.com/webhook"
    }
  }
}

data "b2_bucket_notification_rules" "test" {
  bucket_id = b2_bucket.test.bucket_id
  depends_on = [
    b2_bucket_notification_rules.test,
  ]
}
`, bucketName, ruleName)
}
