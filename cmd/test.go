package cmd

import (
	"os"
	"os/exec"

	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run all project tests (go test ./...)",
	Long: `Runs go mod tidy then go test ./... in the current project directory.

Supports common Go test flags like -v and -cover.

Examples:
  drp test          # run all tests, compact output
  drp test -v       # verbose output with each test name
  drp test -cover   # include coverage report`,
	RunE: func(c *cobra.Command, args []string) error {
		verbose, _ := c.Flags().GetBool("verbose")
		cover, _ := c.Flags().GetBool("cover")

		// Ensure dependencies are resolved before testing.
		tidy := exec.Command("go", "mod", "tidy")
		tidy.Stdout = os.Stdout
		tidy.Stderr = os.Stderr
		tidy.Env = os.Environ()
		if err := tidy.Run(); err != nil {
			output.Fail("go mod tidy failed")
			return err
		}

		goArgs := []string{"test"}
		if verbose {
			goArgs = append(goArgs, "-v")
		}
		if cover {
			goArgs = append(goArgs, "-cover")
		}
		goArgs = append(goArgs, "./...")

		cmd := exec.Command("go", goArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = os.Environ()

		output.Info("go %v", goArgs)
		if err := cmd.Run(); err != nil {
			output.Fail("tests failed")
			return err
		}
		output.Success("all tests passed")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().BoolP("verbose", "v", false, "Verbose test output")
	testCmd.Flags().Bool("cover", false, "Include coverage report")
}
