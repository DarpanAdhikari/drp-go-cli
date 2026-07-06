package cmd

import (
	"os"
	"os/exec"

	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
	"github.com/spf13/cobra"
)

var docsGenerateCmd = &cobra.Command{
	Use:   "docs:generate",
	Short: "Regenerate swagger documentation (swag init)",
	Long: `Parses all Swaggo annotations in your source files and regenerates
the three documentation files under docs/:

  docs/docs.go       — full swag.Spec with paths and definitions
  docs/swagger.json  — OpenAPI 2.0 JSON spec
  docs/swagger.yaml  — OpenAPI 2.0 YAML spec

If the swag CLI is not installed it will be fetched automatically.

Run this after adding or modifying any // @Summary, @Router, @Param,
@Success, @Failure annotations on your handlers.`,
	RunE: func(c *cobra.Command, _ []string) error {
		// Check if swag is available; install if missing.
		if _, err := exec.LookPath("swag"); err != nil {
			output.Info("swag CLI not found — installing...")
			install := exec.Command("go", "install", "github.com/swaggo/swag/cmd/swag@latest")
			install.Stdout = os.Stdout
			install.Stderr = os.Stderr
			install.Env = os.Environ()
			if err := install.Run(); err != nil {
				output.Fail("failed to install swag CLI")
				return err
			}
			output.Success("swag CLI installed")
		}

		cmd := exec.Command("swag", "init", "-g", "cmd/api/main.go")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()

		output.Info("generating swagger documentation...")
		if err := cmd.Run(); err != nil {
			output.Fail("swag init failed")
			return err
		}
		output.Success("swagger documentation regenerated")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(docsGenerateCmd)
}
