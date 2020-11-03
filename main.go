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
	"fmt"
	"io"
	"os"

	"github.com/Backblaze/terraform-provider-b2/b2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/markbates/pkger"
)

var (
	exec = "/tmp/py-terraform-provider-b2"
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "dev"

	// goreleaser can also pass the specific commit if you want
	// commit  string = ""
)

func pybinding(s string, d string) error {
	f1, err := pkger.Open(s)
	if err != nil {
		return err
	}
	defer f1.Close()

	f2, err := os.Create(d)
	if err != nil {
		return err
	}
	defer f2.Close()

	reader := bufio.NewReader(f1)
	buf := make([]byte, 2048)

	for {
		_, err := reader.Read(buf)

		if err != nil {
			if err != io.EOF {
				return err
			}
			f2.Seek(0, 0)
			break
		}
		f2.Write(buf)
	}
	f2.Close()
	os.Chmod(d, 0770)

	return nil
}

func main() {
	err := pybinding("/python-bindings/dist/py-terraform-provider-b2", exec)
	if err != nil {
		fmt.Println(err)
	} else {
		plugin.Serve(&plugin.ServeOpts{ProviderFunc: b2.New(version, exec)})
	}
}
