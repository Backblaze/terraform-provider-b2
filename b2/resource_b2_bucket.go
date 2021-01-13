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

		Schema: map[string]*schema.Schema{
			"bucket_name": {
				Description:  "The name of the bucket.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"bucket_type": {
				Description:  "The bucket type.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
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
			"bucket_info": {
				Description: "The bucket info.",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"cors_rules": {
				Description: "CORS rules.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"lifecycle_rules": {
				Description: "Lifecycle rules.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
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

	input := map[string]interface{}{
		"bucket_name":     d.Get("bucket_name").(string),
		"bucket_type":     d.Get("bucket_type").(string),
		"bucket_info":     d.Get("bucket_info").(map[string]interface{}),
		"cors_rules":      d.Get("cors_rules").(*schema.Set).List(),
		"lifecycle_rules": d.Get("lifecycle_rules").(*schema.Set).List(),
	}

	output, err := client.apply("bucket", RESOURCE_CREATE, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output["bucket_id"].(string))

	d.Set("bucket_id", output["bucket_id"])
	d.Set("account_id", output["account_id"])
	d.Set("bucket_type", output["bucket_type"])
	d.Set("revision", output["revision"])

	if err := d.Set("bucket_info", output["bucket_info"]); err != nil {
		return diag.Errorf("error setting bucket_info: %s", err)
	}

	if err := d.Set("cors_rules", output["cors_rules"]); err != nil {
		return diag.Errorf("error setting cors_rules: %s", err)
	}

	if err := d.Set("lifecycle_rules", output["lifecycle_rules"]); err != nil {
		return diag.Errorf("error setting lifecycle_rules: %s", err)
	}

	if err := d.Set("options", output["options"]); err != nil {
		return diag.Errorf("error setting options: %s", err)
	}

	return nil
}

func resourceB2BucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := map[string]interface{}{
		"bucket_id": d.Id(),
	}

	output, err := client.apply("bucket", RESOURCE_READ, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("account_id", output["account_id"])
	d.Set("bucket_type", output["bucket_type"])
	d.Set("revision", output["revision"])

	if err := d.Set("bucket_info", output["bucket_info"]); err != nil {
		return diag.Errorf("error setting bucket_info: %s", err)
	}

	if err := d.Set("cors_rules", output["cors_rules"]); err != nil {
		return diag.Errorf("error setting cors_rules: %s", err)
	}

	if err := d.Set("lifecycle_rules", output["lifecycle_rules"]); err != nil {
		return diag.Errorf("error setting lifecycle_rules: %s", err)
	}

	if err := d.Set("options", output["options"]); err != nil {
		return diag.Errorf("error setting options: %s", err)
	}

	return nil
}

func resourceB2BucketUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := map[string]interface{}{
		"bucket_id":       d.Id(),
		"account_id":      d.Get("account_id").(string),
		"bucket_type":     d.Get("bucket_type").(string),
		"bucket_info":     d.Get("bucket_info").(map[string]interface{}),
		"cors_rules":      d.Get("cors_rules").(*schema.Set).List(),
		"lifecycle_rules": d.Get("lifecycle_rules").(*schema.Set).List(),
	}

	output, err := client.apply("bucket", RESOURCE_UPDATE, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("bucket_type", output["bucket_type"])

	if err := d.Set("bucket_info", output["bucket_info"]); err != nil {
		return diag.Errorf("error setting bucket_info: %s", err)
	}

	if err := d.Set("cors_rules", output["cors_rules"]); err != nil {
		return diag.Errorf("error setting cors_rules: %s", err)
	}

	if err := d.Set("lifecycle_rules", output["lifecycle_rules"]); err != nil {
		return diag.Errorf("error setting lifecycle_rules: %s", err)
	}

	if err := d.Set("options", output["options"]); err != nil {
		return diag.Errorf("error setting options: %s", err)
	}

	return nil
}

func resourceB2BucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := map[string]interface{}{
		"bucket_id": d.Id(),
	}

	_, err := client.apply("bucket", RESOURCE_DELETE, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
