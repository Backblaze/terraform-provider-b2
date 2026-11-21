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
