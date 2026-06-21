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
				Description: "File versions in the folder.",
				Type:        schema.TypeList,
				Elem:        getDataSourceFileVersionsElem(),
				Computed:    true,
			},
		},
	}
}

func dataSourceB2BucketFilesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := BucketFilesInput{
		BucketId:     d.Get("bucket_id").(string),
		FolderName:   d.Get("folder_name").(string),
		ShowVersions: d.Get("show_versions").(bool),
		Recursive:    d.Get("recursive").(bool),
	}

	var output BucketFilesOutput
	err := client.Apply(ctx, OpDataSourceRead, &input, &output)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output.Sha1)

	err = client.Populate(ctx, OpDataSourceRead, &output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
