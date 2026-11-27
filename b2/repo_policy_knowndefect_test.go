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
	"regexp"
	"testing"
)

// Known-defect repository checks. Compiled only with the "knowndefect" build tag (see regression_knowndefect_test.go
// for how to run them and how to move a test into the default suite once its defect is fixed). The workflow
// checks are text heuristics: they pin the shape of the files they read, so adjust them with the fix.

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
