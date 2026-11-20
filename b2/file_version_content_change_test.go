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
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestFileVersionContentChangePlansReplacement(t *testing.T) {
	// This default-suite test replaces the formerly opt-in known-defect case.
	t.Run("TestKnownDefect_C15_FileVersionContentChangeProducesDiff", func(t *testing.T) {
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
	})
}

func TestFileVersionLargeFileContentChangePlansReplacement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	state := &terraform.InstanceState{ID: "file-1", Attributes: map[string]string{
		"id": "file-1", "file_id": "file-1", "bucket_id": "bucket-1",
		"file_name": "file.txt", "source": path, "content_sha1": "none",
		"file_info.%": "2", "file_info.author": "writer",
		"file_info.large_file_sha1": "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d",
		"server_side_encryption.#":  "0",
	}, RawConfig: cty.ObjectVal(map[string]cty.Value{
		"file_info": cty.MapVal(map[string]cty.Value{"author": cty.StringVal("writer")}),
	})}
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"bucket_id": "bucket-1", "file_name": "file.txt", "source": path,
		"file_info": map[string]interface{}{"author": "writer"},
	})
	resource := resourceB2BucketFileVersion()
	for i := 0; i < 2; i++ {
		diff, err := resource.Diff(context.Background(), state, config, nil)
		if err != nil || (diff != nil && !diff.Empty()) {
			t.Fatalf("unchanged large file must remain stable: diff=%#v, err=%v", diff, err)
		}
	}
	if err := os.WriteFile(path, []byte("world"), 0o600); err != nil {
		t.Fatal(err)
	}
	diff, err := resource.Diff(context.Background(), state, config, nil)
	if err != nil || diff == nil || !diff.RequiresNew() {
		t.Fatalf("changed large file must require replacement: diff=%#v, err=%v", diff, err)
	}
}

func TestFileVersionUnavailableContentPreservesExistingVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(path, []byte("world"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source, checksum string
	}{
		{"missing source and checksum", path + ".missing", "none"},
		{"missing source with checksum", path + ".missing", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{"directory", dir, "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{"missing checksum", path, "none"},
		{"invalid checksum", path, "invalid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := &terraform.InstanceState{ID: "file-1", Attributes: map[string]string{
				"id": "file-1", "file_id": "file-1", "bucket_id": "bucket-1",
				"file_name": "file.txt", "source": tc.source, "content_sha1": tc.checksum,
				"file_info.%": "0", "server_side_encryption.#": "0",
			}}
			config := terraform.NewResourceConfigRaw(map[string]interface{}{
				"bucket_id": "bucket-1", "file_name": "file.txt", "source": tc.source,
			})
			diff, err := resourceB2BucketFileVersion().Diff(context.Background(), state, config, nil)
			if err != nil || (diff != nil && !diff.Empty()) {
				t.Fatalf("unavailable content must not replace the version: diff=%#v, err=%v", diff, err)
			}
		})
	}
}

func TestFileVersionPlanHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	state := &terraform.InstanceState{ID: "file-1", Attributes: map[string]string{
		"id": "file-1", "file_id": "file-1", "bucket_id": "bucket-1",
		"file_name": "file.txt", "source": "source.txt",
		"content_sha1": "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d",
		"file_info.%":  "0", "server_side_encryption.#": "0",
	}}
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"bucket_id": "bucket-1", "file_name": "file.txt", "source": "source.txt",
	})
	_, err := resourceB2BucketFileVersion().Diff(ctx, state, config, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled plan must return context.Canceled, got %v", err)
	}
}

func TestFileVersionUnreadableSourcePreservesExistingVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(path, []byte("world"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
	if file, err := os.Open(path); err == nil {
		_ = file.Close()
		t.Skip("file permissions are not enforced for this user or platform")
	}
	state := &terraform.InstanceState{ID: "file-1", Attributes: map[string]string{
		"id": "file-1", "file_id": "file-1", "bucket_id": "bucket-1",
		"file_name": "file.txt", "source": path,
		"content_sha1": "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d",
		"file_info.%":  "0", "server_side_encryption.#": "0",
	}}
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"bucket_id": "bucket-1", "file_name": "file.txt", "source": path,
	})
	diff, err := resourceB2BucketFileVersion().Diff(context.Background(), state, config, nil)
	if err != nil || (diff != nil && !diff.Empty()) {
		t.Fatalf("unreadable source must not replace the version: diff=%#v, err=%v", diff, err)
	}
}

type cancelOnRead func([]byte) (int, error)

func (f cancelOnRead) Read(p []byte) (int, error) { return f(p) }

func TestFileVersionReadStopsAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reader := &fileVersionContextReader{ctx: ctx, reader: cancelOnRead(func(p []byte) (int, error) {
		cancel()
		return copy(p, "a"), nil
	})}
	_, err := io.Copy(io.Discard, reader)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled read must return context.Canceled, got %v", err)
	}
}
