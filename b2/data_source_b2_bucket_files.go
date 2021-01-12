//####################################################################
//
// File: b2/data_source_b2_bucket_files.go
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

func dataSourceB2BucketFiles() *schema.Resource {
	return &schema.Resource{
		Description: "B2 bucket files data source.",

		ReadContext: dataSourceB2BucketFilesRead,

		Schema: map[string]*schema.Schema{
			"bucket_id": {
				Description:  "The ID of the bucket.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"folder_name": {
				Description: "The folder name (B2 file name prefix).",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"show_versions": {
				Description: "Show all file versions.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"recursive": {
				Description: "Recursive mode.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"file_versions": {
				Description: "File versions in the bucket.",
				Type:        schema.TypeList,
				Elem: &schema.Resource{
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
				},
				Computed: true,
			},
		},
	}
}

func dataSourceB2BucketFilesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := map[string]interface{}{
		"bucket_id":     d.Get("bucket_id").(string),
		"folder_name":   d.Get("folder_name").(string),
		"show_versions": d.Get("show_versions").(bool),
		"recursive":     d.Get("recursive").(bool),
	}

	output, err := client.apply("bucket_files", DATA_SOURCE_READ, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("file_versions", output["file_versions"])
	d.SetId(output["bucket_id"].(string))

	return nil
}
