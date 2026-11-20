//####################################################################
//
// File: b2/data_source_b2_bucket_file_signed_url_duration_test.go
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
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Use the test binary as the bindings process to capture the real JSON request on every OS.
func init() {
	capture := os.Getenv("GO_WANT_B2_DURATION_CAPTURE")
	if capture == "" {
		return
	}
	var input BucketFileSignedUrlInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		os.Exit(2)
	}
	raw, err := json.Marshal(input)
	if err != nil {
		os.Exit(2)
	}
	if err := os.WriteFile(capture, raw, 0o600); err != nil {
		os.Exit(2)
	}
	if err := json.NewEncoder(os.Stdout).Encode(BucketFileSignedUrlOutput{
		BucketFileSignedUrlInput: input,
		SignedUrl:                "https://example.invalid/file?Authorization=test-token",
	}); err != nil {
		os.Exit(2)
	}
	os.Exit(0)
}

func TestDataSourceB2BucketFileSignedURLOmittedDuration(t *testing.T) {
	testSignedURLRequestDuration(t, nil, 86400)
}

func TestDataSourceB2BucketFileSignedURLExplicitDuration(t *testing.T) {
	testSignedURLRequestDuration(t, 3600, 3600)
}

func testSignedURLRequestDuration(t *testing.T, duration interface{}, want int) {
	t.Helper()
	res := dataSourceB2BucketFileSignedUrl()
	config := map[string]interface{}{
		"bucket_id": "bucket-123",
		"file_name": "reports/file.txt",
	}
	if duration != nil {
		config["duration"] = duration
	}
	d := schema.TestResourceDataRaw(t, res.Schema, config)
	capture := filepath.Join(t.TempDir(), "request.json")
	t.Setenv("GO_WANT_B2_DURATION_CAPTURE", capture)
	client := &Client{
		Exec:           os.Args[0],
		DataSourcesMap: map[string]*schema.Resource{"b2_bucket_file_signed_url": res},
	}
	if diags := res.ReadContext(context.Background(), d, client); diags.HasError() {
		t.Fatalf("read diagnostics: %v", diags)
	}
	raw, err := os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	var request map[string]interface{}
	if err := json.Unmarshal(raw, &request); err != nil {
		t.Fatal(err)
	}
	if got := request["duration"]; got != float64(want) {
		t.Fatalf("duration sent to bindings = %v, want %v seconds", got, want)
	}
	if got := d.Get("duration"); got != want {
		t.Errorf("duration in state = %v, want %v", got, want)
	}
}
