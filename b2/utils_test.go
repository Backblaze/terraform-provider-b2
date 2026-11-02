//####################################################################
//
// File: b2/utils_test.go
//
// Copyright 2025 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"strconv"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestOrderLifecycleRules(t *testing.T) {
	cases := map[string]struct {
		rules   []LifecycleRule
		desired []interface{}
		want    []string // expected order of file name prefixes
	}{
		"empty": {
			rules:   nil,
			desired: nil,
			want:    []string{},
		},
		"single rule": {
			rules:   []LifecycleRule{{FileNamePrefix: "a/"}},
			desired: []interface{}{map[string]interface{}{"file_name_prefix": "a/"}},
			want:    []string{"a/"},
		},
		"already ordered": {
			rules: []LifecycleRule{
				{FileNamePrefix: "a/"},
				{FileNamePrefix: "b/"},
				{FileNamePrefix: "c/"},
			},
			desired: []interface{}{
				map[string]interface{}{"file_name_prefix": "a/"},
				map[string]interface{}{"file_name_prefix": "b/"},
				map[string]interface{}{"file_name_prefix": "c/"},
			},
			want: []string{"a/", "b/", "c/"},
		},
		"reordered by B2": {
			// B2 returns rules in an arbitrary order; the desired order must win
			rules: []LifecycleRule{
				{FileNamePrefix: "avatars/"},
				{FileNamePrefix: "thumbnails/"},
				{FileNamePrefix: "uploads/"},
				{FileNamePrefix: "files/"},
			},
			desired: []interface{}{
				map[string]interface{}{"file_name_prefix": "uploads/"},
				map[string]interface{}{"file_name_prefix": "files/"},
				map[string]interface{}{"file_name_prefix": "thumbnails/"},
				map[string]interface{}{"file_name_prefix": "avatars/"},
			},
			want: []string{"uploads/", "files/", "thumbnails/", "avatars/"},
		},
		"no desired rules": {
			// e.g. empty state (import); API order is kept
			rules: []LifecycleRule{
				{FileNamePrefix: "b/"},
				{FileNamePrefix: "a/"},
			},
			desired: nil,
			want:    []string{"b/", "a/"},
		},
		"rule added out-of-band": {
			// rules missing from desired keep API order at the end
			rules: []LifecycleRule{
				{FileNamePrefix: "a/"},
				{FileNamePrefix: "new/"},
				{FileNamePrefix: "b/"},
			},
			desired: []interface{}{
				map[string]interface{}{"file_name_prefix": "b/"},
				map[string]interface{}{"file_name_prefix": "a/"},
			},
			want: []string{"b/", "a/", "new/"},
		},
		"rule removed out-of-band": {
			// desired rules missing from B2 are skipped
			rules: []LifecycleRule{
				{FileNamePrefix: "a/"},
			},
			desired: []interface{}{
				map[string]interface{}{"file_name_prefix": "a/"},
				map[string]interface{}{"file_name_prefix": "removed/"},
			},
			want: []string{"a/"},
		},
		"duplicate prefixes": {
			rules: []LifecycleRule{
				{FileNamePrefix: "a/", DaysFromHidingToDeleting: 2},
				{FileNamePrefix: "a/", DaysFromHidingToDeleting: 1},
			},
			desired: []interface{}{
				map[string]interface{}{"file_name_prefix": "a/"},
				map[string]interface{}{"file_name_prefix": "a/"},
			},
			want: []string{"a/", "a/"},
		},
		"non-rule desired entry": {
			rules: []LifecycleRule{
				{FileNamePrefix: "a/"},
				{FileNamePrefix: "b/"},
			},
			desired: []interface{}{
				"not a rule",
				map[string]interface{}{"file_name_prefix": "b/"},
			},
			want: []string{"b/", "a/"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := orderLifecycleRules(tc.rules, tc.desired)

			prefixes := make([]string, len(got))
			for i, rule := range got {
				prefixes[i] = rule.FileNamePrefix
			}

			if len(prefixes) != len(tc.want) {
				t.Fatalf("expected prefixes %v, got %v", tc.want, prefixes)
			}
			for i := range tc.want {
				if prefixes[i] != tc.want[i] {
					t.Fatalf("expected prefixes %v, got %v", tc.want, prefixes)
				}
			}
		})
	}
}

func TestOrderLifecycleRulesKeepsRuleFields(t *testing.T) {
	rules := []LifecycleRule{
		{FileNamePrefix: "avatars/", DaysFromHidingToDeleting: 1},
		{FileNamePrefix: "uploads/", DaysFromUploadingToHiding: 1},
	}
	desired := []interface{}{
		map[string]interface{}{"file_name_prefix": "uploads/", "days_from_uploading_to_hiding": 1},
		map[string]interface{}{"file_name_prefix": "avatars/", "days_from_hiding_to_deleting": 1},
	}

	got := orderLifecycleRules(rules, desired)

	if got[0].FileNamePrefix != "uploads/" || got[0].DaysFromUploadingToHiding != 1 || got[0].DaysFromHidingToDeleting != 0 {
		t.Errorf("first rule does not match B2 data for 'uploads/': %+v", got[0])
	}
	if got[1].FileNamePrefix != "avatars/" || got[1].DaysFromHidingToDeleting != 1 || got[1].DaysFromUploadingToHiding != 0 {
		t.Errorf("second rule does not match B2 data for 'avatars/': %+v", got[1])
	}
}

