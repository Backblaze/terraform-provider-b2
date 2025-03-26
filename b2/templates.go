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
			"server_side_encryption": {
				Description: "Server-side encryption settings.",
				Type:        schema.TypeList,
				Elem:        getServerSideEncryptionElem(true),
				Computed:    true,
			},
			"upload_timestamp": {
				Description: "This is a UTC time when this file was uploaded.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"bucket_id": {
				Description: "The ID of the bucket.",
				Type:        schema.TypeString,
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

func getResourceFileEncryptionElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"mode": {
				Description:  "Server-side encryption mode.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"none", "SSE-B2", "SSE-C"}, false),
			},
			"algorithm": {
				Description:  "Server-side encryption algorithm. AES256 is the only one supported.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"AES256"}, false),
			},
			"key": {
				Description: "Key used in SSE-C mode.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"secret_b64": {
							Description:  "Secret key value, in standard Base 64 encoding (RFC 4648)",
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validateBase64Key,
						},
						"key_id": {
							Description: "Key identifier stored in file info metadata",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// The API does not return the key, so we need to suppress diff for existing resources
					if k == "server_side_encryption.0.key.#" && d.Id() != "" {
						return true
					}
					return false
				},
			},
		},
	}
}

func getServerSideEncryptionElem(ds bool) *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"mode": {
				Description:  "Server-side encryption mode.",
				Type:         schema.TypeString,
				Computed:     If(ds, true, false),
				Optional:     If(ds, false, true),
				ValidateFunc: If(ds, nil, validation.StringInSlice([]string{"none", "SSE-B2"}, false)),
			},
			"algorithm": {
				Description:  "Server-side encryption algorithm. AES256 is the only one supported.",
				Type:         schema.TypeString,
				Computed:     If(ds, true, false),
				Optional:     If(ds, false, true),
				ValidateFunc: If(ds, nil, validation.StringInSlice([]string{"AES256"}, false)),
			},
		},
	}
}

func getCorsRulesElem(ds bool) *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cors_rule_name": {
				Description:  "A name for humans to recognize the rule in a user interface.",
				Type:         schema.TypeString,
				Computed:     If(ds, true, false),
				Required:     If(ds, false, true),
				ValidateFunc: If(ds, nil, validation.NoZeroValues),
			},
			"allowed_origins": {
				Description: "A non-empty list specifying which origins the rule covers. ",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: If(ds, true, false),
				Required: If(ds, false, true),
			},
			"allowed_operations": {
				Description: "A list specifying which operations the rule allows.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: If(ds, true, false),
				Required: If(ds, false, true),
			},
			"max_age_seconds": {
				Description: "This specifies the maximum number of seconds that a browser may cache the response to a preflight request.",
				Type:        schema.TypeInt,
				Computed:    If(ds, true, false),
				Required:    If(ds, false, true),
			},
			"allowed_headers": {
				Description: "If present, this is a list of headers that are allowed in a pre-flight OPTIONS's request's Access-Control-Request-Headers header value.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: If(ds, true, false),
				Optional: If(ds, false, true),
			},
			"expose_headers": {
				Description: "If present, this is a list of headers that may be exposed to an application inside the client.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: If(ds, true, false),
				Optional: If(ds, false, true),
			},
		},
	}
}

