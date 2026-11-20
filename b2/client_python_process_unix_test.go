//go:build !windows

//####################################################################
//
// File: b2/client_python_process_unix_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

type cancellationPIDs struct {
	Parent     int `json:"parent"`
	Grandchild int `json:"grandchild"`
}

func TestApplyStopsBindingsProcessTree(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Fatalf("the default process-tree regression requires python3: %v", err)
	}
	fixture, err := filepath.Abs("testdata/cancellation_worker.py")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	launcher := filepath.Join(dir, "python-launcher")
	if err := os.WriteFile(launcher, []byte("#!/bin/sh\nexec \"$B2_TEST_CANCEL_PYTHON\" \"$B2_TEST_CANCEL_SCRIPT\" \"$@\"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	if err := listener.(*net.TCPListener).SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	pidFile := filepath.Join(dir, "pids.json")
	t.Setenv("B2_TEST_CANCEL_PYTHON", python)
	t.Setenv("B2_TEST_CANCEL_SCRIPT", fixture)
	t.Setenv("B2_TEST_CANCEL_ADDRESS", listener.Addr().String())
	t.Setenv("B2_TEST_CANCEL_PIDS", pidFile)
	t.Setenv("B2_TEST_PLAIN_LAUNCHER", "1")
	provider := New("test", launcher)()
	client := Client{Exec: launcher, DataSourcesMap: provider.DataSourcesMap}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	var applyErr error
	var pids cancellationPIDs
	exited := make(map[int]bool)
	t.Cleanup(func() {
		cancel()
		if pids.Parent == 0 || pids.Grandchild == 0 {
			if raw, err := os.ReadFile(pidFile); err == nil {
				_ = json.Unmarshal(raw, &pids)
			}
		}
		// The mutant must not leave its surviving grandchild behind.
		for _, pid := range []int{pids.Grandchild, pids.Parent} {
			if pid > 0 && pid != os.Getpid() && !exited[pid] {
				if process, err := os.FindProcess(pid); err == nil {
					_ = process.Kill()
				}
			}
		}
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("Apply is still running during cleanup")
		}
	})
	go func() {
		applyErr = client.Apply(ctx, OpDataSourceRead, &AccountInfoInput{}, nil)
		close(done)
	}()
	conn, err := listener.Accept()
	if err != nil {
		t.Fatalf("Python worker did not announce startup: %v", err)
	}
	defer func() { _ = conn.Close() }()
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := json.NewDecoder(conn).Decode(&pids); err != nil {
		t.Fatalf("read launcher and grandchild PIDs: %v", err)
	}
	if pids.Parent <= 0 || pids.Grandchild <= 0 || pids.Parent == pids.Grandchild ||
		pids.Parent == os.Getpid() || pids.Grandchild == os.Getpid() {
		t.Fatalf("invalid fixture PIDs: %+v", pids)
	}
	for name, pid := range map[string]int{"launcher": pids.Parent, "grandchild": pids.Grandchild} {
		if err := syscall.Kill(pid, 0); err != nil {
			t.Fatalf("%s was not alive before cancellation: %v", name, err)
		}
		if group, err := syscall.Getpgid(pid); err != nil || group != pids.Parent {
			t.Fatalf("%s process group = %d, error = %v; want launcher PID %d", name, group, err, pids.Parent)
		}
	}
	t.Logf("ready: launcher PID %d, grandchild PID %d", pids.Parent, pids.Grandchild)
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Apply did not return within 5s of cancellation")
	}
	if !errors.Is(applyErr, context.Canceled) {
		t.Errorf("Apply error = %v, want context.Canceled", applyErr)
	}
	for name, pid := range map[string]int{"launcher": pids.Parent, "grandchild": pids.Grandchild} {
		ticker := time.NewTicker(10 * time.Millisecond)
		deadline := time.NewTimer(5 * time.Second)
		gone := false
		for !gone {
			if errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) {
				exited[pid] = true
				gone = true
				break
			}
			select {
			case <-ticker.C:
			case <-deadline.C:
				t.Errorf("%s PID %d still exists after Apply returned", name, pid)
				gone = true
			}
		}
		ticker.Stop()
		deadline.Stop()
	}
}
