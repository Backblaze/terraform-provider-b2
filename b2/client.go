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

const (
	DATA_SOURCE_READ string = "data_source_read"

	RESOURCE_CREATE string = "resource_create"
	RESOURCE_READ   string = "resource_read"
	RESOURCE_UPDATE string = "resource_update"
	RESOURCE_DELETE string = "resource_delete"
)

type Client struct {
	Exec                 string
	UserAgentAppend      string
	ApplicationKeyId     string
	ApplicationKey       string
	Endpoint             string
	DataSources          map[string][]string
	Resources            map[string][]string
	SensitiveDataSources map[string]map[string]bool
	SensitiveResources   map[string]map[string]bool
}

func (c Client) apply(ctx context.Context, name string, op string, input map[string]interface{}) (map[string]interface{}, error) {
	tflog.Info(ctx, "Executing pybindings", map[string]interface{}{
		"name": name,
		"op":   op,
	})

	tflog.Debug(ctx, "Input for pybindings", map[string]interface{}{
		"input": input,
	})

	cmd := exec.Command(c.Exec, name, op)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("B2_USER_AGENT_APPEND=%s", c.UserAgentAppend))

	input["provider_application_key_id"] = c.ApplicationKeyId
	input["provider_application_key"] = c.ApplicationKey
	input["provider_endpoint"] = c.Endpoint

	inputJson, err := json.Marshal(input)
	if err != nil {
		// Should never happen
		return nil, err
	}
	cmd.Stdin = bytes.NewReader(inputJson)

	outputJson, err := cmd.Output()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.Stderr != nil && len(exitErr.Stderr) > 0 {
				tflog.Error(ctx, "Error in pybindings", map[string]interface{}{
					"stderr": fmt.Errorf(string(exitErr.Stderr)),
				})
				return nil, fmt.Errorf(string(exitErr.Stderr))
			}
			return nil, fmt.Errorf("failed to execute")
		} else {
			tflog.Error(ctx, "Error", map[string]interface{}{
				"err": err,
			})
			return nil, err
		}
	}

	output := map[string]interface{}{}
	err = json.Unmarshal(outputJson, &output)
	if err != nil {
		return nil, err
	}

	resourceName := "b2_" + name
	var sensitiveSchemaMap map[string]bool
	if op == DATA_SOURCE_READ {
		sensitiveSchemaMap = c.SensitiveDataSources[resourceName]
	} else {
		sensitiveSchemaMap = c.SensitiveResources[resourceName]
	}

	// Do not log application_key
	safeOutput := map[string]interface{}{}
	for k, v := range output {
		if sensitiveSchemaMap[k] {
			safeOutput[k] = "***"
		} else {
			safeOutput[k] = v
		}
	}
	tflog.Debug(ctx, "Safe output from pybindings", map[string]interface{}{
		"output": safeOutput,
	})

	return output, nil
}

func (c Client) populate(ctx context.Context, name string, op string, output map[string]interface{}, d *schema.ResourceData) error {
	resourceName := "b2_" + name
	var schemaList []string
	if op == DATA_SOURCE_READ {
		schemaList = c.DataSources[resourceName]
	} else {
		schemaList = c.Resources[resourceName]
	}

	for _, k := range schemaList {
		v, ok := output[k]
		if !ok {
			return fmt.Errorf("error getting %s", k)
		}
		if err := d.Set(k, v); err != nil {
			return fmt.Errorf("error setting %s: %s", k, err)
		}
	}

	return nil
}
