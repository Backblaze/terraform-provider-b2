//####################################################################
//
// File: b2/validators.go
//
// Copyright 2021 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func validateBase64Key(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if ok {
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err == nil {
			// AES256 (which is the only supported algorithm for now) key should be 256 bits (32 bytes)
			if len(decoded) != 32 {
				errors = append(errors, fmt.Errorf("AES256 key should be 32 bytes, got %d bytes instead",
					len(decoded)))
			}
		} else {
			errors = append(errors, err)
		}
	} else {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
	}

	return warnings, errors
}

// StringLenExact returns a SchemaValidateFunc which tests if the provided value
// is of type string and has given length
//
//nolint:staticcheck // Using SchemaValidateFunc for backward compatibility; migrate to SchemaValidateDiagFunc later.
func StringLenExact(length int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return warnings, errors
		}

		if len(v) != length {
			errors = append(errors, fmt.Errorf("expected length of %s must be %d, got %d", k, length, len(v)))
		}

		return warnings, errors
	}
}

// validateMapKeys warns about map keys that B2 stores in lower case, and about keys listed in
// serverAddedKeys, which B2 sets on its own.
func validateMapKeys(serverAddedKeys ...string) schema.SchemaValidateDiagFunc {
	serverAdded := make(map[string]struct{}, len(serverAddedKeys))
	for _, key := range serverAddedKeys {
		serverAdded[strings.ToLower(key)] = struct{}{}
	}

	return func(i interface{}, path cty.Path) diag.Diagnostics {
		// Values that are not maps are already rejected by the SDK before this runs
		m, ok := i.(map[string]interface{})
		if !ok {
			return nil
		}

		keys := make([]string, 0, len(m))
		for key := range m {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		diags := make(diag.Diagnostics, 0, len(keys))
		for _, key := range keys {
			lowerCaseKey := strings.ToLower(key)

			if key != lowerCaseKey {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Key will be stored in lower case",
					Detail: fmt.Sprintf("B2 converts keys to lower case, so %q will be stored and returned as %q.",
						key, lowerCaseKey),
					AttributePath: path.Copy().IndexString(key),
				})
			}

			if _, isServerAdded := serverAdded[lowerCaseKey]; isServerAdded {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Key is set by B2",
					Detail: fmt.Sprintf("B2 sets %q on its own, so the configured value can be replaced on "+
						"upload, and removing the key from the configuration later is not detected as a change.",
						lowerCaseKey),
					AttributePath: path.Copy().IndexString(key),
				})
			}
		}

		return diags
	}
}
