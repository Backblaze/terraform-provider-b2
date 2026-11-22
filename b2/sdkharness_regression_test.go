//####################################################################
//
// File: b2/sdkharness_regression_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm())
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestSDKHarnessContract must accept reordered rows and additional valid conformance scenarios. The test copies the contract test and the .sdkharness tree to a
// scratch module, adds a scenario row, and runs the contract test there (a nested `go test`).
func TestSDKHarnessContractToleratesNewScenario(t *testing.T) {
	requireCheckout(t)
	if testing.Short() {
		t.Skip("spawns a nested go test")
	}
	if runtime.GOOS == "windows" {
		t.Skip("not exercised on Windows")
	}
	tmp := t.TempDir()
	copyTree(t, filepath.Join(repoRoot, ".sdkharness"), filepath.Join(tmp, ".sdkharness"))
	if err := os.MkdirAll(filepath.Join(tmp, "b2"), 0o755); err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile("sdkharness_contract_test.go")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "b2", "sdkharness_contract_test.go"), src, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module scratchcontract\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tsvPath := filepath.Join(tmp, ".sdkharness", "tests.tsv")
	raw, err := os.ReadFile(tsvPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	newRow := "conformance\tzz.new_scenario\tsimulator\t./.sdkharness/tests/run-conformance"
	var out []string
	inserted := false
	for _, line := range lines {
		if !inserted && strings.HasPrefix(line, "health\t") {
			out = append(out, newRow)
			inserted = true
		}
		out = append(out, line)
	}
	if !inserted {
		t.Fatal("health row not found in tests.tsv")
	}
	for i, j := 1, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	if err := os.WriteFile(tsvPath, []byte(strings.Join(out, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "test", "-count=1", "-run", "^TestSDKHarnessContract$", "./b2")
	cmd.Dir = tmp
	if combined, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("C26: contract test fails after adding one valid conformance scenario row:\n%s", combined)
	}
}
