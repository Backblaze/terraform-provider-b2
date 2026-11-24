//####################################################################
//
// File: b2/webhook_header_plan_test.go
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
	"strconv"
	"strings"
	"testing"
)

// TestWebhookHeaderPlanRedaction exercises Terraform's actual human-readable plan
// using a local provider build and no B2 API calls or credentials.
func TestWebhookHeaderPlanRedaction(t *testing.T) {
	terraform, err := exec.LookPath("terraform")
	if err != nil {
		t.Skip("terraform CLI is required for plan output test")
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "terraform-provider-b2")
	build := exec.Command("go", "build", "-o", binary, "..")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build provider: %v\n%s", err, out)
	}
	config := `terraform {
  required_providers {
    b2 = { source = "Backblaze/b2" }
  }
}
provider "b2" {}
resource "b2_bucket_notification_rules" "test" {
  bucket_id = "test-bucket"
  notification_rules {
    name = "test-rule"
    event_types = ["b2:ObjectCreated:*"]
    target_configuration {
      target_type = "webhook"
      url = "https://example.com/webhook"
      custom_headers {
        name = "Authorization"
        value = "Bearer synthetic-test-token"
      }
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	cliConfig := "provider_installation {\n  dev_overrides {\n    \"registry.terraform.io/backblaze/b2\" = " + strconv.Quote(dir) + "\n  }\n  direct {}\n}\n"
	cliPath := filepath.Join(dir, "terraform.rc")
	if err := os.WriteFile(cliPath, []byte(cliConfig), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(terraform, "plan", "-no-color", "-input=false", "-lock=false")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "TF_CLI_CONFIG_FILE="+cliPath, "B2_APPLICATION_KEY_ID=", "B2_APPLICATION_KEY=", "TF_LOG=", "TF_LOG_CORE=", "TF_LOG_PROVIDER=")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("terraform plan: %v\n%s", err, out)
	}
	plan := string(out)
	if strings.Contains(plan, "Bearer synthetic-test-token") {
		t.Errorf("plan exposed custom header value:\n%s", plan)
	}
	if !strings.Contains(plan, "(sensitive value)") {
		t.Errorf("plan did not mark header value sensitive:\n%s", plan)
	}
	if !strings.Contains(plan, "Authorization") {
		t.Errorf("plan omitted header name:\n%s", plan)
	}
}

func TestWebhookHeaderValueSensitiveInResourceSchema(t *testing.T) {
	// Pin what Terraform actually receives (CoreConfigSchema), not the Go struct flag.
	// In the resource, custom_headers is a nested block whose `value` is sensitive.
	provider := New("test", "unused")()
	headers := provider.ResourcesMap["b2_bucket_notification_rules"].CoreConfigSchema().
		BlockTypes["notification_rules"].Block.
		BlockTypes["target_configuration"].Block.
		BlockTypes["custom_headers"].Block
	if !headers.Attributes["value"].Sensitive {
		t.Error("resource custom header value must be sensitive in the protocol schema")
	}
	if headers.Attributes["name"].Sensitive {
		t.Error("resource custom header name should remain visible")
	}

	// In the data source, notification_rules is computed-only, so Plugin SDK v2 sends it
	// as one list(object) attribute and drops nested Sensitive flags: the header value is
	// NOT redacted there. This test documents that known limitation (see CHANGELOG); the
	// fix is scoped to the resource.
	dataSource := provider.DataSourcesMap["b2_bucket_notification_rules"].CoreConfigSchema()
	if dataSource.Attributes["notification_rules"].Sensitive {
		t.Error("data source notification_rules unexpectedly became sensitive; update the changelog and docs")
	}
}
