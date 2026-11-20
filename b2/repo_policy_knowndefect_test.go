//go:build knowndefect

//####################################################################
//
// File: b2/repo_policy_knowndefect_test.go
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
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// Known-defect repository checks. Compiled only with the "knowndefect" build tag (see regression_knowndefect_test.go
// for how to run them and how to move a test into the default suite once its defect is fixed). The workflow
// checks are text heuristics: they pin the shape of the files they read, so adjust them with the fix.

// C25: CI must run Python unit tests, and must not stay green when the acceptance secrets are missing.
// The checks parse ci.yml and inspect executable `run` steps (comments and step names do not count; see the
// fixtures in ci_workflow_check_test.go). A secrets guard is accepted when a step fails with `exit 1` for each
// of B2_TEST_APPLICATION_KEY and B2_TEST_APPLICATION_KEY_ID being empty: an `if:` on the step or a shell guard
// inside its script. Adjust ciWorkflowFailsOnMissingKeys if the fix uses another form.
func TestKnownDefect_C25_CIRunsPythonTestsAndFailsWithoutSecrets(t *testing.T) {
	ci := readRepoFile(t, ".github", "workflows", "ci.yml")
	t.Run("python unit tests run in CI", func(t *testing.T) {
		ok, err := ciWorkflowRunsPythonTests(ci)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Error("C25: no executable ci.yml step runs Python tests (pytest); python-bindings only gets ruff")
		}
	})
	t.Run("missing acceptance secrets fail loudly", func(t *testing.T) {
		ok, err := ciWorkflowFailsOnMissingKeys(ci)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Error("C25: no executable ci.yml step fails (exit 1) when B2_TEST_APPLICATION_KEY or B2_TEST_APPLICATION_KEY_ID is missing; testacc is skipped and the job stays green")
		}
	})
}

// C27: the Go version stated in the .sdkharness README must match go.mod.
func TestKnownDefect_C27_SdkharnessReadmeGoVersionMatchesGoMod(t *testing.T) {
	goMod := regexp.MustCompile(`(?m)^go (\d+\.\d+)`).FindStringSubmatch(readRepoFile(t, "go.mod"))
	readme := regexp.MustCompile(`go\.mod` + "`" + ` \((\d+\.\d+)\.x\)`).FindStringSubmatch(readRepoFile(t, ".sdkharness", "README.md"))
	if goMod == nil || readme == nil {
		t.Fatalf("could not parse Go versions: go.mod=%v readme=%v", goMod, readme)
	}
	if goMod[1] != readme[1] {
		t.Errorf("C27: .sdkharness/README.md says Go %s.x but go.mod declares %s", readme[1], goMod[1])
	}
}

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

// C26: TestSDKHarnessContract must not pin tests.tsv by row position/count. Adding a legitimate conformance
// scenario must keep the contract test green. The test copies the contract test and the .sdkharness tree to a
// scratch module, adds a scenario row, and runs the contract test there (a nested `go test`).
func TestKnownDefect_C26_SDKHarnessContractToleratesNewScenario(t *testing.T) {
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
	if err := os.WriteFile(tsvPath, []byte(strings.Join(out, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "test", "-count=1", "-run", "^TestSDKHarnessContract$", "./b2")
	cmd.Dir = tmp
	if combined, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("C26: contract test fails after adding one valid conformance scenario row:\n%s", combined)
	}
}
