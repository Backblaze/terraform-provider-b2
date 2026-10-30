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

	tflog.Debug(ctx, "Input for pybindings", map[string]interface{}{
		"input": sanitizeOutput(inputMap, schemaMap),
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
				err := fmt.Errorf("%s", string(exitErr.Stderr))
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
			"output": sanitizeOutput(output, schemaMap),
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

// signed_url remains non-sensitive in the Terraform schema for compatibility,
// but it is redacted from debug logs.
var logOnlySensitiveKeys = map[string]struct{}{
	"signed_url": {},
}

func sanitizeOutput(output interface{}, schemaMap map[string]*schema.Schema) map[string]interface{} {
	outputMap, ok := output.(map[string]interface{})
	if !ok {
		outputMap = convertStructToMap(output)
	}

	return sanitizeMap(outputMap, schemaMap)
}

func sanitizeMap(values map[string]interface{}, schemaMap map[string]*schema.Schema) map[string]interface{} {
	safeValues := make(map[string]interface{}, len(values))
	for key, value := range values {
		if _, ok := logOnlySensitiveKeys[key]; ok {
			safeValues[key] = "***"
			continue
		}

		if field, ok := schemaMap[key]; ok {
			safeValues[key] = sanitizeSchemaValue(value, field)
			continue
		}

		safeValues[key] = value
	}

	return safeValues
}

func sanitizeSchemaValue(value interface{}, field *schema.Schema) interface{} {
	if field.Sensitive {
		return "***"
	}

	switch field.Type {
	case schema.TypeList, schema.TypeSet:
		return sanitizeCollection(value, field.Elem)
	case schema.TypeMap:
		return sanitizeMapValues(value, field.Elem)
	default:
		return value
	}
}

func sanitizeCollection(value interface{}, elem interface{}) interface{} {
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
		safeValues[index] = sanitizeElement(item, elem)
	}
	return safeValues
}

func sanitizeMapValues(value interface{}, elem interface{}) interface{} {
	switch values := value.(type) {
	case map[string]interface{}:
		safeValues := make(map[string]interface{}, len(values))
		for key, item := range values {
			safeValues[key] = sanitizeElement(item, elem)
		}
		return safeValues
	case map[string]string:
		safeValues := make(map[string]interface{}, len(values))
		for key, item := range values {
			safeValues[key] = sanitizeElement(item, elem)
		}
		return safeValues
	default:
		return value
	}
}

func sanitizeElement(value interface{}, elem interface{}) interface{} {
	switch field := elem.(type) {
	case *schema.Resource:
		if values, ok := value.(map[string]interface{}); ok {
			return sanitizeMap(values, field.Schema)
		}
	case *schema.Schema:
		return sanitizeSchemaValue(value, field)
	}

	return value
}
