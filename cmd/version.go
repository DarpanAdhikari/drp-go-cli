package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags "-X github.com/yourorg/drp/cmd.Version=..."
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the drp CLI version",
	RunE: func(c *cobra.Command, args []string) error {
		fmt.Println("drp version", Version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
