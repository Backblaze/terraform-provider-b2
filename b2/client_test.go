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
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// stubBindings writes an executable that stands in for the Python bindings. It reads the request from
// stdin, then runs body, which is a POSIX shell snippet.
func stubBindings(t *testing.T, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the stub bindings are a POSIX shell script")
	}
	path := filepath.Join(t.TempDir(), "stub-bindings")
	script := "#!/bin/sh\ncat > \"$0.stdin\"\n" + body + "\n"
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func stubClient(t *testing.T, exec string) *Client {
	t.Helper()
	p := New("test", exec)()
	return &Client{
		Exec:             exec,
		ApplicationKeyId: "app-key-id-value",
		ApplicationKey:   "app-key-value",
		Endpoint:         "production",
		DataSourcesMap:   p.DataSourcesMap,
		ResourcesMap:     p.ResourcesMap,
	}
}

func TestApplyStopsWhenContextIsCancelled(t *testing.T) {
	// exec makes sleep the process itself, so that killing it also releases the stdout pipe
	exec := stubBindings(t, "exec sleep 20")
	client := stubClient(t, exec)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- client.Apply(ctx, OpResourceRead, &BucketInput{BucketId: "bucket"}, &BucketOutput{})
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected an error from the cancelled bindings call")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Apply kept waiting for the bindings process after its context was cancelled")
	}
}

func TestSanitizeOutputMasksSensitiveFields(t *testing.T) {
	// Top-level Sensitive fields: this is the behaviour that already works and must keep working
	schemaMap := map[string]*schema.Schema{
		"bucket_id": {Type: schema.TypeString},
		"token":     {Type: schema.TypeString, Sensitive: true},
	}
	type out struct {
		BucketId string `json:"bucketId"`
		Token    string `json:"token"`
	}
	got := sanitizeOutput(&out{BucketId: "b", Token: "secret"}, schemaMap)
	if got["bucket_id"] != "b" || got["token"] != "***" {
		t.Errorf("unexpected sanitized output: %v", got)
	}
}
