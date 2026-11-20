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
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
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

// C8: cancelling the context must stop the python child process and make Apply return a cancellation error.
// The test does not depend on timing: the fake child announces that it has started (its pid, written to a
// file) and then stays alive until it is killed; the timeouts below are only failure bounds for a hang.
func TestKnownDefect_C8_ApplyStopsChildOnContextCancel(t *testing.T) {
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

	// control: the same state, config and path with the original content must plan no change at all
	control, err := res.Diff(context.Background(), state, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if control != nil && (!control.Empty() || control.RequiresNew()) {
		t.Fatalf("control: unchanged content already produces a diff (requires new=%v): %v", control.RequiresNew(), control)
	}

	// only the content changes on disk; path and every config attribute stay the same
	if err := os.WriteFile(source, []byte("hello, changed content"), 0o600); err != nil {
		t.Fatal(err)
	}

	diff, err := res.Diff(context.Background(), state, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if diff == nil || diff.Empty() {
		t.Fatal("C15: changing the content of the source file yields no diff, so the new content is never uploaded")
	}
	if !diff.RequiresNew() {
		t.Errorf("C15: the content change does not require replacement (the new content would not be uploaded): %v", diff)
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

// C2: the extracted file must be byte-identical to the embedded one (no stale bytes appended). The defect
// shows only when the embedded size is not a multiple of the 2048-byte copy buffer, so the test does not use
// whatever binary happens to be embedded in this build: it copies bindings.go into a scratch module next to a
// deterministic 5000-byte fixture, and runs a nested test that extracts it in a fresh process (so the cached
// extraction path is empty and GetBindings must do the extraction itself).
func TestKnownDefect_C2_GetBindingsExtractsExactBytes(t *testing.T) {
	if testing.Short() {
		t.Skip("spawns a nested go test")
	}
	if runtime.GOOS == "windows" {
		t.Skip("not exercised on Windows")
	}
	scratch := t.TempDir()
	pkg := filepath.Join(scratch, "b2")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile("bindings.go")
	if err != nil {
		t.Fatal(err)
	}
	fixture := make([]byte, 5000) // 5000 = 2*2048 + 904: the last block is partial
	for i := range fixture {
		fixture[i] = byte(i*7 + i/251 + 1)
	}
	for name, data := range map[string][]byte{
		"bindings.go":              src,
		"py-terraform-provider-b2": fixture,
		"go.mod":                   []byte("module scratchbindings\n\ngo 1.21\n"),
		"extract_test.go": []byte(`package b2

import (
	"bytes"
	"os"
	"testing"
)

func TestExtractedBindingsEqualFixture(t *testing.T) {
	want, err := os.ReadFile("py-terraform-provider-b2")
	if err != nil || len(want) != 5000 {
		t.Fatalf("bad fixture: %d bytes, err=%v", len(want), err)
	}
	path, err := GetBindings()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("extracted binary has %d bytes, embedded fixture has %d", len(got), len(want))
	}
}
`),
	} {
		dir := pkg
		if name == "go.mod" {
			dir = scratch
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	tmp := t.TempDir() // the nested process extracts here, away from the real temp dir
	cmd := exec.Command("go", "test", "-count=1", "-run", "^TestExtractedBindingsEqualFixture$", "./b2")
	cmd.Dir = scratch
	cmd.Env = append(os.Environ(), "TMPDIR="+tmp)
	out, err := cmd.CombinedOutput()
	switch {
	case err == nil:
	case strings.Contains(string(out), "--- FAIL: TestExtractedBindingsEqualFixture"):
		t.Errorf("C2: GetBindings does not extract the embedded bytes exactly (nested go test):\n%s", out)
	default:
		t.Fatalf("nested go test could not run: %v\n%s", err, out)
	}
}

// C9: concurrent first use of GetBindings must be race-free and every caller must get the same working
// result. The race part only fails under the race detector; `make test-known-defects` runs with -race. When
// fixed, add -race to the default suite or keep this test in a job that runs with -race, because the data race
// is invisible to a plain run.
func TestKnownDefect_C9_GetBindingsConcurrentIsRaceFree(t *testing.T) {
	saved := bindings
	t.Cleanup(func() { bindings = saved })

	type result struct {
		path string
		err  error
	}
	const callers = 8
	for iter := 0; iter < 30; iter++ {
		bindings = nil
		results := make(chan result, callers)
		var wg sync.WaitGroup
		for i := 0; i < callers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				p, err := GetBindings()
				results <- result{p, err}
			}()
		}
		wg.Wait()
		close(results)

		var first string
		for r := range results {
			if r.err != nil {
				t.Errorf("C9: iteration %d: GetBindings failed: %v", iter, r.err)
				continue
			}
			if r.path == "" {
				t.Errorf("C9: iteration %d: GetBindings returned an empty path", iter)
				continue
			}
			if _, err := os.ReadFile(r.path); err != nil {
				t.Errorf("C9: iteration %d: extracted file %q is not readable: %v", iter, r.path, err)
			}
			if first == "" {
				first = r.path
			} else if r.path != first {
				t.Errorf("C9: iteration %d: concurrent callers got different paths: %q and %q", iter, first, r.path)
			}
			if saved == nil || r.path != *saved {
				defer func(path string) { _ = os.Remove(path) }(r.path)
			}
		}
	}
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
