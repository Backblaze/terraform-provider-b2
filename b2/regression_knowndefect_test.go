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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
