//####################################################################
//
// File: b2/provider_test.go
//
// Copyright 2020 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"b2": func() (*schema.Provider, error) {
		pybindings, err := GetBindings()
		if err != nil {
			log.Fatal(err.Error())
			return nil, err
		}
		return New("test", pybindings)(), nil
	},
}

func TestProvider(t *testing.T) {
	pybindings, err := GetBindings()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := New("test", pybindings)().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	_, present := os.LookupEnv("B2_TEST_APPLICATION_KEY_ID")
	if !present {
		t.Fatal("B2_TEST_APPLICATION_KEY_ID is not set")
	}
	_ = os.Setenv("B2_APPLICATION_KEY_ID", os.Getenv("B2_TEST_APPLICATION_KEY_ID"))

	_, present = os.LookupEnv("B2_TEST_APPLICATION_KEY")
	if !present {
		t.Fatal("B2_TEST_APPLICATION_KEY is not set")
	}
	_ = os.Setenv("B2_APPLICATION_KEY", os.Getenv("B2_TEST_APPLICATION_KEY"))
}

// Utility functions

func createTempFile(t *testing.T, data string) string {
	tmpFile, err := ioutil.TempFile("", "test-b2-tfp")
	if err != nil {
		t.Fatal(err)
	}
	filename := tmpFile.Name()

	err = ioutil.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		os.Remove(filename)
		t.Fatal(err)
	}

	return filepath.ToSlash(filename)
}
