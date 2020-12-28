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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceB2ApplicationKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceB2ApplicationKeyRead,

		Schema: map[string]*schema.Schema{
			"key_name": {
				Description: "The name assigned when the key was created.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"application_key_id": {
				Description: "The ID of the key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceB2ApplicationKeyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	input := map[string]string{
		"key_name": d.Get("key_name").(string),
	}

	output, err := client.apply("data_source", "application_key_id", input)
	if err != nil {
		return err
	}

	d.Set("application_key_id", output["application_key_id"])
	d.SetId(output["application_key_id"])

	return nil
}
