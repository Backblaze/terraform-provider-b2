//go:build linux || darwin

//####################################################################
//
// File: b2/file_version_open_unix_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestFileVersionOpenDoesNotBlockOnFIFO(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source-fifo")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		file, err := openFileVersionSource(path)
		if file != nil {
			_ = file.Close()
		}
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("nonblocking open failed: %v", err)
		}
	case <-time.After(time.Second):
		// Release a blocking reader before failing.
		if fd, err := syscall.Open(path, syscall.O_WRONLY|syscall.O_NONBLOCK, 0); err == nil {
			_ = syscall.Close(fd)
		}
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
		t.Fatal("opening a source changed to a FIFO blocked")
	}
}
