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
	"fmt"
	"regexp"
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

// The checks below parse the workflow files. They pin the shape of the acceptance jobs so that a missing
// runner, a fork pull request or a missing secret makes the run fail loudly instead of skipping acceptance.

func workflowJobs(t *testing.T, name string) map[string]map[string]interface{} {
	t.Helper()
	raw := workflowMap(t, readWorkflow(t, name)["jobs"], name+" jobs")
	jobs := make(map[string]map[string]interface{}, len(raw))
	for jobName, job := range raw {
		jobs[jobName] = workflowMap(t, job, jobName)
	}
	return jobs
}

func jobStepList(job map[string]interface{}) []map[string]interface{} {
	raw, _ := job["steps"].([]interface{})
	steps := make([]map[string]interface{}, 0, len(raw))
	for _, step := range raw {
		if m, ok := step.(map[string]interface{}); ok {
			steps = append(steps, m)
		}
	}
	return steps
}

func stepString(step map[string]interface{}, key string) string {
	value, _ := step[key].(string)
	return value
}

func jobRunsCommand(job map[string]interface{}, command string) bool {
	for _, step := range jobStepList(job) {
		if strings.Contains(executableScript(stepString(step, "run")), command) {
			return true
		}
	}
	return false
}

func jobNeeds(job map[string]interface{}) []string {
	switch needs := job["needs"].(type) {
	case string:
		return []string{needs}
	case []interface{}:
		var names []string
		for _, item := range needs {
			if name, ok := item.(string); ok {
				names = append(names, name)
			}
		}
		return names
	}
	return nil
}

// jobPlatforms returns the os/arch pairs of a job's matrix.conf entries.
func jobPlatforms(t *testing.T, jobName string, job map[string]interface{}) map[string]bool {
	t.Helper()
	strategy, _ := job["strategy"].(map[string]interface{})
	matrix, _ := strategy["matrix"].(map[string]interface{})
	confs, _ := matrix["conf"].([]interface{})
	platforms := map[string]bool{}
	for _, raw := range confs {
		conf := workflowMap(t, raw, jobName+" matrix entry")
		platforms[conf["os"].(string)+"/"+conf["arch"].(string)] = true
	}
	return platforms
}

