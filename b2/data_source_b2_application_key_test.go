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

func TestAccDataSourceB2ApplicationKey(t *testing.T) {
	resourceName := "b2_application_key.test"
	dataSourceName := "data.b2_application_key.test"

	keyName := fmt.Sprintf("test-datasource-b2-application-key-%d", acctest.RandInt())

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceB2ApplicationKeyConfig(keyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "key_name", keyName),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.0", "readFiles"),
				),
			},
		},
	})
}

func testAccDataSourceB2ApplicationKeyConfig(keyName string) string {
	return fmt.Sprintf(`
resource "b2_application_key" "test" {
  key_name = "%s"
  capabilities = ["readFiles"]
}

data "b2_application_key" "test" {
  key_name = b2_application_key.test.key_name
}
`, keyName)
}
