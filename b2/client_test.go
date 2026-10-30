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
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	switch os.Getenv("GO_WANT_B2_CLIENT_HELPER") {
	case "notification-rules":
		input, err := io.ReadAll(os.Stdin)
		if err != nil || !bytes.Contains(input, []byte("Bearer custom-header-sentinel")) {
			os.Exit(2)
		}
		_, _ = io.WriteString(os.Stdout, `{"bucketId":"bucket-123","notificationRules":[{"name":"rule","targetConfiguration":{"targetType":"webhook","customHeaders":[{"name":"X-Webhook-Token","value":"Bearer custom-header-sentinel"}]}}]}`)
		os.Exit(0)
	case "stderr":
		_, _ = io.Copy(os.Stderr, os.Stdin)
		os.Exit(1)
	}
	if os.Getenv("GO_WANT_B2_SIGNED_URL_HELPER") == "1" {
		_, _ = io.WriteString(os.Stdout, `{"bucketId":"bucket-123","fileName":"reports/file.txt","duration":3600,"signedUrl":"https://example.invalid/file?Authorization=bearer-credential"}`)
		os.Exit(0)
	}
}

func TestSanitizeForLogRedactsSensitiveInputValues(t *testing.T) {
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

	sanitizer := logSanitizer{}
	safe := sanitizer.sanitizeForLog(input, schemaMap)
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

func TestSanitizeForLogRedactsSignedURL(t *testing.T) {
	sanitizer := logSanitizer{}
	safe := sanitizer.sanitizeForLog(map[string]interface{}{
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

func TestApplyRedactsCustomHeaderLogs(t *testing.T) {
	var logs bytes.Buffer
	ctx := tflogtest.RootLogger(context.Background(), &logs)
	t.Setenv("GO_WANT_B2_CLIENT_HELPER", "notification-rules")
	input := BucketNotificationRulesInput{
		BucketId: "bucket-123",
		NotificationRules: []interface{}{
			map[string]interface{}{
				"name": "rule",
				"target_configuration": []interface{}{
					map[string]interface{}{
						"target_type": "webhook",
						"custom_headers": []interface{}{
							map[string]interface{}{
								"name":  "X-Webhook-Token",
								"value": "Bearer custom-header-sentinel",
							},
						},
					},
				},
			},
		},
	}
	client := Client{
		Exec: os.Args[0],
		ResourcesMap: map[string]*schema.Resource{
			"b2_bucket_notification_rules": resourceB2BucketNotificationRules(),
		},
	}
	var output BucketNotificationRulesOutput
	if err := client.Apply(ctx, OpResourceCreate, &input, &output); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(logs.String(), "custom-header-sentinel") {
		t.Error("debug log contains custom header value sentinel")
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
		var record map[string]interface{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		for _, field := range []string{"input", "output"} {
			if value, ok := record[field]; ok {
				seen[field] = true
				rules := value.(map[string]interface{})["notification_rules"].([]interface{})
				target := rules[0].(map[string]interface{})["target_configuration"].([]interface{})
				headers := target[0].(map[string]interface{})["custom_headers"].([]interface{})
				header := headers[0].(map[string]interface{})
				if header["value"] != "***" || header["name"] != "X-Webhook-Token" {
					t.Errorf("%s log must mask header value and preserve header name", field)
				}
			}
		}
	}
	if !seen["input"] || !seen["output"] {
		t.Error("expected both input and output debug records")
	}
	if value := output.NotificationRules[0].TargetConfiguration.CustomHeaders[0].Value; value != "Bearer custom-header-sentinel" {
		t.Error("logging changed the returned custom header value")
	}
}

func TestApplyRedactsBindingStderr(t *testing.T) {
	cases := []struct {
		name     string
		input    ResourceSchema
		resource *schema.Resource
		secrets  []string
	}{
		{
			name: "SSE-C key",
			input: &BucketFileVersionInput{
				BucketId: "bucket-123",
				ServerSideEncryption: []interface{}{
					map[string]interface{}{
						"key": []interface{}{
							map[string]interface{}{"secret_b64": "sse-key-sentinel"},
						},
					},
				},
			},
			resource: resourceB2BucketFileVersion(),
			secrets:  []string{"sse-key-sentinel"},
		},
		{
			name: "webhook headers and signing secret",
			input: &BucketNotificationRulesInput{
				BucketId: "bucket-123",
				NotificationRules: []interface{}{
					map[string]interface{}{
						"target_configuration": []interface{}{
							map[string]interface{}{
								"hmac_sha256_signing_secret": "hmac-secret-sentinel",
								"custom_headers": []interface{}{
									map[string]interface{}{
										"name":  "X-Webhook-Token",
										"value": "Bearer header-\"quoted\"\nsentinel",
									},
								},
							},
						},
					},
				},
			},
			resource: resourceB2BucketNotificationRules(),
			secrets:  []string{"hmac-secret-sentinel", "header-"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var logs bytes.Buffer
			ctx := tflogtest.RootLogger(context.Background(), &logs)
			t.Setenv("GO_WANT_B2_CLIENT_HELPER", "stderr")
			client := Client{
				Exec:             os.Args[0],
				ApplicationKey:   "provider-key-sentinel",
				ApplicationKeyId: "provider-id-sentinel",
				ResourcesMap: map[string]*schema.Resource{
					"b2_" + tc.input.ResourceName(): tc.resource,
				},
			}
			err := client.Apply(ctx, OpResourceCreate, tc.input, nil)
			if err == nil {
				t.Fatal("expected a binding error")
			}
			for _, secret := range append(tc.secrets, "provider-key-sentinel", "provider-id-sentinel") {
				if strings.Contains(err.Error(), secret) || strings.Contains(logs.String(), secret) {
					t.Errorf("binding error or log contains %s", secret)
				}
			}
			if !strings.Contains(err.Error(), "bucket-123") || !strings.Contains(logs.String(), "Error in pybindings") {
				t.Error("binding error must retain non-sensitive context and be logged")
			}
		})
	}
}

func TestLogSanitizerRedactsEscapedAndOverlappingValues(t *testing.T) {
	sanitizer := logSanitizer{}
	sanitizer.mask("secret")
	sanitizer.mask("secret-with-suffix")
	sanitizer.mask("Bearer header-\"quoted\"\nsentinel")
	sanitizer.mask("single'quote\nsecret")
	cases := []struct {
		text string
		want string
	}{
		{"failure: secret-with-suffix", "failure: ***"},
		{`{"value":"Bearer header-\"quoted\"\nsentinel"}`, `{"value":"***"}`},
		{`InvalidHeader('Bearer header-"quoted"\nsentinel')`, `InvalidHeader('***')`},
		{`ValueError('single\'quote\nsecret')`, `ValueError('***')`},
	}
	for _, tc := range cases {
		if got := sanitizer.redactText(tc.text); got != tc.want {
			t.Errorf("redacted error = %q, want %q", got, tc.want)
		}
	}
}
