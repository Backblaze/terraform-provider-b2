//####################################################################
//
// File: b2/bindings_execution_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

func TestC19_ApplyBindingsExecutionErrors(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixtures exercise Unix execute permissions and executable format errors")
	}
	for _, tt := range []struct {
		name  string
		mode  os.FileMode
		cause error
		hint  string
	}{
		{"permission", 0o600, os.ErrPermission, "noexec"},
		{"format", 0o700, syscall.ENOEXEC, "format"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "py-terraform-provider-b2")
			if err := os.WriteFile(path, []byte("invalid executable"), tt.mode); err != nil {
				t.Fatal(err)
			}
			provider := New("test", path)()
			client := Client{Exec: path, DataSourcesMap: provider.DataSourcesMap}
			err := client.Apply(context.Background(), OpDataSourceRead, &AccountInfoInput{}, nil)
			if !errors.Is(err, tt.cause) {
				t.Fatalf("Apply error = %v, want wrapped %v", err, tt.cause)
			}
			if !strings.Contains(err.Error(), "TMPDIR") || !strings.Contains(err.Error(), tt.hint) {
				t.Fatalf("execution error lacks actionable %s/TMPDIR guidance: %v", tt.hint, err)
			}
		})
	}
}
