package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [target]",
	Short: "Build a production binary from ./cmd/<target>",
	Long: `Compiles ./cmd/<target> into ./dist/<target> (or a custom path with --output).

If no target is given it defaults to "api".

Examples:
  drp build           # builds ./cmd/api -> ./dist/api
  drp build worker    # builds ./cmd/worker -> ./dist/worker
  drp build -o myapp  # builds ./cmd/api -> ./myapp`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		target := "api"
		if len(args) > 0 {
			target = args[0]
		}
		if !regexp.MustCompile(`^[A-Za-z0-9_-]+$`).MatchString(target) {
			return fmt.Errorf("build: invalid target %q", target)
		}
		cmdDir := filepath.Join("cmd", target)
		if stat, err := os.Stat(cmdDir); err != nil || !stat.IsDir() {
			return fmt.Errorf("build: %s not found", cmdDir)
		}

		outputPath, _ := c.Flags().GetString("output")
		if outputPath == "" {
			outputPath = filepath.Join("dist", target)
		}

		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return fmt.Errorf("build: creating output directory: %w", err)
		}

		goArgs := []string{"build", "-o", outputPath, "./" + filepath.ToSlash(cmdDir)}
		cmd := exec.Command("go", goArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()

		output.Info("go %v", goArgs)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("build failed")
		}
		output.Success("built %s", outputPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringP("output", "o", "", "Output path (default: ./dist/<target>)")
}