// acceptanceJobs returns the CI jobs that run the acceptance tests.
func acceptanceJobs(t *testing.T, jobs map[string]map[string]interface{}) []string {
	t.Helper()
	var names []string
	for name, job := range jobs {
		if jobRunsCommand(job, "make testacc") {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		t.Fatal("ci.yml has no job that runs make testacc")
	}
	return names
}

func TestSimulatorRevisionIsStatedOnceInCI(t *testing.T) {
	raw := readRepoFile(t, ".github", "workflows", "ci.yml")
	ci := readWorkflow(t, "ci.yml")
	env, _ := ci["env"].(map[string]interface{})
	revision, _ := env["SIMULATOR_REVISION"].(string)
	if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(revision) {
		t.Fatalf("ci.yml must state the Simulator commit once, as a full SHA in the workflow env SIMULATOR_REVISION; got %q", revision)
	}
	if n := strings.Count(raw, revision); n != 1 {
		t.Errorf("the Simulator revision appears %d times in ci.yml; state it only in env.SIMULATOR_REVISION", n)
	}
	checkouts, verified := 0, 0
	for jobName, job := range workflowJobs(t, "ci.yml") {
		for _, step := range jobStepList(job) {
			with, _ := step["with"].(map[string]interface{})
			if with["repository"] == "backblaze-labs/b2-simulator" {
				checkouts++
				if with["ref"] != "${{ env.SIMULATOR_REVISION }}" {
					t.Errorf("%s checks out the Simulator at %v, not ${{ env.SIMULATOR_REVISION }}", jobName, with["ref"])
				}
			}
			script := executableScript(stepString(step, "run"))
			if strings.Contains(script, "rev-parse HEAD") && strings.Contains(script, "$SIMULATOR_REVISION") {
				verified++
			}
		}
	}
	if checkouts == 0 || verified < checkouts {
		t.Errorf("each Simulator checkout must be verified against $SIMULATOR_REVISION (%d checkouts, %d verifications)", checkouts, verified)
	}
}

// An acceptance job must wait only for the build of its own platform: a matrix job waits for every leg of
// every job it needs, so a single unavailable runner used to skip all acceptance legs.
func TestAcceptanceJobsNeedOnlyTheirOwnPlatformBuild(t *testing.T) {
	jobs := workflowJobs(t, "ci.yml")
	builds := map[string]map[string]bool{}
	for name, job := range jobs {
		if jobRunsCommand(job, "make build") {
			builds[name] = jobPlatforms(t, name, job)
		}
	}
	tested := map[string]bool{}
	for _, name := range acceptanceJobs(t, jobs) {
		platforms := jobPlatforms(t, name, jobs[name])
		if len(platforms) != 1 {
			t.Errorf("acceptance job %s covers %d platforms; give each platform its own job", name, len(platforms))
		}
		for platform := range platforms {
			tested[platform] = true
		}
		var needed []string
		for _, need := range jobNeeds(jobs[name]) {
			if _, ok := builds[need]; ok {
				needed = append(needed, need)
			}
		}
		if len(needed) != 1 {
			t.Errorf("acceptance job %s needs build jobs %v; it must need exactly one", name, needed)
			continue
		}
		if fmt.Sprint(builds[needed[0]]) != fmt.Sprint(platforms) {
			t.Errorf("acceptance job %s (%v) needs build %s, which builds %v", name, platforms, needed[0], builds[needed[0]])
		}
	}
	for _, platform := range []string{"linux/amd64", "darwin/amd64", "darwin/arm64", "windows/amd64"} {
		if !tested[platform] {
			t.Errorf("no acceptance job covers %s", platform)
		}
	}
}

// A skipped or cancelled acceptance job must not leave the run green.
func TestAcceptanceGateFailsUnlessEveryAcceptanceJobSucceeded(t *testing.T) {
	jobs := workflowJobs(t, "ci.yml")
	acceptance := acceptanceJobs(t, jobs)
	for name, job := range jobs {
		needs := strings.Join(jobNeeds(job), " ")
		covers := true
		for _, acc := range acceptance {
			if !regexp.MustCompile(`(^| )` + regexp.QuoteMeta(acc) + `( |$)`).MatchString(needs) {
				covers = false
			}
		}
		if !covers || !strings.Contains(fmt.Sprint(job["if"]), "always()") {
			continue
		}
		for _, step := range jobStepList(job) {
			cond := stepString(step, "if")
			if strings.Contains(cond, "&&") || !exitOne.MatchString(executableScript(stepString(step, "run"))) {
				continue
			}
			ok := true
			for _, result := range []string{"failure", "cancelled", "skipped"} {
				if !strings.Contains(cond, "contains(needs.*.result, '"+result+"')") {
					ok = false
				}
			}
			if ok {
				return
			}
		}
		t.Errorf("gate job %s does not fail when an acceptance job fails, is cancelled or is skipped", name)
		return
	}
	t.Errorf("no job with if: always() needs every acceptance job %v and fails unless all succeeded", acceptance)
}

func collectRunnerLabels(value interface{}, key string, labels *[]string) {
	switch v := value.(type) {
	case map[string]interface{}:
		for k, child := range v {
			collectRunnerLabels(child, k, labels)
		}
	case []interface{}:
		for _, child := range v {
			collectRunnerLabels(child, key, labels)
		}
	case string:
		if key == "runner" || key == "runs-on" {
			*labels = append(*labels, v)
		}
	}
}

// GitHub rejects the macOS 14 labels ("The configuration 'macos-14-arm64-us-default' is not supported"), so
// those jobs were cancelled before they started. Platforms must be kept, on a supported label.
func TestWorkflowsUseSupportedMacOSRunners(t *testing.T) {
	retired := regexp.MustCompile(`^macos-(1[0-4])(-|$)`)
	for _, name := range []string{"ci.yml", "cd.yml", "live-b2.yml"} {
		var labels []string
		collectRunnerLabels(readWorkflow(t, name), "", &labels)
		for _, label := range labels {
			if retired.MatchString(label) {
				t.Errorf("%s uses the retired runner label %s", name, label)
			}
		}
	}
	cdJobs := workflowJobs(t, "cd.yml")
	release := jobPlatforms(t, "build-pybindings", cdJobs["build-pybindings"])
	for _, platform := range []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64"} {
		if !release[platform] {
			t.Errorf("release bindings build no longer covers %s", platform)
		}
	}
}

