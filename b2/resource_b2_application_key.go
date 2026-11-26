//####################################################################
//
// File: b2/resource_b2_application_key.go
//
// Copyright 2020 Backblaze Inc. All Rights Reserved.
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

func resourceB2ApplicationKey() *schema.Resource {
	return &schema.Resource{
		Description: "B2 application key resource.",

		CreateContext: resourceB2ApplicationKeyCreate,
		ReadContext:   resourceB2ApplicationKeyRead,
		DeleteContext: resourceB2ApplicationKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"capabilities": {
				Description: "The key capabilities. See https://www.backblaze.com/docs/cloud-storage-application-key-capabilities for valid names.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				ForceNew: true,
			},
			"key_name": {
				Description:  "The name of the key.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"bucket_ids": {
				Description: "When provided, the new key can only access the specified buckets.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"bucket_id"},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Only hide the mirrored set when it represents the configured legacy ID.
					config := d.GetRawConfig()
					if config.IsNull() || !config.IsKnown() || !config.Type().HasAttribute("bucket_id") {
						return false
					}
					id := config.GetAttr("bucket_id")
					if id.IsNull() || !id.IsKnown() || id.AsString() == "" {
						return false
					}
					previous, _ := d.GetChange("bucket_ids")
					buckets := previous.(*schema.Set)
					return buckets.Len() == 1 && buckets.Contains(id.AsString())
				},
			},
			"name_prefix": {
				Description: "When present, restricts access to files whose names start with the prefix.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"application_key": {
				Description: "The secret key, returned only when created and stored in Terraform state. It is empty after import.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"application_key_id": {
				Description: "The ID of the key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expiration_timestamp": {
				Description: "When present, says when this key will expire, in milliseconds since 1970.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"options": {
				Description: "The application key options returned by B2.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"valid_duration_in_seconds": {
				Description:  "The key lifetime in seconds, from 1 to 86400000. When set, `expiration_timestamp` gives the expiry time.",
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 86400000),
			},
			"bucket_id": {
				Description:   "When present, restricts access to one bucket.",
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"bucket_ids"},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// The deprecated alias is populated from the first allowed bucket on read.
					// Consult configuration so a stale alias cannot hide a restriction edit.
					config := d.GetRawConfig()
					if new != "" || old == "" || config.IsNull() || !config.IsKnown() || !config.Type().HasAttribute("bucket_ids") {
						return false
					}
					ids := config.GetAttr("bucket_ids")
					if ids.IsNull() || !ids.IsWhollyKnown() {
						return false
					}
					for it := ids.ElementIterator(); it.Next(); {
						_, id := it.Element()
						if !id.IsNull() && id.AsString() == old {
							return true
						}
					}
					return false
				},
				Deprecated: "This argument is deprecated in favor of 'bucket_ids' argument",
			},
		},
	}
}

func resourceB2ApplicationKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := ApplicationKeyInput{
		KeyName:                d.Get("key_name").(string),
		Capabilities:           d.Get("capabilities").(*schema.Set).List(),
		NamePrefix:             d.Get("name_prefix").(string),
		BucketIds:              d.Get("bucket_ids").(*schema.Set).List(),
		BucketId:               d.Get("bucket_id").(string), // deprecated
		ValidDurationInSeconds: d.Get("valid_duration_in_seconds").(int),
	}

	// Deprecated bucket_id requires the B2 Native API v2 path
	if input.BucketId != "" {
		input.Apiver = "v2"
	}

	var output ApplicationKeyOutput
	err := client.Apply(ctx, OpResourceCreate, &input, &output)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output.ApplicationKeyId)

	err = client.Populate(ctx, OpResourceCreate, &output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Preserve valid_duration_in_seconds in state
	if err := d.Set("valid_duration_in_seconds", input.ValidDurationInSeconds); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceB2ApplicationKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := ApplicationKeyInput{
		ApplicationKeyId: d.Id(),
	}

	var output ApplicationKeyOutput
	err := client.Apply(ctx, OpResourceRead, &input, &output)
	if err != nil {
		return diag.FromErr(err)
	}
	if output.ApplicationKeyId == "" && !d.IsNewResource() {
		// deleted application key
		tflog.Warn(ctx, "Application Key not found, possible resource drift", map[string]interface{}{
			"application_key_id": d.Id(),
		})
		d.SetId("")
		return nil
	}

	output.ApplicationKey = d.Get("application_key").(string)
	validDuration := d.Get("valid_duration_in_seconds").(int)

	err = client.Populate(ctx, OpResourceRead, &output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Restore valid_duration_in_seconds in state
	if err := d.Set("valid_duration_in_seconds", validDuration); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceB2ApplicationKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := ApplicationKeyInput{
		ApplicationKeyId: d.Id(),
	}

	err := client.Apply(ctx, OpResourceDelete, &input, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
