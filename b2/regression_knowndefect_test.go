//go:build knowndefect

//####################################################################
//
// File: b2/regression_knowndefect_test.go
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
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Known-defect tests: each test describes the correct behaviour and fails today because the defect is still
// present. They are compiled only with the "knowndefect" build tag and are run with `make test-known-defects`;
// the default `make test` does not build them.
//
// When a defect is fixed: delete the "TestKnownDefect_" prefix from the test name and move the test to a file
// without the build tag (for example regression_test.go) so that it joins the default suite.
//
// Several tests use the fake child process from regression_test.go in place of the PyInstaller binary.

// C3: with duration omitted, the provider must not send 0 seconds to B2. The expected default is any
// positive value; the owner may pick the exact value.
func TestKnownDefect_C3_SignedUrlOmittedDurationIsNotZero(t *testing.T) {
	capture := filepath.Join(t.TempDir(), "stdin.json")
	client := newFakeClient(t, fmt.Sprintf(`cat > %q; echo '{"bucketId":"b1","fileName":"f.txt","duration":3600,"signedUrl":"https://f.example.invalid/f"}'`, capture))
	res := dataSourceB2BucketFileSignedUrl()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"bucket_id": "b1", "file_name": "f.txt",
	})
	if diags := dataSourceB2BucketFileSignedUrlRead(context.Background(), d, client); diags.HasError() {
		t.Fatal(diags)
	}
	raw, err := os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	var sent map[string]interface{}
	if err := json.Unmarshal(raw, &sent); err != nil {
		t.Fatal(err)
	}
	if dur, _ := sent["duration"].(float64); dur <= 0 {
		t.Errorf("C3: omitted duration was sent to the python layer as %v (want a positive default)", sent["duration"])
	}
}

// C8: cancelling the context must stop the python child process.
func TestKnownDefect_C8_ApplyStopsChildOnContextCancel(t *testing.T) {
	client := newFakeClient(t, `exec sleep 2`)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := client.Apply(ctx, OpDataSourceRead, &AccountInfoInput{}, nil)
	elapsed := time.Since(start)

	if elapsed > time.Second {
		t.Errorf("C8: Apply kept running %v after the context was cancelled at 200ms (err=%v)", elapsed, err)
	}
}

// C10: the deprecated allowed.bucket_id / bucket_name must be set from the first bucket.
func TestKnownDefect_C10_AccountInfoPopulatesDeprecatedBucketFields(t *testing.T) {
	client := newFakeClient(t, `cat >/dev/null; echo '{"accountId":"acc1","accountAuthToken":"tok","apiUrl":"https://api.example.invalid","allowed":[{"buckets":[{"id":"bid1","name":"bname1"}],"bucketId":"","bucketName":"","capabilities":["listBuckets"],"namePrefix":""}],"downloadUrl":"d","s3ApiUrl":"s","recommendedPartSize":1,"absoluteMinimumPartSize":1}'`)
	d := schema.TestResourceDataRaw(t, dataSourceB2AccountInfo().Schema, map[string]interface{}{})

	if diags := dataSourceB2AccountInfoRead(context.Background(), d, client); diags.HasError() {
		t.Fatal(diags)
	}

	if got := d.Get("allowed.0.bucket_id"); got != "bid1" {
		t.Errorf("C10: allowed.0.bucket_id = %q, want %q", got, "bid1")
	}
	if got := d.Get("allowed.0.bucket_name"); got != "bname1" {
		t.Errorf("C10: allowed.0.bucket_name = %q, want %q", got, "bname1")
	}
}

