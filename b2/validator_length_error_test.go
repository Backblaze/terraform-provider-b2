//####################################################################
//
// File: b2/validator_length_error_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestStringLenExactOmitsValueFromError(t *testing.T) {
	target := getNotificationRulesElem(false).Schema["target_configuration"].Elem.(*schema.Resource)
	field := target.Schema["hmac_sha256_signing_secret"]
	if !field.Sensitive || field.ValidateFunc == nil {
		t.Fatal("signing secret must be sensitive and have a validator")
	}

	for _, tc := range []struct {
		name   string
		value  string
		length int
	}{
		{name: "short", value: "abcde", length: 5},
		{name: "long", value: strings.Repeat("x", 33), length: 33},
		{name: "unicode byte length", value: "é", length: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			warnings, errs := field.ValidateFunc(tc.value, "signing_secret")
			if len(warnings) != 0 || len(errs) != 1 {
				t.Fatalf("expected one validation error and no warnings, got %d errors and %d warnings", len(errs), len(warnings))
			}
			message := errs[0].Error()
			if strings.Contains(message, tc.value) {
				t.Fatal("validation error contains the supplied value")
			}
			want := fmt.Sprintf("expected length of signing_secret must be 32, got %d", tc.length)
			if message != want {
				t.Fatalf("validation error does not report the expected and actual lengths")
			}
		})
	}

	t.Run("valid length", func(t *testing.T) {
		warnings, errs := field.ValidateFunc(strings.Repeat("v", 32), "signing_secret")
		if len(warnings) != 0 || len(errs) != 0 {
			t.Fatalf("expected no warnings or errors, got %d warnings and %d errors", len(warnings), len(errs))
		}
	})
	t.Run("non-string", func(t *testing.T) {
		warnings, errs := field.ValidateFunc(42, "signing_secret")
		if len(warnings) != 0 || len(errs) != 1 || errs[0].Error() != "expected type of signing_secret to be string" {
			t.Fatal("expected a type error and no warnings")
		}
	})
}
