//####################################################################
//
// File: b2/simulator_ci_policy_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"strings"
	"testing"
)

// Pull-request and release CI run the acceptance tests against a pinned B2 Simulator; real B2 runs only in
// live-b2.yml, on a weekly schedule or by manual dispatch. These are text checks on the workflow files.

func TestPullRequestAcceptanceUsesSimulatorNotLiveB2(t *testing.T) {
	ci := readRepoFile(t, ".github", "workflows", "ci.yml")
	for _, forbidden := range []string{
		"secrets.B2_TEST_APPLICATION_KEY",
		"env.B2_TEST_APPLICATION_KEY != ''",
		"env.B2_TEST_APPLICATION_KEY_ID != ''",
	} {
		if strings.Contains(ci, forbidden) {
			t.Errorf("ci.yml still contains the live-B2 credential gate %q", forbidden)
		}
	}
	for _, required := range []string{
		"SIMULATOR_DEPLOY_KEY",
		"repository: backblaze-labs/b2-simulator",
		"persist-credentials: false",
		"B2_ENDPOINT",
		"B2_TEST_APPLICATION_KEY_ID: sim-test-id",
		"B2_TEST_APPLICATION_KEY: sim-test-secret",
		"make testacc",
	} {
		if !strings.Contains(ci, required) {
			t.Errorf("ci.yml does not contain the Simulator acceptance setting %q", required)
		}
	}
}

// Replaces the fail-closed half of the former known defect C25: acceptance against real B2 must fail, not
// stay green, when either credential is missing. The Python-tests half is TestCIWorkflowRunsPythonBindingsTests.
func TestLiveB2AcceptanceIsWeeklyManualAndFailsClosed(t *testing.T) {
	live := readRepoFile(t, ".github", "workflows", "live-b2.yml")
	trigger := live
	if end := strings.Index(live, "\npermissions:"); end >= 0 {
		trigger = live[:end]
	}
	for _, required := range []string{"schedule:", "workflow_dispatch:"} {
		if !strings.Contains(trigger, required) {
			t.Errorf("live-b2.yml triggers do not include %q", required)
		}
	}
	for _, forbidden := range []string{"pull_request", "push:", "workflow_call"} {
		if strings.Contains(trigger, forbidden) {
			t.Errorf("live-b2.yml must not be triggered by %q", forbidden)
		}
	}
	for _, required := range []string{
		"secrets.B2_TEST_APPLICATION_KEY }}",
		"secrets.B2_TEST_APPLICATION_KEY_ID }}",
		"make testacc",
	} {
		if !strings.Contains(live, required) {
			t.Errorf("live-b2.yml does not contain %q", required)
		}
	}
	ok, err := ciWorkflowFailsOnMissingKeys(live)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("live-b2.yml does not fail (exit 1) when B2_TEST_APPLICATION_KEY or B2_TEST_APPLICATION_KEY_ID is missing")
	}
}
