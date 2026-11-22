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
	if len(lines) == 0 || lines[0] != "test_level\tscenario\ttarget\texecutable" {
		t.Fatalf("unexpected tests.tsv schema: %q", string(contract))
	}
	seen := make(map[string]bool)
	levels := make(map[string]bool)
	for _, line := range lines[1:] {
		fields := strings.Split(line, "\t")
		if len(fields) != 4 || fields[1] == "" || fields[2] != "simulator" {
			t.Fatalf("unexpected contract row: %q", line)
		}
		key := fields[0] + "\t" + fields[1]
		if seen[key] {
			t.Fatalf("duplicate scenario: %s", key)
		}
		seen[key] = true
		levels[fields[0]] = true
		switch fields[0] {
		case "health":
			if fields[1] != "golden-path" || fields[3] != "./.sdkharness/tests/health-golden-path" {
				t.Fatalf("unexpected health contract row: %q", line)
			}
		case "conformance", "resilience":
			if fields[3] != "./.sdkharness/tests/run-"+fields[0] {
				t.Fatalf("unexpected %s contract row: %q", fields[0], line)
			}
		default:
			t.Fatalf("unexpected contract row: %q", line)
		}
	}
	for _, level := range []string{"health", "conformance", "resilience"} {
		if !levels[level] {
			t.Fatalf("missing %s contract row", level)
		}
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
