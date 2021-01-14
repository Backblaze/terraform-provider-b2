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
				"endpoint": {
					Description: "B2 endpoint (production or custom URL)",
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("B2_ENDPOINT", "production"),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"b2_application_key": dataSourceB2ApplicationKey(),
				"b2_bucket":          dataSourceB2Bucket(),
				"b2_bucket_file":     dataSourceB2BucketFile(),
				"b2_bucket_files":    dataSourceB2BucketFiles(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"b2_application_key":     resourceB2ApplicationKey(),
				"b2_bucket":              resourceB2Bucket(),
				"b2_bucket_file_version": resourceB2BucketFileVersion(),
			},
		}

		p.ConfigureContextFunc = configure(version, exec, p)

		return p
	}
}

func configure(version string, exec string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		dataSources := map[string][]string{}
		for k, v := range p.DataSourcesMap {
			for kk := range v.Schema {
				dataSources[k] = append(dataSources[k], kk)
			}
		}

		resources := map[string][]string{}
		for k, v := range p.ResourcesMap {
			for kk := range v.Schema {
				resources[k] = append(resources[k], kk)
			}
		}

		client := &Client{
			Exec:             exec,
			Version:          version,
			ApplicationKeyId: d.Get("application_key_id").(string),
			ApplicationKey:   d.Get("application_key").(string),
			Endpoint:         d.Get("endpoint").(string),
			DataSources:      dataSources,
			Resources:        resources,
		}

		return client, nil
	}
}
