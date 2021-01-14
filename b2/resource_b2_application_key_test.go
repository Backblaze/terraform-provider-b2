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

func TestAccResourceB2ApplicationKey(t *testing.T) {
	resourceName := "b2_application_key.test"

	keyName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2ApplicationKeyConfig(keyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "application_key", regexp.MustCompile("^[\x20-\x7E]{31}$")),
					resource.TestMatchResourceAttr(resourceName, "application_key_id", regexp.MustCompile("^[a-zA-Z0-9]{25}$")),
					resource.TestCheckResourceAttr(resourceName, "bucket_id", ""),
					resource.TestCheckResourceAttr(resourceName, "bucket_id", ""),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.0", "readFiles"),
					resource.TestCheckResourceAttr(resourceName, "key_name", keyName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0", "s3"),
				),
			},
		},
	})
}

func testAccResourceB2ApplicationKeyConfig(keyName string) string {
	return fmt.Sprintf(`
resource "b2_application_key" "test" {
  key_name = "%s"
  capabilities = ["readFiles"]
}
`, keyName)
}