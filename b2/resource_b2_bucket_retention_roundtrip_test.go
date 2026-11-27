//####################################################################
//
// File: b2/resource_b2_bucket_retention_roundtrip_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type retentionSimulator struct{ url string }

func (s retentionSimulator) call(action string, fields map[string]interface{}, out interface{}) error {
	fields["action"] = action
	body, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(s.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var response struct {
		Result json.RawMessage `json:"result"`
		Error  string          `json:"error"`
	}
	if err = json.Unmarshal(raw, &response); err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("simulator %s: %s", action, response.Error)
	}
	return json.Unmarshal(response.Result, out)
}

// retentionSimulatorExecutable returns a path that can be executed directly. Windows cannot execute a
// .py file through its shebang, so there the helper is started by a .cmd wrapper that runs python.
func retentionSimulatorExecutable(t *testing.T, script string) string {
	t.Helper()
	if runtime.GOOS != "windows" {
		return script
	}
	python, err := exec.LookPath("python")
	if err != nil {
		t.Fatalf("the retention simulator needs python on PATH: %v", err)
	}
	wrapper := filepath.Join(t.TempDir(), "retention_simulator.cmd")
	content := fmt.Sprintf("@\"%s\" \"%s\" %%*\r\n", python, script)
	if err := os.WriteFile(wrapper, []byte(content), 0o700); err != nil {
		t.Fatal(err)
	}
	return wrapper
}

