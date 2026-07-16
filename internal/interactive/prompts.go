package interactive

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
)

type CRUDSelections struct {
	Model      bool
	Repository bool
	Service    bool
	Handler    bool
	Routes     bool
	Migration  bool
	Seeder     bool
	Driver     string
}

type InitOptions struct {
	Name     string
	Module   string
	Auth     bool
	Driver   string
	Infra    []string
}

func PromptInit(name string) (InitOptions, error) {
	var opts InitOptions

	if name != "" {
		opts.Name = name
	}

	authChoice := "auth"

	groups := []*huh.Group{}

	if opts.Name == "" {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Description("Name of the Go project directory and binary.").
				Value(&opts.Name).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("project name cannot be empty")
					}
					return nil
				}),
		))
	}

	groups = append(groups, huh.NewGroup(
		huh.NewInput().
			Title("Module path").
			Description("Go module path (e.g. github.com/yourorg/myapp). Leave empty to use project name.").
			Value(&opts.Module).
			Placeholder("(default: project name)"),

		huh.NewSelect[string]().
			Title("Authentication").
			Options(
				huh.NewOption("Full JWT auth (register, login, refresh, device sessions)", "auth"),
				huh.NewOption("None (minimal project skeleton)", "none"),
			).
			Value(&authChoice),
	))

	groups = append(groups, huh.NewGroup(
		huh.NewSelect[string]().
			Title("Database driver").
			Options(
				huh.NewOption("PostgreSQL", "postgres"),
				huh.NewOption("MySQL", "mysql"),
			).
			Value(&opts.Driver),

		huh.NewMultiSelect[string]().
			Title("Infrastructure files").
			Description("Select the infrastructure files to generate.").
			Options(
				huh.NewOption("Docker (Dockerfile, compose, .dockerignore)", "docker"),
				huh.NewOption("CI (GitHub Actions)", "ci"),
				huh.NewOption("Makefile", "make"),
				huh.NewOption("Lint (.editorconfig, .golangci.yml)", "lint"),
			).
			Value(&opts.Infra),
	))

	form := huh.NewForm(groups...).
		WithTheme(huh.ThemeCatppuccin()).
		WithShowHelp(true).
		WithShowErrors(true)

	if err := form.Run(); err != nil {
		return opts, err
	}

	opts.Auth = authChoice == "auth"

	if opts.Module == "" {
		opts.Module = opts.Name
	}

	return opts, nil
}

func CRUDLayerSelection(resourceName string) (CRUDSelections, error) {
	var selections CRUDSelections
	var selected []string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(fmt.Sprintf("Generate CRUD for %q — select layers", resourceName)).
				Description("Choose which layers to generate.").
				Options(
					huh.NewOption("Model", "model").Selected(true),
					huh.NewOption("Repository", "repository").Selected(true),
					huh.NewOption("Service", "service").Selected(true),
					huh.NewOption("Handler", "handler").Selected(true),
					huh.NewOption("Routes", "routes").Selected(true),
					huh.NewOption("Migration (up/down SQL)", "migration").Selected(true),
					huh.NewOption("Seeder (SQL)", "seeder").Selected(true),
				).
				Value(&selected),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Database driver").
				Description("SQL dialect for generated migrations.").
				Options(
					huh.NewOption("PostgreSQL", "postgres").Selected(true),
					huh.NewOption("MySQL", "mysql"),
				).
				Value(&selections.Driver),
		),
	).
		WithTheme(huh.ThemeCatppuccin()).
		WithShowHelp(true).
		WithShowErrors(true)

	if err := form.Run(); err != nil {
		return selections, err
	}

	sel := make(map[string]bool)
	for _, s := range selected {
		sel[s] = true
	}
	selections.Model = sel["model"]
	selections.Repository = sel["repository"]
	selections.Service = sel["service"]
	selections.Handler = sel["handler"]
	selections.Routes = sel["routes"]
	selections.Migration = sel["migration"]
	selections.Seeder = sel["seeder"]

	if !selections.Model && !selections.Repository && !selections.Service && !selections.Handler && !selections.Routes && !selections.Migration && !selections.Seeder {
		fmt.Fprintln(os.Stderr, "No layers selected. Aborting.")
		os.Exit(2)
	}

	return selections, nil
}

func PromptModule() (string, error) {
	var module string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Go module name").
				Description("e.g. github.com/yourorg/myapp").
				Value(&module).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("module name cannot be empty")
					}
					return nil
				}),
		),
	).
		WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(module), nil
}

func ConfirmGeneration(files []string) (bool, error) {
	var confirmed bool

	summary := strings.Builder{}
	for _, f := range files {
		summary.WriteString(fmt.Sprintf("  • %s\n", f))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Files to be created").
				Description(summary.String()),

			huh.NewConfirm().
				Title("Proceed with generation?").
				Affirmative("Yes, generate").
				Negative("Cancel").
				Value(&confirmed),
		),
	).
		WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		return false, err
	}

	return confirmed, nil
}