// C4 (Go side): when the file version no longer exists (python returns an empty result), refresh must
// drop the resource from state instead of keeping an empty object.
func TestKnownDefect_C4_FileVersionReadRemovesMissingFromState(t *testing.T) {
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

// C15: changing the content of the local source file must produce a diff (replacement).
func TestKnownDefect_C15_FileVersionContentChangeProducesDiff(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(source, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha1.Sum([]byte("hello"))
	oldSha1 := hex.EncodeToString(sum[:])

	res := resourceB2BucketFileVersion()
	state := &terraform.InstanceState{
		ID: "4_zfile1",
		Attributes: map[string]string{
			"id": "4_zfile1", "bucket_id": "b1", "file_name": "f.txt", "source": source,
			"content_sha1": oldSha1, "file_id": "4_zfile1", "action": "upload",
			"content_type": "text/plain", "file_info.%": "0", "server_side_encryption.#": "0",
		},
	}
	cfg := terraform.NewResourceConfigRaw(map[string]interface{}{
		"bucket_id": "b1", "file_name": "f.txt", "source": source,
	})

	// the content changes on disk; path and every config attribute stay the same
	if err := os.WriteFile(source, []byte("hello, changed content"), 0o600); err != nil {
		t.Fatal(err)
	}

	diff, err := res.Diff(context.Background(), state, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if diff == nil || diff.Empty() {
		t.Error("C15: changing the content of the source file yields no diff, so the new content is never uploaded")
	}
}

// C19: a non-executable / noexec extracted binary must give an actionable error, not a raw exec error.
// The expected hint (the message mentions TMPDIR or noexec) is an assumption; adjust it to the chosen message.
func TestKnownDefect_C19_ApplyNonExecutableBindingsGivesActionableError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("mode bits are not meaningful on Windows")
	}
	path := filepath.Join(t.TempDir(), "py-terraform-provider-b2")
	if err := os.WriteFile(path, []byte("not executable"), 0o600); err != nil {
		t.Fatal(err)
	}
	provider := New("test", path)()
	client := &Client{Exec: path, DataSourcesMap: provider.DataSourcesMap, ResourcesMap: provider.ResourcesMap}

	err := client.Apply(context.Background(), OpDataSourceRead, &AccountInfoInput{}, nil)
	if err == nil {
		t.Fatal("expected an error executing a non-executable file")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "tmpdir") && !strings.Contains(msg, "noexec") {
		t.Errorf("C19: raw exec error without any hint about TMPDIR/noexec: %q", err.Error())
	}
}

// embeddedBindings returns the embedded file. It skips when that is the empty placeholder used by
// `make test`; `make test-known-defects` fills the placeholder with bytes for the duration of the run.
func embeddedBindings(t *testing.T) []byte {
	t.Helper()
	data, err := content.ReadFile("py-terraform-provider-b2")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Skip("embedded py-terraform-provider-b2 is the empty placeholder; fill it with bytes or build the real binary")
	}
	return data
}

// C2: the extracted file must be byte-identical to the embedded one (no stale bytes appended). The embedded
// size must not be a multiple of the 2048-byte copy buffer for the defect to show.
func TestKnownDefect_C2_GetBindingsExtractsExactBytes(t *testing.T) {
	want := embeddedBindings(t)
	path, err := GetBindings()
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("C2: extracted binary has %d bytes, embedded has %d", len(got), len(want))
	}
}

// C9: concurrent first use of GetBindings must be race-free. This only fails under the race detector;
// `make test-known-defects` runs with -race. When fixed, add -race to the default suite or keep this test
// in a job that runs with -race, because a plain run passes either way.
func TestKnownDefect_C9_GetBindingsConcurrentIsRaceFree(t *testing.T) {
	saved := bindings
	t.Cleanup(func() { bindings = saved })

	var created sync.Map
	for iter := 0; iter < 30; iter++ {
		bindings = nil
		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if p, err := GetBindings(); err == nil {
					created.Store(p, true)
				}
			}()
		}
		wg.Wait()
	}
	created.Range(func(k, _ interface{}) bool {
		if k.(string) != "" && (saved == nil || k.(string) != *saved) {
			_ = os.Remove(k.(string))
		}
		return true
	})
}

// C18: running the provider's own unit tests must not leave the extracted binary in the temp dir.
// Only the test-leak part is covered; an error-path leak in GetBindings is not.
func TestKnownDefect_C18_ProviderTestsLeaveNoExtractedBinary(t *testing.T) {
	if testing.Short() {
		t.Skip("spawns a nested go test")
	}
	if runtime.GOOS == "windows" {
		t.Skip("not exercised on Windows")
	}
	tmp := t.TempDir()
	cmd := exec.Command("go", "test", "-count=1", "-run", "^TestProvider$", ".")
	cmd.Env = append(os.Environ(), "TMPDIR="+tmp)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("nested go test failed: %v\n%s", err, out)
	}
	leftovers, _ := filepath.Glob(filepath.Join(tmp, "py-terraform-provider*"))
	if len(leftovers) > 0 {
		t.Errorf("C18: TestProvider left extracted binaries behind: %v", leftovers)
	}
}
