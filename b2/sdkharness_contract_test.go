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
	if len(lines) != 32 || lines[0] != "test_level\tscenario\ttarget\texecutable" {
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
		if info.Mode()&0o111 == 0 {
			t.Fatalf("%s is not executable", executable)
		}
	}
}
