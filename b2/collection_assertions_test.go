//####################################################################
//
// File: b2/collection_assertions_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"reflect"
	"strings"
	"testing"
)

func TestCollectionComparisonRejectsDifferentValues(t *testing.T) {
	state := terraform.NewState()
	state.RootModule().Resources = map[string]*terraform.ResourceState{
		"data.b2_bucket.test": {Primary: &terraform.InstanceState{ID: "a", Attributes: map[string]string{"bucket_info.%": "1", "bucket_info.name": "wrong"}}},
		"b2_bucket.test":      {Primary: &terraform.InstanceState{ID: "a", Attributes: map[string]string{"bucket_info.%": "1", "bucket_info.name": "correct"}}},
	}
	check := testCheckCollectionPair("data.b2_bucket.test", "bucket_info", "b2_bucket.test", "bucket_info", "%")
	if err := check(state); err == nil {
		t.Fatal("different map values were accepted")
	}
	state.RootModule().Resources["data.b2_bucket.test"].Primary.Attributes["bucket_info.name"] = "correct"
	if err := check(state); err != nil {
		t.Fatalf("equal maps rejected: %v", err)
	}
	state.RootModule().Resources["data.b2_bucket.test"].Primary.Attributes["bucket_info.%"] = "2"
	if err := check(state); err == nil {
		t.Fatal("different collection counts were accepted")
	}
	delete(state.RootModule().Resources, "b2_bucket.test")
	if err := check(state); err == nil {
		t.Fatal("missing resource was accepted")
	}
}

// testCheckCollectionPair compares flattened collection counts and every child.
func testCheckCollectionPair(left, leftKey, right, rightKey, countSuffix string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resources := state.RootModule().Resources
		a, ok := resources[left]
		if !ok || a.Primary == nil {
			return fmt.Errorf("missing resource %s", left)
		}
		b, ok := resources[right]
		if !ok || b.Primary == nil {
			return fmt.Errorf("missing resource %s", right)
		}
		collection := func(attrs map[string]string, prefix string) map[string]string {
			result := map[string]string{}
			for key, value := range attrs {
				if strings.HasPrefix(key, prefix+".") {
					suffix := strings.TrimPrefix(key, prefix)
					if (strings.HasSuffix(suffix, ".#") || strings.HasSuffix(suffix, ".%")) && value == "0" {
						continue
					}
					result[suffix] = value
				}
			}
			if _, ok := result["."+countSuffix]; !ok {
				result["."+countSuffix] = "0"
			}
			return result
		}
		first, second := collection(a.Primary.Attributes, leftKey), collection(b.Primary.Attributes, rightKey)
		if !reflect.DeepEqual(first, second) {
			return fmt.Errorf("collections %s.%s and %s.%s differ", left, leftKey, right, rightKey)
		}
		return nil
	}
}
