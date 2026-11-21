//go:build knowndefect

//####################################################################
//
// File: b2/regression_knowndefect_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Known-defect tests: each test describes the correct behaviour and fails today because the defect is still
// present. They are compiled only with the "knowndefect" build tag and are run with `make test-known-defects`;
// the default `make test` does not build them.
//
// When a defect is fixed: delete the "TestKnownDefect_" prefix from the test name and move the test to a file
// without the build tag (for example regression_test.go) so that it joins the default suite.
//
// Several tests use the fake child process from regression_test.go in place of the PyInstaller binary.

// C3: with duration omitted, the provider must not send 0 seconds to B2. The expected default is any
// positive value; the owner may pick the exact value.
func TestKnownDefect_C3_SignedUrlOmittedDurationIsNotZero(t *testing.T) {
	capture := filepath.Join(t.TempDir(), "stdin.json")
	client := newFakeClient(t, fmt.Sprintf(`cat > %q; echo '{"bucketId":"b1","fileName":"f.txt","duration":3600,"signedUrl":"https://f.example.invalid/f"}'`, capture))
	res := dataSourceB2BucketFileSignedUrl()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"bucket_id": "b1", "file_name": "f.txt",
	})
	if diags := dataSourceB2BucketFileSignedUrlRead(context.Background(), d, client); diags.HasError() {
		t.Fatal(diags)
	}
	raw, err := os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	var sent map[string]interface{}
	if err := json.Unmarshal(raw, &sent); err != nil {
		t.Fatal(err)
	}
	if dur, _ := sent["duration"].(float64); dur <= 0 {
		t.Errorf("C3: omitted duration was sent to the python layer as %v (want a positive default)", sent["duration"])
	}
}

// C18: running the provider's own unit tests must not leave the extracted binary in the temp dir.
// Only the test-leak part is covered; an error-path leak in GetBindings is not.
func TestKnownDefect_C18_ProviderTestsLeaveNoExtractedBinary(t *testing.T) {
	if testing.Short() {
		t.Skip("spawns a nested go test")
	}
	if runtime.GOOS == "windows" {
		t.Skip("not exercised on Windows")
	}
	tmp := t.TempDir()
	cmd := exec.Command("go", "test", "-count=1", "-run", "^TestProvider$", ".")
	cmd.Env = append(os.Environ(), "TMPDIR="+tmp)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("nested go test failed: %v\n%s", err, out)
	}
	leftovers, _ := filepath.Glob(filepath.Join(tmp, "py-terraform-provider*"))
	if len(leftovers) > 0 {
		t.Errorf("C18: TestProvider left extracted binaries behind: %v", leftovers)
	}
}
