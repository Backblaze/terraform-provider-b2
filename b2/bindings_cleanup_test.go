//####################################################################
//
// File: b2/bindings_cleanup_test.go
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
	"testing"
)

func TestProviderTestsRemoveExtractedBindings(t *testing.T) {
	tmp := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestProvider$")
	cmd.Env = append(os.Environ(), "TMPDIR="+tmp, "TMP="+tmp, "TEMP="+tmp)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("provider test failed: %v\n%s", err, out)
	}
	leftovers, err := filepath.Glob(filepath.Join(tmp, "py-terraform-provider*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(leftovers) != 0 {
		t.Fatalf("provider test left extracted bindings: %v", leftovers)
	}
}
