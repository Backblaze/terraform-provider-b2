//####################################################################
//
// File: b2/resource_b2_bucket_notification_rules_test.go
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

func TestAccResourceB2BucketNotificationRules_basic(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_bucket_notification_rules.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	ruleName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2BucketNotificationRulesConfig_basic(bucketName, ruleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.event_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.event_types.0", "b2:ObjectCreated:*"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.object_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.target_type", "webhook"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.url", "https://example.com/webhook"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.hmac_sha256_signing_secret", ""),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.custom_headers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.is_suspended", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.suspension_reason", ""),
				),
			},
		},
	})
}

func TestAccResourceB2BucketNotificationRules_all(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_bucket_notification_rules.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	ruleName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2BucketNotificationRulesConfig_all(bucketName, ruleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.event_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.event_types.0", "b2:ObjectCreated:*"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.event_types.1", "b2:ObjectDeleted:*"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.is_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.object_name_prefix", "prefix/"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.target_type", "webhook"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.url", "https://example.com/webhook"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.hmac_sha256_signing_secret", "sWhtNHMFntMPukWYacpMmJbrhsuylxTg"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.custom_headers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.custom_headers.0.name", "myCustomHeader1"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.custom_headers.0.value", "myCustomHeaderVal1"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.custom_headers.1.name", "myCustomHeader2"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.custom_headers.1.value", "myCustomHeaderVal2"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.is_suspended", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.suspension_reason", ""),
				),
			},
		},
	})
}

func TestAccResourceB2BucketNotificationRules_update(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_bucket_notification_rules.test"

	bucketName := acctest.RandomWithPrefix("test-b2-tfp")
	ruleName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2BucketNotificationRulesConfig_basic(bucketName, ruleName),
			},
			{
				Config: testAccResourceB2BucketNotificationRulesConfig_all(bucketName, ruleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.event_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.event_types.0", "b2:ObjectCreated:*"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.event_types.1", "b2:ObjectDeleted:*"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.is_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.object_name_prefix", "prefix/"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.target_type", "webhook"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.url", "https://example.com/webhook"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.hmac_sha256_signing_secret", "sWhtNHMFntMPukWYacpMmJbrhsuylxTg"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.custom_headers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.custom_headers.0.name", "myCustomHeader1"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.custom_headers.0.value", "myCustomHeaderVal1"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.custom_headers.1.name", "myCustomHeader2"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.target_configuration.0.custom_headers.1.value", "myCustomHeaderVal2"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.is_suspended", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_rules.0.suspension_reason", ""),
				),
			},
		},
	})
}

func testAccResourceB2BucketNotificationRulesConfig_basic(bucketName string, ruleName string) string {
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
`, bucketName, ruleName)
}

func testAccResourceB2BucketNotificationRulesConfig_all(bucketName string, ruleName string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPublic"
}

resource "b2_bucket_notification_rules" "test" {
  bucket_id = b2_bucket.test.id
  notification_rules {
    name               = "%s"
    event_types        = ["b2:ObjectCreated:*", "b2:ObjectDeleted:*"]
    is_enabled         = false
    object_name_prefix = "prefix/"
    target_configuration {
      target_type                = "webhook"
      url                        = "https://example.com/webhook"
      hmac_sha256_signing_secret = "sWhtNHMFntMPukWYacpMmJbrhsuylxTg"
      custom_headers {
        name  = "myCustomHeader1"
        value = "myCustomHeaderVal1"
      }
      custom_headers {  # optional
        name  = "myCustomHeader2"
        value = "myCustomHeaderVal2"
      }
    }
  }
}
`, bucketName, ruleName)
}
