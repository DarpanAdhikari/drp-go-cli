package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCRUD_AllLayers(t *testing.T) {
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
	// 5 source + 3 migration/seeder + 4 test = 12 files
	if len(written) != 12 {
		t.Errorf("expected 12 files written, got %d: %v", len(written), written)
	}

	// Check source files exist.
	expected := []string{
		filepath.Join("internal", "product", "model.go"),
		filepath.Join("internal", "product", "repository.go"),
		filepath.Join("internal", "product", "service.go"),
		filepath.Join("internal", "product", "handler.go"),
		filepath.Join("internal", "routes", "product_routes.go"),
	}
	for _, path := range expected {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %q not found", path)
		}
	}

	// Check test files exist.
	testExpected := []string{
		filepath.Join("tests", "product", "model_test.go"),
		filepath.Join("tests", "product", "repository_test.go"),
		filepath.Join("tests", "product", "service_test.go"),
		filepath.Join("tests", "product", "handler_test.go"),
	}
	for _, path := range testExpected {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected test file %q not found", path)
		}
	}

	// Check migration files exist (dynamic timestamp prefix).
	migrationDir := filepath.Join("database", "migrations")
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		t.Fatalf("reading migrations dir: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 migration files, got %d", len(entries))
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), "_create_products_table.up.sql") &&
			!strings.HasSuffix(e.Name(), "_create_products_table.down.sql") {
			t.Errorf("unexpected migration file: %q", e.Name())
		}
	}

	// Check seeder file exists.
	seederDir := filepath.Join("database", "seeders")
	seederEntries, err := os.ReadDir(seederDir)
	if err != nil {
		t.Fatalf("reading seeders dir: %v", err)
	}
	if len(seederEntries) != 1 {
		t.Errorf("expected 1 seeder file, got %d", len(seederEntries))
	}
	if !strings.HasSuffix(seederEntries[0].Name(), "_seed_products.sql") {
		t.Errorf("unexpected seeder file: %q", seederEntries[0].Name())
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
	// model.go + model_test.go = 2 files
	if len(written) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(written), written)
	}
	if written[0] != filepath.Join("internal", "product", "model.go") {
		t.Errorf("first file should be model.go, got %q", written[0])
	}
	if written[1] != filepath.Join("tests", "product", "model_test.go") {
		t.Errorf("second file should be model_test.go, got %q", written[1])
	}
}

func TestCRUD_MigrationOnly(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	written, err := CRUD(CRUDOptions{
		RawName:    "product",
		ModuleName: "github.com/test/myapp",
		Migration:  true,
	})
	if err != nil {
		t.Fatalf("CRUD: %v", err)
	}
	// Only migration up + down should be generated (2 files, no test files for migration)
	if len(written) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(written), written)
	}
	for _, path := range written {
		if !strings.Contains(path, "database/migrations") {
			t.Errorf("expected migration file, got %q", path)
		}
	}
}

func TestCRUD_SeederOnly(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	written, err := CRUD(CRUDOptions{
		RawName:    "product",
		ModuleName: "github.com/test/myapp",
		Seeder:     true,
	})
	if err != nil {
		t.Fatalf("CRUD: %v", err)
	}
	if len(written) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(written), written)
	}
	if !strings.Contains(written[0], "database/seeders") {
		t.Errorf("expected seeder file in database/seeders/, got %q", written[0])
	}
}

func TestCRUD_RepositoryOnly(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	written, err := CRUD(CRUDOptions{
		RawName:    "product",
		ModuleName: "github.com/test/myapp",
		Repository: true,
	})
	if err != nil {
		t.Fatalf("CRUD: %v", err)
	}
	// repository.go + repository_test.go = 2 files
	if len(written) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(written), written)
	}
}

func TestCRUD_HandlerOnly(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	written, err := CRUD(CRUDOptions{
		RawName:    "product",
		ModuleName: "github.com/test/myapp",
		Handler:    true,
	})
	if err != nil {
		t.Fatalf("CRUD: %v", err)
	}
	// handler.go + handler_test.go = 2 files
	if len(written) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(written), written)
	}
}

func TestCRUD_ServiceOnly(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	written, err := CRUD(CRUDOptions{
		RawName:    "product",
		ModuleName: "github.com/test/myapp",
		Service:    true,
	})
	if err != nil {
		t.Fatalf("CRUD: %v", err)
	}
	// service.go + service_test.go = 2 files
	if len(written) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(written), written)
	}
}

func TestCRUD_RoutesOnly(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	written, err := CRUD(CRUDOptions{
		RawName:    "product",
		ModuleName: "github.com/test/myapp",
		Routes:     true,
	})
	if err != nil {
		t.Fatalf("CRUD: %v", err)
	}
	// routes only — no test file for routes
	if len(written) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(written), written)
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
	// Generate again — must refuse at the first existing file.
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
	if !contains(content, "package invoice") {
		t.Errorf("model file doesn't contain 'package invoice':\n%s", content)
	}
}

func TestCRUD_ModelTestContent(t *testing.T) {
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

	// Second file should be the model_test.go
	data, _ := os.ReadFile(written[1])
	content := string(data)
	if !contains(content, "package product_test") {
		t.Errorf("test file doesn't contain 'package product_test':\n%s", content)
	}
	if !contains(content, "TestProductModel") {
		t.Errorf("test file doesn't contain 'TestProductModel':\n%s", content)
	}
}

func TestCRUD_MigrationContent_DriverDefault(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	written, err := CRUD(CRUDOptions{
		RawName:    "product",
		ModuleName: "github.com/test/myapp",
		Migration:  true,
	})
	if err != nil {
		t.Fatalf("CRUD: %v", err)
	}

	// Find the up migration file and check for Postgres syntax.
	for _, path := range written {
		if strings.HasSuffix(path, ".up.sql") {
			data, _ := os.ReadFile(path)
			content := string(data)
			if !contains(content, "BIGSERIAL") {
				t.Errorf("default driver migration doesn't contain BIGSERIAL:\n%s", content)
			}
		}
	}
}

func TestCRUD_MigrationContent_MySQL(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	written, err := CRUD(CRUDOptions{
		RawName:    "product",
		ModuleName: "github.com/test/myapp",
		Migration:  true,
		DBDriver:   "mysql",
	})
	if err != nil {
		t.Fatalf("CRUD: %v", err)
	}

	for _, path := range written {
		if strings.HasSuffix(path, ".up.sql") {
			data, _ := os.ReadFile(path)
			content := string(data)
			if !contains(content, "BIGINT UNSIGNED AUTO_INCREMENT") {
				t.Errorf("mysql migration doesn't contain BIGINT UNSIGNED:\n%s", content)
			}
			if !contains(content, "DATETIME(6)") {
				t.Errorf("mysql migration doesn't contain DATETIME(6):\n%s", content)
			}
		}
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
