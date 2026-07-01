package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeEnv writes a .env file to a temp dir and returns its path.
func writeEnv(t *testing.T, contents string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(f, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return f
}

// clearEnv unsets all DB_* env vars and restores them after the test.
func clearEnv(t *testing.T) {
	t.Helper()
	keys := []string{"DB_DRIVER", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"}
	saved := make(map[string]string, len(keys))
	for _, k := range keys {
		saved[k] = os.Getenv(k)
		os.Unsetenv(k)
	}
	t.Cleanup(func() {
		for k, v := range saved {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	})
}

func TestLoad_ValidPostgresConfig(t *testing.T) {
	clearEnv(t)
	path := writeEnv(t, "DB_DRIVER=postgres\nDB_HOST=localhost\nDB_PORT=5432\nDB_USER=myuser\nDB_PASSWORD=secret\nDB_NAME=mydb\nDB_SSLMODE=disable\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DBDriver != "postgres" {
		t.Errorf("DBDriver = %q, want postgres", cfg.DBDriver)
	}
	if cfg.DBName != "mydb" {
		t.Errorf("DBName = %q, want mydb", cfg.DBName)
	}
}

func TestLoad_ValidMySQLConfig(t *testing.T) {
	clearEnv(t)
	path := writeEnv(t, "DB_DRIVER=mysql\nDB_HOST=127.0.0.1\nDB_USER=root\nDB_PASSWORD=secret\nDB_NAME=testdb\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DBDriver != "mysql" {
		t.Errorf("DBDriver = %q, want mysql", cfg.DBDriver)
	}
	if cfg.DBPort != "3306" {
		t.Errorf("DBPort = %q, want 3306 (default)", cfg.DBPort)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	clearEnv(t)
	_, err := Load("/nonexistent/.env")
	if err == nil {
		t.Error("expected error for missing .env, got nil")
	}
}

func TestLoad_MissingDBUser(t *testing.T) {
	clearEnv(t)
	path := writeEnv(t, "DB_NAME=mydb\n")
	_, err := Load(path)
	if err == nil {
		t.Error("expected validation error for missing DB_USER, got nil")
	}
}

func TestLoad_MissingDBName(t *testing.T) {
	clearEnv(t)
	path := writeEnv(t, "DB_USER=myuser\n")
	_, err := Load(path)
	if err == nil {
		t.Error("expected validation error for missing DB_NAME, got nil")
	}
}

func TestLoad_UnknownDriver(t *testing.T) {
	clearEnv(t)
	path := writeEnv(t, "DB_DRIVER=sqlite\nDB_USER=myuser\nDB_NAME=mydb\n")
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for unknown driver, got nil")
	}
}

func TestConfig_DSN_Postgres(t *testing.T) {
	cfg := &Config{DBDriver: "postgres", DBHost: "localhost", DBPort: "5432",
		DBUser: "user", DBPassword: "pass", DBName: "mydb", DBSSLMode: "disable"}
	if cfg.DSN() == "" {
		t.Error("DSN returned empty string for postgres")
	}
}

func TestConfig_DSN_MySQL(t *testing.T) {
	cfg := &Config{DBDriver: "mysql", DBHost: "localhost", DBPort: "3306",
		DBUser: "root", DBPassword: "pass", DBName: "mydb"}
	if cfg.DSN() == "" {
		t.Error("DSN returned empty string for mysql")
	}
}

func TestConfig_AdminDSN_Postgres(t *testing.T) {
	cfg := &Config{DBDriver: "postgres", DBHost: "localhost", DBPort: "5432",
		DBUser: "user", DBPassword: "pass", DBName: "mydb", DBSSLMode: "disable"}
	dsn := cfg.AdminDSN()
	if dsn == "" {
		t.Error("AdminDSN returned empty string for postgres")
	}
}

func TestConfig_AdminDSN_MySQL(t *testing.T) {
	cfg := &Config{DBDriver: "mysql", DBHost: "localhost", DBPort: "3306",
		DBUser: "root", DBPassword: "pass", DBName: "mydb"}
	dsn := cfg.AdminDSN()
	if dsn == "" {
		t.Error("AdminDSN returned empty string for mysql")
	}
}
