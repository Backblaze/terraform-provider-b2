//####################################################################
//
// File: b2/data_source_b2_account_info_test.go
//
// Copyright 2020 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceB2AccountInfo_basic(t *testing.T) {
	dataSourceName := "data.b2_account_info.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceB2AccountInfoConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "account_id", regexp.MustCompile("^[a-z0-9]{12}$")),
					resource.TestMatchResourceAttr(dataSourceName, "account_auth_token", regexp.MustCompile("^[-=_a-zA-Z0-9]{77}$")),
					resource.TestMatchResourceAttr(dataSourceName, "api_url", regexp.MustCompile("https://api00[0-9].backblazeb2.com")),
					resource.TestCheckResourceAttr(dataSourceName, "allowed.#", "1"),
					resource.TestMatchResourceAttr(dataSourceName, "allowed.0.capabilities.#", regexp.MustCompile("[1-9][0-9]*")),
					resource.TestCheckResourceAttr(dataSourceName, "allowed.0.buckets.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "allowed.0.bucket_name", ""), // deprecated
					resource.TestCheckResourceAttr(dataSourceName, "allowed.0.bucket_id", ""),   // deprecated
					resource.TestMatchResourceAttr(dataSourceName, "download_url", regexp.MustCompile("https://f00[0-9].backblazeb2.com")),
					resource.TestMatchResourceAttr(dataSourceName, "s3_api_url", regexp.MustCompile("https://s3.(us-west|eu-central)-00[0-9].backblazeb2.com")),
					resource.TestMatchResourceAttr(dataSourceName, "recommended_part_size", regexp.MustCompile("^[1-9][0-9]*$")),
					resource.TestMatchResourceAttr(dataSourceName, "absolute_minimum_part_size", regexp.MustCompile("^[1-9][0-9]*$")),
				),
			},
		},
	})
}

func testAccDataSourceB2AccountInfoConfig_basic() string {
	return `
data "b2_account_info" "test" {
}
`
}
