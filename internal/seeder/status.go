package seeder

import (
	"fmt"
	"strings"
	"time"
)

// StatusEntry describes one seeder's applied/pending state.
type StatusEntry struct {
	Seeder    string
	Applied   bool
	AppliedAt time.Time
}

// FormatStatus renders a slice of StatusEntry into an aligned, human-readable
// table suitable for CLI output.
//
// Example output:
//
//	[✓] 20240115120000_seed_users   (applied at 2024-01-15 12:01:00)
//	[ ] 20240116080000_seed_companies
func FormatStatus(entries []StatusEntry) string {
	if len(entries) == 0 {
		return "  No seeder files found."
	}

	var sb strings.Builder
	for _, e := range entries {
		mark := " "
		suffix := ""
		if e.Applied {
			mark = "✓"
			suffix = fmt.Sprintf("   (applied at %s)", e.AppliedAt.Format("2006-01-02 15:04:05"))
		}
		sb.WriteString(fmt.Sprintf("  [%s] %s%s\n", mark, e.Seeder, suffix))
	}
	return sb.String()
}

// Status returns a list of all known seeder files with their applied/pending
// state, determined by comparing on-disk files against the seed_history table.
func (e *Engine) Status() ([]StatusEntry, error) {
	if err := EnsureSeedHistoryTable(e.DB, e.DriverName); err != nil {
		return nil, err
	}

	files, err := DiscoverFiles(e.SeedersDir)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}

	applied, err := AppliedSeeders(e.DB)
	if err != nil {
		return nil, err
	}
	appliedMap := make(map[string]time.Time, len(applied))
	for _, a := range applied {
		appliedMap[a.Seeder] = a.AppliedAt
	}

	entries := make([]StatusEntry, 0, len(files))
	for _, f := range files {
		id := f.Identifier()
		at, ok := appliedMap[id]
		entries = append(entries, StatusEntry{
			Seeder:    id,
			Applied:   ok,
			AppliedAt: at,
		})
	}
	return entries, nil
}
