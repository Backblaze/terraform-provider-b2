//####################################################################
//
// File: b2/resource_b2_bucket.go
//
// Copyright 2021 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceB2Bucket() *schema.Resource {
	return &schema.Resource{
		Description: "B2 bucket resource.",

		CreateContext: resourceB2BucketCreate,
		ReadContext:   resourceB2BucketRead,
		UpdateContext: resourceB2BucketUpdate,
		DeleteContext: resourceB2BucketDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bucket_name": {
				Description:  "The name of the bucket.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"bucket_type": {
				Description:  "The bucket type. Either 'allPublic', meaning that files in this bucket can be downloaded by anybody, or 'allPrivate'.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"bucket_info": {
				Description: "User-defined information to be stored with the bucket.",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"cors_rules": {
				Description: "The initial list of CORS rules for this bucket.",
				Type:        schema.TypeList,
				Elem:        getCorsRulesElem(false),
				Optional:    true,
			},
			"file_lock_configuration": {
				Description: "File lock enabled flag, and default retention settings.",
				Type:        schema.TypeList,
				Elem:        getFileLockConfigurationElem(false),
				Optional:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// The API sets default value
					if k == "file_lock_configuration.#" {
						return old == "1" && new == "0"
					}
					return old == "none" && new == ""
				},
			},
			"default_server_side_encryption": {
				Description: "The default server-side encryption settings for this bucket.",
				Type:        schema.TypeList,
				Elem:        getServerSideEncryptionElem(false),
				Optional:    true,
				MaxItems:    1,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// The API sets default value
					if k == "default_server_side_encryption.#" {
						return old == "1" && new == "0"
					}
					return old == "none" && new == ""
				},
			},
			"lifecycle_rules": {
				Description: "The initial list of lifecycle rules for this bucket.",
				Type:        schema.TypeList,
				Elem:        getLifecycleRulesElem(false),
				Optional:    true,
			},
			"bucket_id": {
				Description: "The ID of the bucket.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"account_id": {
				Description: "Account ID that the bucket belongs to.",
				Type:        schema.TypeString,
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

func resourceB2BucketCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := BucketInput{
		BucketName:                  d.Get("bucket_name").(string),
		BucketType:                  d.Get("bucket_type").(string),
		BucketInfo:                  d.Get("bucket_info").(map[string]interface{}),
		CorsRules:                   d.Get("cors_rules").([]interface{}),
		FileLockConfiguration:       d.Get("file_lock_configuration").([]interface{}),
		DefaultServerSideEncryption: d.Get("default_server_side_encryption").([]interface{}),
		LifecycleRules:              d.Get("lifecycle_rules").([]interface{}),
	}

	var output BucketOutput
	err := client.Apply(ctx, OpResourceCreate, &input, &output)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output.BucketId)

	err = client.Populate(ctx, OpResourceCreate, &output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceB2BucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := BucketInput{
		BucketId: d.Id(),
	}

	var output BucketOutput
	err := client.Apply(ctx, OpResourceRead, &input, &output)
	if err != nil {
		return diag.FromErr(err)
	}
	if output.BucketId == "" && !d.IsNewResource() {
		// deleted bucket
		tflog.Warn(ctx, "Bucket not found, possible resource drift", map[string]interface{}{
			"bucket_id": d.Id(),
		})
		d.SetId("")
		return nil
	}

	err = client.Populate(ctx, OpResourceRead, &output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceB2BucketUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := BucketInput{
		BucketId:                    d.Id(),
		AccountId:                   d.Get("account_id").(string),
		BucketType:                  d.Get("bucket_type").(string),
		BucketInfo:                  d.Get("bucket_info").(map[string]interface{}),
		CorsRules:                   d.Get("cors_rules").([]interface{}),
		FileLockConfiguration:       d.Get("file_lock_configuration").([]interface{}),
		DefaultServerSideEncryption: d.Get("default_server_side_encryption").([]interface{}),
		LifecycleRules:              d.Get("lifecycle_rules").([]interface{}),
	}

	var output BucketOutput
	err := client.Apply(ctx, OpResourceUpdate, &input, &output)
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.Populate(ctx, OpResourceUpdate, &output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceB2BucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := BucketInput{
		BucketId: d.Id(),
	}

	err := client.Apply(ctx, OpResourceDelete, &input, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
