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
	"bufio"
	"github.com/markbates/pkger"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	bindings *string
	lock     = &sync.Mutex{}
)

func GetBindings(sourcePath string, testing bool) (string, error) {
	if bindings == nil {
		lock.Lock()
		defer lock.Unlock()
	}
	if bindings != nil {
		return *bindings, nil
	}

	var sourceFile io.ReadCloser
	if testing == true {
		src, err := os.Open(sourcePath)
		if err != nil {
			return "", err
		}
		sourceFile = src
	} else {
		src, err := pkger.Open(sourcePath)
		if err != nil {
			return "", err
		}
		sourceFile = src
	}
	defer sourceFile.Close()

	var tmpPattern string
	if runtime.GOOS == "windows" {
		tmpPattern = "py-terraform-provider*.exe"
	} else {
		tmpPattern = "py-terraform-provider*"
	}

	destinationFile, err := ioutil.TempFile("", tmpPattern)
	if err != nil {
		return "", err
	}
	defer destinationFile.Close()

	destinationPath := filepath.ToSlash(destinationFile.Name())
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

	bindings = &destinationPath
	log.Printf("[TRACE] Extracted pybindings: %s\n", *bindings)
	return *bindings, nil
}
