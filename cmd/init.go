package cmd

import (
	"os/exec"
	"strings"

	"github.com/DarpanAdhikari/drp-go-cli/internal/generator"
	"github.com/DarpanAdhikari/drp-go-cli/internal/interactive"
	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Scaffold a new DRP-backed Go project",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		module, _ := c.Flags().GetString("module")
		force, _ := c.Flags().GetBool("force")
		authFlag, _ := c.Flags().GetBool("auth")
		auth := authFlag
		infra, _ := c.Flags().GetString("infra")
		dbDriver, _ := c.Flags().GetString("driver")
		noInt, _ := c.Flags().GetBool("no-interaction")
		hasFlags := c.Flags().Changed("auth") || c.Flags().Changed("infra") || c.Flags().Changed("driver") || module != ""

		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		if (!hasFlags || name == "") && output.IsTTY() && !noInt {
			iopts, err := interactive.PromptInit(name)
			if err != nil {
				output.Fail("%v", err)
				return err
			}
			name = iopts.Name
			if module == "" {
				module = iopts.Module
			}
			if !c.Flags().Changed("auth") {
				auth = iopts.Auth
			}
			if !c.Flags().Changed("driver") {
				dbDriver = iopts.Driver
			}
			if !c.Flags().Changed("infra") {
				infra = strings.Join(iopts.Infra, ",")
			}
		}

		if name == "" {
			output.Fail("Project name is required.")
			return nil
		}

		opts := generator.ProjectOptions{
			Name:       name,
			ModuleName: module,
			Force:      force,
			Auth:       auth,
			Infra:      infra,
			DBDriver:   dbDriver,
		}
		if opts.ModuleName == "" {
			opts.ModuleName = name
		}

		if err := generator.Project(opts); err != nil {
			output.Fail("%v", err)
			return err
		}

		// Initialize git repository for the new project.
		if err := exec.Command("git", "init", name).Run(); err != nil {
			output.Warn("could not initialize git repo: %v", err)
		}

		output.Success("Project %q created", name)
		output.Info("Next steps:")
		output.Info("  cd %s", name)
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
	initCmd.Flags().Bool("auth", true, "Scaffold JWT auth, users, token handling, and auth routes")
	initCmd.Flags().String("infra", "", "Generate infrastructure files: all, docker,ci,make,lint (comma-separated)")
	initCmd.Flags().String("driver", "postgres", "Database driver: postgres or mysql")
	initCmd.Flags().Bool("no-interaction", false, "Skip interactive prompts")
}
