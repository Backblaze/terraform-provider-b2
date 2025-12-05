//####################################################################
//
// File: b2/resource_b2_application_key_test.go
//
// Copyright 2020 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceB2ApplicationKey_basic(t *testing.T) {
	resourceName := "b2_application_key.test"

	keyName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2ApplicationKeyConfig_basic(keyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "application_key", regexp.MustCompile("^[\x20-\x7E]{31}$")),
					resource.TestMatchResourceAttr(resourceName, "application_key_id", regexp.MustCompile("^[a-zA-Z0-9]{25}$")),
					resource.TestCheckResourceAttr(resourceName, "bucket_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "bucket_id", ""), // deprecated
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.0", "readFiles"),
					resource.TestCheckResourceAttr(resourceName, "expiration_timestamp", "0"),
					resource.TestCheckResourceAttr(resourceName, "key_name", keyName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0", "s3"),
					resource.TestCheckResourceAttr(resourceName, "valid_duration_in_seconds", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"application_key", "valid_duration_in_seconds"},
			},
		},
	})
}

func TestAccResourceB2ApplicationKey_all(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_application_key.test"

	keyName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2ApplicationKeyConfig_all(keyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "application_key", regexp.MustCompile("^[\x20-\x7E]{31}$")),
					resource.TestMatchResourceAttr(resourceName, "application_key_id", regexp.MustCompile("^[a-zA-Z0-9]{25}$")),
					resource.TestCheckResourceAttr(resourceName, "bucket_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_ids.0", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"), // deprecated
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.0", "writeFiles"),
					resource.TestMatchResourceAttr(resourceName, "expiration_timestamp", regexp.MustCompile("^[1-9][0-9]{12,}$")),
					resource.TestCheckResourceAttr(resourceName, "key_name", keyName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "prefix"),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0", "s3"),
					resource.TestCheckResourceAttr(resourceName, "valid_duration_in_seconds", "86400"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"application_key", "valid_duration_in_seconds"},
			},
		},
	})
}

func TestAccResourceB2ApplicationKey_deprecatedBucketId(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_application_key.test"

	keyName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2ApplicationKeyConfig_deprecatedBucketId(keyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "application_key", regexp.MustCompile("^[\x20-\x7E]{31}$")),
					resource.TestMatchResourceAttr(resourceName, "application_key_id", regexp.MustCompile("^[a-zA-Z0-9]{25}$")),
					resource.TestCheckResourceAttr(resourceName, "bucket_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_ids.0", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"), // deprecated
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.0", "writeFiles"),
					resource.TestMatchResourceAttr(resourceName, "expiration_timestamp", regexp.MustCompile("^[1-9][0-9]{12,}$")),
					resource.TestCheckResourceAttr(resourceName, "key_name", keyName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "prefix"),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0", "s3"),
					resource.TestCheckResourceAttr(resourceName, "valid_duration_in_seconds", "86400"),
				),
			},
			{
				Config: testAccResourceB2ApplicationKeyConfig_all(keyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "application_key", regexp.MustCompile("^[\x20-\x7E]{31}$")),
					resource.TestMatchResourceAttr(resourceName, "application_key_id", regexp.MustCompile("^[a-zA-Z0-9]{25}$")),
					resource.TestCheckResourceAttr(resourceName, "bucket_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_ids.0", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"), // deprecated
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.0", "writeFiles"),
					resource.TestMatchResourceAttr(resourceName, "expiration_timestamp", regexp.MustCompile("^[1-9][0-9]{12,}$")),
					resource.TestCheckResourceAttr(resourceName, "key_name", keyName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "prefix"),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0", "s3"),
					resource.TestCheckResourceAttr(resourceName, "valid_duration_in_seconds", "86400"),
				),
			},
		},
	})
}

func testAccResourceB2ApplicationKeyConfig_basic(keyName string) string {
	return fmt.Sprintf(`
resource "b2_application_key" "test" {
  key_name = "%s"
  capabilities = ["readFiles"]
}
`, keyName)
}

func testAccResourceB2ApplicationKeyConfig_all(keyName string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPrivate"
}

resource "b2_application_key" "test" {
  key_name = "%s"
  capabilities = ["writeFiles"]
  bucket_ids = [b2_bucket.test.bucket_id]
  name_prefix = "prefix"
  valid_duration_in_seconds = 86400
}
`, keyName, keyName)
}

func testAccResourceB2ApplicationKeyConfig_deprecatedBucketId(keyName string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPrivate"
}

resource "b2_application_key" "test" {
  key_name = "%s"
  capabilities = ["writeFiles"]
  bucket_id = b2_bucket.test.bucket_id
  name_prefix = "prefix"
  valid_duration_in_seconds = 86400
}
`, keyName, keyName)
}
