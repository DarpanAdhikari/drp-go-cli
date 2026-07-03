package cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [target] [args...]",
	Short: "Run a Go command from ./cmd/<target>",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		target := args[0]
		if !regexp.MustCompile(`^[A-Za-z0-9_-]+$`).MatchString(target) {
			return fmt.Errorf("run: invalid target %q", target)
		}
		cmdDir := filepath.Join("cmd", target)
		if stat, err := os.Stat(cmdDir); err != nil || !stat.IsDir() {
			return fmt.Errorf("run: %s not found", cmdDir)
		}
		goArgs := append([]string{"run", "./" + filepath.ToSlash(cmdDir)}, args[1:]...)
		watch, _ := c.Flags().GetBool("watch")
		if watch {
			return runWithWatch(goArgs)
		}
		goCmd := exec.Command("go", goArgs...)
		goCmd.Stdout = os.Stdout
		goCmd.Stderr = os.Stderr
		goCmd.Stdin = os.Stdin
		goCmd.Env = os.Environ()
		output.Info("go %v", goArgs)
		return goCmd.Run()
	},
}

// isPortFree does a real bind-test rather than trusting process state
// alone. It's a genuine extra safety net (e.g. some *other* unrelated
// process could be squatting on the port), but it is NOT a substitute
// for actually terminating the process tree — see stopProcess in
// run_unix.go / run_windows.go, which is what does that job. This
// function only tells you whether that job succeeded.
func isPortFree(port string) bool {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

func getAppPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}

// waitForPortFree polls until the port is free or attempts are
// exhausted. It returns whether the port ended up free — the caller
// MUST check this and refuse to start a new process if false, or you
// get exactly the crash-loop seen in testing (old server still bound,
// new server started anyway, bind fails).
func waitForPortFree(port string, maxAttempts int, delay time.Duration) bool {
	for i := 0; i < maxAttempts; i++ {
		if isPortFree(port) {
			return true
		}
		if i < maxAttempts-1 {
			output.Warn("Port %s still in use, waiting...", port)
			time.Sleep(delay)
		}
	}
	return false
}

func runWithWatch(goArgs []string) error {
	var handle *procHandle
	var mu sync.Mutex
	restartCh := make(chan struct{}, 1)
	port := getAppPort()

	startProcess := func() {
		mu.Lock()
		if handle != nil {
			stopProcess(handle)

			output.Info("Waiting for port %s to be released...", port)
			if !waitForPortFree(port, 10, 500*time.Millisecond) {
				output.Fail("Port %s did not free up after stopping the previous process. Skipping restart — check for a leaked/unrelated process on this port.", port)
				handle = nil
				mu.Unlock()
				return
			}
		} else {
			output.Info("Watching for changes...")
			output.Info("Starting go %v", goArgs)
		}

		h, err := startProcessTree(goArgs)
		if err != nil {
			output.Fail("Failed to start: %v", err)
			handle = nil
			mu.Unlock()
			return
		}
		handle = h
		mu.Unlock()
	}

	// Initial start
	startProcess()
	lastModTime := getLatestModTime()
	ticker := time.NewTicker(800 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		t := getLatestModTime()
		if t.After(lastModTime) {
			lastModTime = t
			select {
			case restartCh <- struct{}{}:
			default:
			}
		}
	}

	// Process restart requests - only latest wins
	go func() {
		for range restartCh {
			// Drain any pending requests, keeping only the latest
			done := false
			for !done {
				select {
				case <-restartCh:
				default:
					done = true
				}
			}
			startProcess()
		}
	}()

	// Keep main goroutine alive
	select {}
}

func getLatestModTime() time.Time {
	var latest time.Time
	_ = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "tmp" || name == "node_modules" || name == "database" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(info.Name(), ".go") || info.Name() == ".env" {
			if info.ModTime().After(latest) {
				latest = info.ModTime()
			}
		}
		return nil
	})
	return latest
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolP("watch", "w", false, "Watch for changes in .go and .env files and auto-restart")
}