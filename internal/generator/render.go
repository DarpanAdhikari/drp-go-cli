package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/DarpanAdhikari/drp-go-cli/internal/generator/embedded"
)

// Renderer resolves and executes templates with a single, consistent
// precedence rule: user-overridable files in UserTemplatesDir first,
// embedded fallback second. This is the only place that rule lives.
type Renderer struct {
	UserTemplatesDir string // e.g. "./templates" — checked before embedded
}

// NewRenderer constructs a Renderer with the conventional templates dir.
func NewRenderer() *Renderer {
	return &Renderer{UserTemplatesDir: "templates"}
}

// Render executes the named template (e.g. "model.tpl") with data and
// returns the rendered bytes.
func (r *Renderer) Render(name string, data any) ([]byte, error) {
	src, err := r.load(name)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(name).Parse(src)
	if err != nil {
		return nil, fmt.Errorf("renderer: parsing template %q: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("renderer: executing template %q: %w", name, err)
	}
	return buf.Bytes(), nil
}

// load returns the raw template source for name, preferring a file on disk
// in UserTemplatesDir over the embedded fallback.
func (r *Renderer) load(name string) (string, error) {
	// 1. Check for a user-provided override on disk.
	diskPath := filepath.Join(r.UserTemplatesDir, name)
	if data, err := os.ReadFile(diskPath); err == nil {
		return string(data), nil
	}

	// 2. Fall back to the embedded copy.
	embeddedPath := filepath.Join("files", name)
	data, err := embedded.FS.ReadFile(embeddedPath)
	if err != nil {
		return "", fmt.Errorf("renderer: template %q not found on disk (%s) or embedded", name, diskPath)
	}
	return string(data), nil
}
