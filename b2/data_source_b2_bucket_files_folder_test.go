//####################################################################
//
// File: b2/data_source_b2_bucket_files_folder_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestBucketFilesFolderEntryPopulatesState(t *testing.T) {
	var output BucketFilesOutput
	if err := json.Unmarshal([]byte(`{"bucketId":"bucket-1","folderName":"","showVersions":false,"recursive":false,"fileVersions":[{"action":"folder","bucketId":"bucket-1","fileName":"sub/"}]}`), &output); err != nil {
		t.Fatal(err)
	}
	resource := dataSourceB2BucketFiles()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{"bucket_id": "bucket-1"})
	client := Client{DataSourcesMap: map[string]*schema.Resource{"b2_bucket_files": resource}}
	if err := client.Populate(context.Background(), OpDataSourceRead, &output, d); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]interface{}{
		"file_versions.#":           1,
		"file_versions.0.action":    "folder",
		"file_versions.0.bucket_id": "bucket-1",
		"file_versions.0.file_name": "sub/",
		"file_versions.0.file_id":   "",
	} {
		if got := d.Get(key); got != want {
			t.Errorf("%s = %v, want %v", key, got, want)
		}
	}
}

func TestBucketFilesSchemaDescribesFolderEntries(t *testing.T) {
	description := strings.ToLower(dataSourceB2BucketFiles().Schema["file_versions"].Description)
	for _, detail := range []string{"non-recursive", "subfolder", "action", "folder", "bucket_id", "file_name", "empty or zero"} {
		if !strings.Contains(description, detail) {
			t.Errorf("file_versions description does not explain %q: %s", detail, description)
		}
	}
}
