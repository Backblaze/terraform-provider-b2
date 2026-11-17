//####################################################################
//
// File: b2/regression_test.go
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
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Regression tests for defects that are already fixed. They use a fake child process (a shell script) in
// place of the PyInstaller binary, so no B2 credentials, network or Terraform CLI are needed. Tests for
// defects that are still present live in the *_knowndefect_test.go files (build tag "knowndefect").

// writeFakeExec writes an executable /bin/sh script that stands in for the python bindings.
func writeFakeExec(t *testing.T, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake shell child process is not available on Windows")
	}
	path := filepath.Join(t.TempDir(), "fake-pybindings")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func newFakeClient(t *testing.T, body string) *Client {
	t.Helper()
	exec := writeFakeExec(t, body)
	p := New("test", exec)()
	return &Client{
		Exec:           exec,
		DataSourcesMap: p.DataSourcesMap,
		ResourcesMap:   p.ResourcesMap,
	}
}

func signedUrlFakeBody() string {
	return `cat >/dev/null; echo '{"bucketId":"b1","fileName":"f.txt","duration":60,"signedUrl":"https://f.example.invalid/file/b/f.txt?Authorization=TOKEN123"}'`
}

// C1: Apply must not log secrets from the input at Debug level.
func TestApplyDoesNotLogInputSecrets_C1(t *testing.T) {
	const signingSecret = "SIGNING-SECRET-0123456789abcdef!"
	const appKey = "APP-KEY-NEVER-IN-LOGS"
	client := newFakeClient(t, `cat >/dev/null; echo '{"bucketId":"b1","notificationRules":[]}'`)
	client.ApplicationKey = appKey

	var logs bytes.Buffer
	ctx := tflogtest.RootLogger(context.Background(), &logs)

	input := &BucketNotificationRulesInput{
		BucketId: "b1",
		NotificationRules: []interface{}{map[string]interface{}{
			"name": "rule",
			"target_configuration": []interface{}{map[string]interface{}{
				"url":                        "https://example.invalid/hook",
				"hmac_sha256_signing_secret": signingSecret,
			}},
		}},
	}
	var output BucketNotificationRulesOutput
	if err := client.Apply(ctx, OpResourceCreate, input, &output); err != nil {
		t.Fatal(err)
	}

	for _, secret := range []string{signingSecret, appKey} {
		if strings.Contains(logs.String(), secret) {
			t.Errorf("C1: a secret value appears in the provider log (%d bytes of log)", len(logs.String()))
		}
	}
}

// C14: sensitive fields nested in lists/objects must be masked in the sanitized log copy.
func TestSanitizeForLogMasksNestedSecrets_C14(t *testing.T) {
	const signingSecret = "NESTED-SIGNING-SECRET-0123456789"
	output := &BucketNotificationRulesOutput{
		BucketId: "b1",
		NotificationRules: []NotificationRule{{
			Name: "rule",
			TargetConfiguration: &TargetConfiguration{
				Url:                     "https://example.invalid/hook",
				TargetType:              "webhook",
				HmacSha256SigningSecret: signingSecret,
			},
		}},
	}
	schemaMap := resourceB2BucketNotificationRules().Schema

	safe := (&logSanitizer{}).sanitizeForLog(output, schemaMap)

	if strings.Contains(fmt.Sprintf("%v", safe), signingSecret) {
		t.Errorf("C14: nested sensitive hmac_sha256_signing_secret is not masked in the sanitized copy: %v", safe)
	}
}

// C5: the data source id is shown in plan/state output, so it must not contain the URL query string.
func TestSignedUrlIsNotUsedAsID_C5(t *testing.T) {
	client := newFakeClient(t, signedUrlFakeBody())
	res := dataSourceB2BucketFileSignedUrl()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"bucket_id": "b1", "file_name": "f.txt", "duration": 60,
	})
	if diags := dataSourceB2BucketFileSignedUrlRead(context.Background(), d, client); diags.HasError() {
		t.Fatal(diags)
	}
	if strings.ContainsAny(d.Id(), "?=") || strings.Contains(d.Id(), "TOKEN123") {
		t.Errorf("C5: data source id contains the URL query string: %q", d.Id())
	}
}
