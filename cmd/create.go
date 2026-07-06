package cmd

import (
	"github.com/DarpanAdhikari/drp-go-cli/internal/generator"
	"github.com/DarpanAdhikari/drp-go-cli/internal/interactive"
	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
	"github.com/spf13/cobra"
)

var createCrudCmd = &cobra.Command{
	Use:   "create:crud [name]",
	Short: "Generate model/handler/repository/service/routes files for a resource",
	Long: `Generates all CRUD layers for a resource under the current project directory.

Run without any layer flags to enter interactive mode (recommended):
  drp create:crud products

With specific flags (non-interactive):
  -m   model only
  -r   repository only
  -s   service only
  --handler   handler only
  --routes    routes only
  --no-interaction   skip interactive prompts (implies all layers)

Generated files (domain-based layout):
  internal/<domain>/model.go
  internal/<domain>/repository.go
  internal/<domain>/service.go
  internal/<domain>/handler.go
  internal/routes/<domain>_routes.go`,
	Args: cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		module, _ := c.Flags().GetString("module")
		noInt, _ := c.Flags().GetBool("no-interaction")

		// Auto-detect module name from go.mod when not provided.
		if module == "" {
			detected, err := generator.DetectModuleName("")
			if err != nil {
				if noInt || !output.IsTTY() {
					output.Fail("%v", err)
					return err
				}
				output.Warn("%v", err)
				detected, err = interactive.PromptModule()
				if err != nil {
					output.Fail("Module name is required.")
					output.Info("Run with --no-interaction and --module to skip interactive mode.")
					return err
				}
				module = detected
			} else {
				module = detected
			}
		}

		model, _ := c.Flags().GetBool("model")
		repo, _ := c.Flags().GetBool("repository")
		svc, _ := c.Flags().GetBool("service")
		handler, _ := c.Flags().GetBool("handler")
		routes, _ := c.Flags().GetBool("routes")
		anyFlag := model || repo || svc || handler || routes

		if !anyFlag && !noInt {
			if !output.IsTTY() {
				output.Warn("Not a terminal — generating all layers. Use --no-interaction to silence this warning.")
				model = true
				repo = true
				svc = true
				handler = true
				routes = true
			} else {
				selections, err := interactive.CRUDLayerSelection(args[0])
				if err != nil {
					output.Fail("%v", err)
					output.Info("Run with --no-interaction to skip interactive mode.")
					return err
				}
				model = selections.Model
				repo = selections.Repository
				svc = selections.Service
				handler = selections.Handler
				routes = selections.Routes
			}
		} else if !anyFlag {
			model = true
			repo = true
			svc = true
			handler = true
			routes = true
		}

		opts := generator.CRUDOptions{
			RawName:    args[0],
			ModuleName: module,
			Model:      model,
			Repository: repo,
			Service:    svc,
			Handler:    handler,
			Routes:     routes,
		}

		if !noInt && !anyFlag && output.IsTTY() {
			previewOpts := opts
			previewOpts.DryRun = true
			previewFiles, err := generator.CRUD(previewOpts)
			if err != nil {
				output.Fail("%v", err)
				return err
			}

			if len(previewFiles) == 0 {
				output.Info("No files to generate for the selected layers.")
				return nil
			}

			confirmed, err := interactive.ConfirmGeneration(previewFiles)
			if err != nil {
				output.Fail("%v", err)
				return err
			}
			if !confirmed {
				output.Info("Cancelled.")
				return nil
			}
		}

		written, err := generator.CRUD(opts)
		if err != nil {
			output.Fail("%v", err)
			return err
		}

		for _, path := range written {
			output.Success("Created %s", path)
		}

		if len(written) > 0 {
			names := generator.NewNames(args[0], module)
			output.Info("Register routes in cmd/api/main.go:")
			output.Info("  routes.Register%sRoutes(mux, db)", names.Name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCrudCmd)
	createCrudCmd.Flags().BoolP("model", "m", false, "Generate model only")
	createCrudCmd.Flags().BoolP("repository", "r", false, "Generate repository only")
	createCrudCmd.Flags().BoolP("service", "s", false, "Generate service only")
	// NOTE: -h conflicts with Cobra's built-in --help shorthand; long flag only.
	createCrudCmd.Flags().Bool("handler", false, "Generate handler only")
	createCrudCmd.Flags().Bool("routes", false, "Generate routes only")
	createCrudCmd.Flags().String("module", "", "Go module name (auto-detected from go.mod if not set)")
	createCrudCmd.Flags().Bool("no-interaction", false, "Skip interactive prompts (implies all layers)")
}
