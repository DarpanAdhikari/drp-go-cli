// Package migration implements the core migration engine: file discovery,
// schema_history bookkeeping, and up/down/rollback/fresh execution.
// This is the riskiest, most stateful component in drp and is tested hardest.
package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// filePattern matches migration file names in the form:
// <digits>_some_migration_name.up.sql
// <digits>_some_migration_name.down.sql
// Supports arbitrary digit length (e.g. 14 for legacy, 16 for microsecond).
var filePattern = regexp.MustCompile(`^(\d+)_([a-z0-9_]+)\.(up|down)\.sql$`)

// File represents a discovered migration on disk.
type File struct {
	Timestamp int64

	Name      string // e.g. "create_users_table"
	UpPath    string // absolute path to the .up.sql file
	DownPath  string // absolute path to the .down.sql file (may be empty warning)
}

// Identifier returns the canonical migration ID stored in schema_history,
// e.g. "1719929853123456_create_users_table".
func (f File) Identifier() string {
	return fmt.Sprintf("%d_%s", f.Timestamp, f.Name)
}

// DiscoverFiles scans dir and returns all migration File entries sorted by
// timestamp ascending. Returns an error if an up file exists without a
// matching down file or vice versa.
func DiscoverFiles(dir string) ([]File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("migration: directory %q not found — run `drp init` first", dir)
		}
		return nil, fmt.Errorf("migration: reading directory %q: %w", dir, err)
	}

	type half struct {
		ts   int64
		name string
		up   string
		down string
	}
	index := map[string]*half{}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := filePattern.FindStringSubmatch(e.Name())
		if m == nil {
			continue // skip files that don't match the naming convention
		}

		ts, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("migration: invalid timestamp %q: %w", m[1], err)
		}

		key := m[1] + "_" + m[2]
		if _, ok := index[key]; !ok {
			index[key] = &half{ts: ts, name: m[2]}
		}
		abs := filepath.Join(dir, e.Name())
		switch m[3] {
		case "up":
			index[key].up = abs
		case "down":
			index[key].down = abs
		}
	}

	var files []File
	for key, h := range index {
		if h.up == "" {
			return nil, fmt.Errorf("migration: %q has a down file but no up file", key)
		}
		if h.down == "" {
			return nil, fmt.Errorf("migration: %q has an up file but no down file", key)
		}
		files = append(files, File{
			Timestamp: h.ts,
			Name:      h.name,
			UpPath:    h.up,
			DownPath:  h.down,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Timestamp < files[j].Timestamp
	})
	return files, nil
}

// NewFile creates a new timestamped up/down migration file pair in dir.
// Returns an error if a migration with the same timestamp-and-name already
// exists (collision guard — callers should not generate two in the same second).
func NewFile(dir, name string) (File, error) {
	return NewFileForTable(dir, name, "")
}

// NewFileForTable creates a new migration pair with starter table SQL. If
// tableName is empty, the table name is inferred from common migration names.
func NewFileForTable(dir, name, tableName string) (File, error) {
	// Sanitise name: lowercase, spaces → underscores, strip anything non-alnum.
	name = strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	name = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(name, "")
	if name == "" {
		return File{}, fmt.Errorf("migration: name %q is invalid after sanitisation", name)
	}
	tableName = sanitizeIdentifier(tableName)
	if tableName == "" {
		tableName = inferTableName(name)
	}

	ts := time.Now().UnixMicro()
	base := fmt.Sprintf("%d_%s", ts, name)
	upPath := filepath.Join(dir, base+".up.sql")
	downPath := filepath.Join(dir, base+".down.sql")

	if _, err := os.Stat(upPath); err == nil {
		return File{}, fmt.Errorf("migration: %q already exists — wait a second and retry", base)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return File{}, fmt.Errorf("migration: creating directory %q: %w", dir, err)
	}

	upContent := upTemplate(base, tableName)
	downContent := downTemplate(base, tableName)

	if err := os.WriteFile(upPath, []byte(upContent), 0o644); err != nil {
		return File{}, fmt.Errorf("migration: writing %q: %w", upPath, err)
	}
	if err := os.WriteFile(downPath, []byte(downContent), 0o644); err != nil {
		os.Remove(upPath) // clean up the up file so we don't leave an orphan
		return File{}, fmt.Errorf("migration: writing %q: %w", downPath, err)
	}

	return File{Timestamp: ts, Name: name, UpPath: upPath, DownPath: downPath}, nil
}

func sanitizeIdentifier(name string) string {
	name = strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	name = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(name, "")
	return strings.Trim(name, "_")
}

func inferTableName(migrationName string) string {
	switch {
	case strings.HasPrefix(migrationName, "create_") && strings.HasSuffix(migrationName, "_table"):
		return strings.TrimSuffix(strings.TrimPrefix(migrationName, "create_"), "_table")
	case strings.HasPrefix(migrationName, "create_"):
		return strings.TrimPrefix(migrationName, "create_")
	case strings.Contains(migrationName, "_to_"):
		parts := strings.Split(migrationName, "_to_")
		return sanitizeIdentifier(parts[len(parts)-1])
	case strings.Contains(migrationName, "_from_"):
		parts := strings.Split(migrationName, "_from_")
		return sanitizeIdentifier(parts[len(parts)-1])
	default:
		return migrationName
	}
}

func upTemplate(base, tableName string) string {
	return fmt.Sprintf(`-- Migration: %s (up)
CREATE TABLE IF NOT EXISTS %s (
  id BIGSERIAL PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`, base, tableName)
}

func downTemplate(base, tableName string) string {
	return fmt.Sprintf(`-- Migration: %s (down)
DROP TABLE IF EXISTS %s;
`, base, tableName)
}
