package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/DarpanAdhikari/drp-go-cli/internal/doctor"
	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check your environment for common DRP setup issues",
	Long:  `Checks Go installation, .env presence and validity, database reachability, and project directory structure. All checks run even if an earlier one fails.`,
	RunE: func(c *cobra.Command, _ []string) error {
		results := doctor.Run(globalEnvFile)

		allOK := true
		for _, r := range results {
			if r.OK {
				output.Success("%-28s %s", r.Label, r.Detail)
			} else {
				output.Fail("%-28s %s", r.Label, r.Detail)
				allOK = false
			}
		}

		fmt.Println()
		if allOK {
			output.Success("All checks passed — your environment looks good")
			return nil
		}
		output.Fail("Some checks failed — fix the issues above and re-run `drp doctor`")
		// Return nil so Cobra doesn't double-print the error; we already printed it.
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
