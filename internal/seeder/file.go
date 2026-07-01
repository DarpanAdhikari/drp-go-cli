// Package seeder implements the seeder engine: creating seeder files and
// running them against the database in order. Seeders use the same
// timestamp-prefixed naming convention as migrations but are tracked in
// a separate seed_history table so migrate:fresh can re-seed cleanly.
package seeder

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// filePattern matches seeder file names:
// 20060102150405_seed_users.sql
var filePattern = regexp.MustCompile(`^(\d{14})_([a-z0-9_]+)\.sql$`)

const timestampLayout = "20060102150405"

// File represents a single discovered seeder file on disk.
type File struct {
	Timestamp time.Time
	Name      string
	Path      string
}

// Identifier returns the canonical seeder identifier stored in seed_history,
// e.g. "20240115120000_seed_users".
func (f File) Identifier() string {
	return fmt.Sprintf("%s_%s", f.Timestamp.Format(timestampLayout), f.Name)
}

// DiscoverFiles scans dir and returns all seeder File entries sorted by
// timestamp ascending. Non-matching files are silently skipped.
func DiscoverFiles(dir string) ([]File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("seeder: directory %q not found — run `drp init` first", dir)
		}
		return nil, fmt.Errorf("seeder: reading directory %q: %w", dir, err)
	}

	var files []File
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := filePattern.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		ts, err := time.ParseInLocation(timestampLayout, m[1], time.UTC)
		if err != nil {
			return nil, fmt.Errorf("seeder: invalid timestamp in %q: %w", e.Name(), err)
		}
		files = append(files, File{
			Timestamp: ts,
			Name:      m[2],
			Path:      filepath.Join(dir, e.Name()),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Timestamp.Before(files[j].Timestamp)
	})
	return files, nil
}

// NewFile creates a new timestamped seeder SQL file in dir with the given name.
// Returns an error if a file with the same timestamp-and-name already exists.
func NewFile(dir, name string) (File, error) {
	name = strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	name = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(name, "")
	if name == "" {
		return File{}, fmt.Errorf("seeder: name is invalid after sanitisation")
	}

	ts := time.Now().UTC()
	filename := fmt.Sprintf("%s_%s.sql", ts.Format(timestampLayout), name)
	path := filepath.Join(dir, filename)

	if _, err := os.Stat(path); err == nil {
		return File{}, fmt.Errorf("seeder: %q already exists — wait a second and retry", filename)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return File{}, fmt.Errorf("seeder: creating directory %q: %w", dir, err)
	}

	content := fmt.Sprintf("-- Seeder: %s\n-- TODO: write your seed INSERT statements here\n\n", filename)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return File{}, fmt.Errorf("seeder: writing %q: %w", path, err)
	}

	return File{Timestamp: ts, Name: name, Path: path}, nil
}
