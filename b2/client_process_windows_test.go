//####################################################################
//
// File: b2/client_process_windows_test.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"strings"
	"testing"
)

func TestApplyStopsBindingsProcessTree(t *testing.T) {
	t.Skip("plain Python launcher uses Unix process groups and PID probes; TestApplyStopsGoProcessTree covers the Windows Job Object")
}

func TestBindingsStderrRetainsPrefixAndSuffix(t *testing.T) {
	prefix, suffix := strings.Repeat("a", 32<<10), strings.Repeat("z", 32<<10)
	for _, tc := range []struct{ name, input, want string }{
		{name: "short", input: "plain error", want: "plain error"},
		{name: "at limit", input: prefix + suffix, want: prefix + suffix},
		{name: "truncated", input: prefix + "mid" + suffix, want: prefix + "\n... omitting 3 bytes ...\n" + suffix},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, chunk := range []int{17, 32 << 10, len(tc.input)} {
				var stderr bindingsStderr
				for remaining := tc.input; remaining != ""; {
					n := min(chunk, len(remaining))
					if written, err := stderr.Write([]byte(remaining[:n])); err != nil || written != n {
						t.Fatalf("Write = %d, %v; want %d, nil", written, err, n)
					}
					remaining = remaining[n:]
				}
				if string(stderr.Bytes()) != tc.want {
					t.Errorf("chunk size %d: stderr did not preserve the expected prefix and suffix", chunk)
				}
			}
		})
	}
}
