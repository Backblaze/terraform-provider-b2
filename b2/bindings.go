//####################################################################
//
// File: b2/bindings.go
//
// Copyright 2024 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"embed"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	bindings *string
	//go:embed py-terraform-provider-b2
	content embed.FS
	lock    = &sync.Mutex{}
)

func GetBindings() (string, error) {
	if bindings == nil {
		lock.Lock()
		defer lock.Unlock()
	}
	if bindings != nil {
		return *bindings, nil
	}

	sourceFile, err := content.Open("py-terraform-provider-b2")
	if err != nil {
		return "", err
	}
	defer func() { _ = sourceFile.Close() }()

	var tmpPattern string
	if runtime.GOOS == "windows" {
		tmpPattern = "py-terraform-provider*.exe"
	} else {
		tmpPattern = "py-terraform-provider*"
	}

	destinationFile, err := os.CreateTemp("", tmpPattern)
	if err != nil {
		return "", err
	}
	defer func() { _ = destinationFile.Close() }()

	destinationPath := filepath.ToSlash(destinationFile.Name())

	if _, err := io.Copy(destinationFile, sourceFile); err != nil {
		return destinationPath, err
	}

	if err := destinationFile.Close(); err != nil {
		return destinationPath, err
	}

	err = os.Chmod(destinationPath, 0770)
	if err != nil {
		return destinationPath, err
	}

	bindings = &destinationPath
	log.Printf("Extracted pybindings: %s\n", *bindings)
	return *bindings, nil
}
