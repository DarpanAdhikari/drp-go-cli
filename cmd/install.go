package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the current binary as drp",
	RunE: func(c *cobra.Command, args []string) error {
		binDir, _ := c.Flags().GetString("bin-dir")
		noShell, _ := c.Flags().GetBool("no-shell")
		if binDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			binDir = filepath.Join(home, ".local", "bin")
		}

		exe, err := os.Executable()
		if err != nil {
			return err
		}
		dest := filepath.Join(binDir, executableName("drp"))
		if err := copyExecutable(exe, dest); err != nil {
			output.Fail("%v", err)
			return err
		}

		output.Success("Installed drp to %s", dest)
		if !noShell {
			if err := ensureLocalBinOnPath(binDir); err != nil {
				output.Info("Could not update shell profile: %v", err)
				output.Info("Add this to your shell profile: export PATH=\"%s:$PATH\"", binDir)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().String("bin-dir", "", "Directory to install drp into (default: ~/.local/bin)")
	installCmd.Flags().Bool("no-shell", false, "Do not update shell profile PATH")
}

func copyExecutable(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("install: create bin directory: %w", err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("install: open current executable: %w", err)
	}
	defer in.Close()

	tmp := dest + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return fmt.Errorf("install: create %q: %w", tmp, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return fmt.Errorf("install: copy executable: %w", err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("install: close executable: %w", err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		return fmt.Errorf("install: replace %q: %w", dest, err)
	}
	return os.Chmod(dest, 0o755)
}

func ensureLocalBinOnPath(binDir string) error {
	pathEntries := filepath.SplitList(os.Getenv("PATH"))
	for _, entry := range pathEntries {
		if samePath(entry, binDir) {
			return nil
		}
	}

	profile, shell, err := shellProfilePath()
	if err != nil {
		return err
	}
	line := fmt.Sprintf("\n# Added by drp CLI\nexport PATH=\"%s:$PATH\"\n", binDir)
	if shell == "fish" {
		line = fmt.Sprintf("\n# Added by drp CLI\nfish_add_path %s\n", binDir)
	}
	content, _ := os.ReadFile(profile)
	if strings.Contains(string(content), binDir) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(profile), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(profile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line)
	return err
}

func shellProfilePath() (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	shell := filepath.Base(os.Getenv("SHELL"))
	switch shell {
	case "zsh":
		return filepath.Join(home, ".zshrc"), shell, nil
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish"), shell, nil
	default:
		return filepath.Join(home, ".bashrc"), shell, nil
	}
}

func samePath(a, b string) bool {
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	return errA == nil && errB == nil && aa == bb
}

func executableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
