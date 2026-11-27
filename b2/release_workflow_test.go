//####################################################################
//
// File: b2/release_workflow_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func readWorkflow(t *testing.T, name string) map[string]interface{} {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", ".github", "workflows", name))
	if err != nil {
		t.Fatal(err)
	}
	var workflow map[string]interface{}
	if err := yaml.Unmarshal(raw, &workflow); err != nil {
		t.Fatal(err)
	}
	return workflow
}

func workflowMap(t *testing.T, value interface{}, label string) map[string]interface{} {
	t.Helper()
	result, ok := value.(map[string]interface{})
	if !ok {
		t.Fatalf("%s is not a mapping", label)
	}
	return result
}

func TestReleaseRequiresCIForTaggedCommit(t *testing.T) {
	ci := readWorkflow(t, "ci.yml")
	ciEvents := workflowMap(t, ci["on"], "CI triggers")
	ciCall, ok := ciEvents["workflow_call"].(map[string]interface{})
	if !ok {
		t.Error("CI must be callable by the release workflow")
	} else {
		secrets := workflowMap(t, ciCall["secrets"], "callable CI secrets")
		for _, name := range []string{"B2_TEST_APPLICATION_KEY", "B2_TEST_APPLICATION_KEY_ID"} {
			secret := workflowMap(t, secrets[name], "callable CI secret "+name)
			if secret["required"] != false {
				t.Errorf("callable CI secret %s must remain optional", name)
			}
		}
	}

	cd := readWorkflow(t, "cd.yml")
	events := workflowMap(t, cd["on"], "release triggers")
	push := workflowMap(t, events["push"], "release push trigger")
	tags, ok := push["tags"].([]interface{})
	if !ok || len(tags) != 1 || tags[0] != "v*" {
		t.Error("release must remain limited to v* tag pushes")
	}
	jobs := workflowMap(t, cd["jobs"], "release jobs")
	ciJob, ok := jobs["ci"].(map[string]interface{})
	if !ok || ciJob["uses"] != "./.github/workflows/ci.yml" {
		t.Error("release CI job must call this revision's CI workflow")
	} else {
		secrets := workflowMap(t, ciJob["secrets"], "release CI secrets")
		if len(secrets) != 3 || secrets["B2_TEST_APPLICATION_KEY"] != "${{ secrets.B2_TEST_APPLICATION_KEY }}" ||
			secrets["B2_TEST_APPLICATION_KEY_ID"] != "${{ secrets.B2_TEST_APPLICATION_KEY_ID }}" ||
			secrets["SIMULATOR_DEPLOY_KEY"] != "${{ secrets.SIMULATOR_DEPLOY_KEY }}" {
			t.Error("release CI must receive only its two acceptance-test secrets and the B2 Simulator deploy key")
		}
	}
	build := workflowMap(t, jobs["build-pybindings"], "bindings build job")
	if !needsJob(build["needs"], "ci") || build["if"] != nil {
		t.Error("bindings build must depend on successful CI without overriding the success condition")
	}
	deploy := workflowMap(t, jobs["build-and-deploy"], "release publication job")
	if !needsJob(deploy["needs"], "build-pybindings") || deploy["if"] != nil {
		t.Error("release publication must depend on a successful bindings build")
	}
}

func needsJob(value interface{}, name string) bool {
	if value == name {
		return true
	}
	if list, ok := value.([]interface{}); ok {
		for _, item := range list {
			if item == name {
				return true
			}
		}
	}
	return false
}

func TestWorkflowPermissionsAndActionPins(t *testing.T) {
	pin := regexp.MustCompile(`^[^@]+@[0-9a-f]{40}$`)
	for _, name := range []string{"ci.yml", "cd.yml"} {
		t.Run(name, func(t *testing.T) {
			workflow := readWorkflow(t, name)
			permissions := workflowMap(t, workflow["permissions"], "workflow permissions")
			if len(permissions) != 1 || permissions["contents"] != "read" {
				t.Error("workflow token must have read-only contents permission by default")
			}
			jobs := workflowMap(t, workflow["jobs"], "workflow jobs")
			for jobName, rawJob := range jobs {
				job := workflowMap(t, rawJob, jobName)
				steps, _ := job["steps"].([]interface{})
				for _, rawStep := range steps {
					step := workflowMap(t, rawStep, jobName+" step")
					uses, _ := step["uses"].(string)
					if uses != "" && !strings.HasPrefix(uses, "./") && !pin.MatchString(uses) {
						t.Errorf("%s uses an unpinned action: %s", jobName, uses)
					}
				}
			}
			if name == "cd.yml" {
				call := workflowMap(t, jobs["ci"], "callable CI job")
				callPermissions := workflowMap(t, call["permissions"], "callable CI permissions")
				if len(callPermissions) != 2 || callPermissions["contents"] != "read" || callPermissions["pull-requests"] != "read" {
					t.Error("callable CI must have read-only contents and pull-request permissions")
				}
				deploy := workflowMap(t, jobs["build-and-deploy"], "publication job")
				deployPermissions := workflowMap(t, deploy["permissions"], "publication permissions")
				if len(deployPermissions) != 1 || deployPermissions["contents"] != "write" {
					t.Error("publication job requires contents: write")
				}
			}
		})
	}
}

func TestReusableCIArtifactsDoNotCollideWithReleaseArtifacts(t *testing.T) {
	ci := readWorkflow(t, "ci.yml")
	cd := readWorkflow(t, "cd.yml")
	ciJobs := workflowMap(t, ci["jobs"], "CI jobs")
	cdJobs := workflowMap(t, cd["jobs"], "release jobs")
	ciUpload := artifactName(t, ciJobs["build"], "actions/upload-artifact@")
	ciDownload := artifactName(t, ciJobs["testacc"], "actions/download-artifact@")
	cdUpload := artifactName(t, cdJobs["build-pybindings"], "actions/upload-artifact@")
	if ciUpload == "" || ciUpload != ciDownload {
		t.Errorf("CI upload and acceptance download must use the same artifact name: %q, %q", ciUpload, ciDownload)
	}
	if ciUpload == cdUpload {
		t.Errorf("reusable CI and release builds upload the same artifact name: %q", ciUpload)
	}
	deploy := workflowMap(t, cdJobs["build-and-deploy"], "publication job")
	steps, _ := deploy["steps"].([]interface{})
	for _, rawStep := range steps {
		step := workflowMap(t, rawStep, "publication step")
		uses, _ := step["uses"].(string)
		if strings.HasPrefix(uses, "actions/download-artifact@") {
			with := workflowMap(t, step["with"], "release download inputs")
			if with["pattern"] != "py-terraform-provider-b2-*" {
				t.Error("publication must download only release artifacts")
			}
			return
		}
	}
	t.Error("publication artifact download is missing")
}

func artifactName(t *testing.T, rawJob interface{}, action string) string {
	t.Helper()
	job := workflowMap(t, rawJob, "artifact job")
	steps, _ := job["steps"].([]interface{})
	for _, rawStep := range steps {
		step := workflowMap(t, rawStep, "artifact step")
		uses, _ := step["uses"].(string)
		if strings.HasPrefix(uses, action) {
			with := workflowMap(t, step["with"], "artifact action inputs")
			name, _ := with["name"].(string)
			return name
		}
	}
	t.Fatalf("artifact action %s is missing", action)
	return ""
}
