//####################################################################
//
// File: b2/data_source_b2_bucket.go
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceB2Bucket() *schema.Resource {
	return &schema.Resource{
		Description: "B2 bucket data source.",

		ReadContext: dataSourceB2BucketRead,

		Schema: map[string]*schema.Schema{
			"bucket_name": {
				Description:  "The name of the bucket.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"account_id": {
				Description: "Account ID that the bucket belongs to.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"bucket_id": {
				Description: "The ID of the bucket.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"bucket_info": {
				Description: "User-defined information to be stored with the bucket.",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"bucket_type": {
				Description: "The bucket type. Either 'allPublic', meaning that files in this bucket can be downloaded by anybody, or 'allPrivate'.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cors_rules": {
				Description: "The initial list of CORS rules for this bucket.",
				Type:        schema.TypeList,
				Elem:        getDataSourceCorsRulesElem(),
				Computed:    true,
			},
			"lifecycle_rules": {
				Description: "The initial list of lifecycle rules for this bucket.",
				Type:        schema.TypeList,
				Elem:        getDataSourceLifecycleRulesElem(),
				Computed:    true,
			},
			"options": {
				Description: "List of bucket options.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"revision": {
				Description: "Bucket revision.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func dataSourceB2BucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "bucket"
	const op = DATA_SOURCE_READ

	input := map[string]interface{}{
		"bucket_name": d.Get("bucket_name").(string),
	}

	output, err := client.apply(name, op, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output["bucket_id"].(string))

	err = client.populate(name, op, output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
