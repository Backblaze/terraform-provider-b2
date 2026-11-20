//####################################################################
//
// File: b2/file_version_content_change_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestFileVersionContentChangePlansReplacement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	state := &terraform.InstanceState{ID: "file-1", Attributes: map[string]string{
		"id": "file-1", "file_id": "file-1", "bucket_id": "bucket-1",
		"file_name": "file.txt", "source": path, "action": "upload",
		"content_sha1": "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d",
		"content_type": "text/plain", "file_info.%": "0", "server_side_encryption.#": "0",
	}}
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"bucket_id": "bucket-1", "file_name": "file.txt", "source": path,
	})
	resource := resourceB2BucketFileVersion()
	unchanged, err := resource.Diff(context.Background(), state, config, nil)
	if err != nil || (unchanged != nil && !unchanged.Empty()) {
		t.Fatalf("unchanged content must have no diff: diff=%#v, err=%v", unchanged, err)
	}
	if err := os.WriteFile(path, []byte("world"), 0o600); err != nil {
		t.Fatal(err)
	}
	diff, err := resource.Diff(context.Background(), state, config, nil)
	if err != nil || diff == nil || diff.Empty() || !diff.RequiresNew() {
		t.Fatalf("changed content must require replacement: diff=%#v, err=%v", diff, err)
	}
}
