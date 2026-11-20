//####################################################################
//
// File: b2/resource_b2_bucket_retention_state_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestRetainExplicitNoDefaultRetention(t *testing.T) {
	cases := []struct {
		name       string
		configured string
		enabled    bool
		remote     *DefaultRetention
		want       string
	}{
		{name: "explicit none and empty remote", configured: "none", enabled: true, want: "none"},
		{name: "explicit none and remote governance", configured: "none", enabled: true, remote: &DefaultRetention{Mode: "governance"}, want: "governance"},
		{name: "governance and empty remote", configured: "governance", enabled: true, want: ""},
		{name: "omitted config and empty remote", configured: "", enabled: true, want: ""},
		{name: "disabled and empty remote", configured: "none", enabled: false, want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config := map[string]interface{}{}
			if tc.configured != "" {
				config["file_lock_configuration"] = []interface{}{map[string]interface{}{
					"is_file_lock_enabled": true,
					"default_retention":    []interface{}{map[string]interface{}{"mode": tc.configured}},
				}}
			}
			data := schema.TestResourceDataRaw(t, resourceB2Bucket().Schema, config)
			output := BucketOutput{FileLockConfiguration: &FileLockConfiguration{IsFileLockEnabled: tc.enabled, DefaultRetention: tc.remote}}
			retainExplicitNoDefaultRetention(data, &output)
			got := ""
			if output.FileLockConfiguration.DefaultRetention != nil {
				got = output.FileLockConfiguration.DefaultRetention.Mode
			}
			if got != tc.want {
				t.Fatalf("retention mode = %q, want %q", got, tc.want)
			}
		})
	}
}