func TestAccResourceB2BucketDefaultRetentionNoneRoundTrip(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("acceptance test requires TF_ACC=1")
	}
	// The simulator holds all state in memory; the helper is only for acceptance
	// testing and calls the actual Python bindings and b2sdk RawSimulator.
	script, err := filepath.Abs("testdata/retention_simulator.py")
	if err != nil {
		t.Fatal(err)
	}
	script = retentionSimulatorExecutable(t, script)
	cmd := exec.Command(script, "--serve")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	line, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil {
		t.Fatalf("start b2sdk simulator: %v (%s)", err, stderr.String())
	}
	port, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || port < 1 || port > 65535 {
		t.Fatalf("invalid simulator port %q", line)
	}
	sim := retentionSimulator{url: fmt.Sprintf("http://127.0.0.1:%d", port)}
	t.Setenv("B2_RETENTION_SIMULATOR_URL", sim.url)
	// This test exercises a local simulator. Never pass ambient B2 credentials to it.
	for _, key := range []string{"B2_APPLICATION_KEY_ID", "B2_APPLICATION_KEY", "B2_TEST_APPLICATION_KEY_ID", "B2_TEST_APPLICATION_KEY"} {
		t.Setenv(key, "")
	}
	factories := map[string]func() (*schema.Provider, error){"b2": func() (*schema.Provider, error) { return New("test", script)(), nil }}

	for _, mode := range []string{"governance", "compliance"} {
		t.Run(mode, func(t *testing.T) {
			bucketName := acctest.RandomWithPrefix("test-b2-retention")
			var bucketID, fileID string
			var initialRetention map[string]interface{}
			t.Cleanup(func() {
				if bucketID != "" {
					var result map[string]bool
					if err := sim.call("cleanup_bucket", map[string]interface{}{"bucket_id": bucketID}, &result); err != nil {
						t.Errorf("cleanup disposable simulator bucket: %v", err)
					}
				}
			})
			seed := func(state *terraform.State) error {
				bucketID = state.RootModule().Resources["b2_bucket.test"].Primary.ID
				var result struct {
					FileID    string                 `json:"file_id"`
					Retention map[string]interface{} `json:"retention"`
				}
				if err := sim.call("create_file", map[string]interface{}{"bucket_id": bucketID, "mode": mode}, &result); err != nil {
					return err
				}
				fileID, initialRetention = result.FileID, result.Retention
				if fileID == "" || initialRetention["mode"] != mode || initialRetention["retainUntilTimestamp"] == nil {
					return fmt.Errorf("file was not created with %s retention: %v", mode, initialRetention)
				}
				t.Logf("%s: initial bucket default and retained file verified", mode)
				return nil
			}
			verify := func(state *terraform.State) error {
				var bucketRetention map[string]interface{}
				if err := sim.call("bucket_retention", map[string]interface{}{"bucket_id": bucketID}, &bucketRetention); err != nil {
					return err
				}
				serverMode, ok := bucketRetention["mode"]
				if !ok || serverMode != nil || bucketRetention["period"] != nil {
					return fmt.Errorf("bucket default retention was not cleared: %v", bucketRetention)
				}
				var current map[string]interface{}
				if err := sim.call("get_file", map[string]interface{}{"file_id": fileID}, &current); err != nil {
					return err
				}
				if current["mode"] != initialRetention["mode"] || current["retainUntilTimestamp"] != initialRetention["retainUntilTimestamp"] {
					return fmt.Errorf("file retention changed: before=%v after=%v", initialRetention, current)
				}
				t.Logf("%s: Terraform state mode=none, period absent; simulator bucket mode=null; file retention before=%v after=%v", mode, initialRetention, current)
				var deleted map[string]bool
				if err := sim.call("delete_file", map[string]interface{}{"file_id": fileID}, &deleted); err != nil {
					return err
				}
				if !deleted["deleted"] {
					return fmt.Errorf("retained file was not deleted")
				}
				return nil
			}
			checkDestroy := func(_ *terraform.State) error {
				var result map[string]bool
				if err := sim.call("bucket_exists", map[string]interface{}{"bucket_id": bucketID}, &result); err != nil {
					return err
				}
				if result["exists"] {
					return fmt.Errorf("disposable bucket still exists")
				}
				t.Logf("%s: disposable file and bucket deleted", mode)
				return nil
			}
			current := retentionRoundTripConfig(bucketName, mode)
			cleared := retentionRoundTripConfig(bucketName, "none")
			invalid := strings.Replace(cleared, `mode = "none"`, `mode = "none"
      period {
        duration = 1
        unit = "days"
      }`, 1)
			invalidSecondBlock := strings.TrimSuffix(current, "\n}") + `
  file_lock_configuration {
    is_file_lock_enabled = true
    default_retention {
      mode = "none"
      period {
        duration = 1
        unit = "days"
      }
    }
  }
}`
			resource.Test(t, resource.TestCase{
				ProviderFactories: factories,
				CheckDestroy:      checkDestroy,
				Steps: []resource.TestStep{
					{Config: invalid, ExpectError: regexp.MustCompile(`period must be omitted when default retention mode is none`)},
					{Config: invalidSecondBlock, ExpectError: regexp.MustCompile(`period must be omitted when default retention mode is none`)},
					{Config: current, Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("b2_bucket.test", "file_lock_configuration.0.default_retention.0.mode", mode),
						resource.TestCheckResourceAttr("b2_bucket.test", "file_lock_configuration.0.default_retention.0.period.0.duration", "1"),
						seed,
					)},
					{Config: invalid, ExpectError: regexp.MustCompile(`period must be omitted when default retention mode is none`)},
					{Config: current, PlanOnly: true},
					{Config: cleared, Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("b2_bucket.test", "file_lock_configuration.0.default_retention.0.mode", "none"),
						resource.TestCheckResourceAttr("b2_bucket.test", "file_lock_configuration.0.default_retention.0.period.#", "0"),
						verify,
					)},
					{Config: cleared, PlanOnly: true},
				},
			})
		})
	}
}

func retentionRoundTripConfig(name, mode string) string {
	period := ""
	if mode != "none" {
		period = `period {
        duration = 1
        unit = "days"
      }`
	}
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = %q
  bucket_type = "allPrivate"
  file_lock_configuration {
    is_file_lock_enabled = true
    default_retention {
      mode = %q
      %s
    }
  }
}`, name, mode, period)
}
