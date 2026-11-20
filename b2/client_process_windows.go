//####################################################################
//
// File: b2/client_process_windows.go
//
// Copyright 2026 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func bindingsOutput(cmd *exec.Cmd) ([]byte, error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = windows.CloseHandle(job) }()
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	var pinned runtime.Pinner
	pinned.Pin(&limits)
	defer pinned.Unpin()
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		return nil, err
	}

	// Contain the launcher before it can create PyInstaller's Python worker.
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_SUSPENDED}
	var startup sync.Mutex
	startup.Lock()
	cmd.Cancel = func() error {
		startup.Lock()
		defer startup.Unlock()
		return windows.TerminateJobObject(job, 1)
	}
	var stdout bytes.Buffer
	var stderr bindingsStderr
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Start(); err != nil {
		startup.Unlock()
		return nil, err
	}
	err = assignAndResumeBindings(job, uint32(cmd.Process.Pid))
	startup.Unlock()
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, fmt.Errorf("prepare bindings process: %w", err)
	}
	err = cmd.Wait()
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitErr.Stderr = stderr.Bytes()
	}
	return stdout.Bytes(), err
}

func assignAndResumeBindings(job windows.Handle, pid uint32) error {
	process, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, pid)
	if err != nil {
		return err
	}
	defer func() { _ = windows.CloseHandle(process) }()
	if err := windows.AssignProcessToJobObject(job, process); err != nil {
		return err
	}
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPTHREAD, 0)
	if err != nil {
		return err
	}
	defer func() { _ = windows.CloseHandle(snapshot) }()
	entry := windows.ThreadEntry32{}
	entry.Size = uint32(unsafe.Sizeof(entry))
	for err = windows.Thread32First(snapshot, &entry); err == nil; err = windows.Thread32Next(snapshot, &entry) {
		if entry.OwnerProcessID != pid {
			continue
		}
		thread, err := windows.OpenThread(windows.THREAD_SUSPEND_RESUME, false, entry.ThreadID)
		if err != nil {
			return err
		}
		_, err = windows.ResumeThread(thread)
		_ = windows.CloseHandle(thread)
		return err
	}
	return fmt.Errorf("find suspended bindings thread: %w", err)
}

// Match exec.Cmd.Output's bounded capture: retain the first and last 32 KiB.
type bindingsStderr struct {
	prefix, suffix []byte
	omitted        int64
}

func (b *bindingsStderr) Write(p []byte) (int, error) {
	const limit = 32 << 10
	n := len(p)
	first := min(limit-len(b.prefix), len(p))
	b.prefix = append(b.prefix, p[:first]...)
	p = p[first:]
	if overflow := len(b.suffix) + len(p) - limit; overflow > 0 {
		b.omitted += int64(overflow)
		if overflow >= len(b.suffix) {
			p = p[overflow-len(b.suffix):]
			b.suffix = b.suffix[:0]
		} else {
			b.suffix = b.suffix[overflow:]
		}
	}
	b.suffix = append(b.suffix, p...)
	return n, nil
}

func (b *bindingsStderr) Bytes() []byte {
	output := append([]byte(nil), b.prefix...)
	if b.omitted > 0 {
		output = fmt.Appendf(output, "\n... omitting %d bytes ...\n", b.omitted)
	}
	return append(output, b.suffix...)
}
