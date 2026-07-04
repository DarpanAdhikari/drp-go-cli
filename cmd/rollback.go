package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Restore the previous drp binary from backup (.old)",
	RunE: func(c *cobra.Command, _ []string) error {
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("rollback: cannot find current executable: %w", err)
		}

		backup := exe + ".old"
		if _, err := os.Stat(backup); err != nil {
			if os.IsNotExist(err) {
				output.Fail("No backup found at %s — nothing to roll back", backup)
				return fmt.Errorf("rollback: no backup at %s", backup)
			}
			return fmt.Errorf("rollback: stat %q: %w", backup, err)
		}

		tmp := exe + ".tmp"

		// mv current → .tmp
		if err := os.Rename(exe, tmp); err != nil {
			return fmt.Errorf("rollback: rename %q → .tmp: %w", exe, err)
		}

		// mv .old → current
		if err := os.Rename(backup, exe); err != nil {
			// Try to restore: mv .tmp back
			os.Rename(tmp, exe)
			return fmt.Errorf("rollback: rename .old → %q: %w", filepath.Base(exe), err)
		}

		// mv .tmp → .old
		if err := os.Rename(tmp, backup); err != nil {
			// Non-critical: .tmp is orphaned; warn but don't fail
			output.Warn("Could not move .tmp to .old: %v", err)
		}

		output.Success("Rolled back to previous version at %s", exe)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}
