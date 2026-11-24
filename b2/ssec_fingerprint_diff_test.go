//####################################################################
//
// File: b2/ssec_fingerprint_diff_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"
	"encoding/base64"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strings"
	"testing"
)

func TestSSECFingerprintDiff(t *testing.T) {
	const fingerprint = "3ba3f5f43b92602683c19aee62a20342b084dd5971ddd33808d81a328879a547"
	for _, tc := range []struct {
		name, prior, key string
		replacement      bool
	}{
		{"unchanged", fingerprint, "a", false},
		{"rotation", fingerprint, "b", true},
		{"identifier", fingerprint, "a", true},
		{"legacy identifier", "", "a", false},
		{"legacy unchanged", "", "a", false},
		{"legacy unknown prior key", "", "b", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			configuredID := "key-id"
			if tc.name == "identifier" || tc.name == "legacy identifier" {
				configuredID = "new-id"
			}
			r := resourceB2BucketFileVersion()
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"bucket_id": "bucket-1", "file_name": "file.txt", "source": "missing-source"})
			d.SetId("file-1")
			if err := d.Set("file_id", "file-1"); err != nil {
				t.Fatal(err)
			}
			encryption := map[string]interface{}{"mode": "SSE-C", "algorithm": "AES256", "key": []interface{}{}}
			if tc.prior != "" {
				encryption["key"] = []interface{}{map[string]interface{}{"key_id": "key-id", "secret_b64": ""}}
			}
			if err := d.Set("server_side_encryption", []interface{}{encryption}); err != nil {
				t.Fatal(err)
			}
			if err := d.Set("file_info", map[string]interface{}{}); err != nil {
				t.Fatal(err)
			}
			config := terraform.NewResourceConfigRaw(map[string]interface{}{"bucket_id": "bucket-1", "file_name": "file.txt", "source": "missing-source", "server_side_encryption": []interface{}{map[string]interface{}{"mode": "SSE-C", "algorithm": "AES256", "key": []interface{}{map[string]interface{}{"key_id": configuredID, "secret_b64": base64.StdEncoding.EncodeToString([]byte(strings.Repeat(tc.key, 32)))}}}}})
			config.CtyValue = cty.ObjectVal(map[string]cty.Value{"server_side_encryption": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"key": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"secret_b64": cty.StringVal(base64.StdEncoding.EncodeToString([]byte(strings.Repeat(tc.key, 32)))), "key_id": cty.StringVal(configuredID)})})})})})
			state := d.State()
			// The protocol server attaches the actual configuration to prior state.
			state.RawConfig = config.CtyValue
			if tc.prior != "" {
				state.Attributes["server_side_encryption.0.key_fingerprint"] = tc.prior
			}
			diff, err := r.Diff(context.Background(), state, config, nil)
			if err != nil {
				t.Fatal(err)
			}
			if got := diff != nil && diff.RequiresNew(); got != tc.replacement {
				t.Fatalf("replacement=%t, want %t", got, tc.replacement)
			}
			if !tc.replacement && diff != nil && !diff.Empty() {
				t.Fatal("unchanged or legacy state must have no diff")
			}
		})
	}
}

func TestSSECCreateStateContainsOnlyFingerprint(t *testing.T) {
	r := resourceB2BucketFileVersion()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"bucket_id": "bucket-1", "file_name": "file.txt", "source": "missing-source", "server_side_encryption": []interface{}{map[string]interface{}{"mode": "SSE-C", "algorithm": "AES256", "key": []interface{}{map[string]interface{}{"key_id": "key-id", "secret_b64": base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32)))}}}}})
	client := newFakeClient(t, `cat >/dev/null; printf '%s' '{"fileId":"file-1","fileName":"file.txt","serverSideEncryption":{"mode":"SSE-C","algorithm":"AES256"}}'`)
	if diags := resourceB2BucketFileVersionCreate(context.Background(), d, client); diags.HasError() {
		t.Fatal(diags)
	}
	attrs := d.State().Attributes
	if attrs["server_side_encryption.0.key.0.secret_b64"] != "" {
		t.Fatal("create persisted raw SSE-C key")
	}
	if attrs["server_side_encryption.0.key_fingerprint"] != "3ba3f5f43b92602683c19aee62a20342b084dd5971ddd33808d81a328879a547" {
		t.Fatal("create did not persist the SHA-256 key fingerprint")
	}
}
