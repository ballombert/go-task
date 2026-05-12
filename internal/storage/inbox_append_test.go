package storage
package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/beaallombert/gotask/internal/domain"
)

func TestAppendTasksSafelyPreservesExistingContent(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "inbox.md")

	if err := os.WriteFile(path, []byte("- [ ] Existing task 🔼"), 0644); err != nil {
		t.Fatalf("write initial file: %v", err)
	}

	due := time.Date(2026, 5, 20, 0, 0, 0, 0, time.Local)
	w := NewInboxWriter(path)
	count, err := w.AppendTasksSafely([]*domain.Task{{
		Description: "Generated recurring",
		Status:      domain.StatusPaused,
		Priority:    domain.PriorityHigh,
		DueDate:     &due,
	}})
	if err != nil {
		t.Fatalf("append tasks safely: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 appended task, got %d", count)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read result file: %v", err)
	}
	content := string(raw)

	if !strings.Contains(content, "Existing task") {
		t.Fatalf("existing content should be preserved")
	}
	if !strings.Contains(content, "Generated recurring") {
		t.Fatalf("appended content missing")
	}
	if !strings.Contains(content, "\n- [ ] Generated recurring") {
		t.Fatalf("expected appended task on a new line")
	}
}
