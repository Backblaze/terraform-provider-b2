//go:build !windows

//####################################################################
//
// File: b2/client_process_unix.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func bindingsOutput(cmd *exec.Cmd) ([]byte, error) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	killGroup := func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.Cancel = killGroup
	output, err := cmd.Output()
	if err != nil {
		// Also clean up workers if the launcher exits before closing its pipes.
		_ = killGroup()
	}
	return output, err
}
