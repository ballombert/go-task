package rules

import (
	"fmt"
	"time"

	"github.com/beaallombert/gotask/internal/domain"
)

// RecurrenceGenerator creates future task instances for recurring tasks.
type RecurrenceGenerator struct{}

func NewRecurrenceGenerator() *RecurrenceGenerator {
	return &RecurrenceGenerator{}
}

// Generate creates instances up to horizon for tasks with due date + recurrence.
// It updates task.Recurrence.LastGenerated to avoid duplicate generation.
func (g *RecurrenceGenerator) Generate(tasks []*domain.Task, now time.Time, horizon time.Time) []*domain.Task {
	if horizon.Before(now) {
		return nil
	}

	generated := make([]*domain.Task, 0)

	for _, task := range tasks {
		if task == nil || task.Recurrence == nil || task.DueDate == nil {
			continue
		}
		if task.Recurrence.Type == domain.RecurrenceNone {
			continue
		}

		cursor := *task.DueDate
		if !task.Recurrence.LastGenerated.IsZero() {
			cursor = task.Recurrence.LastGenerated
		}

		last := cursor
		for {
			next, ok := nextOccurrence(cursor, task.Recurrence)
			if !ok {
				break
			}
			if next.After(horizon) {
				break
			}
			cursor = next
			last = next
			if next.Before(now) {
				continue
			}

			instance := cloneRecurringTask(task, next, now)
			generated = append(generated, instance)
		}

		if !last.IsZero() {
			task.Recurrence.LastGenerated = last
		}
	}

	return generated
}

func cloneRecurringTask(base *domain.Task, dueDate time.Time, now time.Time) *domain.Task {
	clone := &domain.Task{
		ID:          fmt.Sprintf("recur-%d", now.UnixNano()),
		Description: base.Description,
		Status:      domain.StatusPaused,
		Priority:    base.Priority,
		DueDate:     &dueDate,
		Duration:    0,
		CreatedAt:   now,
		LineNumber:  0,
		ParentID:    "",
	}
	if base.Recurrence != nil {
		clone.Recurrence = &domain.Recurrence{
			Type:       base.Recurrence.Type,
			Interval:   base.Recurrence.Interval,
			DaysOfWeek: append([]int(nil), base.Recurrence.DaysOfWeek...),
			DaysOfMonth: append([]int(nil),
				base.Recurrence.DaysOfMonth...),
		}
	}
	return clone
}

func nextOccurrence(from time.Time, recurrence *domain.Recurrence) (time.Time, bool) {
	if recurrence == nil {
		return time.Time{}, false
	}

	interval := recurrence.Interval
	if interval <= 0 {
		interval = 1
	}

	switch recurrence.Type {
	case domain.RecurrenceDaily:
		return from.AddDate(0, 0, interval), true
	case domain.RecurrenceWeekly:
		return from.AddDate(0, 0, 7*interval), true
	case domain.RecurrenceMonthly:
		return from.AddDate(0, interval, 0), true
	case domain.RecurrenceYearly:
		return from.AddDate(interval, 0, 0), true
	default:
		return time.Time{}, false
	}
}
