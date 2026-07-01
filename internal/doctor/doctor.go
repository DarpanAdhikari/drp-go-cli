// Package doctor implements the environment checks for `drp doctor`.
// Each check is independent and all checks always run — one failure
// does not skip subsequent checks, so the user gets a complete picture.
package doctor

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/yourorg/drp/internal/config"
	"github.com/yourorg/drp/internal/db"
)

// Result holds the outcome of a single doctor check.
type Result struct {
	Label  string
	OK     bool
	Detail string // shown on both pass and fail for context
}

// Run executes all doctor checks and returns a slice of Results.
// envFile is the path to .env (empty means default ".env").
func Run(envFile string) []Result {
	var results []Result

	// 1. Go installed
	results = append(results, checkGoVersion())

	// 2. .env file present
	envResult, envOK := checkEnvFile(envFile)
	results = append(results, envResult)

	// 3. Config validity (only if .env was found)
	var cfg *config.Config
	if envOK {
		cfgResult, loadedCfg := checkConfig(envFile)
		results = append(results, cfgResult)
		cfg = loadedCfg
	} else {
		results = append(results, Result{
			Label:  "Config valid",
			OK:     false,
			Detail: "skipped — .env not found",
		})
	}

	// 4. DB reachability (only if config loaded)
	if cfg != nil {
		results = append(results, checkDB(cfg))
	} else {
		results = append(results, Result{
			Label:  "Database reachable",
			OK:     false,
			Detail: "skipped — config could not be loaded",
		})
	}

	// 5. Migrations directory
	results = append(results, checkDir("database/migrations", "Migrations directory"))

	// 6. Seeders directory
	results = append(results, checkDir("database/seeders", "Seeders directory"))

	// 7. go.mod present
	results = append(results, checkFile("go.mod", "go.mod present"))

	return results
}

func checkGoVersion() Result {
	version := runtime.Version()
	goCmd, err := exec.LookPath("go")
	detail := version
	if err == nil {
		detail = fmt.Sprintf("%s (%s)", version, goCmd)
	}
	return Result{Label: "Go installed", OK: true, Detail: detail}
}

func checkEnvFile(envFile string) (Result, bool) {
	if envFile == "" {
		envFile = ".env"
	}
	if _, err := os.Stat(envFile); err != nil {
		return Result{
			Label:  fmt.Sprintf(".env present (%s)", envFile),
			OK:     false,
			Detail: fmt.Sprintf("not found — create %q or run `drp init`", envFile),
		}, false
	}
	return Result{
		Label:  fmt.Sprintf(".env present (%s)", envFile),
		OK:     true,
		Detail: envFile,
	}, true
}

func checkConfig(envFile string) (Result, *config.Config) {
	cfg, err := config.Load(envFile)
	if err != nil {
		msg := strings.ReplaceAll(err.Error(), "\n  → run `drp doctor` for a full check", "")
		return Result{Label: "Config valid", OK: false, Detail: msg}, nil
	}
	return Result{
		Label:  "Config valid",
		OK:     true,
		Detail: fmt.Sprintf("driver=%s host=%s port=%s db=%s", cfg.DBDriver, cfg.DBHost, cfg.DBPort, cfg.DBName),
	}, cfg
}

func checkDB(cfg *config.Config) Result {
	// Doctor uses a plain sql.Open + Ping — it must only observe, not modify.
	// It does NOT call db.Connect() so it won't auto-create databases or
	// schema_history tables.
	driverName := cfg.DBDriver
	if driverName != "mysql" {
		driverName = "postgres"
	}

	database, err := sql.Open(driverName, cfg.DSN())
	if err != nil {
		return Result{
			Label:  "Database reachable",
			OK:     false,
			Detail: fmt.Sprintf("open failed: %v", err),
		}
	}
	defer database.Close()

	if err := db.Ping(database); err != nil {
		return Result{
			Label:  "Database reachable",
			OK:     false,
			Detail: fmt.Sprintf("cannot reach %s:%s — %v", cfg.DBHost, cfg.DBPort, err),
		}
	}
	return Result{
		Label:  "Database reachable",
		OK:     true,
		Detail: fmt.Sprintf("%s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName),
	}
}

func checkDir(path, label string) Result {
	fi, err := os.Stat(path)
	if err != nil || !fi.IsDir() {
		return Result{
			Label:  label,
			OK:     false,
			Detail: fmt.Sprintf("%q not found — run `drp init` to create it", path),
		}
	}
	return Result{Label: label, OK: true, Detail: path}
}

func checkFile(path, label string) Result {
	if _, err := os.Stat(path); err != nil {
		return Result{
			Label:  label,
			OK:     false,
			Detail: fmt.Sprintf("%q not found", path),
		}
	}
	return Result{Label: label, OK: true, Detail: path}
}