func TestMapStoresConfig(t *testing.T) {
	serverAdded := map[string]struct{}{"large_file_sha1": {}}

	cases := map[string]struct {
		config map[string]string
		state  map[string]string
		want   bool
	}{
		"identical":                {map[string]string{"a": "1"}, map[string]string{"a": "1"}, true},
		"both empty":               {map[string]string{}, map[string]string{}, true},
		"differs only by key case": {map[string]string{"Owner": "x"}, map[string]string{"owner": "x"}, true},
		"configured key missing":   {map[string]string{"a": "1", "b": "2"}, map[string]string{"a": "1"}, false},
		"value differs":            {map[string]string{"a": "1"}, map[string]string{"a": "2"}, false},
		"stored key not configured and not server added": {
			map[string]string{"a": "1"}, map[string]string{"a": "1", "b": "2"}, false,
		},
		"stored key added by the server": {
			map[string]string{"a": "1"}, map[string]string{"a": "1", "large_file_sha1": "abc"}, true,
		},
		"server added key configured too": {
			map[string]string{"large_file_sha1": "abc"}, map[string]string{"large_file_sha1": "abc"}, true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := mapStoresConfig(tc.config, tc.state, serverAdded); got != tc.want {
				t.Errorf("mapStoresConfig(%v, %v) = %v, want %v", tc.config, tc.state, got, tc.want)
			}
		})
	}
}

func TestValidateMapKeys(t *testing.T) {
	validate := validateMapKeys("large_file_sha1")

	t.Run("lower case keys are accepted", func(t *testing.T) {
		if diags := validate(map[string]interface{}{"owner": "x"}, cty.Path{}); len(diags) != 0 {
			t.Errorf("unexpected diagnostics: %v", diags)
		}
	})
	t.Run("mixed case key warns about lower casing", func(t *testing.T) {
		diags := validate(map[string]interface{}{"Owner": "x"}, cty.Path{})
		if len(diags) != 1 || diags[0].Severity != diag.Warning || diags[0].Summary != "Key will be stored in lower case" {
			t.Errorf("unexpected diagnostics: %v", diags)
		}
	})
	t.Run("server added key warns, whatever its case", func(t *testing.T) {
		diags := validate(map[string]interface{}{"Large_File_SHA1": "x"}, cty.Path{})
		if len(diags) != 2 || diags[1].Summary != "Key is set by B2" {
			t.Errorf("unexpected diagnostics: %v", diags)
		}
	})
	t.Run("a map without server added keys never reports one", func(t *testing.T) {
		if diags := validateMapKeys()(map[string]interface{}{"large_file_sha1": "x"}, cty.Path{}); len(diags) != 0 {
			t.Errorf("unexpected diagnostics: %v", diags)
		}
	})
}

// suppressionFor runs the DiffSuppressFunc of a map attribute for a state and a configuration.
func suppressionFor(t *testing.T, serverAdded []string, state, config map[string]string, key string) bool {
	t.Helper()

	attribute := "file_info"
	sm := map[string]*schema.Schema{
		attribute: {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
	}

	attrs := map[string]string{attribute + ".%": strconv.Itoa(len(state))}
	for k, v := range state {
		attrs[attribute+"."+k] = v
	}
	configValues := map[string]cty.Value{}
	for k, v := range config {
		configValues[k] = cty.StringVal(v)
	}
	var configMap cty.Value
	if len(configValues) == 0 {
		configMap = cty.MapValEmpty(cty.String)
	} else {
		configMap = cty.MapVal(configValues)
	}
	raw := cty.ObjectVal(map[string]cty.Value{attribute: configMap})

	d, err := schema.InternalMap(sm).Data(&terraform.InstanceState{Attributes: attrs}, &terraform.InstanceDiff{RawConfig: raw})
	if err != nil {
		t.Fatal(err)
	}

	return suppressMapKeyCaseDiff(attribute, serverAdded...)(attribute+"."+key, "", "", d)
}

func TestSuppressMapKeyCaseDiff(t *testing.T) {
	serverAdded := []string{"large_file_sha1"}

	cases := map[string]struct {
		serverAdded []string
		state       map[string]string
		config      map[string]string
		key         string
		want        bool
	}{
		"a key stored in lower case is not a change": {
			nil, map[string]string{"owner": "x"}, map[string]string{"Owner": "x"}, "Owner", true,
		},
		"a changed value is a change": {
			nil, map[string]string{"owner": "x"}, map[string]string{"Owner": "y"}, "Owner", false,
		},
		"a key removed from the config is a change": {
			nil, map[string]string{"owner": "x"}, map[string]string{}, "owner", false,
		},
		"a key added by the server is not a change": {
			serverAdded, map[string]string{"large_file_sha1": "abc"}, map[string]string{}, "large_file_sha1", true,
		},
		"without server added keys the same key is a change": {
			nil, map[string]string{"large_file_sha1": "abc"}, map[string]string{}, "large_file_sha1", false,
		},
		"the element count of a map that holds the config is suppressed": {
			serverAdded, map[string]string{"owner": "x", "large_file_sha1": "abc"}, map[string]string{"Owner": "x"}, "%", true,
		},
		"the element count is kept when a key was removed": {
			serverAdded, map[string]string{"owner": "x", "other": "y", "large_file_sha1": "abc"}, map[string]string{"Owner": "x"}, "%", false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := suppressionFor(t, tc.serverAdded, tc.state, tc.config, tc.key); got != tc.want {
				t.Errorf("suppress %q = %v, want %v", tc.key, got, tc.want)
			}
		})
	}
}
