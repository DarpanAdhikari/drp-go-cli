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
}

// allLayers reports whether the options mean "generate everything".
func (o CRUDOptions) allLayers() bool {
	return !o.Model && !o.Handler && !o.Repository && !o.Service && !o.Routes
}

// layer groups a template name, output directory, and file suffix.
type layer struct {
	templateName string
	dir          string
	fileSuffix   string // appended after the snake_case name, e.g. "_handler.go"
	enabled      func(CRUDOptions) bool
}

// blueprintLayers defines the generated project's internal directory structure.
// This is the canonical "what drp create:crud writes where" mapping.
var blueprintLayers = []layer{
	{
		templateName: "model.tpl",
		dir:          filepath.Join("internal", "models"),
		fileSuffix:   ".go",
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Model },
	},
	{
		templateName: "repository.tpl",
		dir:          filepath.Join("internal", "repositories"),
		fileSuffix:   "_repository.go",
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Repository },
	},
	{
		templateName: "service.tpl",
		dir:          filepath.Join("internal", "services"),
		fileSuffix:   "_service.go",
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Service },
	},
	{
		templateName: "handler.tpl",
		dir:          filepath.Join("internal", "handlers"),
		fileSuffix:   "_handler.go",
		enabled:      func(o CRUDOptions) bool { return o.allLayers() || o.Handler },
	},
	{
		templateName: "routes.tpl",
		dir:          filepath.Join("internal", "routes"),
		fileSuffix:   "_routes.go",
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
	var written []string

	for _, l := range blueprintLayers {
		if !l.enabled(opts) {
			continue
		}

		outPath := filepath.Join(l.dir, names.TableName+l.fileSuffix)

		// Refuse to overwrite existing files — idempotency rule.
		if _, err := os.Stat(outPath); err == nil {
			return written, fmt.Errorf("crud: %q already exists — remove it first or check for name collision", outPath)
		}

		rendered, err := renderer.Render(l.templateName, names)
		if err != nil {
			return written, fmt.Errorf("crud: rendering %s: %w", l.templateName, err)
		}

		if err := os.MkdirAll(l.dir, 0o755); err != nil {
			return written, fmt.Errorf("crud: creating directory %q: %w", l.dir, err)
		}

		if err := os.WriteFile(outPath, rendered, 0o644); err != nil {
			return written, fmt.Errorf("crud: writing %q: %w", outPath, err)
		}

		written = append(written, outPath)
	}

	return written, nil
}
