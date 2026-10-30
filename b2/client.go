//####################################################################
//
// File: b2/client.go
//
// Copyright 2020 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Operation represents a Terraform operation type
type Operation string

const (
	OpDataSourceRead Operation = "data_source_read"
	OpResourceCreate Operation = "resource_create"
	OpResourceRead   Operation = "resource_read"
	OpResourceUpdate Operation = "resource_update"
	OpResourceDelete Operation = "resource_delete"
)

type Client struct {
	Exec             string
	UserAgentAppend  string
	ApplicationKeyId string
	ApplicationKey   string
	Endpoint         string
	DataSourcesMap   map[string]*schema.Resource
	ResourcesMap     map[string]*schema.Resource
}

// Apply executes a provider operation with typed input and output.
func (c Client) Apply(ctx context.Context, op Operation, input ResourceSchema, output ResourceSchema) error {
	name := input.ResourceName()

	tflog.Info(ctx, "Executing pybindings", map[string]interface{}{
		"name": name,
		"op":   op,
	})

	// Convert input struct to map for backward compatibility with Python bindings
	inputMap := convertStructToMap(input)
	schemaMap := c.getSchemaMap(name, op)
	if schemaMap == nil {
		// Should never happen
		return fmt.Errorf("schema not found for resource: b2_%s", name)
	}

	sanitizer := logSanitizer{}
	tflog.Debug(ctx, "Input for pybindings", map[string]interface{}{
		"input": sanitizer.sanitizeForLog(inputMap, schemaMap),
	})

	cmd := exec.Command(c.Exec, name, string(op))
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("B2_USER_AGENT_APPEND=%s", c.UserAgentAppend))

	inputMap["provider_application_key_id"] = c.ApplicationKeyId
	inputMap["provider_application_key"] = c.ApplicationKey
	inputMap["provider_endpoint"] = c.Endpoint

	inputJson, err := json.Marshal(inputMap)
	if err != nil {
		// Should never happen
		return err
	}
	cmd.Stdin = bytes.NewReader(inputJson)

	outputJson, err := cmd.Output()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if len(exitErr.Stderr) > 0 {
				err := fmt.Errorf("%s", sanitizer.redactText(string(exitErr.Stderr), c.ApplicationKey, c.ApplicationKeyId))
				tflog.Error(ctx, "Error in pybindings", map[string]interface{}{
					"stderr": err,
				})
				return err
			}
			return fmt.Errorf("failed to execute")
		} else {
			tflog.Error(ctx, "Error", map[string]interface{}{
				"err": err,
			})
			return err
		}
	}

	if output != nil {
		err = json.Unmarshal(outputJson, output)
		if err != nil {
			return err
		}

		tflog.Debug(ctx, "Safe output from pybindings", map[string]interface{}{
			"output": sanitizer.sanitizeForLog(output, schemaMap),
		})
	}

	return nil
}

// Populate fills the Terraform ResourceData with values from the typed output.
func (c Client) Populate(ctx context.Context, op Operation, output ResourceSchema, d *schema.ResourceData) error {
	name := output.ResourceName()

	tflog.Info(ctx, "Populating data from pybindings", map[string]interface{}{
		"name": name,
		"op":   op,
	})

	schemaMap := c.getSchemaMap(name, op)
	if schemaMap == nil {
		// Should never happen
		return fmt.Errorf("schema not found for resource: b2_%s", name)
	}

	outputMap := convertStructToMap(output)

	for k := range schemaMap {
		v, ok := outputMap[k]
		if !ok {
			return fmt.Errorf("error getting %s", k)
		}
		if err := d.Set(k, v); err != nil {
			return fmt.Errorf("error setting %s: %s", k, err)
		}
	}

	return nil
}

func (c Client) getSchemaMap(name string, op Operation) map[string]*schema.Schema {
	resourceName := "b2_" + name
	if op == OpDataSourceRead {
		if ds, ok := c.DataSourcesMap[resourceName]; ok {
			return ds.Schema
		}
	} else {
		if res, ok := c.ResourcesMap[resourceName]; ok {
			return res.Schema
		}
	}
	return nil
}

type logSanitizer struct {
	secrets []string
}

// sanitizeForLog copies inputs or outputs, recursively masking Sensitive schema
// fields, signed_url, and custom_headers[].value. The last two are log-only masks
// to preserve Terraform output compatibility. Masked strings are also collected
// for redacting binding error text.
func (s *logSanitizer) sanitizeForLog(output interface{}, schemaMap map[string]*schema.Schema) map[string]interface{} {
	outputMap, ok := output.(map[string]interface{})
	if !ok {
		outputMap = convertStructToMap(output)
	}

	return s.sanitizeMap(outputMap, schemaMap, "")
}

