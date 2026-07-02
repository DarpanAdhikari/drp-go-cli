//go:build windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"unsafe"

	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
	"golang.org/x/sys/windows"
)

type procHandle struct {
	job     windows.Handle
	process windows.Handle
	pid     uint32
}

// startProcessTree launches "go run <args>" SUSPENDED, assigns it to a
// Job Object configured with KILL_ON_JOB_CLOSE, then resumes it. This
// guarantees the compiled binary that go run execs is contained too —
// this is the piece equivalent to Setpgid on Unix, and its absence is
// why a plain cmd.Process.Kill() on Windows leaves the real server
// running and still bound to its port.
func startProcessTree(goArgs []string) (*procHandle, error) {
	goPath, err := findTrustedGo()
	if err != nil {
		return nil, err
	}

	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("create job object: %w", err)
	}

	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}
	if _, err := windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	); err != nil {
		windows.CloseHandle(job)
		return nil, fmt.Errorf("configure job object: %w", err)
	}

	cmdLine, err := windows.UTF16PtrFromString(buildCommandLine(append([]string{goPath}, goArgs...)))
	if err != nil {
		windows.CloseHandle(job)
		return nil, err
	}

	si := &windows.StartupInfo{
		Cb:        uint32(unsafe.Sizeof(windows.StartupInfo{})),
		StdInput:  windows.Handle(os.Stdin.Fd()),
		StdOutput: windows.Handle(os.Stdout.Fd()),
		StdErr:    windows.Handle(os.Stderr.Fd()),
		Flags:     windows.STARTF_USESTDHANDLES,
	}
	pi := &windows.ProcessInformation{}

	err = windows.CreateProcess(
		nil,
		cmdLine,
		nil,
		nil,
		true,
		windows.CREATE_SUSPENDED|windows.CREATE_NEW_PROCESS_GROUP,
		nil,
		nil,
		si,
		pi,
	)
	if err != nil {
		windows.CloseHandle(job)
		return nil, fmt.Errorf("create process: %w", err)
	}
	defer windows.CloseHandle(pi.Thread)

	if err := windows.AssignProcessToJobObject(job, pi.Process); err != nil {
		_ = windows.TerminateProcess(pi.Process, 1)
		windows.CloseHandle(pi.Process)
		windows.CloseHandle(job)
		return nil, fmt.Errorf("assign process to job: %w", err)
	}

	if _, err := windows.ResumeThread(pi.Thread); err != nil {
		_ = windows.TerminateProcess(pi.Process, 1)
		windows.CloseHandle(pi.Process)
		windows.CloseHandle(job)
		return nil, fmt.Errorf("resume thread: %w", err)
	}

	return &procHandle{job: job, process: pi.Process, pid: pi.ProcessId}, nil
}

// stopProcess terminates the job, which atomically kills every process
// in the tree in one kernel call via JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE.
func stopProcess(h *procHandle) {
	if h == nil {
		return
	}
	output.Info("Stopping previous process...")

	_ = windows.TerminateJobObject(h.job, 0)
	_, _ = windows.WaitForSingleObject(h.process, uint32(5*time.Second/time.Millisecond))

	windows.CloseHandle(h.process)
	windows.CloseHandle(h.job)
	output.Success("Previous process stopped.")
}

func findTrustedGo() (string, error) {
	root := os.Getenv("GOROOT")
	if root != "" {
		candidate := root + `\bin\go.exe`
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return exec.LookPath("go.exe")
}

func buildCommandLine(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		if strings.ContainsAny(a, " \t\"") {
			a = `"` + strings.ReplaceAll(a, `"`, `\"`) + `"`
		}
		quoted[i] = a
	}
	return strings.Join(quoted, " ")
}
