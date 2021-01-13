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
			"bucket_id": {
				Description: "The ID of the bucket.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"capabilities": {
				Description: "A list of capabilities.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"name_prefix": {
				Description: "A prefix to restrict access to files",
				Type:        schema.TypeString,
				Optional:    true,
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

func dataSourceB2ApplicationKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := map[string]interface{}{
		"key_name": d.Get("key_name").(string),
	}

	output, err := client.apply("application_key", DATA_SOURCE_READ, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output["application_key_id"].(string))

	d.Set("application_key_id", output["application_key_id"])
	d.Set("bucket_id", output["bucket_id"])
	d.Set("name_prefix", output["name_prefix"])

	if err := d.Set("capabilities", output["capabilities"]); err != nil {
		return diag.Errorf("error setting capabilities: %s", err)
	}

	if err := d.Set("options", output["options"]); err != nil {
		return diag.Errorf("error setting options: %s", err)
	}

	return nil
}
