//####################################################################
//
// File: b2/resource_b2_bucket_file_lock_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceB2BucketFileLockDiff(t *testing.T) {
	cases := []struct {
		name            string
		stateEnabled    string
		configEnabled   bool
		wantRequiresNew bool
	}{
		{name: "enable on existing bucket updates in place", stateEnabled: "false", configEnabled: true, wantRequiresNew: false},
		{name: "disable forces a new bucket", stateEnabled: "true", configEnabled: false, wantRequiresNew: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := &terraform.InstanceState{
				ID: "bucket-id",
				Attributes: map[string]string{
					"id":                        "bucket-id",
					"bucket_id":                 "bucket-id",
					"bucket_name":               "my-bucket",
					"bucket_type":               "allPrivate",
					"revision":                  "1",
					"file_lock_configuration.#": "1",
					"file_lock_configuration.0.is_file_lock_enabled": tc.stateEnabled,
					"file_lock_configuration.0.default_retention.#":  "0",
				},
			}
			config := terraform.NewResourceConfigRaw(map[string]interface{}{
				"bucket_name": "my-bucket",
				"bucket_type": "allPrivate",
				"file_lock_configuration": []interface{}{
					map[string]interface{}{"is_file_lock_enabled": tc.configEnabled},
				},
			})

			diff, err := resourceB2Bucket().SimpleDiff(context.Background(), state, config, nil)
			if err != nil {
				t.Fatal(err)
			}
			if diff == nil {
				t.Fatal("expected a diff")
			}
			attr, ok := diff.Attributes["file_lock_configuration.0.is_file_lock_enabled"]
			if !ok {
				t.Fatalf("expected is_file_lock_enabled in diff, got %v", diff.Attributes)
			}
			if attr.RequiresNew != tc.wantRequiresNew {
				t.Fatalf("RequiresNew = %v, want %v", attr.RequiresNew, tc.wantRequiresNew)
			}
		})
	}
}
