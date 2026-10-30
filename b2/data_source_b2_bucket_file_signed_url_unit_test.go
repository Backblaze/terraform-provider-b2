//####################################################################
//
// File: b2/data_source_b2_bucket_file_signed_url_unit_test.go
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
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceB2BucketFileSignedURLUsesNonCredentialID(t *testing.T) {
	resource := dataSourceB2BucketFileSignedUrl()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"bucket_id": "bucket-123",
		"file_name": "reports/file.txt",
		"duration":  3600,
	})
	t.Setenv("GO_WANT_B2_SIGNED_URL_HELPER", "1")

	diagnostics := dataSourceB2BucketFileSignedUrlRead(context.Background(), d, &Client{
		Exec: os.Args[0],
		DataSourcesMap: map[string]*schema.Resource{
			"b2_bucket_file_signed_url": resource,
		},
	})
	if diagnostics.HasError() {
		t.Fatalf("read diagnostics: %#v", diagnostics)
	}

	id := d.Id()
	if strings.Contains(id, "Authorization") || strings.Contains(id, "https://") {
		t.Fatalf("data source ID contains signed URL data: %q", id)
	}
	if want := bucketFileSignedURLID("bucket-123", "reports/file.txt", 3600); id != want {
		t.Fatalf("data source ID = %q, want %q", id, want)
	}
}
