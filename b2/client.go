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
	"os/exec"
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
	Version          string
	ApplicationKeyId string
	ApplicationKey   string
}

func (c Client) apply(name string, op string, input map[string]interface{}) (map[string]interface{}, error) {
	cmd := exec.Command(c.Exec, name, op)

	input["_application_key_id"] = c.ApplicationKeyId
	input["_application_key"] = c.ApplicationKey

	inputJson, err := json.Marshal(input)
	if err != nil {
		// Should never happen
		return nil, err
	}

	cmd.Stdin = bytes.NewReader(inputJson)

	log.Printf("[TRACE] Executing pybindings for '%s' and '%s' operation\n", name, op)

	outputJson, err := cmd.Output()

	log.Printf("[TRACE] Output from pybindings: %+v\n", string(outputJson))

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

	return output, nil
}
