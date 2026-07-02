//go:build linux || darwin

package cmd

import (
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
)

type procHandle struct {
	cmd *exec.Cmd
}

// startProcessTree starts "go run <args>" in a new process group so the
// go run wrapper and its child binary can be signaled together later.
// This is the piece that was missing in the version that crash-looped:
// without Setpgid, signaling cmd.Process only reaches the wrapper, and
// the actual server binary is left running and still bound to its port.
func startProcessTree(goArgs []string) (*procHandle, error) {
	cmd := exec.Command("go", goArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &procHandle{cmd: cmd}, nil
}

// stopProcess gracefully stops the whole process group, escalating to
// SIGKILL if it doesn't exit within the timeout.
func stopProcess(h *procHandle) {
	if h == nil || h.cmd == nil || h.cmd.Process == nil {
		return
	}

	pgid, err := syscall.Getpgid(h.cmd.Process.Pid)
	if err != nil {
		_ = h.cmd.Wait()
		return
	}

	output.Info("Stopping previous process...")
	_ = syscall.Kill(-pgid, syscall.SIGTERM) // negative pid = whole group

	done := make(chan error, 1)
	go func() { done <- h.cmd.Wait() }()

	select {
	case <-done:
		output.Success("Previous process stopped.")
	case <-time.After(5 * time.Second):
		output.Warn("Graceful shutdown timed out. Killing process group...")
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		<-done
	}
}