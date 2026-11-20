//####################################################################
//
// File: b2/ci_workflow_check_test.go
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

	"gopkg.in/yaml.v3"
)

type workflowStep struct {
	Name string `yaml:"name"`
	If   string `yaml:"if"`
	Run  string `yaml:"run"`
}

type workflowFile struct {
	Jobs map[string]struct {
		Steps []workflowStep `yaml:"steps"`
	} `yaml:"jobs"`
}

// workflowSteps returns the steps of every job. YAML comments are dropped by the parser.
func workflowSteps(workflow string) ([]workflowStep, error) {
	var wf workflowFile
	if err := yaml.Unmarshal([]byte(workflow), &wf); err != nil {
		return nil, fmt.Errorf("parse workflow: %w", err)
	}
	var steps []workflowStep
	for _, job := range wf.Jobs {
		steps = append(steps, job.Steps...)
	}
	return steps, nil
}

var shellComment = regexp.MustCompile(`(?m)(?:^|\s)#.*$`)

// executableScript removes shell comments from a step's `run` script.
func executableScript(run string) string { return shellComment.ReplaceAllString(run, "") }

var pytestCommand = regexp.MustCompile(`(?m)(?:^|[\s;&|(/])pytest(?:\s|$)`)

// ciWorkflowRunsPythonTests reports whether some executable `run` step invokes pytest.
func ciWorkflowRunsPythonTests(workflow string) (bool, error) {
	steps, err := workflowSteps(workflow)
	if err != nil {
		return false, err
	}
	for _, step := range steps {
		if pytestCommand.MatchString(executableScript(step.Run)) {
			return true, nil
		}
	}
	return false, nil
}

var (
	exitOne = regexp.MustCompile(`\bexit\s+1\b`)
	fiWord  = regexp.MustCompile(`\bfi\b`)
	thenKw  = regexp.MustCompile(`\bthen\b`)
)

// stepGuardsSecret reports whether the step itself fails (exit 1) when the named variable is empty:
// either its `if:` condition tests the variable against ” (joined only with ||) and its script exits 1,
// or its script has a shell guard for the variable that ends in exit 1 inside the guard.
func stepGuardsSecret(step workflowStep, name string) bool {
	script := executableScript(step.Run)

	ifForm := regexp.MustCompile(`(?:env|secrets)\.` + name + `\b\s*==\s*(?:''|"")`)
	if ifForm.MatchString(step.If) && !strings.Contains(step.If, "&&") && exitOne.MatchString(script) {
		return true
	}

	if regexp.MustCompile(`\$\{` + name + `:\?`).MatchString(script) {
		return true
	}
	if regexp.MustCompile(`-n\s+"?\$\{?` + name + `\}?"?\s*\]\]?\s*\|\|\s*(?:\{\s*)?(?:echo[^;]*;\s*)?exit\s+1\b`).MatchString(script) {
		return true
	}
	empty := regexp.MustCompile(`-z\s+"?\$\{?` + name + `\b\}?"?|"?\$\{?` + name + `\b\}?"?\s*==?\s*(?:''|"")`)
	for _, loc := range empty.FindAllStringIndex(script, -1) {
		rest := script[loc[1]:]
		fi := fiWord.FindStringIndex(rest)
		if fi == nil {
			continue
		}
		block := rest[:fi[0]]
		then := thenKw.FindStringIndex(block)
		if then == nil || !exitOne.MatchString(block[then[1]:]) {
			continue
		}
		condStart := strings.LastIndex(script[:loc[0]], "if")
		if condStart < 0 || strings.Contains(script[condStart:loc[1]+then[0]], "&&") {
			continue
		}
		return true
	}
	return false
}

// ciWorkflowFailsOnMissingKeys reports whether executable steps fail when either acceptance secret is empty.
func ciWorkflowFailsOnMissingKeys(workflow string) (bool, error) {
	steps, err := workflowSteps(workflow)
	if err != nil {
		return false, err
	}
	for _, name := range []string{"B2_TEST_APPLICATION_KEY", "B2_TEST_APPLICATION_KEY_ID"} {
		guarded := false
		for _, step := range steps {
			if stepGuardsSecret(step, name) {
				guarded = true
				break
			}
		}
		if !guarded {
			return false, nil
		}
	}
	return true, nil
}

