//####################################################################
//
// File: b2/data_source_b2_application_key.go
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

func dataSourceB2ApplicationKey() *schema.Resource {
	return &schema.Resource{
		Description: "B2 application key data source.",

		ReadContext: dataSourceB2ApplicationKeyRead,

		Schema: map[string]*schema.Schema{
			"key_name": {
				Description:  "The name assigned when the key was created.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"application_key_id": {
				Description: "The ID of the key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"bucket_ids": {
				Description: "When present, restricts access to specified buckets.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"capabilities": {
				Description: "A set of strings, each one naming a capability the key has.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"name_prefix": {
				Description: "When present, restricts access to files whose names start with the prefix.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"options": {
				Description: "A list of application key options.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"bucket_id": {
				Description: "When present, restricts access to one bucket.",
				Type:        schema.TypeString,
				Computed:    true,
				Deprecated:  "This argument is deprecated in favor of 'bucket_ids' argument",
			},
		},
	}
}

func dataSourceB2ApplicationKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "application_key"
	const op = DATA_SOURCE_READ

	input := map[string]interface{}{
		"key_name": d.Get("key_name").(string),
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

	return nil
}
