//####################################################################
//
// File: main.go
//
// Copyright 2020 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package main

import (
	"flag"
	"github.com/Backblaze/terraform-provider-b2/b2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"log"
	"os"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	pybindingsSource string = "/python-bindings/dist/py-terraform-provider-b2"
	version          string = "dev"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	pybindings, err := b2.GetBindings(pybindingsSource, false)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer os.Remove(pybindings)

	opts := &plugin.ServeOpts{ProviderFunc: b2.New(version, pybindings), Debug: debugMode}
	plugin.Serve(opts)
}
