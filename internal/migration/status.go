package migration

import (
	"fmt"
	"strings"
)

// StatusEntry describes one migration's applied/pending state.
type StatusEntry struct {
	Migration string
	Applied   bool
	Batch     int
}

// FormatStatus renders a slice of StatusEntry into an aligned, human-readable
// table suitable for CLI output.
//
// Example output:
//
//	[✓] 20240115120000_create_users_table   (batch 1)
//	[ ] 20240116080000_add_email_index
func FormatStatus(entries []StatusEntry) string {
	if len(entries) == 0 {
		return "  No migration files found."
	}

	var sb strings.Builder
	for _, e := range entries {
		mark := " "
		suffix := ""
		if e.Applied {
			mark = "✓"
			suffix = fmt.Sprintf("   (batch %d)", e.Batch)
		}
		sb.WriteString(fmt.Sprintf("  [%s] %s%s\n", mark, e.Migration, suffix))
	}
	return sb.String()
}
