package generator

import (
	"os"
	"path/filepath"
	"strings"
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
		filepath.Join("myapp", "internal", "auth"),
		filepath.Join("myapp", "internal", "config"),
		filepath.Join("myapp", "internal", "middleware"),
		filepath.Join("myapp", "internal", "routes"),
		filepath.Join("myapp", "internal", "shared"),
		filepath.Join("myapp", "internal", "user"),
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

func TestProject_WithAuthCreatesAuthPreset(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	err := Project(ProjectOptions{Name: "myapp", ModuleName: "github.com/test/myapp", Auth: true})
	if err != nil {
		t.Fatalf("Project: %v", err)
	}

	expectedFiles := []string{
		filepath.Join("myapp", "internal", "auth", "jwt.go"),
		filepath.Join("myapp", "internal", "auth", "token_store.go"),
		filepath.Join("myapp", "internal", "auth", "handler.go"),
		filepath.Join("myapp", "internal", "auth", "docs.go"),
		filepath.Join("myapp", "docs", "docs.go"),
		filepath.Join("myapp", "internal", "middleware", "auth.go"),
		filepath.Join("myapp", "internal", "middleware", "cors.go"),
		filepath.Join("myapp", "internal", "middleware", "requestid.go"),
		filepath.Join("myapp", "internal", "middleware", "rate_limiter.go"),
		filepath.Join("myapp", "internal", "middleware", "rate_limit.go"),
		filepath.Join("myapp", "internal", "shared", "base.go"),
		filepath.Join("myapp", "internal", "shared", "context.go"),
		filepath.Join("myapp", "internal", "shared", "response.go"),
		filepath.Join("myapp", "internal", "user", "model.go"),
		filepath.Join("myapp", "internal", "user", "repository.go"),
		filepath.Join("myapp", "internal", "user", "service.go"),
		filepath.Join("myapp", "internal", "user", "handler.go"),
		filepath.Join("myapp", "internal", "user", "repository_test.go"),
		filepath.Join("myapp", "internal", "user", "service_test.go"),
		filepath.Join("myapp", "internal", "auth", "handler_test.go"),
		filepath.Join("myapp", "internal", "routes", "routes.go"),
		filepath.Join("myapp", "database", "migrations", "000001_create_users.up.sql"),
	}
	for _, f := range expectedFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("expected auth file %q not found", f)
		}
	}

	env, err := os.ReadFile(filepath.Join("myapp", ".env"))
	if err != nil {
		t.Fatalf("read .env: %v", err)
	}
	for _, want := range []string{"JWT_SECRET=", "ACCESS_TOKEN_EXPIRY_MINUTES=", "REFRESH_TOKEN_EXPIRY_DAYS=", "ENCRYPTION_KEY="} {
		if !strings.Contains(string(env), want) {
			t.Errorf(".env missing %q", want)
		}
	}

	routes, err := os.ReadFile(filepath.Join("myapp", "internal", "routes", "routes.go"))
	if err != nil {
		t.Fatalf("read routes: %v", err)
	}
	for _, want := range []string{"/auth/register", "/auth/login", "/auth/refresh", "/auth/logout", "/me"} {
		if !strings.Contains(string(routes), want) {
			t.Errorf("routes missing %q", want)
		}
	}

	mainGo, err := os.ReadFile(filepath.Join("myapp", "cmd", "api", "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	if !strings.Contains(string(mainGo), "/swagger/") {
		t.Error("main.go missing swagger route")
	}
	if !strings.Contains(string(mainGo), "http-swagger") {
		t.Error("main.go missing http-swagger import")
	}
}
