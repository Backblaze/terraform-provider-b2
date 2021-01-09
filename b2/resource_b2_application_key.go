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
)

func resourceB2ApplicationKey() *schema.Resource {
	return &schema.Resource{
		Description: "B2 application key resource.",

		CreateContext: resourceB2ApplicationKeyCreate,
		ReadContext:   resourceB2ApplicationKeyRead,
		// 		UpdateContext: resourceB2ApplicationKeyUpdate,
		DeleteContext: resourceB2ApplicationKeyDelete,

		Schema: map[string]*schema.Schema{
			"key_name": {
				Description: "The name of the key.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"capabilities": {
				Description: "A list of capabilities.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				ForceNew: true,
			},
			"application_key_id": {
				Description: "The ID of the key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"application_key": {
				Description: "The key.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceB2ApplicationKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := map[string]interface{}{
		"key_name":     d.Get("key_name").(string),
		"capabilities": d.Get("capabilities").(*schema.Set).List(),
	}

	output, err := client.apply(TYPE_RESOURCE, "application_key", CRUD_CREATE, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("application_key_id", output["application_key_id"])
	d.Set("application_key", output["application_key"])
	d.SetId(output["application_key_id"].(string))

	return nil
}

func resourceB2ApplicationKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := map[string]interface{}{
		"application_key_id": d.Id(),
	}

	output, err := client.apply(TYPE_RESOURCE, "application_key", CRUD_READ, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("capabilities", output["capabilities"])

	return nil
}

func resourceB2ApplicationKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := map[string]interface{}{
		"application_key_id": d.Id(),
	}

	_, err := client.apply(TYPE_RESOURCE, "application_key", CRUD_DELETE, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
