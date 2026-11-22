//####################################################################
//
// File: b2/data_source_bucket_ids_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	if os.Getenv("B2_BUCKET_ID_TEST_HELPER") != "1" {
		return
	}
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(2)
	}
	var input map[string]interface{}
	if json.Unmarshal(raw, &input) != nil {
		os.Exit(2)
	}
	digest := sha1.Sum(raw)
	input["_sha1"] = hex.EncodeToString(digest[:])
	input["bucketId"] = input["bucket_id"]
	input["fileName"] = input["file_name"]
	input["folderName"] = input["folder_name"]
	input["showVersions"] = input["show_versions"]
	input["maxResults"] = input["max_results"]
	input["fileVersions"] = []interface{}{}
	if json.NewEncoder(os.Stdout).Encode(input) != nil {
		os.Exit(2)
	}
	os.Exit(0)
}

func TestBucketDataSourceIDsIgnoreCredentials(t *testing.T) {
	t.Setenv("B2_BUCKET_ID_TEST_HELPER", "1")
	tests := []struct {
		name     string
		resource *schema.Resource
		config   map[string]interface{}
		wantID   string
	}{
		{"b2_bucket_file", dataSourceB2BucketFile(), map[string]interface{}{"bucket_id": "bucket-id", "file_name": "path/file.txt"}, "bucket-id/path/file.txt"},
		{"b2_bucket_files", dataSourceB2BucketFiles(), map[string]interface{}{"bucket_id": "bucket-id", "folder_name": "path/", "max_results": 7}, "bucket-id/path/"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var ids []string
			for _, auth := range []struct{ key, id, endpoint string }{
				{"first-key", "first-id", "first-endpoint"},
				{"rotated-key", "rotated-id", "second-endpoint"},
			} {
				data := schema.TestResourceDataRaw(t, tc.resource.Schema, tc.config)
				client := &Client{Exec: os.Args[0], ApplicationKey: auth.key, ApplicationKeyId: auth.id, Endpoint: auth.endpoint, DataSourcesMap: map[string]*schema.Resource{tc.name: tc.resource}}
				if diags := tc.resource.ReadContext(context.Background(), data, client); diags.HasError() {
					t.Fatalf("read: %v", diags)
				}
				ids = append(ids, data.Id())
				if tc.name == "b2_bucket_files" && data.Get("max_results") != 7 {
					t.Errorf("max_results = %v, want 7", data.Get("max_results"))
				}
			}
			if ids[0] == "" || ids[0] != ids[1] {
				t.Errorf("data source ID changed when credentials rotated: %q != %q", ids[0], ids[1])
			}
			if ids[0] != tc.wantID {
				t.Errorf("data source ID = %q, want %q", ids[0], tc.wantID)
			}
		})
	}
}

func TestBucketFilesMaxResultsSchema(t *testing.T) {
	resource := dataSourceB2BucketFiles()
	field := resource.Schema["max_results"]
	if field == nil || !field.Optional || field.Default != 10000 {
		t.Fatal("max_results must be optional with a default of 10000")
	}
	for _, tc := range []struct {
		value int
		valid bool
	}{{0, false}, {1, true}, {10000, true}, {100000, true}, {100001, false}} {
		_, errs := field.ValidateFunc(tc.value, "max_results")
		if (len(errs) == 0) != tc.valid {
			t.Errorf("max_results %d: errors = %v, valid = %t", tc.value, errs, tc.valid)
		}
	}
}
