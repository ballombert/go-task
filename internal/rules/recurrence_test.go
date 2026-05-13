package rules

import (
	"testing"
	"time"

	"github.com/beaallombert/gotask/internal/domain"
)

func TestRecurrenceGenerateDaily(t *testing.T) {
	now := time.Date(2026, 5, 12, 9, 0, 0, 0, time.UTC)
	firstDue := time.Date(2026, 5, 13, 9, 0, 0, 0, time.UTC)
	horizon := time.Date(2026, 5, 16, 9, 0, 0, 0, time.UTC)

	task := &domain.Task{
		ID:          "base-1",
		Description: "Daily focus",
		Status:      domain.StatusPaused,
		Priority:    domain.PriorityHigh,
		DueDate:     &firstDue,
		Recurrence: &domain.Recurrence{
			Type:     domain.RecurrenceDaily,
			Interval: 1,
		},
	}

	g := NewRecurrenceGenerator()
	got := g.Generate([]*domain.Task{task}, now, horizon)

	if len(got) != 3 {
		t.Fatalf("expected 3 generated tasks, got %d", len(got))
	}

	if got[0].DueDate == nil || got[0].DueDate.Format("2006-01-02") != "2026-05-14" {
		t.Fatalf("unexpected first generated due date: %+v", got[0].DueDate)
	}
	if got[2].DueDate == nil || got[2].DueDate.Format("2006-01-02") != "2026-05-16" {
		t.Fatalf("unexpected last generated due date: %+v", got[2].DueDate)
	}

	if task.Recurrence.LastGenerated.IsZero() {
		t.Fatalf("expected LastGenerated to be updated")
	}
}

func TestRecurrenceGenerateSkipsUnsupportedAndPastHorizon(t *testing.T) {
	now := time.Date(2026, 5, 12, 9, 0, 0, 0, time.UTC)
	firstDue := time.Date(2026, 5, 13, 9, 0, 0, 0, time.UTC)
	horizon := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)

	task := &domain.Task{
		ID:          "base-2",
		Description: "Custom recurrence",
		Status:      domain.StatusPaused,
		Priority:    domain.PriorityMedium,
		DueDate:     &firstDue,
		Recurrence: &domain.Recurrence{
			Type: domain.RecurrenceCustom,
		},
	}

	g := NewRecurrenceGenerator()
	got := g.Generate([]*domain.Task{task}, now, horizon)
	if len(got) != 0 {
		t.Fatalf("expected no generated tasks for unsupported custom recurrence, got %d", len(got))
	}
}
