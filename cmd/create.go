package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yourorg/drp/internal/generator"
	"github.com/yourorg/drp/internal/output"
)

var createCrudCmd = &cobra.Command{
	Use:   "create:crud [name]",
	Short: "Generate model/handler/repository/service/routes files for a resource",
	Long: `Generates all CRUD layers for a resource under the current project directory.

With no flags, all layers are generated:
  internal/models/<name>.go
  internal/repositories/<name>_repository.go
  internal/services/<name>_service.go
  internal/handlers/<name>_handler.go
  internal/routes/<name>_routes.go

Use flags to generate specific layers only:
  -m   model only
  -r   repository only
  -s   service only
  --handler   handler only
  --routes    routes only`,
	Args: cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		module, _ := c.Flags().GetString("module")

		// Auto-detect module name from go.mod when not provided.
		if module == "" {
			detected, err := generator.DetectModuleName("")
			if err != nil {
				output.Fail("%v", err)
				return err
			}
			module = detected
		}

		model, _ := c.Flags().GetBool("model")
		repo, _ := c.Flags().GetBool("repository")
		svc, _ := c.Flags().GetBool("service")
		handler, _ := c.Flags().GetBool("handler")
		routes, _ := c.Flags().GetBool("routes")

		opts := generator.CRUDOptions{
			RawName:    args[0],
			ModuleName: module,
			Model:      model,
			Repository: repo,
			Service:    svc,
			Handler:    handler,
			Routes:     routes,
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
}
