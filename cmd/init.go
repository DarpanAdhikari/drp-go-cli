package cmd

import (
	"github.com/DarpanAdhikari/drp-go-cli/internal/generator"
	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Scaffold a new DRP-backed Go project",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		module, _ := c.Flags().GetString("module")
		force, _ := c.Flags().GetBool("force")
		auth, _ := c.Flags().GetBool("auth")

		opts := generator.ProjectOptions{
			Name:       args[0],
			ModuleName: module,
			Force:      force,
			Auth:       auth,
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
		output.Info("  drp run api")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String("module", "", "Go module path (default: project name)")
	initCmd.Flags().Bool("force", false, "Overwrite non-empty directory")
	initCmd.Flags().Bool("auth", false, "Scaffold JWT auth, users, token handling, and auth routes")
}
