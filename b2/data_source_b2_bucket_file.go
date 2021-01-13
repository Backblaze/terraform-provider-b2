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

func getFileVersionsElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"action": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_md5": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_sha1": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_info": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"file_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"upload_timestamp": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceB2BucketFileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := map[string]interface{}{
		"bucket_id":     d.Get("bucket_id").(string),
		"file_name":     d.Get("file_name").(string),
		"show_versions": d.Get("show_versions").(bool),
	}

	output, err := client.apply("bucket_file", DATA_SOURCE_READ, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output["_sha1"].(string))

	if err := d.Set("file_versions", output["file_versions"]); err != nil {
		return diag.Errorf("error setting file_versions: %s", err)
	}

	return nil
}
