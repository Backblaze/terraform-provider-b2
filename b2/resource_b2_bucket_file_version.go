//####################################################################
//
// File: b2/resource_b2_bucket_file_version.go
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

func resourceB2BucketFileVersion() *schema.Resource {
	return &schema.Resource{
		Description: "B2 bucket file version resource.",

		CreateContext: resourceB2BucketFileVersionCreate,
		ReadContext:   resourceB2BucketFileVersionRead,
		DeleteContext: resourceB2BucketFileVersionDelete,

		Schema: map[string]*schema.Schema{
			"bucket_id": {
				Description:  "The ID of the bucket.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"file_name": {
				Description:  "The name of the B2 file.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"source": {
				Description:  "Path to the local file.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"content_md5": {
				Description: "MD5 sum of the content.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"content_sha1": {
				Description: "SHA1 hash of the content.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"content_type": {
				Description: "Content type. If not set, it will be set based on the file extension.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "" {
						return true // The API sets default value
					}
					return false
				},
			},
			"file_id": {
				Description: "The file ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"file_info": {
				Description: "Additional file info.",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: true,
			},
			"size": {
				Description: "File size.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"upload_timestamp": {
				Description: "Upload timestamp.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func resourceB2BucketFileVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := map[string]interface{}{
		"bucket_id":    d.Get("bucket_id").(string),
		"file_name":    d.Get("file_name").(string),
		"source":       d.Get("source").(string),
		"content_type": d.Get("content_type").(string),
		"file_info":    d.Get("file_info").(map[string]interface{}),
	}

	output, err := client.apply("bucket_file_version", RESOURCE_CREATE, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output["file_id"].(string))

	d.Set("file_id", output["file_id"])
	d.Set("content_md5", output["content_md5"])
	d.Set("content_sha1", output["content_sha1"])
	d.Set("content_type", output["content_type"])
	d.Set("size", output["size"])
	d.Set("upload_timestamp", output["upload_timestamp"])

	if err := d.Set("file_info", output["file_info"]); err != nil {
		return diag.Errorf("error setting file_info: %s", err)
	}

	return nil
}

func resourceB2BucketFileVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := map[string]interface{}{
		"file_id": d.Id(),
	}

	output, err := client.apply("bucket_file_version", RESOURCE_READ, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("content_md5", output["content_md5"])
	d.Set("content_sha1", output["content_sha1"])
	d.Set("content_type", output["content_type"])
	d.Set("size", output["size"])
	d.Set("upload_timestamp", output["upload_timestamp"])

	if err := d.Set("file_info", output["file_info"]); err != nil {
		return diag.Errorf("error setting file_info: %s", err)
	}

	return nil
}

func resourceB2BucketFileVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := map[string]interface{}{
		"file_id":   d.Id(),
		"file_name": d.Get("file_name").(string),
	}

	_, err := client.apply("bucket_file_version", RESOURCE_DELETE, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
