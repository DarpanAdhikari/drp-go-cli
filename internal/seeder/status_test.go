package seeder

import (
	"strings"
	"testing"
	"time"
)

func TestFormatStatus_EmptyList(t *testing.T) {
	out := FormatStatus(nil)
	if !strings.Contains(out, "No seeder files") {
		t.Errorf("expected empty message, got: %q", out)
	}
}

func TestFormatStatus_ShowsAppliedAndPending(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 1, 0, 0, time.UTC)
	entries := []StatusEntry{
		{Seeder: "20240115120000_seed_users", Applied: true, AppliedAt: now},
		{Seeder: "20240116080000_seed_companies", Applied: false},
	}
	out := FormatStatus(entries)

	if !strings.Contains(out, "✓") {
		t.Error("expected ✓ for applied seeder")
	}
	if !strings.Contains(out, "applied at 2024-01-15 12:01:00") {
		t.Error("expected applied_at timestamp for applied seeder")
	}
	if !strings.Contains(out, "[ ]") {
		t.Error("expected [ ] for pending seeder")
	}
}
