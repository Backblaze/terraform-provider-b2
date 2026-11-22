//####################################################################
//
// File: b2/bindings_regression_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func TestC20_GetBindingsOwnerOnlyPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits are not enforced on Windows")
	}
	path, err := GetBindings()
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Fatalf("extracted bindings mode = %04o, want 0700", got)
	}
}

// C2: the extracted file must be byte-identical to the embedded one (no stale bytes appended). The defect
// shows only when the embedded size is not a multiple of the 2048-byte copy buffer, so the test does not use
// whatever binary happens to be embedded in this build: it copies bindings.go into a scratch module next to a
// deterministic 5000-byte fixture, and runs a nested test that extracts it in a fresh process (so the cached
// extraction path is empty and GetBindings must do the extraction itself).
func TestC2_GetBindingsExtractsExactBytes(t *testing.T) {
	if testing.Short() {
		t.Skip("spawns a nested go test")
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
	cmd := exec.Command("go", "test", "-race", "-count=1", "-run", "^TestExtractedBindingsEqualFixture$", "./b2")
	cmd.Dir = scratch
	cmd.Env = append(os.Environ(), "TMPDIR="+tmp, "TMP="+tmp, "TEMP="+tmp, "GOWORK=off")
	out, err := cmd.CombinedOutput()
	switch {
	case err == nil:
	case strings.Contains(string(out), "--- FAIL: TestExtractedBindingsEqualFixture"):
		t.Errorf("C2: GetBindings does not extract the embedded bytes exactly (nested go test):\n%s", out)
	default:
		t.Fatalf("nested go test could not run: %v\n%s", err, out)
	}
}

// C9: concurrent first use must return the same readable file. Run with -race to
// additionally detect unsynchronized access to the cached path.
func TestC9_GetBindingsConcurrentIsRaceFree(t *testing.T) {
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
				t.Cleanup(func() { _ = os.Remove(r.path) })
			}
		}
	}
}
