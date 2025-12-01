//####################################################################
//
// File: b2/resource_b2_bucket_notification_rules.go
//
// Copyright 2024 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceB2BucketNotificationRules() *schema.Resource {
	return &schema.Resource{
		Description: "B2 bucket notification rules resource.",

		CreateContext: resourceB2BucketNotificationRulesCreate,
		ReadContext:   resourceB2BucketNotificationRulesRead,
		UpdateContext: resourceB2BucketNotificationRulesUpdate,
		DeleteContext: resourceB2BucketNotificationRulesDelete,

		Schema: map[string]*schema.Schema{
			"bucket_id": {
				Description:  "The ID of the bucket.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"notification_rules": {
				Description: "An array of Event Notification Rules.",
				Type:        schema.TypeList,
				Elem:        getNotificationRulesElem(false),
				Required:    true,
				MinItems:    1,
			},
		},
	}
}

func resourceB2BucketNotificationRulesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "bucket_notification_rules"
	const op = RESOURCE_CREATE

	input := map[string]interface{}{
		"bucket_id":          d.Get("bucket_id").(string),
		"notification_rules": d.Get("notification_rules").([]interface{}),
	}

	var notificationRules BucketNotificationRulesSchema
	err := client.apply(ctx, name, op, input, &notificationRules)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(notificationRules.BucketId)

	err = client.populate(ctx, name, op, &notificationRules, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceB2BucketNotificationRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "bucket_notification_rules"
	const op = RESOURCE_READ

	input := map[string]interface{}{
		"bucket_id": d.Id(),
	}

	var notificationRules BucketNotificationRulesSchema
	err := client.apply(ctx, name, op, input, &notificationRules)
	if err != nil {
		return diag.FromErr(err)
	}
	if notificationRules.BucketId == "" && !d.IsNewResource() {
		// deleted bucket, thus notification rules no longer exist
		tflog.Warn(ctx, "Bucket not found for Event Notifications, possible resource drift", map[string]interface{}{
			"bucket_id": d.Id(),
		})
		d.SetId("")
		return nil
	}

	err = client.populate(ctx, name, op, &notificationRules, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceB2BucketNotificationRulesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "bucket_notification_rules"
	const op = RESOURCE_UPDATE

	input := map[string]interface{}{
		"bucket_id":          d.Id(),
		"notification_rules": d.Get("notification_rules").([]interface{}),
	}

	var notificationRules BucketNotificationRulesSchema
	err := client.apply(ctx, name, op, input, &notificationRules)
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.populate(ctx, name, op, &notificationRules, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceB2BucketNotificationRulesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	const name = "bucket_notification_rules"
	const op = RESOURCE_DELETE

	input := map[string]interface{}{
		"bucket_id": d.Id(),
	}

	err := client.apply(ctx, name, op, input, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