// The Simulator certificate is added to a bundle of public roots, for the acceptance step only, so other
// HTTPS calls in that step (Go module downloads, the Terraform registry) still verify.
func TestSimulatorCertificateExtendsPublicRootsInAcceptanceStep(t *testing.T) {
	jobs := workflowJobs(t, "ci.yml")
	appendCert := regexp.MustCompile(`cat\s+[^\n]*loopback-cert\.pem[^\n]*>>`)
	for _, name := range acceptanceJobs(t, jobs) {
		found := false
		for _, step := range jobStepList(jobs[name]) {
			script := executableScript(stepString(step, "run"))
			if strings.Contains(script, "GITHUB_ENV") {
				t.Errorf("%s step %q writes to GITHUB_ENV; keep Simulator settings in the step that needs them", name, stepString(step, "name"))
			}
			if !strings.Contains(script, "loopback-cert.pem") {
				continue
			}
			found = true
			if !strings.Contains(script, "rootCertificates") || !appendCert.MatchString(script) {
				t.Errorf("%s must append loopback-cert.pem to a bundle of public root certificates", name)
			}
			if regexp.MustCompile(`\bcp\s+[^\n]*loopback-cert\.pem`).MatchString(script) {
				t.Errorf("%s copies loopback-cert.pem as the whole CA bundle", name)
			}
			for _, variable := range []string{"SSL_CERT_FILE=", "REQUESTS_CA_BUNDLE=", "B2_ENDPOINT"} {
				if !strings.Contains(script, variable) {
					t.Errorf("%s does not set %s in the acceptance step", name, variable)
				}
			}
			if !strings.Contains(script, "make testacc") {
				t.Errorf("%s sets the Simulator CA bundle in a step that does not run make testacc", name)
			}
		}
		if !found {
			t.Errorf("%s does not trust the Simulator certificate", name)
		}
	}
}

// Fork pull requests cannot read the deploy key; they must fail at once with a clear message.
func TestForkPullRequestsFailEarlyWithClearMessage(t *testing.T) {
	jobs := workflowJobs(t, "ci.yml")
	policy := ""
	for name, job := range jobs {
		if len(jobNeeds(job)) != 0 {
			continue
		}
		for _, step := range jobStepList(job) {
			env, _ := step["env"].(map[string]interface{})
			script := executableScript(stepString(step, "run"))
			if strings.Contains(fmt.Sprint(env), "github.event.pull_request.head.repo.full_name") &&
				strings.Contains(script, "GITHUB_REPOSITORY") && exitOne.MatchString(script) && strings.Contains(script, "echo") {
				policy = name
			}
		}
	}
	if policy == "" {
		t.Fatal("ci.yml has no first-stage job that fails pull requests from forks with a message")
	}
	for _, name := range acceptanceJobs(t, jobs) {
		found := false
		for _, need := range jobNeeds(jobs[name]) {
			found = found || need == policy
		}
		if !found {
			t.Errorf("acceptance job %s does not need the fork check %s", name, policy)
		}
	}
}

