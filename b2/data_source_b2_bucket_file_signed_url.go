//####################################################################
//
// File: b2/data_source_b2_bucket_file_signed_url.go
//
// Copyright 2022 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceB2BucketFileSignedUrl() *schema.Resource {
	return &schema.Resource{
		Description: "B2 signed URL for a bucket file data source.",

		ReadContext: dataSourceB2BucketFileSignedUrlRead,

		Schema: map[string]*schema.Schema{
			"bucket_id": {
				Description:  "The ID of the bucket.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"file_name": {
				Description:  "The file name.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"duration": {
				Description: "The duration for which the presigned URL is valid",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"signed_url": {
				Description: "The signed URL for the given file",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceB2BucketFileSignedUrlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "bucket_file_signed_url"
	const op = DATA_SOURCE_READ

	input := map[string]interface{}{
		"bucket_id": d.Get("bucket_id").(string),
		"file_name": d.Get("file_name").(string),
		"duration":  d.Get("duration").(int),
	}

	output, err := client.apply(name, op, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output["signed_url"].(string))

	err = client.populate(name, op, output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
