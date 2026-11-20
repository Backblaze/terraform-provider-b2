//####################################################################
//
// File: b2/bindings_lifecycle_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// C8: cancelling the context must stop the python child process and make Apply return a cancellation error.
// The test does not depend on timing: the fake child announces that it has started (its pid, written to a
// file) and then stays alive until it is killed; the timeouts below are only failure bounds for a hang.
func TestApplyStopsChildOnContextCancel_C8(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake child is an /bin/sh script")
	}
	const bound = 10 * time.Second
	pidFile := filepath.Join(t.TempDir(), "child.pid")
	client := newFakeClient(t, fmt.Sprintf(`echo $$ > %q.tmp; mv %q.tmp %q; exec sleep 300`, pidFile, pidFile, pidFile))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- client.Apply(ctx, OpDataSourceRead, &AccountInfoInput{}, nil) }()

	// wait until the child is really running (or Apply gave up before starting it)
	var pid int
	deadline := time.After(bound)
	for pid == 0 {
		select {
		case err := <-done:
			t.Fatalf("C8: Apply returned (err=%v) before the child process started", err)
		case <-deadline:
			t.Fatalf("C8: the child process did not start within %v", bound)
		case <-time.After(10 * time.Millisecond):
			if raw, err := os.ReadFile(pidFile); err == nil {
				pid, _ = strconv.Atoi(strings.TrimSpace(string(raw)))
			}
		}
	}
	childAlive := func() bool {
		proc, err := os.FindProcess(pid)
		return err == nil && proc.Signal(syscall.Signal(0)) == nil
	}
	t.Cleanup(func() { // never leave the sleeping child behind, whatever the outcome
		if proc, err := os.FindProcess(pid); err == nil {
			_ = proc.Kill()
		}
	})

	cancel()

	var err error
	select {
	case err = <-done:
	case <-time.After(bound):
		t.Fatalf("C8: Apply did not return within %v after the context was cancelled; the child (pid %d) keeps running", bound, pid)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("C8: Apply returned %v after cancellation, want an error wrapping context.Canceled", err)
	}
	exited := false
	for end := time.Now().Add(bound); time.Now().Before(end); time.Sleep(10 * time.Millisecond) {
		if !childAlive() {
			exited = true
			break
		}
	}
	if !exited {
		t.Errorf("C8: the child process (pid %d) is still running after Apply returned", pid)
	}
}

// C4 (Go side): when the file version no longer exists (python returns an empty result), refresh must
// drop the resource from state instead of keeping an empty object.
func TestFileVersionReadRemovesMissingFromState_C4(t *testing.T) {
	client := newFakeClient(t, `cat >/dev/null; echo '{"_sha1":"x"}'`)
	res := resourceB2BucketFileVersion()
	d := res.TestResourceData()
	d.SetId("4_zfile_missing")
	_ = d.Set("bucket_id", "b1")
	_ = d.Set("source", "/tmp/whatever")

	if diags := resourceB2BucketFileVersionRead(context.Background(), d, client); diags.HasError() {
		t.Fatal(diags)
	}
	if d.Id() != "" {
		t.Errorf("C4: file version that no longer exists was kept in state with id %q", d.Id())
	}
}

func TestFileVersionReadPreservesExistingStateAndErrors(t *testing.T) {
	for _, tc := range []struct {
		name, helper string
		wantError    bool
	}{
		{name: "existing", helper: "file-version-existing"},
		{name: "API error", helper: "stderr-plain", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GO_WANT_B2_CLIENT_HELPER", tc.helper)
			provider := New("test", os.Args[0])()
			client := &Client{Exec: os.Args[0], ResourcesMap: provider.ResourcesMap}
			d := resourceB2BucketFileVersion().TestResourceData()
			d.SetId("4_existing")
			if err := d.Set("bucket_id", "bucket-1"); err != nil {
				t.Fatal(err)
			}
			if err := d.Set("source", "source.txt"); err != nil {
				t.Fatal(err)
			}
			if err := d.Set("file_name", "file.txt"); err != nil {
				t.Fatal(err)
			}
			if err := d.Set("size", 3); err != nil {
				t.Fatal(err)
			}
			diags := resourceB2BucketFileVersionRead(context.Background(), d, client)
			if diags.HasError() != tc.wantError {
				t.Fatalf("diagnostics = %v, want error = %v", diags, tc.wantError)
			}
			if d.Id() != "4_existing" || d.Get("bucket_id") != "bucket-1" || d.Get("source") != "source.txt" {
				t.Fatal("read lost the existing ID, bucket, or source")
			}
			if d.Get("file_name") != "file.txt" || d.Get("size") != 3 {
				t.Fatal("read did not populate the existing file's attributes")
			}
		})
	}
}
