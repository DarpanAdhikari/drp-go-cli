package cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
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

func isPortFree(port string) bool {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

func getAppPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}

func waitForPortFree(port string, maxAttempts int, delayMs time.Duration) {
	for i := 0; i < maxAttempts; i++ {
		if isPortFree(port) {
			return
		}
		if i < maxAttempts-1 {
			output.Warn("Port %s still in use, waiting...", port)
			time.Sleep(delayMs)
		}
	}
}

func runWithWatch(goArgs []string) error {
	var currentCmd *exec.Cmd
	var mu sync.Mutex
	port := getAppPort()

	startProcess := func() {
		mu.Lock()
		defer mu.Unlock()
		if currentCmd != nil && currentCmd.Process != nil {
			output.Info("Stopping previous process...")

			if runtime.GOOS == "windows" {
				// Windows: direct kill (no graceful shutdown available)
				_ = currentCmd.Process.Kill()
				_ = currentCmd.Wait()
				output.Success("Previous process stopped.")
			} else {
				// Unix/Linux/macOS: graceful shutdown with timeout
				_ = currentCmd.Process.Signal(syscall.SIGTERM)

				done := make(chan error, 1)
				go func() {
					done <- currentCmd.Wait()
				}()

				select {
				case <-done:
					output.Success("Previous process stopped.")
				case <-time.After(5 * time.Second):
					output.Warn("Graceful shutdown timed out. Killing process...")
					_ = currentCmd.Process.Kill()
					<-done
				}
			}
			// Wait for port to be free
			output.Info("Waiting for port %s to be released...", port)
			waitForPortFree(port, 10, 500*time.Millisecond)
		} else {
			output.Info("Watching for changes...")
			output.Info("Starting go %v", goArgs)
		}

		currentCmd = exec.Command("go", goArgs...)
		currentCmd.Stdout = os.Stdout
		currentCmd.Stderr = os.Stderr
		currentCmd.Env = os.Environ()
		// Disable Stdin to prevent background process issues

		if err := currentCmd.Start(); err != nil {
			output.Fail("Failed to start: %v", err)
		}
	}

	startProcess()

	lastModTime := getLatestModTime()
	ticker := time.NewTicker(800 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		t := getLatestModTime()
		if t.After(lastModTime) {
			lastModTime = t
			startProcess()
		}
	}
	return nil
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
