//####################################################################
//
// File: b2/account_info_deprecated_fields_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	fixture := os.Getenv("GO_WANT_ACCOUNT_INFO_FIXTURE")
	if fixture == "" {
		return
	}
	if _, err := io.Copy(io.Discard, os.Stdin); err != nil {
		os.Exit(2)
	}
	switch fixture {
	case "restricted":
		_, _ = io.WriteString(os.Stdout, `{"accountId":"account-1","allowed":[{"buckets":[{"id":"bucket-1","name":"first-bucket"},{"id":"bucket-2","name":"second-bucket"}],"capabilities":["listBuckets"],"namePrefix":"reports/"}]}`)
	case "unrestricted":
		_, _ = io.WriteString(os.Stdout, `{"accountId":"account-1","allowed":[{"buckets":[],"capabilities":["listBuckets"],"namePrefix":""}]}`)
	default:
		os.Exit(2)
	}
	os.Exit(0)
}

func TestAccountInfoPopulatesDeprecatedBucketFields(t *testing.T) {
	t.Setenv("GO_WANT_ACCOUNT_INFO_FIXTURE", "restricted")
	resource := dataSourceB2AccountInfo()
	d := schema.TestResourceDataRaw(t, resource.Schema, nil)
	client := &Client{Exec: os.Args[0], DataSourcesMap: map[string]*schema.Resource{
		"b2_account_info": resource,
	}}
	if diags := dataSourceB2AccountInfoRead(context.Background(), d, client); diags.HasError() {
		t.Fatalf("account info read failed: %v", diags)
	}
	if got := d.Id(); got != "account-1" {
		t.Errorf("account ID = %q, want account-1", got)
	}
	for key, want := range map[string]interface{}{
		"allowed.0.bucket_id":      "bucket-1",
		"allowed.0.bucket_name":    "first-bucket",
		"allowed.0.buckets.0.id":   "bucket-1",
		"allowed.0.buckets.0.name": "first-bucket",
		"allowed.0.buckets.1.id":   "bucket-2",
		"allowed.0.buckets.1.name": "second-bucket",
		"allowed.0.name_prefix":    "reports/",
	} {
		if got := d.Get(key); got != want {
			t.Errorf("%s = %v, want %v", key, got, want)
		}
	}
	if got := d.Get("allowed.0.capabilities").(*schema.Set).List(); len(got) != 1 || got[0] != "listBuckets" {
		t.Errorf("capabilities = %v, want [listBuckets]", got)
	}
}

func TestAccountInfoUnrestrictedHasEmptyDeprecatedBucketFields(t *testing.T) {
	t.Setenv("GO_WANT_ACCOUNT_INFO_FIXTURE", "unrestricted")
	resource := dataSourceB2AccountInfo()
	d := schema.TestResourceDataRaw(t, resource.Schema, nil)
	client := &Client{Exec: os.Args[0], DataSourcesMap: map[string]*schema.Resource{
		"b2_account_info": resource,
	}}
	if diags := dataSourceB2AccountInfoRead(context.Background(), d, client); diags.HasError() {
		t.Fatalf("account info read failed: %v", diags)
	}
	for _, key := range []string{"allowed.0.bucket_id", "allowed.0.bucket_name"} {
		if got := d.Get(key); got != "" {
			t.Errorf("%s = %v, want empty string", key, got)
		}
	}
	if got := d.Get("allowed.0.buckets.#"); got != 0 {
		t.Errorf("bucket count = %v, want 0", got)
	}
}
