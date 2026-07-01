package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yourorg/drp/internal/generator"
	"github.com/yourorg/drp/internal/output"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Scaffold a new DRP-backed Go project",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		module, _ := c.Flags().GetString("module")
		force, _ := c.Flags().GetBool("force")

		opts := generator.ProjectOptions{
			Name:       args[0],
			ModuleName: module,
			Force:      force,
		}
		if opts.ModuleName == "" {
			opts.ModuleName = args[0]
		}

		if err := generator.Project(opts); err != nil {
			output.Fail("%v", err)
			return err
		}

		output.Success("Project %q created", args[0])
		output.Info("Next steps:")
		output.Info("  cd %s", args[0])
		output.Info("  edit .env")
		output.Info("  drp migrate:seed")
		output.Info("  go run ./cmd/api")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String("module", "", "Go module path (default: project name)")
	initCmd.Flags().Bool("force", false, "Overwrite non-empty directory")
}