// The deploy key reaches only actions/checkout; its presence is checked by an expression, never in a shell.
func TestSimulatorDeployKeyNeverEntersAShell(t *testing.T) {
	jobs := workflowJobs(t, "ci.yml")
	presence := regexp.MustCompile(`^\$\{\{\s*secrets\.SIMULATOR_DEPLOY_KEY\s*!=\s*''\s*\}\}$`)
	var walk func(value interface{}, path string)
	walk = func(value interface{}, path string) {
		switch v := value.(type) {
		case map[string]interface{}:
			for k, child := range v {
				walk(child, path+"."+k)
			}
		case []interface{}:
			for i, child := range v {
				walk(child, fmt.Sprintf("%s[%d]", path, i))
			}
		case string:
			shellRef := strings.HasSuffix(path, ".run") && regexp.MustCompile(`\$\{?SIMULATOR_DEPLOY_KEY\b`).MatchString(v)
			if !strings.Contains(v, "secrets.SIMULATOR_DEPLOY_KEY") && !shellRef {
				return
			}
			switch {
			case strings.HasSuffix(path, ".with.ssh-key") && v == "${{ secrets.SIMULATOR_DEPLOY_KEY }}":
			case presence.MatchString(v):
			default:
				t.Errorf("%s exposes the deploy key: %q", path, v)
			}
		}
	}
	walk(readWorkflow(t, "ci.yml"), "")

	for name, job := range jobs {
		steps := jobStepList(job)
		checkout := -1
		for i, step := range steps {
			if with, _ := step["with"].(map[string]interface{}); with["repository"] == "backblaze-labs/b2-simulator" {
				checkout = i
			}
		}
		if checkout < 0 {
			continue
		}
		env, _ := job["env"].(map[string]interface{})
		guarded := false
		for _, step := range steps[:checkout] {
			cond := stepString(step, "if")
			for variable, value := range env {
				if s, _ := value.(string); presence.MatchString(s) && strings.Contains(cond, "env."+variable) &&
					exitOne.MatchString(executableScript(stepString(step, "run"))) {
					guarded = true
				}
			}
		}
		if !guarded {
			t.Errorf("%s does not fail with an if: check on the deploy key before checking out the Simulator", name)
		}
	}
}

func TestLiveB2ReportsScheduledFailuresAndRunsOneAtATime(t *testing.T) {
	live := readWorkflow(t, "live-b2.yml")
	concurrency, _ := live["concurrency"].(map[string]interface{})
	if concurrency["group"] == nil || concurrency["cancel-in-progress"] != false {
		t.Error("live-b2.yml must queue runs in one concurrency group without cancelling a run in progress")
	}
	jobs := workflowJobs(t, "live-b2.yml")
	reporters := 0
	for name, job := range jobs {
		permissions, _ := job["permissions"].(map[string]interface{})
		if permissions["issues"] == nil {
			continue
		}
		reporters++
		if len(permissions) != 1 || permissions["issues"] != "write" {
			t.Errorf("%s must have only issues: write, got %v", name, permissions)
		}
		if jobRunsCommand(job, "make testacc") {
			t.Errorf("%s runs the live tests with issues: write", name)
		}
		cond := fmt.Sprint(job["if"])
		if !strings.Contains(cond, "failure()") || !strings.Contains(cond, "github.event_name == 'schedule'") {
			t.Errorf("%s must run only when a scheduled run fails, got if: %s", name, cond)
		}
		needsLive := false
		for _, need := range jobNeeds(job) {
			needsLive = needsLive || jobRunsCommand(jobs[need], "make testacc")
		}
		if !needsLive {
			t.Errorf("%s does not need the live acceptance job", name)
		}
		if !jobRunsCommand(job, "gh issue create") || !jobRunsCommand(job, "gh issue comment") {
			t.Errorf("%s must open an issue, or comment on the open one", name)
		}
	}
	if reporters != 1 {
		t.Errorf("live-b2.yml has %d jobs with issues permission; want one failure reporter", reporters)
	}
}

// A release is published only when the live B2 workflow has passed on the tagged commit.
func TestReleaseRequiresGreenLiveB2RunForTaggedCommit(t *testing.T) {
	jobs := workflowJobs(t, "cd.yml")
	gate := ""
	for name, job := range jobs {
		permissions, _ := job["permissions"].(map[string]interface{})
		if len(permissions) != 1 || permissions["actions"] != "read" {
			continue
		}
		for _, step := range jobStepList(job) {
			script := executableScript(stepString(step, "run"))
			env, _ := step["env"].(map[string]interface{})
			if strings.Contains(script, "actions/workflows/live-b2.yml/runs") && strings.Contains(script, "head_sha") &&
				strings.Contains(script, "status=success") && exitOne.MatchString(script) &&
				fmt.Sprint(env["RELEASE_SHA"]) == "${{ github.sha }}" && job["if"] == nil {
				gate = name
			}
		}
	}
	if gate == "" {
		t.Fatal("cd.yml has no job that fails unless live-b2.yml succeeded on the tagged commit")
	}
	if !needsJob(jobs["build-and-deploy"]["needs"], gate) {
		t.Errorf("release publication does not need the live B2 gate %s", gate)
	}
}