func getFileLockConfigurationElem(ds bool) *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"is_file_lock_enabled": {
				Description: "If present, the boolean value specifies whether bucket is File Lock-enabled.",
				Type:        schema.TypeBool,
				Computed:    If(ds, true, false),
				Optional:    If(ds, false, true),
				DefaultFunc: If(ds,
					nil,
					func() (any, error) { return false, nil },
				),
				ForceNew: true,
			},
			"default_retention": {
				Description: "Default retention settings for files uploaded to this bucket",
				Type:        schema.TypeList,
				Computed:    If(ds, true, false),
				Optional:    If(ds, false, true),
				MaxItems:    If(ds, 0, 1),
				DiffSuppressFunc: If(ds,
					nil,
					func(k, old, new string, d *schema.ResourceData) bool {
						// The API sets default value
						if k == "default_retention.#" {
							return old == "1" && new == "0"
						}
						return old == "none" && new == ""
					},
				),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Description:  "Default retention mode (compliance|governance|none).",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"compliance", "governance", "none"}, false),
						},
						"period": {
							Description: "How long for to make files immutable",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"duration": {
										Description: "Duration",
										Type:        schema.TypeInt,
										Required:    true,
									},
									"unit": {
										Description:  "Unit for duration (days|years)",
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"days", "years"}, false),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func getLifecycleRulesElem(ds bool) *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"file_name_prefix": {
				Description: "It specifies which files in the bucket it applies to.",
				Type:        schema.TypeString,
				Computed:    If(ds, true, false),
				Required:    If(ds, false, true),
			},
			"days_from_hiding_to_deleting": {
				Description: "It says how long to keep file versions that are not the current version.",
				Type:        schema.TypeInt,
				Computed:    If(ds, true, false),
				Optional:    If(ds, false, true),
			},
			"days_from_uploading_to_hiding": {
				Description: "It causes files to be hidden automatically after the given number of days.",
				Type:        schema.TypeInt,
				Computed:    If(ds, true, false),
				Optional:    If(ds, false, true),
			},
		},
	}
}

func getNotificationRulesElem(ds bool) *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"event_types": {
				Description: "The list of event types for the event notification rule.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: If(ds,
						nil,
						validation.StringInSlice([]string{
							"b2:ObjectCreated:*", "b2:ObjectCreated:Upload", "b2:ObjectCreated:MultipartUpload",
							"b2:ObjectCreated:Copy", "b2:ObjectCreated:Replica", "b2:ObjectCreated:MultipartReplica",
							"b2:ObjectDeleted:*", "b2:ObjectDeleted:Delete", "b2:ObjectDeleted:LifecycleRule",
							"b2:HideMarkerCreated:*", "b2:HideMarkerCreated:Hide", "b2:HideMarkerCreated:LifecycleRule",
							"b2:MultipartUploadCreated:*", "b2:MultipartUploadCreated:LiveRead",
						}, false),
					),
				},
				Computed: If(ds, true, false),
				Required: If(ds, false, true),
				MinItems: If(ds, 0, 1),
			},
			"is_enabled": {
				Description: "Whether the event notification rule is enabled.",
				Type:        schema.TypeBool,
				Computed:    If(ds, true, false),
				Optional:    If(ds, false, true),
				DefaultFunc: If(ds,
					nil,
					func() (any, error) { return true, nil },
				),
			},
			"name": {
				Description:  "A name for the event notification rule. The name must be unique among the bucket's notification rules.",
				Type:         schema.TypeString,
				Computed:     If(ds, true, false),
				Required:     If(ds, false, true),
				ValidateFunc: If(ds, nil, validation.NoZeroValues),
			},
			"object_name_prefix": {
				Description: "Specifies which object(s) in the bucket the event notification rule applies to.",
				Type:        schema.TypeString,
				Computed:    If(ds, true, false),
				Optional:    If(ds, false, true),
			},
			"target_configuration": {
				Description: "The target configuration for the event notification rule.",
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_type": {
							Description:  "The type of the target configuration, currently \"webhook\" only.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"webhook"}, false),
						},
						"url": {
							Description:  "The URL for the webhook.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
						},
						"custom_headers": {
							Description: "When present, additional header name/value pairs to be sent on the webhook invocation.",
							Type:        schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Description:  "Name of the header.",
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"value": {
										Description:  "Value of the header.",
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
							Optional: true,
							MaxItems: 10,
						},
						"hmac_sha256_signing_secret": {
							Description:  "The signing secret for use in verifying the X-Bz-Event-Notification-Signature.",
							Type:         schema.TypeString,
							Sensitive:    true,
							Optional:     true,
							ValidateFunc: StringLenExact(32),
						},
					},
				},
				Computed: If(ds, true, false),
				Required: If(ds, false, true),
				MaxItems: If(ds, 0, 1),
			},
			"is_suspended": {
				Description: "Whether the event notification rule is suspended.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"suspension_reason": {
				Description: "A brief description of why the event notification rule was suspended.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}
