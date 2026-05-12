package domain

import (
	"time"
)

// Status represents the state of a task
type Status string

const (
	StatusPaused    Status = "paused"     // [ ]
	StatusInProgress Status = "in_progress" // [>]
	StatusCompleted Status = "completed"   // [x]
)

// Priority represents the urgency level of a task
type Priority string

const (
	PriorityHighest Priority = "highest" // 🔺
	PriorityHigh    Priority = "high"    // ⏫
	PriorityMedium  Priority = "medium"  // 🔼
	PriorityLow     Priority = "low"     // 🔽
	PriorityLowest  Priority = "lowest"  // ⏬
)

// RecurrenceType defines how a task repeats
type RecurrenceType string

const (
	RecurrenceNone    RecurrenceType = "none"
	RecurrenceDaily   RecurrenceType = "daily"
	RecurrenceWeekly  RecurrenceType = "weekly"
	RecurrenceMonthly RecurrenceType = "monthly"
	RecurrenceYearly  RecurrenceType = "yearly"
	RecurrenceCustom  RecurrenceType = "custom"
)

// Recurrence defines the recurrence pattern for a task
type Recurrence struct {
	Type          RecurrenceType
	Interval      int       // e.g., every N days/weeks/months
	DaysOfWeek    []int     // 0=Sunday, 1=Monday, etc. (for weekly)
	DaysOfMonth   []int     // 1-31 (for monthly)
	LastGenerated time.Time // Last time an instance was generated
}

// Task represents a single task in the inbox
type Task struct {
	ID              string // Unique identifier (line hash or UUID)
	Description     string
	Status          Status
	Priority        Priority
	DueDate         *time.Time
	Duration        int        // Minutes spent on this task (cumulative)
	Recurrence      *Recurrence
	Subtasks        []*Task
	CreatedAt       time.Time
	CompletedAt     *time.Time
	LineNumber      int // Line in inbox.md (for updates)
	ParentID        string // Empty if root task
}

// TimeSpent returns the duration as a formatted string
func (t *Task) TimeSpent() string {
	if t.Duration == 0 {
		return ""
	}
	return formatDuration(t.Duration)
}

// IsOverdue returns true if task has a due date in the past
func (t *Task) IsOverdue() bool {
	if t.DueDate == nil || t.Status == StatusCompleted {
		return false
	}
	return t.DueDate.Before(time.Now())
}

// IsDueToday returns true if task is due today
func (t *Task) IsDueToday() bool {
	if t.DueDate == nil {
		return false
	}
	today := time.Now().Truncate(24 * time.Hour)
	dueDate := t.DueDate.Truncate(24 * time.Hour)
	return today.Equal(dueDate)
}

// EffectivePriority returns the priority for sorting
// considering overdue status
func (t *Task) EffectivePriority() int {
	// Lower number = higher priority
	priorities := map[Priority]int{
		PriorityHighest: 1,
		PriorityHigh:    2,
		PriorityMedium:  3,
		PriorityLow:     4,
		PriorityLowest:  5,
	}
	score := priorities[t.Priority]
	if t.IsOverdue() {
		score -= 10 // Overdue tasks jump to top
	}
	return score
}

func formatDuration(minutes int) string {
	if minutes < 60 {
		return "" // Don't show minutes below 1 hour in compact format
	}
	hours := minutes / 60
	remainingMins := minutes % 60
	if remainingMins == 0 {
		return "⏱ " + string(rune(hours)) + "h"
	}
	return "⏱ " + string(rune(hours)) + "h" + string(rune(remainingMins)) + "m"
}

// CountCompletedSubtasks returns the number of completed subtasks
func (t *Task) CountCompletedSubtasks() int {
	count := 0
	for _, st := range t.Subtasks {
		if st.Status == StatusCompleted {
			count++
		}
	}
	return count
}

// IsSubtaskProgressVisible returns true if task has subtasks
func (t *Task) IsSubtaskProgressVisible() bool {
	return len(t.Subtasks) > 0
}

// SubtaskProgress returns "X/Y" format
func (t *Task) SubtaskProgress() string {
	completed := t.CountCompletedSubtasks()
	total := len(t.Subtasks)
	if total == 0 {
		return ""
	}
	return string(rune(completed)) + "/" + string(rune(total))
}
