//####################################################################
//
// File: b2/client_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	if os.Getenv("GO_WANT_B2_SIGNED_URL_HELPER") == "1" {
		_, _ = io.WriteString(os.Stdout, `{"bucketId":"bucket-123","fileName":"reports/file.txt","duration":3600,"signedUrl":"https://example.invalid/file?Authorization=bearer-credential"}`)
		os.Exit(0)
	}
}

func TestSanitizeOutputRedactsSensitiveInputValues(t *testing.T) {
	schemaMap := resourceB2BucketFileVersion().Schema
	schemaMap["top_secret"] = &schema.Schema{Type: schema.TypeString, Sensitive: true}

	input := map[string]interface{}{
		"bucket_id":  "bucket-123",
		"file_name":  "reports/file.txt",
		"top_secret": "top-secret",
		"server_side_encryption": []interface{}{
			map[string]interface{}{
				"mode":      "SSE-C",
				"algorithm": "AES256",
				"key": []interface{}{
					map[string]interface{}{
						"key_id":     "key-id",
						"secret_b64": "nested-secret",
					},
				},
			},
		},
	}

	safe := sanitizeOutput(input, schemaMap)
	if got := safe["top_secret"]; got != "***" {
		t.Fatalf("top-level sensitive value = %#v, want redacted", got)
	}
	if got := safe["bucket_id"]; got != "bucket-123" {
		t.Fatalf("non-sensitive bucket_id = %#v, want preserved", got)
	}

	encryption := safe["server_side_encryption"].([]interface{})[0].(map[string]interface{})
	key := encryption["key"].([]interface{})[0].(map[string]interface{})
	if got := key["secret_b64"]; got != "***" {
		t.Fatalf("nested sensitive value = %#v, want redacted", got)
	}
	if got := key["key_id"]; got != "key-id" {
		t.Fatalf("non-sensitive nested value = %#v, want preserved", got)
	}
}

func TestSanitizeOutputRedactsSignedURL(t *testing.T) {
	safe := sanitizeOutput(map[string]interface{}{
		"bucket_id":  "bucket-123",
		"file_name":  "reports/file.txt",
		"signed_url": "https://example.invalid/file?Authorization=bearer-credential",
	}, dataSourceB2BucketFileSignedUrl().Schema)

	if got := safe["signed_url"]; got != "***" {
		t.Fatalf("signed_url = %#v, want redacted", got)
	}
	if got := safe["file_name"]; got != "reports/file.txt" {
		t.Fatalf("non-sensitive file_name = %#v, want preserved", got)
	}
}

func TestApplyRedactsSensitiveInputLog(t *testing.T) {
	var logs bytes.Buffer
	ctx := tflogtest.RootLogger(context.Background(), &logs)
	t.Setenv("GO_WANT_B2_SIGNED_URL_HELPER", "1")

	input := BucketFileVersionInput{
		BucketId: "bucket-123",
		ServerSideEncryption: []interface{}{
			map[string]interface{}{
				"key": []interface{}{
					map[string]interface{}{
						"secret_b64": "nested-secret",
					},
				},
			},
		},
	}
	client := Client{
		Exec: os.Args[0],
		ResourcesMap: map[string]*schema.Resource{
			"b2_bucket_file_version": resourceB2BucketFileVersion(),
		},
	}

	if err := client.Apply(ctx, OpResourceCreate, &input, nil); err != nil {
		t.Fatalf("apply input: %s", err)
	}

	if got := logs.String(); strings.Contains(got, "nested-secret") {
		t.Fatalf("input log contains nested sensitive value: %s", got)
	}
	if got := logs.String(); !strings.Contains(got, `"secret_b64":"***"`) {
		t.Fatalf("input log does not contain redacted nested sensitive value: %s", got)
	}
}
