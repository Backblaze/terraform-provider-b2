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
)

func dataSourceB2ApplicationKey() *schema.Resource {
	return &schema.Resource{
		Description: "B2 application key data source.",

		ReadContext: dataSourceB2ApplicationKeyRead,

		Schema: map[string]*schema.Schema{
			"key_name": {
				Description: "The name assigned when the key was created.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"capabilities": {
				Description: "A list of capabilities.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"application_key_id": {
				Description: "The ID of the key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceB2ApplicationKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := map[string]interface{}{
		"key_name": d.Get("key_name").(string),
	}

	output, err := client.apply(TYPE_DATA_SOURCE, "application_key", CRUD_READ, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("application_key_id", output["application_key_id"])
	d.Set("capabilities", output["capabilities"])
	d.SetId(output["application_key_id"].(string))

	return nil
}