func logOnlySensitive(key, parentKey string) bool {
	return key == "signed_url" || (parentKey == "custom_headers" && key == "value")
}

func (s *logSanitizer) sanitizeMap(values map[string]interface{}, schemaMap map[string]*schema.Schema, parentKey string) map[string]interface{} {
	safeValues := make(map[string]interface{}, len(values))
	for key, value := range values {
		if logOnlySensitive(key, parentKey) {
			safeValues[key] = s.mask(value)
			continue
		}

		if field, ok := schemaMap[key]; ok {
			safeValues[key] = s.sanitizeSchemaValue(value, field, key)
			continue
		}

		safeValues[key] = value
	}

	return safeValues
}

func (s *logSanitizer) sanitizeSchemaValue(value interface{}, field *schema.Schema, parentKey string) interface{} {
	if field.Sensitive {
		return s.mask(value)
	}

	switch field.Type {
	case schema.TypeList, schema.TypeSet:
		return s.sanitizeCollection(value, field.Elem, parentKey)
	case schema.TypeMap:
		return s.sanitizeMapValues(value, field.Elem, parentKey)
	default:
		return value
	}
}

func (s *logSanitizer) sanitizeCollection(value interface{}, elem interface{}, parentKey string) interface{} {
	var values []interface{}
	switch typedValue := value.(type) {
	case []interface{}:
		values = typedValue
	case *schema.Set:
		values = typedValue.List()
	default:
		return value
	}

	safeValues := make([]interface{}, len(values))
	for index, item := range values {
		safeValues[index] = s.sanitizeElement(item, elem, parentKey)
	}
	return safeValues
}

func (s *logSanitizer) sanitizeMapValues(value interface{}, elem interface{}, parentKey string) interface{} {
	switch values := value.(type) {
	case map[string]interface{}:
		safeValues := make(map[string]interface{}, len(values))
		for key, item := range values {
			safeValues[key] = s.sanitizeElement(item, elem, parentKey)
		}
		return safeValues
	case map[string]string:
		safeValues := make(map[string]interface{}, len(values))
		for key, item := range values {
			safeValues[key] = s.sanitizeElement(item, elem, parentKey)
		}
		return safeValues
	default:
		return value
	}
}

func (s *logSanitizer) sanitizeElement(value interface{}, elem interface{}, parentKey string) interface{} {
	switch field := elem.(type) {
	case *schema.Resource:
		if values, ok := value.(map[string]interface{}); ok {
			return s.sanitizeMap(values, field.Schema, parentKey)
		}
	case *schema.Schema:
		return s.sanitizeSchemaValue(value, field, parentKey)
	}

	return value
}

func (s *logSanitizer) mask(value interface{}) string {
	switch value := value.(type) {
	case string:
		if value != "" {
			s.secrets = append(s.secrets, value)
		}
	case []interface{}:
		for _, item := range value {
			s.mask(item)
		}
	case *schema.Set:
		s.mask(value.List())
	case map[string]interface{}:
		for _, item := range value {
			s.mask(item)
		}
	case map[string]string:
		for _, item := range value {
			s.mask(item)
		}
	}
	return "***"
}

func (s *logSanitizer) redactText(text string, additionalSecrets ...string) string {
	secrets := append(append([]string{}, s.secrets...), additionalSecrets...)
	variants := map[string]struct{}{}
	pythonEscape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		variants[secret] = struct{}{}
		encoded, _ := json.Marshal(secret)
		variants[string(encoded[1:len(encoded)-1])] = struct{}{}
		escaped := pythonEscape.Replace(secret)
		variants[escaped] = struct{}{}
		variants[strings.ReplaceAll(escaped, "'", "\\'")] = struct{}{}
	}

	// Replace longer values first so overlapping secrets are fully masked.
	ordered := make([]string, 0, len(variants))
	for value := range variants {
		ordered = append(ordered, value)
	}
	sort.Slice(ordered, func(i, j int) bool { return len(ordered[i]) > len(ordered[j]) })
	replacements := make([]string, 0, len(ordered)*2)
	for _, value := range ordered {
		replacements = append(replacements, value, "***")
	}
	return strings.NewReplacer(replacements...).Replace(text)
}
