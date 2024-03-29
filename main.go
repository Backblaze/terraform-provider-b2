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
	"bufio"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/Backblaze/terraform-provider-b2/b2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/markbates/pkger"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	pkgerInput string = "/python-bindings/dist/py-terraform-provider-b2"
	version    string = "dev"
)

func extractPybindings(sourcePath string) (string, error) {
	sourceFile, err := pkger.Open(sourcePath)
	if err != nil {
		return "", err
	}
	defer sourceFile.Close()

	destinationFile, err := ioutil.TempFile("", "py-terraform-provider")
	if err != nil {
		return "", err
	}
	defer destinationFile.Close()

	destinationPath := destinationFile.Name()
	reader := bufio.NewReader(sourceFile)
	buf := make([]byte, 2048)

	for {
		_, err := reader.Read(buf)

		if err != nil {
			if err != io.EOF {
				return destinationPath, err
			}

			_, err = destinationFile.Seek(0, 0)
			if err != nil {
				return destinationPath, err
			}

			break
		}

		_, err = destinationFile.Write(buf)
		if err != nil {
			return destinationPath, err
		}
	}

	destinationFile.Close()

	err = os.Chmod(destinationPath, 0770)
	if err != nil {
		return destinationPath, err
	}

	log.Printf("[TRACE] Extracted pybindings: %s\n", destinationPath)
	return destinationPath, nil
}

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	pkgerOutput, err := extractPybindings(filepath.FromSlash(pkgerInput))
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer os.Remove(pkgerOutput)

	opts := &plugin.ServeOpts{ProviderFunc: b2.New(version, pkgerOutput), Debug: debugMode}
	plugin.Serve(opts)
}
