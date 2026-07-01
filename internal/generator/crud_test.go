package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCRUD_AllLayers(t *testing.T) {
	// Run CRUD in a temp dir so it writes real files.
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	written, err := CRUD(CRUDOptions{
		RawName:    "product",
		ModuleName: "github.com/test/myapp",
	})
	if err != nil {
		t.Fatalf("CRUD: %v", err)
	}
	if len(written) != 5 {
		t.Errorf("expected 5 files written, got %d: %v", len(written), written)
	}

	expected := []string{
		filepath.Join("internal", "models", "products.go"),
		filepath.Join("internal", "repositories", "products_repository.go"),
		filepath.Join("internal", "services", "products_service.go"),
		filepath.Join("internal", "handlers", "products_handler.go"),
		filepath.Join("internal", "routes", "products_routes.go"),
	}
	for _, path := range expected {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %q not found", path)
		}
	}
}

func TestCRUD_ModelOnly(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	written, err := CRUD(CRUDOptions{
		RawName:    "product",
		ModuleName: "github.com/test/myapp",
		Model:      true,
	})
	if err != nil {
		t.Fatalf("CRUD: %v", err)
	}
	if len(written) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(written), written)
	}
	if written[0] != filepath.Join("internal", "models", "products.go") {
		t.Errorf("unexpected file: %q", written[0])
	}
}

func TestCRUD_RefusesOverwrite(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	// Generate once.
	if _, err := CRUD(CRUDOptions{RawName: "product", ModuleName: "m"}); err != nil {
		t.Fatalf("first CRUD: %v", err)
	}
	// Generate again — must refuse.
	_, err := CRUD(CRUDOptions{RawName: "product", ModuleName: "m"})
	if err == nil {
		t.Error("expected error on duplicate CRUD, got nil")
	}
}

func TestCRUD_EmptyName(t *testing.T) {
	_, err := CRUD(CRUDOptions{RawName: ""})
	if err == nil {
		t.Error("expected error for empty name, got nil")
	}
}

func TestCRUD_RenderedContentContainsResourceName(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	written, err := CRUD(CRUDOptions{
		RawName:    "invoice",
		ModuleName: "github.com/test/myapp",
		Model:      true,
	})
	if err != nil {
		t.Fatalf("CRUD: %v", err)
	}

	data, _ := os.ReadFile(written[0])
	content := string(data)
	if !contains(content, "Invoice") {
		t.Errorf("model file doesn't contain 'Invoice':\n%s", content)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
