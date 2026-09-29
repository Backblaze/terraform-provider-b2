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

func TestSDKHarnessHealthContract(t *testing.T) {
	root := filepath.Join("..", ".sdkharness")
	contract, err := os.ReadFile(filepath.Join(root, "tests.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(contract)), "\n")
	if len(lines) != 2 || lines[0] != "test_level\tscenario\ttarget\texecutable" {
		t.Fatalf("unexpected tests.tsv schema: %q", string(contract))
	}
	wantRow := "health\tgolden-path\tsimulator\t./.sdkharness/tests/health-golden-path"
	if lines[1] != wantRow {
		t.Fatalf("tests.tsv does not contain %q", wantRow)
	}

	info, err := os.Stat(filepath.Join(root, "tests", "health-golden-path"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatal("health-golden-path is not executable")
	}
}
