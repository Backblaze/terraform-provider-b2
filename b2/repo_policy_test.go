//####################################################################
//
// File: b2/repo_policy_test.go
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
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// Static repository checks. Like sdkharness_contract_test.go they read files of the repository next to the
// package, and they are skipped when the package is not tested from a full checkout.

const repoRoot = ".."

func requireCheckout(t *testing.T) {
	t.Helper()
	for _, rel := range []string{"go.mod", ".sdkharness"} {
		if _, err := os.Stat(filepath.Join(repoRoot, rel)); err != nil {
			t.Skipf("not run from a full repository checkout (%s not found)", rel)
		}
	}
}

func readRepoFile(t *testing.T, rel ...string) string {
	t.Helper()
	requireCheckout(t)
	data, err := os.ReadFile(filepath.Join(append([]string{repoRoot}, rel...)...))
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}

// C28 (guard): the retry behaviour named in the changelog is covered by scenarios that are registered in
// tests.tsv and present as executable scripts. This pins the scenario names and the tests.tsv column layout.
func TestRetryScenariosAreRegisteredAndPresent_C28(t *testing.T) {
	tsv := readRepoFile(t, ".sdkharness", "tests.tsv")
	for _, scenario := range []string{"api.retry_after_503", "upload.retry_408"} {
		if !strings.Contains(tsv, "resilience\t"+scenario+"\tsimulator\t./.sdkharness/tests/run-resilience") {
			t.Errorf("C28: scenario %s is not registered in .sdkharness/tests.tsv", scenario)
		}
		info, err := os.Stat(filepath.Join(repoRoot, ".sdkharness", "tests", "resilience", scenario))
		if err != nil {
			t.Errorf("C28: %v", err)
			continue
		}
		if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
			t.Errorf("C28: %s is not executable", scenario)
		}
	}
}

// C29 (guard): every resilience script names the scenario it implements in its header, the name matches the
// file name, and the scenario is registered in tests.tsv. This pins the docstring header format.
func TestResilienceScriptHeadersResolveToRegisteredScenarios_C29(t *testing.T) {
	tsv := readRepoFile(t, ".sdkharness", "tests.tsv")
	dir := filepath.Join(repoRoot, ".sdkharness", "tests", "resilience")
	header := regexp.MustCompile(`(?m)^"""terraform-provider-b2 x (\S+?)\.$`)
	count := 0
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		count++
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		m := header.FindStringSubmatch(string(data))
		if m == nil {
			t.Errorf("C29: %s has no scenario header", d.Name())
			return nil
		}
		if m[1] != d.Name() {
			t.Errorf("C29: %s header names scenario %q", d.Name(), m[1])
		}
		if !strings.Contains(tsv, "resilience\t"+m[1]+"\t") {
			t.Errorf("C29: scenario %s is not registered in tests.tsv", m[1])
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("no resilience scripts found")
	}
}
