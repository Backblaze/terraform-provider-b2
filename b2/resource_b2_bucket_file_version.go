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
			"content_type": {
				Description: "Content type. If not set, it will be set based on the file extension.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return new == "" // The API sets default value
				},
			},
			"file_info": {
				Description: "The custom information that is uploaded with the file.",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return k == "file_info.sse_c_key_id" || old == new
				},
			},
			"server_side_encryption": {
				Description: "Server-side encryption settings.",
				Type:        schema.TypeList,
				Elem:        getResourceFileEncryption(),
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// The API sets default value
					if k == "server_side_encryption.#" {
						return old == "1" && new == "0"
					}
					return old == "none" && new == ""
				},
			},
			"action": {
				Description: "One of 'start', 'upload', 'hide', 'folder', or other values added in the future.",
				Type:        schema.TypeString,
				Computed:    true,
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
			"file_id": {
				Description: "The unique identifier for this version of this file.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"size": {
				Description: "The file size.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"upload_timestamp": {
				Description: "This is a UTC time when this file was uploaded.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func resourceB2BucketFileVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "bucket_file_version"
	const op = RESOURCE_CREATE

	input := map[string]interface{}{
		"bucket_id":              d.Get("bucket_id").(string),
		"file_name":              d.Get("file_name").(string),
		"source":                 d.Get("source").(string),
		"content_type":           d.Get("content_type").(string),
		"file_info":              d.Get("file_info").(map[string]interface{}),
		"server_side_encryption": d.Get("server_side_encryption").([]interface{}),
	}

	output, err := client.apply(ctx, name, op, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output["file_id"].(string))

	err = client.populate(ctx, name, op, output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceB2BucketFileVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "bucket_file_version"
	const op = RESOURCE_READ

	input := map[string]interface{}{
		"file_id": d.Id(),
	}

	output, err := client.apply(ctx, name, op, input)
	if err != nil {
		return diag.FromErr(err)
	}

	output["bucket_id"] = d.Get("bucket_id").(string)
	output["size"] = d.Get("size").(int)
	output["source"] = d.Get("source").(string)

	err = client.populate(ctx, name, op, output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceB2BucketFileVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "bucket_file_version"
	const op = RESOURCE_DELETE

	input := map[string]interface{}{
		"file_id":   d.Id(),
		"file_name": d.Get("file_name").(string),
	}

	_, err := client.apply(ctx, name, op, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
