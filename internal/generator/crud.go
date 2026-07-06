package generator

import (
	"fmt"
	"os"
	"path/filepath"
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

	// DryRun returns the file list without writing anything.
	DryRun bool
}

// allLayers reports whether the options mean "generate everything".
func (o CRUDOptions) allLayers() bool {
	return !o.Model && !o.Handler && !o.Repository && !o.Service && !o.Routes
}

// layer groups a template name with dynamic output directory and filename.
type layer struct {
	templateName string
	dirFunc      func(Names) string
	fileFunc     func(Names) string
	enabled      func(CRUDOptions) bool
}

// blueprintLayers defines the generated project's internal directory structure.
// Model, repository, service, and handler are placed under internal/<domain>/
// to match the domain-based layout used by drp init --auth.
// Routes remain in internal/routes/ with a domain prefix to avoid collisions.
var blueprintLayers = []layer{
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
}

// CRUD generates the requested layer files for a resource under the current
// working directory (assumed to be the root of a drp-managed project).
// Returns a list of file paths that were written.
func CRUD(opts CRUDOptions) ([]string, error) {
	if opts.RawName == "" {
		return nil, fmt.Errorf("crud: resource name cannot be empty")
	}

	names := NewNames(opts.RawName, opts.ModuleName)
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
