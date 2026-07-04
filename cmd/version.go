package cmd

import (
	"fmt"
	"strings"

	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags "-X github.com/DarpanAdhikari/drp-go-cli/cmd.Version=..."
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the drp CLI version",
	RunE: func(c *cobra.Command, args []string) error {
		fmt.Println("drp version", Version)
		return nil
	},
}

var versionCheckCmd = &cobra.Command{
	Use:   "version:check",
	Short: "Check if a newer drp release is available",
	RunE: func(c *cobra.Command, _ []string) error {
		current := Version
		output.Info("Current version: %s", current)

		latest, err := latestReleaseTag(defaultReleaseRepo)
		if err != nil {
			output.Fail("%v", err)
			return err
		}
		output.Info("Latest version:  %s", latest)

		if current == "dev" {
			output.Info("→ Not a release build. Run `drp upgrade` to install the latest stable version")
			return nil
		}

		cmp := compareVersions(current, latest)
		switch {
		case cmp < 0:
			output.Info("→ Run `drp upgrade` to update, or `drp upgrade --version %s` to stay on current", current)
		case cmp > 0:
			output.Info("→ You are ahead of the latest release (local build or pre-release)")
		default:
			output.Success("You're on the latest release")
		}
		return nil
	},
}

// compareVersions compares two semver strings (with optional "v" prefix).
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func compareVersions(a, b string) int {
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")

	pa := strings.Split(a, ".")
	pb := strings.Split(b, ".")

	for i := 0; i < 3; i++ {
		var va, vb int
		if i < len(pa) {
			fmt.Sscanf(pa[i], "%d", &va)
		}
		if i < len(pb) {
			fmt.Sscanf(pb[i], "%d", &vb)
		}
		if va < vb {
			return -1
		}
		if va > vb {
			return 1
		}
	}
	return 0
}

func init() {
	rootCmd.AddCommand(versionCmd, versionCheckCmd)
}
