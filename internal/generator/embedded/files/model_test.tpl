package {{.DomainName}}_test

import (
	"testing"
	"time"

	"{{.ModuleName}}/internal/{{.DomainName}}"
)

func Test{{.Name}}Model(t *testing.T) {
	m := {{.DomainName}}.{{.Name}}{
		ID:        1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if m.ID != 1 {
		t.Errorf("expected ID 1, got %d", m.ID)
	}
	if m.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if m.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}
