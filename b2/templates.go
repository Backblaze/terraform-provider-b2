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
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_md5": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_sha1": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_info": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"file_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"upload_timestamp": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func getDataSourceCorsRulesElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cors_rule_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"allowed_origins": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"allowed_operations": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"max_age_seconds": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"allowed_headers": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"expose_headers": {
				Type: schema.TypeList,
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
				Type:     schema.TypeString,
				Computed: true,
			},
			"days_from_hiding_to_deleting": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"days_from_uploading_to_hiding": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func getDataSourceAllowedElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"bucket_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"bucket_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"capabilities": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"name_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func getResourceCorsRulesElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cors_rule_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"allowed_origins": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"allowed_operations": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"max_age_seconds": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"allowed_headers": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"expose_headers": {
				Type: schema.TypeList,
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
				Type:     schema.TypeString,
				Required: true,
			},
			"days_from_hiding_to_deleting": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"days_from_uploading_to_hiding": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}
