// Package cmd contains all Cobra CLI commands for drp. Commands stay thin:
// parse flags, call internal packages, format output via internal/output.
// Business logic belongs in internal/.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourorg/drp/internal/output"
)

// globalEnvFile is the path to the .env file, set via --env-file flag.
var globalEnvFile string

var rootCmd = &cobra.Command{
	Use:   "drp",
	Short: "DRP — Developer Rapid Productivity CLI",
	Long: `DRP is a lightweight, SQL-first CLI for Go backend development.
It generates plain, idiomatic Go code without locking your project into
a runtime dependency.

Run 'drp doctor' to check your environment.`,
	SilenceUsage:  true, // don't print usage on RunE errors
	SilenceErrors: true, // we print errors ourselves via output.Fail
}

// Execute is called by main.main().
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&globalEnvFile, "env-file", "", "Path to .env file (default: .env)")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable coloured output")

	// Apply --no-color before any command runs.
	cobra.OnInitialize(func() {
		noColor, _ := rootCmd.PersistentFlags().GetBool("no-color")
		output.SetNoColor(noColor)
	})
}
