//####################################################################
//
// File: b2/application_key_bucket_diff_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"
	"strconv"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestApplicationKeyBucketRestrictionPlan(t *testing.T) {
	for _, tc := range []struct {
		name       string
		oldBuckets []interface{}
		config     map[string]interface{}
		replace    bool
	}{
		{"narrow modern", []interface{}{"a", "b"}, map[string]interface{}{"bucket_ids": []interface{}{"a"}}, true},
		{"widen modern", []interface{}{"a", "b"}, map[string]interface{}{"bucket_ids": []interface{}{"a", "b", "c"}}, true},
		{"replace modern", []interface{}{"a"}, map[string]interface{}{"bucket_ids": []interface{}{"b"}}, true},
		{"replace legacy", []interface{}{"a"}, map[string]interface{}{"bucket_id": "b"}, true},
		{"explicitly empty modern restriction", []interface{}{"a"}, map[string]interface{}{"bucket_ids": []interface{}{}}, true},
		{"explicitly empty legacy restriction", []interface{}{"a"}, map[string]interface{}{"bucket_id": ""}, true},
		{"remove restriction", []interface{}{"a"}, map[string]interface{}{}, true},
		{"narrow by migrating to legacy", []interface{}{"a", "b"}, map[string]interface{}{"bucket_id": "a"}, true},
		{"unchanged modern", []interface{}{"a", "b"}, map[string]interface{}{"bucket_ids": []interface{}{"b", "a"}}, false},
		{"equivalent modern migration", []interface{}{"a"}, map[string]interface{}{"bucket_ids": []interface{}{"a"}}, false},
		{"equivalent legacy migration", []interface{}{"a"}, map[string]interface{}{"bucket_id": "a"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := resourceB2ApplicationKey()
			attrs := map[string]string{"options.#": "0", "id": "key-1", "key_name": "test-key", "capabilities.#": "1", "capabilities." + strconv.Itoa(schema.HashString("listFiles")): "listFiles", "bucket_ids.#": strconv.Itoa(len(tc.oldBuckets)), "bucket_id": tc.oldBuckets[0].(string)}
			for _, id := range tc.oldBuckets {
				attrs["bucket_ids."+strconv.Itoa(schema.HashString(id))] = id.(string)
			}
			state := &terraform.InstanceState{ID: "key-1", Attributes: attrs}
			tc.config["key_name"] = "test-key"
			tc.config["capabilities"] = []interface{}{"listFiles"}
			config := terraform.NewResourceConfigRaw(tc.config)
			raw := map[string]cty.Value{"bucket_id": cty.NullVal(cty.String), "bucket_ids": cty.NullVal(cty.Set(cty.String))}
			if id, ok := tc.config["bucket_id"]; ok {
				raw["bucket_id"] = cty.StringVal(id.(string))
			}
			if ids, ok := tc.config["bucket_ids"]; ok {
				values := []cty.Value{}
				for _, id := range ids.([]interface{}) {
					values = append(values, cty.StringVal(id.(string)))
				}
				if len(values) == 0 {
					raw["bucket_ids"] = cty.SetValEmpty(cty.String)
				} else {
					raw["bucket_ids"] = cty.SetVal(values)
				}
			}
			config.CtyValue = cty.ObjectVal(raw)
			diff, err := r.Diff(context.Background(), state, config, nil)
			if err != nil {
				t.Fatal(err)
			}
			if tc.replace {
				if diff == nil || !diff.RequiresNew() {
					t.Fatalf("restriction change must require replacement; diff=%#v", diff)
				}
			} else if diff != nil && !diff.Empty() {
				t.Fatalf("equivalent restriction must have no diff; diff=%#v", diff)
			}
		})
	}
}
