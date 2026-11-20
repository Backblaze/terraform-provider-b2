//####################################################################
//
// File: b2/resource_b2_application_key_diff_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceB2ApplicationKeyBucketIdsDiff(t *testing.T) {
	for _, tc := range []struct {
		name            string
		buckets         []interface{}
		legacy          bool
		unknownLegacy   bool
		wantReplacement bool
	}{
		{"add bucket", []interface{}{"bucket-a", "bucket-b", "bucket-c"}, false, false, true},
		{"remove bucket", []interface{}{"bucket-a"}, false, false, true},
		{"remove restriction", nil, false, false, true},
		{"unchanged", []interface{}{"bucket-a", "bucket-b"}, false, false, false},
		{"legacy bucket_id", nil, true, false, false},
		{"legacy bucket_id unknown", nil, true, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := resourceB2ApplicationKey()
			state := &terraform.InstanceState{
				ID: "key-id",
				Attributes: map[string]string{
					"key_name":       "test-key",
					"capabilities.#": "1",
					fmt.Sprintf("capabilities.%d", schema.HashString("readFiles")): "readFiles",
					"bucket_ids.#": "2",
					fmt.Sprintf("bucket_ids.%d", schema.HashString("bucket-a")): "bucket-a",
					fmt.Sprintf("bucket_ids.%d", schema.HashString("bucket-b")): "bucket-b",
					"bucket_id": "bucket-a", // Populated by refresh, including for bucket_ids keys.
				},
			}
			config := map[string]interface{}{
				"key_name":     "test-key",
				"capabilities": []interface{}{"readFiles"},
			}
			rawBucketID := cty.NullVal(cty.String)
			if tc.legacy {
				config["bucket_id"] = "bucket-a"
				rawBucketID = cty.StringVal("bucket-a")
				if tc.unknownLegacy {
					config["bucket_id"] = "74D93920-ED26-11E3-AC10-0800200C9A66" // hcl2shim.UnknownVariableValue
					rawBucketID = cty.UnknownVal(cty.String)
				}
			} else if tc.buckets != nil {
				config["bucket_ids"] = tc.buckets
			}
			// The SDK carries the request's raw configuration on InstanceState during diff.
			state.RawConfig = cty.ObjectVal(map[string]cty.Value{"bucket_id": rawBucketID})
			diff, err := r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), nil)
			if err != nil {
				t.Fatal(err)
			}
			gotReplacement := diff != nil && diff.RequiresNew()
			if gotReplacement != tc.wantReplacement {
				t.Fatalf("RequiresNew = %t, want %t; diff: %#v", gotReplacement, tc.wantReplacement, diff)
			}
		})
	}
}
