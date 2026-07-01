package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectModuleName(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "go.mod")

	os.WriteFile(path, []byte("module github.com/acme/myapp\n\ngo 1.22\n"), 0o644)
	mod, err := DetectModuleName(path)
	if err != nil {
		t.Fatalf("DetectModuleName: %v", err)
	}
	if mod != "github.com/acme/myapp" {
		t.Errorf("got %q, want %q", mod, "github.com/acme/myapp")
	}
}

func TestDetectModuleName_Missing(t *testing.T) {
	_, err := DetectModuleName("/nonexistent/go.mod")
	if err == nil {
		t.Error("expected error for missing go.mod")
	}
}

func TestDetectModuleName_NoModuleDeclaration(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "go.mod")
	os.WriteFile(path, []byte("go 1.22\n"), 0o644)
	_, err := DetectModuleName(path)
	if err == nil {
		t.Error("expected error for go.mod with no module declaration")
	}
}