// Fixtures for the workflow checks used by TestKnownDefect_C25 (repo_policy_knowndefect_test.go). They prove that
// the checks look at executable steps: a comment or an unrelated `exit 1` must not satisfy them.

const wfHeader = "name: ci\non: push\njobs:\n  testacc:\n    runs-on: ubuntu-latest\n    steps:\n"

func TestCIWorkflowRunsPythonTests(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{"pytest in a run step", "      - run: |\n          cd python-bindings\n          python -m pytest\n", true},
		{"bare pytest command", "      - name: py\n        run: pytest python-bindings\n", true},
		{"only a comment mentions pytest", "      - run: |\n          # pytest is run elsewhere\n          make test\n", false},
		{"trailing comment mentions pytest", "      - run: make test # pytest\n", false},
		{"pytest in a step name only", "      - name: pytest\n        run: make test\n", false},
		{"pytest in a yaml comment", "      # - run: pytest\n      - run: make test\n", false},
		{"pytest as part of another word", "      - run: echo mypytestfile\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ciWorkflowRunsPythonTests(wfHeader + tc.body)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("ciWorkflowRunsPythonTests = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCIWorkflowFailsOnMissingKeys(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{"step guarded by if on both keys", "      - if: ${{ env.B2_TEST_APPLICATION_KEY == '' || env.B2_TEST_APPLICATION_KEY_ID == '' }}\n        run: |\n          echo '::error::missing keys'\n          exit 1\n", true},
		{"shell guard on both keys", "      - run: |\n          if [ -z \"$B2_TEST_APPLICATION_KEY\" ] || [ -z \"$B2_TEST_APPLICATION_KEY_ID\" ]; then\n            echo missing\n            exit 1\n          fi\n", true},
		{"separate shell guards", "      - run: |\n          if [ -z \"${B2_TEST_APPLICATION_KEY}\" ]; then exit 1; fi\n          if [ -z \"${B2_TEST_APPLICATION_KEY_ID}\" ]; then exit 1; fi\n", true},
		{"parameter expansion guards", "      - run: |\n          : \"${B2_TEST_APPLICATION_KEY:?missing}\"\n          : \"${B2_TEST_APPLICATION_KEY_ID:?missing}\"\n", true},
		{"unrelated exit 1 in another step", "      - if: ${{ env.B2_TEST_APPLICATION_KEY == '' || env.B2_TEST_APPLICATION_KEY_ID == '' }}\n        run: echo skipping\n      - run: |\n          make lint || exit 1\n", false},
		{"key comparison in if, exit 1 in an unconditional unrelated step", "      - if: ${{ env.B2_TEST_APPLICATION_KEY == '' || env.B2_TEST_APPLICATION_KEY_ID == '' }}\n        run: echo skipping\n      - run: exit 1\n", false},
		{"shell guard whose exit 1 is outside the if block", "      - run: |\n          if [ -z \"$B2_TEST_APPLICATION_KEY\" ] || [ -z \"$B2_TEST_APPLICATION_KEY_ID\" ]; then\n            echo missing\n          fi\n          exit 1\n", false},
		{"only one key is guarded", "      - if: ${{ env.B2_TEST_APPLICATION_KEY == '' }}\n        run: exit 1\n", false},
		{"shell guard checks only the key ID", "      - run: |\n          if [ -z \"$B2_TEST_APPLICATION_KEY_ID\" ]; then exit 1; fi\n", false},
		{"both keys guarded with and (fails only when both are missing)", "      - if: ${{ env.B2_TEST_APPLICATION_KEY == '' && env.B2_TEST_APPLICATION_KEY_ID == '' }}\n        run: exit 1\n", false},
		{"guard only in a comment", "      - run: |\n          # if [ -z \"$B2_TEST_APPLICATION_KEY\" ] || [ -z \"$B2_TEST_APPLICATION_KEY_ID\" ]; then exit 1; fi\n          make testacc\n", false},
		{"guard that tests for presence is the skip condition", "      - if: ${{ env.B2_TEST_APPLICATION_KEY != '' && env.B2_TEST_APPLICATION_KEY_ID != '' }}\n        run: make testacc\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ciWorkflowFailsOnMissingKeys(wfHeader + tc.body)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("ciWorkflowFailsOnMissingKeys = %v, want %v", got, tc.want)
			}
		})
	}
}
