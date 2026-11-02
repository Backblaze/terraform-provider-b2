//####################################################################
//
// File: b2/resource_b2_application_key_test.go
//
// Copyright 2020 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceB2ApplicationKey_basic(t *testing.T) {
	resourceName := "b2_application_key.test"

	keyName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2ApplicationKeyConfig_basic(keyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "application_key", regexp.MustCompile("^[\x20-\x7E]{31}$")),
					resource.TestMatchResourceAttr(resourceName, "application_key_id", regexp.MustCompile("^[a-zA-Z0-9]{25}$")),
					resource.TestCheckResourceAttr(resourceName, "bucket_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "bucket_id", ""), // deprecated
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.0", "readFiles"),
					resource.TestCheckResourceAttr(resourceName, "expiration_timestamp", "0"),
					resource.TestCheckResourceAttr(resourceName, "key_name", keyName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0", "s3"),
					resource.TestCheckResourceAttr(resourceName, "valid_duration_in_seconds", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"application_key", "valid_duration_in_seconds"},
			},
		},
	})
}

func TestAccResourceB2ApplicationKey_all(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_application_key.test"

	keyName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2ApplicationKeyConfig_all(keyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "application_key", regexp.MustCompile("^[\x20-\x7E]{31}$")),
					resource.TestMatchResourceAttr(resourceName, "application_key_id", regexp.MustCompile("^[a-zA-Z0-9]{25}$")),
					resource.TestCheckResourceAttr(resourceName, "bucket_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_ids.0", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"), // deprecated
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.0", "writeFiles"),
					resource.TestMatchResourceAttr(resourceName, "expiration_timestamp", regexp.MustCompile("^[1-9][0-9]{12,}$")),
					resource.TestCheckResourceAttr(resourceName, "key_name", keyName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "prefix"),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0", "s3"),
					resource.TestCheckResourceAttr(resourceName, "valid_duration_in_seconds", "86400"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"application_key", "valid_duration_in_seconds"},
			},
		},
	})
}

func TestAccResourceB2ApplicationKey_deprecatedBucketId(t *testing.T) {
	parentResourceName := "b2_bucket.test"
	resourceName := "b2_application_key.test"

	keyName := acctest.RandomWithPrefix("test-b2-tfp")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceB2ApplicationKeyConfig_deprecatedBucketId(keyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "application_key", regexp.MustCompile("^[\x20-\x7E]{31}$")),
					resource.TestMatchResourceAttr(resourceName, "application_key_id", regexp.MustCompile("^[a-zA-Z0-9]{25}$")),
					resource.TestCheckResourceAttr(resourceName, "bucket_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_ids.0", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"), // deprecated
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.0", "writeFiles"),
					resource.TestMatchResourceAttr(resourceName, "expiration_timestamp", regexp.MustCompile("^[1-9][0-9]{12,}$")),
					resource.TestCheckResourceAttr(resourceName, "key_name", keyName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "prefix"),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0", "s3"),
					resource.TestCheckResourceAttr(resourceName, "valid_duration_in_seconds", "86400"),
				),
			},
			{
				Config: testAccResourceB2ApplicationKeyConfig_all(keyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "application_key", regexp.MustCompile("^[\x20-\x7E]{31}$")),
					resource.TestMatchResourceAttr(resourceName, "application_key_id", regexp.MustCompile("^[a-zA-Z0-9]{25}$")),
					resource.TestCheckResourceAttr(resourceName, "bucket_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_ids.0", parentResourceName, "bucket_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_id", parentResourceName, "bucket_id"), // deprecated
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.0", "writeFiles"),
					resource.TestMatchResourceAttr(resourceName, "expiration_timestamp", regexp.MustCompile("^[1-9][0-9]{12,}$")),
					resource.TestCheckResourceAttr(resourceName, "key_name", keyName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "prefix"),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0", "s3"),
					resource.TestCheckResourceAttr(resourceName, "valid_duration_in_seconds", "86400"),
				),
			},
		},
	})
}

func testAccResourceB2ApplicationKeyConfig_basic(keyName string) string {
	return fmt.Sprintf(`
resource "b2_application_key" "test" {
  key_name = "%s"
  capabilities = ["readFiles"]
}
`, keyName)
}

func testAccResourceB2ApplicationKeyConfig_all(keyName string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPrivate"
}

resource "b2_application_key" "test" {
  key_name = "%s"
  capabilities = ["writeFiles"]
  bucket_ids = [b2_bucket.test.bucket_id]
  name_prefix = "prefix"
  valid_duration_in_seconds = 86400
}
`, keyName, keyName)
}

func testAccResourceB2ApplicationKeyConfig_deprecatedBucketId(keyName string) string {
	return fmt.Sprintf(`
