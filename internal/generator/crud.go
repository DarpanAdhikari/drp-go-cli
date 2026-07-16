package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CRUDOptions configures what `drp create:crud` generates.
type CRUDOptions struct {
	// Raw resource name as typed by the user, e.g. "product" or "product_category".
	RawName string
	// Go module name from the project's go.mod, e.g. "github.com/yourorg/myapp".
	ModuleName string

	// Layer flags — if all false, all layers are generated.
	Model      bool
	Handler    bool
	Repository bool
	Service    bool
	Routes     bool
	Migration  bool
	Seeder     bool

	// DBDriver selects the SQL dialect for migrations ("postgres" or "mysql").
	DBDriver string

	// DryRun returns the file list without writing anything.
	DryRun bool
}

// allLayers reports whether the options mean "generate everything".
func (o CRUDOptions) allLayers() bool {
	return !o.Model && !o.Handler && !o.Repository && !o.Service && !o.Routes &&
		!o.Migration && !o.Seeder
}

// layer groups a template name with dynamic output directory and filename.
type layer struct {
	templateName string
	dirFunc      func(Names) string
	fileFunc     func(Names) string
	enabled      func(CRUDOptions) bool
}

// blueprintLayers defines all files the CRUD generator can produce.
// Source layers go under internal/<domain>/, tests under tests/<domain>/,
// migrations under database/migrations/, seeders under database/seeders/,
// and routes under internal/routes/.
var blueprintLayers = []layer{
	// ── Source layers ──────────────────────────────────────────────
	{
		templateName: "model.tpl",
		dirFunc:      func(n Names) string { return filepath.Join("internal", n.DomainName) },
		fileFunc:     func(Names) string { return "model.go" },
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Model },
	},
	{
		templateName: "repository.tpl",
		dirFunc:      func(n Names) string { return filepath.Join("internal", n.DomainName) },
		fileFunc:     func(Names) string { return "repository.go" },
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Repository },
	},
	{
		templateName: "service.tpl",
		dirFunc:      func(n Names) string { return filepath.Join("internal", n.DomainName) },
		fileFunc:     func(Names) string { return "service.go" },
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Service },
	},
	{
		templateName: "handler.tpl",
		dirFunc:      func(n Names) string { return filepath.Join("internal", n.DomainName) },
		fileFunc:     func(Names) string { return "handler.go" },
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Handler },
	},
	{
		templateName: "routes.tpl",
		dirFunc:      func(Names) string { return filepath.Join("internal", "routes") },
		fileFunc:     func(n Names) string { return n.DomainName + "_routes.go" },
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Routes },
	},

	// ── Database migration & seeder layers ─────────────────────────
	{
		templateName: "migration_up.tpl",
		dirFunc:      func(Names) string { return filepath.Join("database", "migrations") },
		fileFunc:     func(n Names) string { return fmt.Sprintf("%s_create_%s_table.up.sql", timestamp(), n.TableName) },
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Migration },
	},
	{
		templateName: "migration_down.tpl",
		dirFunc:      func(Names) string { return filepath.Join("database", "migrations") },
		fileFunc:     func(n Names) string { return fmt.Sprintf("%s_create_%s_table.down.sql", timestamp(), n.TableName) },
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Migration },
	},
	{
		templateName: "seeder.tpl",
		dirFunc:      func(Names) string { return filepath.Join("database", "seeders") },
		fileFunc:     func(n Names) string { return fmt.Sprintf("%s_seed_%s.sql", timestamp(), n.TableName) },
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Seeder },
	},

	// ── Test layers (root tests/<domain>/) ─────────────────────────
	{
		templateName: "model_test.tpl",
		dirFunc:      func(n Names) string { return filepath.Join("tests", n.DomainName) },
		fileFunc:     func(Names) string { return "model_test.go" },
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Model },
	},
	{
		templateName: "repository_test.tpl",
		dirFunc:      func(n Names) string { return filepath.Join("tests", n.DomainName) },
		fileFunc:     func(Names) string { return "repository_test.go" },
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Repository },
	},
	{
		templateName: "service_test.tpl",
		dirFunc:      func(n Names) string { return filepath.Join("tests", n.DomainName) },
		fileFunc:     func(Names) string { return "service_test.go" },
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Service },
	},
	{
		templateName: "handler_test.tpl",
		dirFunc:      func(n Names) string { return filepath.Join("tests", n.DomainName) },
		fileFunc:     func(Names) string { return "handler_test.go" },
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Handler },
	},
}

// timestamp returns the current UTC time as YYYYMMDDHHmmss for use in
// migration and seeder filenames.
func timestamp() string {
	return time.Now().UTC().Format("20060102150405")
}

// CRUD generates the requested layer files for a resource under the current
// working directory (assumed to be the root of a drp-managed project).
// Returns a list of file paths that were written.
func CRUD(opts CRUDOptions) ([]string, error) {
	if opts.RawName == "" {
		return nil, fmt.Errorf("crud: resource name cannot be empty")
	}

	names := NewNamesWithDriver(opts.RawName, opts.ModuleName, opts.DBDriver)
	renderer := NewRenderer()
	var files []string

	for _, l := range blueprintLayers {
		if !l.enabled(opts) {
			continue
		}

		outDir := l.dirFunc(names)
		outFile := l.fileFunc(names)
		outPath := filepath.Join(outDir, outFile)

		// Refuse to overwrite existing files — idempotency rule.
		if _, err := os.Stat(outPath); err == nil {
			return files, fmt.Errorf("crud: %q already exists — remove it first or check for name collision", outPath)
		}

		if opts.DryRun {
			files = append(files, outPath)
			continue
		}

		rendered, err := renderer.Render(l.templateName, names)
		if err != nil {
			return files, fmt.Errorf("crud: rendering %s: %w", l.templateName, err)
		}

		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return files, fmt.Errorf("crud: creating directory %q: %w", outDir, err)
		}

		if err := os.WriteFile(outPath, rendered, 0o644); err != nil {
			return files, fmt.Errorf("crud: writing %q: %w", outPath, err)
		}

		files = append(files, outPath)
	}

	return files, nil
}
