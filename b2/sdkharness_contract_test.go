//####################################################################
//
// File: b2/sdkharness_contract_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//

package b2

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSDKHarnessContract(t *testing.T) {
	root := filepath.Join("..", ".sdkharness")
	contract, err := os.ReadFile(filepath.Join(root, "tests.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	normalizedContract := strings.ReplaceAll(string(contract), "\r\n", "\n")
	lines := strings.Split(strings.TrimSpace(normalizedContract), "\n")
	if len(lines) != 31 || lines[0] != "test_level\tscenario\ttarget\texecutable" {
		t.Fatalf("unexpected tests.tsv schema: %q", string(contract))
	}
	wantRow := "health\tgolden-path\tsimulator\t./.sdkharness/tests/health-golden-path"
	if lines[16] != wantRow {
		t.Fatalf("tests.tsv does not contain %q", wantRow)
	}
	seenScenarios := make(map[string]bool)
	for _, line := range lines[1:16] {
		fields := strings.Split(line, "\t")
		if len(fields) != 4 || fields[0] != "conformance" || fields[2] != "simulator" || fields[3] != "./.sdkharness/tests/run-conformance" {
			t.Fatalf("unexpected conformance contract row: %q", line)
		}
		if seenScenarios[fields[1]] {
			t.Fatalf("duplicate conformance scenario: %s", fields[1])
		}
		seenScenarios[fields[1]] = true
	}
	seenScenarios = make(map[string]bool)
	for _, line := range lines[17:] {
		fields := strings.Split(line, "\t")
		if len(fields) != 4 || fields[0] != "resilience" || fields[2] != "simulator" || fields[3] != "./.sdkharness/tests/run-resilience" {
			t.Fatalf("unexpected resilience contract row: %q", line)
		}
		if seenScenarios[fields[1]] {
			t.Fatalf("duplicate resilience scenario: %s", fields[1])
		}
		seenScenarios[fields[1]] = true
	}

	for _, executable := range []string{"health-golden-path", "run-conformance", "run-resilience"} {
		info, statErr := os.Stat(filepath.Join(root, "tests", executable))
		if statErr != nil {
			t.Fatal(statErr)
		}
		// NTFS has no POSIX exec bit, so Mode()&0o111 is always 0 on Windows. The exec
		// bit is a property of the git tree (100755), enforced by the Linux/macOS jobs;
		// on Windows we assert only that the script exists.
		if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
			t.Fatalf("%s is not executable", executable)
		}
	}
}

// The row-count test above breaks whenever a scenario is added. This one checks the shape of the
// contract instead: every row is well formed, scenarios are unique within a level, and every
// executable the contract names exists and is runnable.
func TestSDKHarnessContractRowsAreConsistent(t *testing.T) {
	root := filepath.Join("..", ".sdkharness")
	contract, err := os.ReadFile(filepath.Join(root, "tests.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(strings.ReplaceAll(string(contract), "\r\n", "\n")), "\n")
	if lines[0] != "test_level\tscenario\ttarget\texecutable" {
		t.Fatalf("unexpected tests.tsv header: %q", lines[0])
	}
	byLevel := map[string][][]string{}
	for _, line := range lines[1:] {
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			t.Fatalf("row does not have 4 columns: %q", line)
		}
		for _, f := range fields {
			if strings.TrimSpace(f) == "" || f != strings.TrimSpace(f) {
				t.Fatalf("empty or padded field in row: %q", line)
			}
		}
		byLevel[fields[0]] = append(byLevel[fields[0]], fields)
	}
	for _, level := range []string{"conformance", "health", "resilience"} {
		if len(byLevel[level]) == 0 {
			t.Errorf("no rows for test level %q", level)
		}
	}
	for level, rows := range byLevel {
		seen := map[string]bool{}
		for _, fields := range rows {
			scenario, executable := fields[1], fields[3]
			if seen[scenario] {
				t.Errorf("duplicate %s scenario: %s", level, scenario)
			}
			seen[scenario] = true
			info, statErr := os.Stat(filepath.Join("..", filepath.FromSlash(executable)))
			if statErr != nil {
				t.Errorf("%s/%s: executable %s: %v", level, scenario, executable, statErr)
				continue
			}
			if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
				t.Errorf("%s/%s: %s is not executable", level, scenario, executable)
			}
		}
	}
}
