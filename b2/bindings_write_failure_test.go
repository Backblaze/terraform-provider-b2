//####################################################################
//
// File: b2/bindings_write_failure_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// Build with a nonempty embedded fixture even when the repository uses the empty
// placeholder binary for offline tests. Limit writes in a subprocess so the
// extraction fails after CreateTemp without changing the parent process limits.
func TestGetBindingsRemovesPartialFileOnWriteError(t *testing.T) {
	switch runtime.GOOS {
	case "windows", "plan9", "js", "wasip1":
		t.Skip("file-size limit fixture requires a Unix platform")
	}
	dir := t.TempDir()
	source, err := os.ReadFile("bindings.go")
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		"go.mod":                   []byte("module cleanupfixture\n\ngo 1.21\n"),
		"bindings.go":              source,
		"py-terraform-provider-b2": bytes.Repeat([]byte("x"), 4096),
		"cleanup_test.go": []byte(`package b2
import ("path/filepath"; "syscall"; "testing")
func TestWriteFailure(t *testing.T) {
    tmp := t.TempDir()
    t.Setenv("TMPDIR", tmp)
    if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 1, Max: 1}); err != nil { t.Fatal(err) }
    if _, err := GetBindings(); err == nil { t.Fatal("expected extraction write failure") }
    leftovers, err := filepath.Glob(filepath.Join(tmp, "py-terraform-provider*"))
    if err != nil { t.Fatal(err) }
    if len(leftovers) != 0 { t.Fatalf("failed extraction left partial bindings: %v", leftovers) }
}
`),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command("go", "test", "-run", "^TestWriteFailure$", "-count=1", "-v", ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("write-error test failed: %v\n%s", err, out)
	}
}
