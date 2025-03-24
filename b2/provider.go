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
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown

	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		desc = strings.TrimSpace(desc)

		if !bytes.HasSuffix([]byte(desc), []byte(".")) && desc != "" {
			desc += "."
		}

		if s.Default != nil || s.DefaultFunc != nil {
			if s.DefaultFunc != nil {
				val, err := s.DefaultFunc()
				if err == nil && val != nil {
					desc += fmt.Sprintf(" Defaults to `%v`.", val)
				}
			} else if s.Default == "" {
				desc += " Defaults to `\"\"`."
			} else {
				desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
			}
		}

		if s.RequiredWith != nil && len(s.RequiredWith) > 0 {
			requiredWith := make([]string, len(s.RequiredWith))
			for i, c := range s.RequiredWith {
				requiredWith[i] = fmt.Sprintf("`%s`", c)
			}
			desc += fmt.Sprintf(" Required when using %s.", strings.Join(requiredWith, ", "))
		}

		if s.ConflictsWith != nil && len(s.ConflictsWith) > 0 {
			conflicts := make([]string, len(s.ConflictsWith))
			for i, c := range s.ConflictsWith {
				conflicts[i] = fmt.Sprintf("`%s`", c)
			}
			desc += fmt.Sprintf(" Conflicts with %s.", strings.Join(conflicts, ", "))
		}

		if s.ExactlyOneOf != nil && len(s.ExactlyOneOf) > 0 {
			exactlyOneOfs := make([]string, len(s.ExactlyOneOf))
			for i, c := range s.ExactlyOneOf {
				exactlyOneOfs[i] = fmt.Sprintf("`%s`", c)
			}
			desc += fmt.Sprintf(" Must provide only one of %s.", strings.Join(exactlyOneOfs, ", "))
		}

		if s.AtLeastOneOf != nil && len(s.AtLeastOneOf) > 0 {
			atLeastOneOfs := make([]string, len(s.AtLeastOneOf))
			for i, c := range s.AtLeastOneOf {
				atLeastOneOfs[i] = fmt.Sprintf("`%s`", c)
			}
			desc += fmt.Sprintf(" Must provide at least one of %s.", strings.Join(atLeastOneOfs, ", "))
		}

		if s.ForceNew {
			desc += " **Modifying this attribute will force creation of a new resource.**"
		}

		return strings.TrimSpace(desc)
	}
}

func New(version string, exec string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"application_key_id": {
					Description: "B2 Application Key ID (B2_APPLICATION_KEY_ID env)",
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("B2_APPLICATION_KEY_ID", nil),
				},
				"application_key": {
					Description: "B2 Application Key (B2_APPLICATION_KEY env)",
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("B2_APPLICATION_KEY", nil),
				},
				"endpoint": {
					Description: "B2 endpoint - the string 'production' or a custom B2 API URL (B2_ENDPOINT env)." +
						" You should not need to set this unless you work at Backblaze.",
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("B2_ENDPOINT", "production"),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"b2_account_info":              dataSourceB2AccountInfo(),
				"b2_application_key":           dataSourceB2ApplicationKey(),
				"b2_bucket":                    dataSourceB2Bucket(),
				"b2_bucket_file":               dataSourceB2BucketFile(),
				"b2_bucket_file_signed_url":    dataSourceB2BucketFileSignedUrl(),
				"b2_bucket_files":              dataSourceB2BucketFiles(),
				"b2_bucket_notification_rules": dataSourceB2BucketNotificationRules(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"b2_application_key":           resourceB2ApplicationKey(),
				"b2_bucket":                    resourceB2Bucket(),
				"b2_bucket_file_version":       resourceB2BucketFileVersion(),
				"b2_bucket_notification_rules": resourceB2BucketNotificationRules(),
			},
		}

		p.ConfigureContextFunc = configure(version, exec, p)

		return p
	}
}

func configure(version string, exec string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		dataSources := map[string][]string{}
		sensitiveDataSources := map[string]map[string]bool{}
		for k, v := range p.DataSourcesMap {
			sensitiveDataSources[k] = make(map[string]bool)
			for kk, vv := range v.Schema {
				dataSources[k] = append(dataSources[k], kk)
				if vv.Sensitive {
					sensitiveDataSources[k][kk] = true
				}
			}
		}

		resources := map[string][]string{}
		sensitiveResources := map[string]map[string]bool{}
		for k, v := range p.ResourcesMap {
			sensitiveResources[k] = make(map[string]bool)
			for kk, vv := range v.Schema {
				resources[k] = append(resources[k], kk)
				if vv.Sensitive {
					sensitiveResources[k][kk] = true
				}
			}
		}

		userAgent := p.UserAgent("Terraform-B2-Provider", version)
		client := &Client{
			Exec:                 exec,
			UserAgentAppend:      userAgent,
			ApplicationKeyId:     d.Get("application_key_id").(string),
			ApplicationKey:       d.Get("application_key").(string),
			Endpoint:             d.Get("endpoint").(string),
			DataSources:          dataSources,
			Resources:            resources,
			SensitiveDataSources: sensitiveDataSources,
			SensitiveResources:   sensitiveResources,
		}

		tflog.Info(ctx, "User Agent append", map[string]interface{}{
			"user_agent_append": userAgent,
		})

		return client, nil
	}
}
