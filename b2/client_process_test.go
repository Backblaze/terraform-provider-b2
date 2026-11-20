//####################################################################
//
// File: b2/client_process_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestCIRequiresPackagedBindingsFixture(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestApplyStopsBindingsProcessTree$", "-test.v")
	cmd.Env = append(os.Environ(), "CI=true", "TF_ACC=", "B2_TEST_PACKAGED_BINDINGS=", "GO_WANT_B2_CLIENT_HELPER=")
	output, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 ||
		!bytes.Contains(output, []byte("CI requires B2_TEST_PACKAGED_BINDINGS")) {
		t.Fatalf("missing CI fixture must fail explicitly; error = %v, output:\n%s", err, output)
	}
}

// Offline CI builds testdata/cancellation_worker.py with PyInstaller --onefile
// and must supply its path. Local and live acceptance runs can use the Go fixture.
func TestApplyStopsBindingsProcessTree(t *testing.T) {
	executable := os.Getenv("B2_TEST_PACKAGED_BINDINGS")
	if executable == "" {
		if os.Getenv("CI") != "" && os.Getenv("TF_ACC") == "" {
			t.Fatal("CI requires B2_TEST_PACKAGED_BINDINGS pointing to the one-file PyInstaller fixture")
		}
		executable = os.Args[0]
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	if err := listener.(*net.TCPListener).SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GO_WANT_B2_CLIENT_HELPER", "cancel-launcher")
	t.Setenv("B2_TEST_CANCEL_ADDRESS", listener.Addr().String())
	provider := New("test", executable)()
	client := Client{Exec: executable, DataSourcesMap: provider.DataSourcesMap}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	var applyErr error
	go func() {
		applyErr = client.Apply(ctx, OpDataSourceRead, &AccountInfoInput{}, nil)
		close(done)
	}()
	conn, err := listener.Accept()
	if err != nil {
		t.Fatalf("bindings worker did not start: %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close() // lets the fixture exit even if cancellation is broken
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("Apply is still running during cleanup")
		}
	})
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if ready, err := bufio.NewReader(conn).ReadString('\n'); err != nil || ready != "started\n" {
		t.Fatalf("worker startup marker = %q, error = %v", ready, err)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Apply did not return within 5s of cancellation")
	}
	if !errors.Is(applyErr, context.Canceled) {
		t.Errorf("Apply error = %v, want context.Canceled", applyErr)
	}
	var b [1]byte
	if _, err := conn.Read(b[:]); !errors.Is(err, io.EOF) && !errors.Is(err, syscall.ECONNRESET) {
		t.Errorf("worker connection is still open after Apply returned: %v", err)
	}
}
