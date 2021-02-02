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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

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
	Exec             string
	UserAgentAppend  string
	ApplicationKeyId string
	ApplicationKey   string
	Endpoint         string
	DataSources      map[string][]string
	Resources        map[string][]string
}

func (c Client) apply(name string, op string, input map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("[TRACE] Executing pybindings for '%s' and '%s' operation\n", name, op)
	log.Printf("[TRACE] Input for pybindings: %+v\n", input)

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
				return nil, fmt.Errorf(string(exitErr.Stderr))
			}
			return nil, fmt.Errorf("failed to execute")
		} else {
			return nil, err
		}
	}

	output := map[string]interface{}{}
	err = json.Unmarshal(outputJson, &output)
	if err != nil {
		return nil, err
	}

	// Do not log application_key
	safeOutput := map[string]interface{}{}
	for k, v := range output {
		if k != "application_key" {
			safeOutput[k] = v
		}
	}
	log.Printf("[TRACE] Output from pybindings (without application key): %+v\n", safeOutput)

	return output, nil
}

func (c Client) populate(name string, op string, output map[string]interface{}, d *schema.ResourceData) error {
	resourceName := "b2_" + name
	var schemaKeys []string
	if op == DATA_SOURCE_READ {
		schemaKeys = c.DataSources[resourceName]
	} else {
		schemaKeys = c.Resources[resourceName]
	}

	for _, k := range schemaKeys {
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
