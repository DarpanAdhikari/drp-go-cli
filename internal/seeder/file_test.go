package seeder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFile_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	f, err := NewFile(dir, "seed_users")
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}
	if _, err := os.Stat(f.Path); err != nil {
		t.Errorf("file not created: %v", err)
	}
	if f.Name != "seed_users" {
		t.Errorf("Name = %q, want seed_users", f.Name)
	}
}

func TestNewFile_SanitisesName(t *testing.T) {
	dir := t.TempDir()
	f, err := NewFile(dir, "Seed Users!!")
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}
	if f.Name != "seed_users" {
		t.Errorf("Name = %q, want seed_users", f.Name)
	}
}

func TestDiscoverFiles_SortsAscending(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{
		"20240103000000_seed_c.sql",
		"20240101000000_seed_a.sql",
		"20240102000000_seed_b.sql",
	} {
		os.WriteFile(filepath.Join(dir, name), []byte("SELECT 1;"), 0o644)
	}

	files, err := DiscoverFiles(dir)
	if err != nil {
		t.Fatalf("DiscoverFiles: %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("len = %d, want 3", len(files))
	}
	if files[0].Name != "seed_a" || files[1].Name != "seed_b" || files[2].Name != "seed_c" {
		t.Errorf("wrong order: %v %v %v", files[0].Name, files[1].Name, files[2].Name)
	}
}

func TestDiscoverFiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	files, err := DiscoverFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestDiscoverFiles_IgnoresNonSeederFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "README.md"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(dir, "20240101000000_seed_users.sql"), []byte("SELECT 1;"), 0o644)

	files, err := DiscoverFiles(dir)
	if err != nil {
		t.Fatalf("DiscoverFiles: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1, got %d", len(files))
	}
}

func TestDiscoverFiles_MissingDir(t *testing.T) {
	_, err := DiscoverFiles("/nonexistent/path")
	if err == nil {
		t.Error("expected error for missing directory")
	}
}

func TestFile_Identifier(t *testing.T) {
	dir := t.TempDir()
	f, _ := NewFile(dir, "seed_products")
	id := f.Identifier()
	if id == "" {
		t.Error("Identifier returned empty string")
	}
	if id != f.Timestamp.Format(timestampLayout)+"_seed_products" {
		t.Errorf("Identifier = %q, unexpected format", id)
	}
}
