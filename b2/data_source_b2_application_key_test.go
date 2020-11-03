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
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceB2ApplicationKey(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceB2ApplicationKey,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.b2_application_key.test", "key_name", "API"),
				),
			},
		},
	})
}

const testAccDataSourceB2ApplicationKey = `
data "b2_application_key" "test" {
  key_name = "API"
}
`
