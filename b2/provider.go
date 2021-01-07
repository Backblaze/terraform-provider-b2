//####################################################################
//
// File: b2/provider.go
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

func New(version string, exec string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"application_key_id": {
					Description: "B2 Application Key ID",
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("B2_APPLICATION_KEY_ID", nil),
				},
				"application_key": {
					Description: "B2 Application Key",
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("B2_APPLICATION_KEY", nil),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"b2_application_key": dataSourceB2ApplicationKey(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"b2_application_key": resourceB2ApplicationKey(),
			},
		}

		p.ConfigureContextFunc = configure(version, exec, p)

		return p
	}
}

func configure(version string, exec string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		client := &Client{
			Exec:             exec,
			Version:          version,
			ApplicationKeyId: d.Get("application_key_id").(string),
			ApplicationKey:   d.Get("application_key").(string),
		}

		// 		if version != "test" {
		// 			input := map[string]string{
		// 				"application_key_id": d.Get("application_key_id").(string),
		// 				"application_key":    d.Get("application_key").(string),
		// 			}
		//
		// 			_, err := client.apply("provider", "authorize_account", input)
		// 			if err != nil {
		// 				return nil, err
		// 			}
		// 		}

		return client, nil
	}
}
