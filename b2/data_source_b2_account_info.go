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
			"recommended_part_size": {
				Description: "The recommended number of bytes in a part of a large file.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"absolute_minimum_part_size": {
				Description: "The smallest possible size of a part of a large file (except the last one). This is smaller than the recommendedPartSize. If you use it, you may find that it takes longer overall to upload a large file.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func dataSourceB2AccountInfoRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := AccountInfoInput{}

	var output AccountInfoOutput
	err := client.Apply(ctx, OpDataSourceRead, &input, &output)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output.AccountId)

	err = client.Populate(ctx, OpDataSourceRead, &output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := dataSourceB2AccountInfoPopulateDeprecated(d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func dataSourceB2AccountInfoPopulateDeprecated(d *schema.ResourceData) error {
	if allowed, ok := d.GetOk("allowed"); ok {
		allowedList := allowed.([]interface{})
		if len(allowedList) > 0 {
			allowedMap := allowedList[0].(map[string]interface{})
			if buckets, ok := allowedMap["buckets"].([]interface{}); ok && len(buckets) > 0 {
				firstBucket := buckets[0].(map[string]interface{})
				if bucketId, ok := firstBucket["id"].(string); ok {
					allowedMap["bucket_id"] = bucketId
				}
				if bucketName, ok := firstBucket["name"].(string); ok {
					allowedMap["bucket_name"] = bucketName
				}
			} else {
				// Set empty strings if no buckets
				allowedMap["bucket_id"] = ""
				allowedMap["bucket_name"] = ""
			}
		}
	}
	return nil
}
