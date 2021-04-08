//####################################################################
//
// File: b2/data_source_b2_account_info.go
//
// Copyright 2021 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceB2AccountInfo() *schema.Resource {
	return &schema.Resource{
		Description: "B2 account info data source.",

		ReadContext: dataSourceB2AccountInfoRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Description: "The identifier for the account.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"account_auth_token": {
				Description: "An authorization token to use with all calls, other than b2_authorize_account, that need an Authorization header. This authorization token is valid for at most 24 hours.",
				Type:        schema.TypeString,
				Sensitive:   true,
				Computed:    true,
			},
			"api_url": {
				Description: "The base URL to use for all API calls except for uploading and downloading files.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"allowed": {
				Description: "An object containing the capabilities of this auth token, and any restrictions on using it.",
				Type:        schema.TypeList,
				Elem:        getDataSourceAllowedElem(),
				Computed:    true,
			},
			"download_url": {
				Description: "The base URL to use for downloading files.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"s3_api_url": {
				Description: "The base URL to use for S3-compatible API calls.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceB2AccountInfoRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "account_info"
	const op = DATA_SOURCE_READ

	input := map[string]interface{}{}

	output, err := client.apply(name, op, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output["account_id"].(string))

	err = client.populate(name, op, output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
