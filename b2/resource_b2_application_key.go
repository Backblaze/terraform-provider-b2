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
				Description: "A set of strings, each one naming a capability the key has.",
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
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Suppress diff if bucket_id is set in config (backward compatibility)
					if _, ok := d.GetOk("bucket_id"); ok {
						return true
					}
					return false
				},
			},
			"name_prefix": {
				Description: "When present, restricts access to files whose names start with the prefix.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"application_key": {
				Description: "The key.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"application_key_id": {
				Description: "The ID of the newly created key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expiration_timestamp": {
				Description: "When present, says when this key will expire, in milliseconds since 1970.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"options": {
				Description: "List of application key options.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"valid_duration_in_seconds": {
				Description:  "When provided, the key will expire after the given number of seconds, and will have expirationTimestamp set. Value must be a positive integer, and must be less than 1000 days (in seconds).",
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
					// Suppress diff if bucket_ids is set (bucket_id is auto-populated from bucket_ids)
					if _, ok := d.GetOk("bucket_ids"); ok {
						return true
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
	const name = "application_key"
	const op = RESOURCE_CREATE

	input := map[string]interface{}{
		"key_name":                  d.Get("key_name").(string),
		"capabilities":              d.Get("capabilities").(*schema.Set).List(),
		"name_prefix":               d.Get("name_prefix").(string),
		"valid_duration_in_seconds": d.Get("valid_duration_in_seconds").(int),
	}

	// Handle backward compatibility
	for k, v := range resourceB2ApplicationKeyApplyDeprecatedToCurrent(d) {
		input[k] = v
	}

	var applicationKey ApplicationKeySchema
	err := client.apply(ctx, name, op, input, &applicationKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(applicationKey.ApplicationKeyId)

	err = client.populate(ctx, name, op, &applicationKey, d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Preserve valid_duration_in_seconds in state
	if err := d.Set("valid_duration_in_seconds", input["valid_duration_in_seconds"]); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceB2ApplicationKeyPopulateDeprecatedToCurrent(d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceB2ApplicationKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "application_key"
	const op = RESOURCE_READ

	input := map[string]interface{}{
		"application_key_id": d.Id(),
	}

	var applicationKey ApplicationKeySchema
	err := client.apply(ctx, name, op, input, &applicationKey)
	if err != nil {
		return diag.FromErr(err)
	}
	if applicationKey.ApplicationKeyId == "" && !d.IsNewResource() {
		// deleted application key
		tflog.Warn(ctx, "Application Key not found, possible resource drift", map[string]interface{}{
			"application_key_id": d.Id(),
		})
		d.SetId("")
		return nil
	}

	applicationKey.ApplicationKey = d.Get("application_key").(string)
	validDuration := d.Get("valid_duration_in_seconds").(int)

	err = client.populate(ctx, name, op, &applicationKey, d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Restore valid_duration_in_seconds in state
	if err := d.Set("valid_duration_in_seconds", validDuration); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceB2ApplicationKeyPopulateDeprecatedToCurrent(d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceB2ApplicationKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "application_key"
	const op = RESOURCE_DELETE

	input := map[string]interface{}{
		"application_key_id": d.Id(),
	}

	err := client.apply(ctx, name, op, input, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourceB2ApplicationKeyApplyDeprecatedToCurrent(d *schema.ResourceData) map[string]interface{} {
	bucketIds := d.Get("bucket_ids").(*schema.Set).List()
	if bucketId, ok := d.GetOk("bucket_id"); ok && bucketId.(string) != "" {
		bucketIds = []interface{}{bucketId.(string)}
	}
	return map[string]interface{}{
		"bucket_ids": bucketIds,
	}
}

func resourceB2ApplicationKeyPopulateDeprecatedToCurrent(d *schema.ResourceData) error {
	if bucketIds, ok := d.GetOk("bucket_ids"); ok {
		bucketIdsList := bucketIds.(*schema.Set).List()
		if len(bucketIdsList) > 0 {
			return d.Set("bucket_id", bucketIdsList[0].(string))
		}
	}
	// Set empty string if no bucket_ids
	return d.Set("bucket_id", "")
}
