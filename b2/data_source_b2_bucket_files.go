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
		Description: "B2 bucket files data source. The ID is `<bucket_id>/<folder_name>`.",

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
				Description: "Return all file versions when true; otherwise return the latest version of each file. The default is false.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"recursive": {
				Description: "List files in all subfolders when true; otherwise list the files directly in the folder and the immediate subfolder prefixes. The default is false.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"max_results": {
				Description:  "Maximum number of file versions and folder entries to return (1 to 100000). The read fails if more entries match; increase this value or narrow `folder_name`.",
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      10000,
				ValidateFunc: validation.IntBetween(1, 100000),
			},
			"file_versions": {
				Description: "File versions in the folder. Non-recursive listings also include immediate subfolder prefixes as entries with action 'folder', bucket_id, and file_name. Their other file-version fields are empty or zero.",
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
		MaxResults:   d.Get("max_results").(int),
	}

	var output BucketFilesOutput
	err := client.Apply(ctx, OpDataSourceRead, &input, &output)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(input.BucketId + "/" + input.FolderName)

	err = client.Populate(ctx, OpDataSourceRead, &output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
