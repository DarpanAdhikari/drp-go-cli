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
	"strings"
	"time"
)

// filePattern matches migration file names in the form:
// 20060102150405_some_migration_name.up.sql
// 20060102150405_some_migration_name.down.sql
var filePattern = regexp.MustCompile(`^(\d{14})_([a-z0-9_]+)\.(up|down)\.sql$`)

// timestampLayout is the layout used for migration timestamps.
const timestampLayout = "20060102150405"

// File represents a discovered migration on disk.
type File struct {
	Timestamp time.Time
	Name      string // e.g. "create_users_table"
	UpPath    string // absolute path to the .up.sql file
	DownPath  string // absolute path to the .down.sql file (may be empty warning)
}

// Identifier returns the canonical migration ID stored in schema_history,
// e.g. "20240115120000_create_users_table".
func (f File) Identifier() string {
	return fmt.Sprintf("%s_%s", f.Timestamp.Format(timestampLayout), f.Name)
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
		ts   time.Time
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

		ts, err := time.ParseInLocation(timestampLayout, m[1], time.UTC)
		if err != nil {
			return nil, fmt.Errorf("migration: invalid timestamp in %q: %w", e.Name(), err)
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
		return files[i].Timestamp.Before(files[j].Timestamp)
	})
	return files, nil
}

// NewFile creates a new timestamped up/down migration file pair in dir.
// Returns an error if a migration with the same timestamp-and-name already
// exists (collision guard — callers should not generate two in the same second).
func NewFile(dir, name string) (File, error) {
	// Sanitise name: lowercase, spaces → underscores, strip anything non-alnum.
	name = strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	name = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(name, "")
	if name == "" {
		return File{}, fmt.Errorf("migration: name %q is invalid after sanitisation", name)
	}

	ts := time.Now().UTC()
	base := fmt.Sprintf("%s_%s", ts.Format(timestampLayout), name)
	upPath := filepath.Join(dir, base+".up.sql")
	downPath := filepath.Join(dir, base+".down.sql")

	if _, err := os.Stat(upPath); err == nil {
		return File{}, fmt.Errorf("migration: %q already exists — wait a second and retry", base)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return File{}, fmt.Errorf("migration: creating directory %q: %w", dir, err)
	}

	upContent := fmt.Sprintf("-- Migration: %s (up)\n-- TODO: write your UP migration SQL here\n", base)
	downContent := fmt.Sprintf("-- Migration: %s (down)\n-- TODO: write your DOWN (rollback) SQL here\n", base)

	if err := os.WriteFile(upPath, []byte(upContent), 0o644); err != nil {
		return File{}, fmt.Errorf("migration: writing %q: %w", upPath, err)
	}
	if err := os.WriteFile(downPath, []byte(downContent), 0o644); err != nil {
		os.Remove(upPath) // clean up the up file so we don't leave an orphan
		return File{}, fmt.Errorf("migration: writing %q: %w", downPath, err)
	}

	return File{Timestamp: ts, Name: name, UpPath: upPath, DownPath: downPath}, nil
}
