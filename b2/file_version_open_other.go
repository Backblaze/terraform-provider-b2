//go:build !linux && !darwin

//####################################################################
//
// File: b2/file_version_open_other.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import "os"

func openFileVersionSource(path string) (*os.File, error) {
	return os.Open(path)
}
