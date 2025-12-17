//####################################################################
//
// File: b2/data_source_b2_bucket_notification_rules.go
//
// Copyright 2024 Backblaze Inc. All Rights Reserved.
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

func dataSourceB2BucketNotificationRules() *schema.Resource {
	return &schema.Resource{
		Description: "B2 bucket notification rules data source.",

		ReadContext: dataSourceB2BucketNotificationRulesRead,

		Schema: map[string]*schema.Schema{
			"bucket_id": {
				Description:  "The ID of the bucket.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"notification_rules": {
				Description: "An array of Event Notification Rules.",
				Type:        schema.TypeList,
				Elem:        getNotificationRulesElem(true),
				Computed:    true,
			},
		},
	}
}

func dataSourceB2BucketNotificationRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := BucketNotificationRulesInput{
		BucketId: d.Get("bucket_id").(string),
	}

	var output BucketNotificationRulesOutput
	err := client.Apply(ctx, OpDataSourceRead, &input, &output)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output.BucketId)

	err = client.Populate(ctx, OpDataSourceRead, &output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
