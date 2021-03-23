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
			"bucket_id": {
				Description: "When present, restricts access to one bucket.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"name_prefix": {
				Description:  "When present, restricts access to files whose names start with the prefix.",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"bucket_id"},
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
			"options": {
				Description: "List of application key options.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
		},
	}
}

func resourceB2ApplicationKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "application_key"
	const op = RESOURCE_CREATE

	input := map[string]interface{}{
		"key_name":     d.Get("key_name").(string),
		"capabilities": d.Get("capabilities").(*schema.Set).List(),
		"bucket_id":    d.Get("bucket_id").(string),
		"name_prefix":  d.Get("name_prefix").(string),
	}

	output, err := client.apply(name, op, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output["application_key_id"].(string))

	err = client.populate(name, op, output, d)
	if err != nil {
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

	output, err := client.apply(name, op, input)
	if err != nil {
		return diag.FromErr(err)
	}

	output["application_key"] = d.Get("application_key").(string)

	err = client.populate(name, op, output, d)
	if err != nil {
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

	_, err := client.apply(name, op, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
