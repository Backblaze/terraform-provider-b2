//####################################################################
//
// File: b2/data_source_b2_application_key_test.go
//
// Copyright 2020 Backblaze Inc. All Rights Reserved.
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

func TestAccDataSourceB2ApplicationKey_basic(t *testing.T) {
	resourceName := "b2_application_key.test"
	dataSourceName := "data.b2_application_key.test"

	keyName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceB2ApplicationKeyConfig_basic(keyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "application_key_id", resourceName, "application_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_id", resourceName, "bucket_id"),
					resource.TestCheckResourceAttr(dataSourceName, "capabilities.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "capabilities.0", "readFiles"),
					resource.TestCheckResourceAttrPair(dataSourceName, "capabilities", resourceName, "capabilities"),
					resource.TestCheckResourceAttr(dataSourceName, "key_name", keyName),
					resource.TestCheckResourceAttrPair(dataSourceName, "key_name", resourceName, "key_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name_prefix", resourceName, "name_prefix"),
					resource.TestCheckResourceAttrPair(dataSourceName, "options", resourceName, "options"),
				),
			},
		},
	})
}

func TestAccDataSourceB2ApplicationKey_all(t *testing.T) {
	resourceName := "b2_application_key.test"
	dataSourceName := "data.b2_application_key.test"

	keyName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceB2ApplicationKeyConfig_all(keyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "application_key_id", resourceName, "application_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_id", resourceName, "bucket_id"),
					resource.TestCheckResourceAttr(dataSourceName, "capabilities.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "capabilities.0", "writeFiles"),
					resource.TestCheckResourceAttrPair(dataSourceName, "capabilities", resourceName, "capabilities"),
					resource.TestCheckResourceAttr(dataSourceName, "key_name", keyName),
					resource.TestCheckResourceAttrPair(dataSourceName, "key_name", resourceName, "key_name"),
					resource.TestCheckResourceAttr(dataSourceName, "name_prefix", "prefix"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name_prefix", resourceName, "name_prefix"),
					resource.TestCheckResourceAttrPair(dataSourceName, "options", resourceName, "options"),
				),
			},
		},
	})
}

func testAccDataSourceB2ApplicationKeyConfig_basic(keyName string) string {
	return fmt.Sprintf(`
resource "b2_application_key" "test" {
  key_name = "%s"
  capabilities = ["readFiles"]
}

data "b2_application_key" "test" {
  key_name = b2_application_key.test.key_name

  depends_on = [
    b2_application_key.test,
  ]
}
`, keyName)
}

func testAccDataSourceB2ApplicationKeyConfig_all(keyName string) string {
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
}

data "b2_application_key" "test" {
  key_name = b2_application_key.test.key_name

  depends_on = [
    b2_application_key.test,
  ]
}
`, keyName, keyName)
}
