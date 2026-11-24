//####################################################################
//
// File: b2/ssec_key_state_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestSSECKeyRotationPlan(t *testing.T) {
	terraform, err := exec.LookPath("terraform")
	if err != nil {
		t.Skip("offline plan integration requires Terraform CLI")
	}
	dir := t.TempDir()
	providerDir := filepath.Join(dir, "provider")
	if err := os.Mkdir(providerDir, 0o700); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(providerDir, "terraform-provider-b2")
	if os.PathSeparator == '\\' {
		binary += ".exe"
	}
	buildCtx, cancelBuild := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelBuild()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", binary, "..")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build provider: %v\n%s", err, out)
	}
	rc := filepath.Join(dir, "terraform.rc")
	if err := os.WriteFile(rc, []byte(fmt.Sprintf("provider_installation {\n dev_overrides { \"Backblaze/b2\" = %q }\n direct {}\n}\n", filepath.ToSlash(providerDir))), 0o600); err != nil {
		t.Fatal(err)
	}
	oldKey := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32)))
	newKey := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("b", 32)))
	for _, operation := range []string{"create", "read"} {
		t.Run(operation, func(t *testing.T) {
			r := resourceB2BucketFileVersion()
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
				"bucket_id": "bucket-1", "file_name": "file.txt", "source": "missing-audit-source",
				"server_side_encryption": []interface{}{map[string]interface{}{"mode": "SSE-C", "algorithm": "AES256", "key": []interface{}{map[string]interface{}{"key_id": "old-id", "secret_b64": oldKey}}}},
			})
			// The real API returns encryption mode/algorithm, never the supplied key.
			client := newFakeClient(t, `cat >/dev/null; printf '%s' '{"fileId":"file-1","bucketId":"bucket-1","fileName":"file.txt","source":"missing-audit-source","contentSha1":"none","serverSideEncryption":{"mode":"SSE-C","algorithm":"AES256"}}'`)
			if operation == "read" {
				d.SetId("file-1")
				if diags := resourceB2BucketFileVersionRead(context.Background(), d, client); diags.HasError() {
					t.Fatal(diags)
				}
			} else {
				if diags := resourceB2BucketFileVersionCreate(context.Background(), d, client); diags.HasError() {
					t.Fatal(diags)
				}
			}
			attrs := map[string]interface{}{"id": d.Id()}
			for k := range r.Schema {
				attrs[k] = d.Get(k)
			}
			for _, tc := range []struct {
				name, key, id string
				replace       bool
			}{
				{"unchanged", oldKey, "old-id", false}, {"material", newKey, "old-id", true}, {"identifier", oldKey, "new-id", true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					work := t.TempDir()
					sensitivePath := []map[string]interface{}{
						{"type": "get_attr", "value": "server_side_encryption"},
						{"type": "index", "value": map[string]interface{}{"type": "number", "value": 0}},
						{"type": "get_attr", "value": "key"},
						{"type": "index", "value": map[string]interface{}{"type": "number", "value": 0}},
						{"type": "get_attr", "value": "secret_b64"},
					}
					state := map[string]interface{}{
						"version": 4, "serial": 1,
						"lineage": "ee9ed944-22aa-41cb-8ebc-7a8d30000000",
						"outputs": map[string]interface{}{},
						"resources": []map[string]interface{}{{
							"mode": "managed", "type": "b2_bucket_file_version", "name": "test",
							"provider": "provider[\"registry.terraform.io/backblaze/b2\"]",
							"instances": []map[string]interface{}{{
								"schema_version": 0, "attributes": attrs,
								"sensitive_attributes": []interface{}{sensitivePath},
							}},
						}},
					}
					raw, err := json.Marshal(state)
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(work, "terraform.tfstate"), raw, 0o600); err != nil {
						t.Fatal(err)
					}
					config := fmt.Sprintf(`terraform {
 required_providers { b2 = { source = "Backblaze/b2" } }
}
provider "b2" {
 application_key_id = "synthetic-id"
 application_key = "synthetic-key"
}
resource "b2_bucket_file_version" "test" {
 bucket_id = "bucket-1"
 file_name = "file.txt"
 source = "missing-audit-source"
 server_side_encryption {
  mode = "SSE-C"
  algorithm = "AES256"
  key {
   key_id = %q
   secret_b64 = %q
  }
 }
}
`, tc.id, tc.key)
					if err := os.WriteFile(filepath.Join(work, "main.tf"), []byte(config), 0o600); err != nil {
						t.Fatal(err)
					}
					planCtx, cancelPlan := context.WithTimeout(context.Background(), time.Minute)
					defer cancelPlan()
					plan := exec.CommandContext(planCtx, terraform, "plan", "-refresh=false", "-input=false", "-lock=false", "-no-color", "-out=plan")
					plan.Dir = work
					plan.Env = append(os.Environ(), "TF_CLI_CONFIG_FILE="+rc, "CHECKPOINT_DISABLE=1", "TF_LOG=", "TF_LOG_PATH=", "TF_DATA_DIR="+filepath.Join(work, ".terraform"))
					if out, err := plan.CombinedOutput(); err != nil {
						t.Fatalf("offline plan: %v\n%s", err, out)
					}
					show := exec.CommandContext(planCtx, terraform, "show", "-json", "plan")
					show.Dir = work
					show.Env = plan.Env
					out, err := show.Output()
					if err != nil {
						t.Fatal(err)
					}
					var result struct {
						Changes []struct {
							Change struct {
								Actions []string `json:"actions"`
							} `json:"change"`
						} `json:"resource_changes"`
					}
					if err := json.Unmarshal(out, &result); err != nil {
						t.Fatal(err)
					}
					if len(result.Changes) != 1 {
						t.Fatalf("expected one resource change, got %d", len(result.Changes))
					}
					actions := strings.Join(result.Changes[0].Change.Actions, ",")
					want := "no-op"
					if tc.replace {
						want = "delete,create"
					}
					if actions != want {
						t.Errorf("plan actions = %s, want %s", actions, want)
					}
				})
			}
		})
	}
}

func TestFileVersionReadRetainsOnlySSECKey(t *testing.T) {
	for _, mode := range []string{"SSE-C", "SSE-B2"} {
		t.Run(mode, func(t *testing.T) {
			r := resourceB2BucketFileVersion()
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"bucket_id": "bucket-1", "file_name": "file.txt", "source": "missing-source", "server_side_encryption": []interface{}{map[string]interface{}{"mode": "SSE-C", "algorithm": "AES256", "key": []interface{}{map[string]interface{}{"key_id": "configured-id", "secret_b64": base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32)))}}}}})
			d.SetId("file-1")
			client := newFakeClient(t, fmt.Sprintf(`cat >/dev/null; printf '%%s' '{"fileId":"file-1","fileName":"file.txt","serverSideEncryption":{"mode":%q,"algorithm":"AES256"}}'`, mode))
			if diags := resourceB2BucketFileVersionRead(context.Background(), d, client); diags.HasError() {
				t.Fatal(diags)
			}
			keys := d.Get("server_side_encryption.0.key").([]interface{})
			if mode == "SSE-C" {
				if len(keys) != 1 || keys[0].(map[string]interface{})["key_id"] != "configured-id" || keys[0].(map[string]interface{})["secret_b64"] != base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32))) {
					t.Fatal("configured SSE-C key was not preserved")
				}
			} else if len(keys) != 0 {
				t.Fatal("stale SSE-C key was kept for SSE-B2")
			}
		})
	}
}
