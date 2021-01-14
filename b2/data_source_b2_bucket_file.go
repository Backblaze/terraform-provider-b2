//####################################################################
//
// File: b2/data_source_b2_bucket_file.go
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

func dataSourceB2BucketFile() *schema.Resource {
	return &schema.Resource{
		Description: "B2 bucket file data source.",

		ReadContext: dataSourceB2BucketFileRead,

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
			"show_versions": {
				Description: "Show all file versions.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"file_versions": {
				Description: "File versions.",
				Type:        schema.TypeList,
				Elem:        getFileVersionsElem(),
				Computed:    true,
			},
		},
	}
}

func dataSourceB2BucketFileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "bucket_file"
	const op = DATA_SOURCE_READ

	input := map[string]interface{}{
		"bucket_id":     d.Get("bucket_id").(string),
		"file_name":     d.Get("file_name").(string),
		"show_versions": d.Get("show_versions").(bool),
	}

	output, err := client.apply(name, op, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output["_sha1"].(string))

	err = client.populate(name, op, output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
