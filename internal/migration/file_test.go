package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFile_CreatesUpAndDownFiles(t *testing.T) {
	dir := t.TempDir()
	f, err := NewFile(dir, "create_users_table")
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}

	if _, err := os.Stat(f.UpPath); err != nil {
		t.Errorf("up file not created: %v", err)
	}
	if _, err := os.Stat(f.DownPath); err != nil {
		t.Errorf("down file not created: %v", err)
	}
	if f.Name != "create_users_table" {
		t.Errorf("Name = %q, want %q", f.Name, "create_users_table")
	}

	up, err := os.ReadFile(f.UpPath)
	if err != nil {
		t.Fatalf("read up file: %v", err)
	}
	if !strings.Contains(string(up), "CREATE TABLE IF NOT EXISTS users") {
		t.Errorf("up file did not infer users table:\n%s", up)
	}
	if !strings.Contains(string(up), "id BIGSERIAL PRIMARY KEY") {
		t.Errorf("up file missing id column:\n%s", up)
	}
	if !strings.Contains(string(up), "created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()") {
		t.Errorf("up file missing created_at column:\n%s", up)
	}
	if !strings.Contains(string(up), "updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()") {
		t.Errorf("up file missing updated_at column:\n%s", up)
	}

	down, err := os.ReadFile(f.DownPath)
	if err != nil {
		t.Fatalf("read down file: %v", err)
	}
	if !strings.Contains(string(down), "DROP TABLE IF EXISTS users;") {
		t.Errorf("down file did not infer users table:\n%s", down)
	}
}

func TestNewFile_SanitisesName(t *testing.T) {
	dir := t.TempDir()
	f, err := NewFile(dir, "Create Users Table!!")
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}
	if f.Name != "create_users_table" {
		t.Errorf("sanitised Name = %q, want %q", f.Name, "create_users_table")
	}
}

func TestNewFileForTable_UsesExplicitTable(t *testing.T) {
	dir := t.TempDir()
	f, err := NewFileForTable(dir, "add_profile_fields", "user_profiles")
	if err != nil {
		t.Fatalf("NewFileForTable: %v", err)
	}

	up, err := os.ReadFile(f.UpPath)
	if err != nil {
		t.Fatalf("read up file: %v", err)
	}
	if !strings.Contains(string(up), "CREATE TABLE IF NOT EXISTS user_profiles") {
		t.Errorf("up file did not use explicit table:\n%s", up)
	}

	down, err := os.ReadFile(f.DownPath)
	if err != nil {
		t.Fatalf("read down file: %v", err)
	}
	if !strings.Contains(string(down), "DROP TABLE IF EXISTS user_profiles;") {
		t.Errorf("down file did not use explicit table:\n%s", down)
	}
}

func TestDiscoverFiles_SortsAscending(t *testing.T) {
	dir := t.TempDir()

	// Write three migration pairs out of order.
	pairs := []string{
		"20240103000000_third",
		"20240101000000_first",
		"20240102000000_second",
	}
	for _, name := range pairs {
		os.WriteFile(filepath.Join(dir, name+".up.sql"), []byte("SELECT 1;"), 0o644)
		os.WriteFile(filepath.Join(dir, name+".down.sql"), []byte("SELECT 1;"), 0o644)
	}

	files, err := DiscoverFiles(dir)
	if err != nil {
		t.Fatalf("DiscoverFiles: %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("len(files) = %d, want 3", len(files))
	}
	if files[0].Name != "first" {
		t.Errorf("files[0].Name = %q, want %q", files[0].Name, "first")
	}
	if files[1].Name != "second" {
		t.Errorf("files[1].Name = %q, want %q", files[1].Name, "second")
	}
	if files[2].Name != "third" {
		t.Errorf("files[2].Name = %q, want %q", files[2].Name, "third")
	}
}

func TestDiscoverFiles_ErrorOnOrphanUp(t *testing.T) {
	dir := t.TempDir()
	// Write an up file with no matching down.
	os.WriteFile(filepath.Join(dir, "20240101000000_orphan.up.sql"), []byte("SELECT 1;"), 0o644)

	_, err := DiscoverFiles(dir)
	if err == nil {
		t.Error("expected error for orphan up file, got nil")
	}
}

func TestDiscoverFiles_ErrorOnOrphanDown(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "20240101000000_orphan.down.sql"), []byte("SELECT 1;"), 0o644)

	_, err := DiscoverFiles(dir)
	if err == nil {
		t.Error("expected error for orphan down file, got nil")
	}
}

func TestDiscoverFiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	files, err := DiscoverFiles(dir)
	if err != nil {
		t.Fatalf("expected no error for empty dir, got: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestDiscoverFiles_IgnoresNonMigrationFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# hi"), 0o644)
	os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(dir, "20240101000000_valid.up.sql"), []byte("SELECT 1;"), 0o644)
	os.WriteFile(filepath.Join(dir, "20240101000000_valid.down.sql"), []byte("SELECT 1;"), 0o644)

	files, err := DiscoverFiles(dir)
	if err != nil {
		t.Fatalf("DiscoverFiles: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
}

func TestFile_Identifier(t *testing.T) {
	dir := t.TempDir()
	f, _ := NewFile(dir, "create_products_table")
	id := f.Identifier()
	if id == "" {
		t.Error("Identifier returned empty string")
	}
	if id != fmt.Sprintf("%d_create_products_table", f.Timestamp) {
		t.Errorf("Identifier = %q, unexpected format", id)
	}
}
