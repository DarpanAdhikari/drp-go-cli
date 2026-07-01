package migration

import (
	"strings"
	"testing"
)

func TestFormatStatus_EmptyList(t *testing.T) {
	out := FormatStatus(nil)
	if !strings.Contains(out, "No migration files") {
		t.Errorf("expected empty message, got: %q", out)
	}
}

func TestFormatStatus_ShowsAppliedAndPending(t *testing.T) {
	entries := []StatusEntry{
		{Migration: "20240101000000_create_users", Applied: true, Batch: 1},
		{Migration: "20240102000000_add_email_index", Applied: false},
	}
	out := FormatStatus(entries)

	if !strings.Contains(out, "✓") {
		t.Error("expected ✓ for applied migration")
	}
	if !strings.Contains(out, "batch 1") {
		t.Error("expected batch number for applied migration")
	}
	if !strings.Contains(out, "[ ]") {
		t.Error("expected [ ] for pending migration")
	}
}
