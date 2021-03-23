//####################################################################
//
// File: b2/templates.go
//
// Copyright 2021 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func getDataSourceFileVersionsElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"action": {
				Description: "One of 'start', 'upload', 'hide', 'folder', or other values added in the future.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"content_md5": {
				Description: "MD5 sum of the content.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"content_sha1": {
				Description: "SHA1 hash of the content.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"content_type": {
				Description: "Content type. If not set, it will be set based on the file extension.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"file_id": {
				Description: "The unique identifier for this version of this file.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"file_info": {
				Description: "The custom information that is uploaded with the file.",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"file_name": {
				Description: "The name of the B2 file.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"size": {
				Description: "The file size.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"upload_timestamp": {
				Description: "This is a UTC time when this file was uploaded.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func getDataSourceCorsRulesElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cors_rule_name": {
				Description: "A name for humans to recognize the rule in a user interface.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"allowed_origins": {
				Description: "A non-empty list specifying which origins the rule covers. ",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"allowed_operations": {
				Description: "A list specifying which operations the rule allows.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"max_age_seconds": {
				Description: "This specifies the maximum number of seconds that a browser may cache the response to a preflight request.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"allowed_headers": {
				Description: "If present, this is a list of headers that are allowed in a pre-flight OPTIONS's request's Access-Control-Request-Headers header value.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"expose_headers": {
				Description: "If present, this is a list of headers that may be exposed to an application inside the client.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
		},
	}
}

func getDataSourceLifecycleRulesElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"file_name_prefix": {
				Description: "It specifies which files in the bucket it applies to.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"days_from_hiding_to_deleting": {
				Description: "It says how long to keep file versions that are not the current version.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"days_from_uploading_to_hiding": {
				Description: "It causes files to be hidden automatically after the given number of days.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func getDataSourceAllowedElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"bucket_id": {
				Description: "When present, restricts access to one bucket.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"bucket_name": {
				Description: "When 'bucket_id' is set, and it is a valid bucket that has not been deleted, this field is set to the name of the bucket.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"capabilities": {
				Description: "A list of strings, each one naming a capability the key has.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"name_prefix": {
				Description: "When present, access is restricted to files whose names start with the prefix.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func getResourceCorsRulesElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cors_rule_name": {
				Description:  "A name for humans to recognize the rule in a user interface.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"allowed_origins": {
				Description: "A non-empty list specifying which origins the rule covers. ",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"allowed_operations": {
				Description: "A list specifying which operations the rule allows.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"max_age_seconds": {
				Description: "This specifies the maximum number of seconds that a browser may cache the response to a preflight request.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"allowed_headers": {
				Description: "If present, this is a list of headers that are allowed in a pre-flight OPTIONS's request's Access-Control-Request-Headers header value.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"expose_headers": {
				Description: "If present, this is a list of headers that may be exposed to an application inside the client.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	}
}

func getResourceLifecycleRulesElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"file_name_prefix": {
				Description: "It specifies which files in the bucket it applies to.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"days_from_hiding_to_deleting": {
				Description: "It says how long to keep file versions that are not the current version.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"days_from_uploading_to_hiding": {
				Description: "It causes files to be hidden automatically after the given number of days.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
		},
	}
}
