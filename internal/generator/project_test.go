package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProject_CreatesLayout(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	err := Project(ProjectOptions{Name: "myapp", ModuleName: "github.com/test/myapp"})
	if err != nil {
		t.Fatalf("Project: %v", err)
	}

	expectedDirs := []string{
		filepath.Join("myapp", "cmd", "api"),
		filepath.Join("myapp", "internal", "config"),
		filepath.Join("myapp", "internal", "handlers"),
		filepath.Join("myapp", "internal", "repositories"),
		filepath.Join("myapp", "internal", "services"),
		filepath.Join("myapp", "internal", "routes"),
		filepath.Join("myapp", "internal", "models"),
		filepath.Join("myapp", "database", "migrations"),
		filepath.Join("myapp", "database", "seeders"),
	}
	for _, d := range expectedDirs {
		if fi, err := os.Stat(d); err != nil || !fi.IsDir() {
			t.Errorf("expected directory %q not found", d)
		}
	}

	expectedFiles := []string{
		filepath.Join("myapp", "go.mod"),
		filepath.Join("myapp", ".env"),
		filepath.Join("myapp", ".env.example"),
		filepath.Join("myapp", ".gitignore"),
		filepath.Join("myapp", "cmd", "api", "main.go"),
		filepath.Join("myapp", "internal", "config", "config.go"),
		filepath.Join("myapp", "README.md"),
	}
	for _, f := range expectedFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("expected file %q not found", f)
		}
	}
}

func TestProject_RefusesNonEmptyDir(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	// Create directory with a file in it.
	os.MkdirAll("myapp", 0o755)
	os.WriteFile(filepath.Join("myapp", "existing.txt"), []byte("hi"), 0o644)

	err := Project(ProjectOptions{Name: "myapp"})
	if err == nil {
		t.Error("expected error for non-empty directory, got nil")
	}
}

func TestProject_ForceOverwritesExisting(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	os.MkdirAll("myapp", 0o755)
	os.WriteFile(filepath.Join("myapp", "existing.txt"), []byte("hi"), 0o644)

	err := Project(ProjectOptions{Name: "myapp", Force: true})
	if err != nil {
		t.Errorf("Project with --force should not fail: %v", err)
	}
}

func TestProject_EmptyNameFails(t *testing.T) {
	err := Project(ProjectOptions{Name: ""})
	if err == nil {
		t.Error("expected error for empty name")
	}
}
