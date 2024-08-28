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
	"log"

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
				Elem:        getResourceCorsRulesElem(),
				Optional:    true,
			},
			"file_lock_configuration": {
				Description: "File lock enabled flag, and default retention settings.",
				Type:        schema.TypeList,
				Elem:        getResourceFileLockConfiguration(),
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
				Elem:        getResourceDefaultBucketServerSideEncryption(),
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
				Elem:        getResourceLifecycleRulesElem(),
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
	const name = "bucket"
	const op = RESOURCE_CREATE

	input := map[string]interface{}{
		"bucket_name":                    d.Get("bucket_name").(string),
		"bucket_type":                    d.Get("bucket_type").(string),
		"bucket_info":                    d.Get("bucket_info").(map[string]interface{}),
		"cors_rules":                     d.Get("cors_rules").([]interface{}),
		"file_lock_configuration":        d.Get("file_lock_configuration").([]interface{}),
		"default_server_side_encryption": d.Get("default_server_side_encryption").([]interface{}),
		"lifecycle_rules":                d.Get("lifecycle_rules").([]interface{}),
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

func resourceB2BucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "bucket"
	const op = RESOURCE_READ

	input := map[string]interface{}{
		"bucket_id":  d.Id(),
		"cors_rules": d.Get("cors_rules"),
	}

	output, err := client.apply(name, op, input)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, ok := output["bucket_id"]; !ok && !d.IsNewResource() {
		// deleted bucket
		log.Printf("[WARN] Bucket (%s) not found, possible resource drift", d.Id())
		d.SetId("")
		return nil
	}

	err = client.populate(name, op, output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceB2BucketUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "bucket"
	const op = RESOURCE_UPDATE

	input := map[string]interface{}{
		"bucket_id":                      d.Id(),
		"account_id":                     d.Get("account_id").(string),
		"bucket_type":                    d.Get("bucket_type").(string),
		"bucket_info":                    d.Get("bucket_info").(map[string]interface{}),
		"cors_rules":                     d.Get("cors_rules").([]interface{}),
		"file_lock_configuration":        d.Get("file_lock_configuration").([]interface{}),
		"default_server_side_encryption": d.Get("default_server_side_encryption").([]interface{}),
		"lifecycle_rules":                d.Get("lifecycle_rules").([]interface{}),
	}

	output, err := client.apply(name, op, input)
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.populate(name, op, output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceB2BucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "bucket"
	const op = RESOURCE_DELETE

	input := map[string]interface{}{
		"bucket_id": d.Id(),
	}

	_, err := client.apply(name, op, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