resource "b2_bucket" "test" {
  bucket_name = "%s"
  bucket_type = "allPrivate"
}

resource "b2_application_key" "test" {
  key_name = "%s"
  capabilities = ["writeFiles"]
  bucket_id = b2_bucket.test.bucket_id
  name_prefix = "prefix"
  valid_duration_in_seconds = 86400
}
`, keyName, keyName)
}

// --- plan/diff regression tests --------------------------------------------
//
// The DiffSuppressFunc on bucket_ids (and on the deprecated bucket_id) must
// consult only the configuration being planned, never the merged state. The
// provider mirrors bucketIds[0] into state's bucket_id (and vice versa) on
// every create/read, so a state-based check (d.GetOk) suppressed every real
// change to the bucket restrictions and OpenTofu reported "No changes".

var appKeyDiffCaps = []string{"listBuckets", "listFiles", "readFiles", "writeFiles", "deleteFiles"}

// appKeyDiffState builds the InstanceState as resourceB2ApplicationKeyRead
// (Client.Populate) leaves it after a create/read: bucketIds mirrored into
// bucket_id, or a deprecated bucket_id mirrored into bucket_ids.
func appKeyDiffState(t *testing.T, res *schema.Resource, bucketIds []string, deprecatedBucketId string, rawConfig cty.Value) *terraform.InstanceState {
	t.Helper()

	attrs := map[string]string{
		"application_key_id":        "appKeyId0",
		"application_key":           "appKey0",
		"expiration_timestamp":      "0",
		"key_name":                  "my-v4-key",
		"name_prefix":               "",
		"valid_duration_in_seconds": "0",
	}

	w := &schema.MapFieldWriter{Schema: res.SchemaMap()}
	ids := make([]interface{}, 0, len(bucketIds))
	for _, id := range bucketIds {
		ids = append(ids, id)
	}
	caps := make([]interface{}, 0, len(appKeyDiffCaps))
	for _, c := range appKeyDiffCaps {
		caps = append(caps, c)
	}
	for attr, val := range map[string][]interface{}{
		"bucket_ids":   ids,
		"capabilities": caps,
		"options":      {"s3"},
	} {
		if err := w.WriteField([]string{attr}, val); err != nil {
			t.Fatalf("writing %s to state: %s", attr, err)
		}
	}
	for k, v := range w.Map() {
		attrs[k] = v
	}

	// the mirrors performed by python-bindings ApplicationKey._postprocess:
	// bucketIds[0] -> bucketId, and bucketId -> [bucketId]
	wIds := ids
	switch {
	case len(bucketIds) > 0:
		attrs["bucket_id"] = bucketIds[0]
	case deprecatedBucketId != "":
		attrs["bucket_id"] = deprecatedBucketId
		wIds = []interface{}{deprecatedBucketId}
	default:
		attrs["bucket_id"] = ""
	}
	if len(wIds) != len(ids) {
		if err := w.WriteField([]string{"bucket_ids"}, wIds); err != nil {
			t.Fatalf("writing mirrored bucket_ids to state: %s", err)
		}
		for k, v := range w.Map() {
			attrs[k] = v
		}
	}

	return &terraform.InstanceState{
		ID:         attrs["application_key_id"],
		Attributes: attrs,
		RawConfig:  rawConfig, // as grpc_provider.PlanResourceChange does
	}
}

// appKeyDiffConfig builds the configuration exactly the way
// grpc_provider.PlanResourceChange passes it to Resource.Diff: an object value
// shimmed with NewResourceConfigShimmed.
func appKeyDiffConfig(t *testing.T, res *schema.Resource, bucketIds []string, deprecatedBucketId string) (*terraform.ResourceConfig, cty.Value) {
	t.Helper()

	caps := make([]cty.Value, 0, len(appKeyDiffCaps))
	for _, c := range appKeyDiffCaps {
		caps = append(caps, cty.StringVal(c))
	}

	bucketIdsVal := cty.NullVal(cty.Set(cty.String))
	if bucketIds != nil {
		vals := make([]cty.Value, 0, len(bucketIds))
		for _, id := range bucketIds {
			vals = append(vals, cty.StringVal(id))
		}
		bucketIdsVal = cty.SetVal(vals)
	}

	bucketIdVal := cty.NullVal(cty.String)
	if deprecatedBucketId != "" {
		bucketIdVal = cty.StringVal(deprecatedBucketId)
	}

	cfgVal := cty.ObjectVal(map[string]cty.Value{
		"application_key":           cty.NullVal(cty.String),
		"application_key_id":        cty.NullVal(cty.String),
		"bucket_id":                 bucketIdVal,
		"bucket_ids":                bucketIdsVal,
		"capabilities":              cty.SetVal(caps),
		"expiration_timestamp":      cty.NullVal(cty.Number),
		"key_name":                  cty.StringVal("my-v4-key"),
		"name_prefix":               cty.NullVal(cty.String),
		"options":                   cty.NullVal(cty.Set(cty.String)),
		"valid_duration_in_seconds": cty.NullVal(cty.Number),
	})

	return terraform.NewResourceConfigShimmed(cfgVal, res.CoreConfigSchema()), cfgVal
}

// appKeyDiff runs the same diff that PlanResourceChange runs for the given
// prior state and configuration.
func appKeyDiff(t *testing.T, stateBucketIds []string, stateDeprecatedId, cfgDeprecatedId string, cfgBucketIds []string) *terraform.InstanceDiff {
	t.Helper()

	res := resourceB2ApplicationKey()
	cfg, cfgVal := appKeyDiffConfig(t, res, cfgBucketIds, cfgDeprecatedId)
	state := appKeyDiffState(t, res, stateBucketIds, stateDeprecatedId, cfgVal)

	diff, err := res.Diff(context.Background(), state, cfg, nil)
	if err != nil {
		t.Fatalf("unexpected diff error: %s", err)
	}
	return diff
}

// appKeyDiffAttrRequiresNew reports whether the diff plans a replacement
// because of the given attribute (matched exactly or as a set/list prefix).
func appKeyDiffAttrRequiresNew(diff *terraform.InstanceDiff, attr string) bool {
	if diff == nil {
		return false
	}
	for k, v := range diff.Attributes {
		if v == nil || k != attr && !strings.HasPrefix(k, attr+".") {
			continue
		}
		if v.RequiresNew {
			return true
		}
	}
	return false
}

func TestResourceB2ApplicationKeyDiffSuppress(t *testing.T) {
	t.Run("widening bucket_ids forces replacement", func(t *testing.T) {
		// the reported bug: state holds a key scoped to 2 buckets (plus the
		// mirrored bucket_id), the config widens the scope to 3 buckets
		diff := appKeyDiff(t, []string{"bucket-aaaaaaaa", "bucket-bbbbbbbb"}, "", "",
			[]string{"bucket-aaaaaaaa", "bucket-bbbbbbbb", "bucket-cccccccc"})
		if diff == nil {
			t.Fatal("expected a diff (replacement), got none: update is silently suppressed")
		}
		if !appKeyDiffAttrRequiresNew(diff, "bucket_ids") {
			t.Fatalf("expected bucket_ids to force replacement, got: %#v", diff.Attributes)
		}
	})

	t.Run("narrowing bucket_ids forces replacement", func(t *testing.T) {
		diff := appKeyDiff(t, []string{"bucket-aaaaaaaa", "bucket-bbbbbbbb", "bucket-cccccccc"}, "", "",
			[]string{"bucket-aaaaaaaa", "bucket-bbbbbbbb"})
		if diff == nil || !appKeyDiffAttrRequiresNew(diff, "bucket_ids") {
			t.Fatalf("expected bucket_ids to force replacement, got: %#v", diff)
		}
	})

	t.Run("changing deprecated bucket_id forces replacement", func(t *testing.T) {
		// state's bucket_ids mirror must not suppress a real bucket_id change
		diff := appKeyDiff(t, nil, "bucket-aaaaaaaa", "bucket-cccccccc", nil)
		if diff == nil || !appKeyDiffAttrRequiresNew(diff, "bucket_id") {
			t.Fatalf("expected bucket_id to force replacement, got: %#v", diff)
		}
	})

	t.Run("deprecated bucket_id in config still suppresses the bucket_ids mirror", func(t *testing.T) {
		// unchanged deprecated configuration: the bucket_ids value in state is
		// the mirror written by the provider, not a config value; it must not
		// produce a diff
		diff := appKeyDiff(t, nil, "bucket-aaaaaaaa", "bucket-aaaaaaaa", nil)
		if diff != nil && len(diff.Attributes) != 0 {
			t.Fatalf("expected no diff for an unchanged deprecated configuration, got: %#v", diff.Attributes)
		}
	})

	t.Run("unchanged bucket_ids suppress the bucket_id mirror", func(t *testing.T) {
		// unchanged bucket_ids configuration: state's bucket_id is the mirror
		// written by the provider; it must not produce a diff
		diff := appKeyDiff(t, []string{"bucket-aaaaaaaa", "bucket-bbbbbbbb"}, "", "",
			[]string{"bucket-aaaaaaaa", "bucket-bbbbbbbb"})
		if diff != nil && len(diff.Attributes) != 0 {
			t.Fatalf("expected no diff for an unchanged configuration, got: %#v", diff.Attributes)
		}
	})
}
