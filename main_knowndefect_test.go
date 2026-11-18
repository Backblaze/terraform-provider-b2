//go:build knowndefect

//####################################################################
//
// File: main_knowndefect_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package main

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Known-defect test (build tag "knowndefect"; run with `make test-known-defects`). When the defect is fixed,
// drop the build tag and the "TestKnownDefect_" prefix and rename this file to main_test.go.

// C38: in -debug mode the provider must be served with its registry address, otherwise the
// TF_REATTACH_PROVIDERS entry printed for the debugger is keyed "provider" and Terraform ignores it.
// The expected address string (registry.terraform.io/backblaze/b2) is an assumption; adjust it if the
// provider is published under another address.
func TestKnownDefect_C38_DebugModeAdvertisesProviderAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and runs the provider binary")
	}
	if runtime.GOOS == "windows" {
		t.Skip("not exercised on Windows")
	}
	bin := filepath.Join(t.TempDir(), "terraform-provider-b2")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	cmd := exec.Command(bin, "-debug")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	// The child extracts its python bindings into TMPDIR and is killed below, so it cannot clean up after
	// itself: give it a test-owned directory that is removed afterwards. It is made with os.MkdirTemp, not
	// t.TempDir, because the child's unix socket lives there and the long t.TempDir path exceeds the socket
	// path limit on macOS.
	childTmp, err := os.MkdirTemp("", "c38")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(childTmp) })
	cmd.Env = append(os.Environ(), "TMPDIR="+childTmp)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()

	found := make(chan string, 1)
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			// the line that carries the JSON config (the header line only names the variable)
			if strings.Contains(sc.Text(), `"Protocol"`) {
				found <- sc.Text()
				return
			}
		}
		found <- ""
	}()

	select {
	case line := <-found:
		if !strings.Contains(strings.ToLower(line), "registry.terraform.io/backblaze/b2") {
			t.Errorf("C38: debug reattach config does not carry the provider address: %q", line)
		}
	case <-time.After(60 * time.Second):
		t.Fatal("timed out waiting for the reattach configuration")
	}
}
