package domain

import "time"

// RuleType defines what kind of rule to trigger
type RuleType string

const (
	RuleTypeTimerCompleted       RuleType = "timer_completed"
	RuleTypeTaskOverdue          RuleType = "task_overdue"
	RuleTypeNoActiveTask         RuleType = "no_active_task"
	RuleTypeRecurrenceCheck      RuleType = "recurrence_check"
	RuleTypeFocusCheck           RuleType = "focus_check"
	RuleTypeInactivityWarning    RuleType = "inactivity_warning"
)

// InterventionType defines what action to take
type InterventionType string

const (
	InterventionTypeNotification InterventionType = "notification"
	InterventionTypeRefocus      InterventionType = "refocus"
	InterventionTypeBreakReminder InterventionType = "break_reminder"
	InterventionTypeTaskGeneration InterventionType = "task_generation"
)

// Intervention represents a system decision to act
type Intervention struct {
	ID              string
	Type            InterventionType
	RuleType        RuleType
	Message         string
	TaskID          *string
	CreatedAt       time.Time
	Dismissed       bool
	DismissedAt     *time.Time
	Confidence      float64 // 0.0 to 1.0
}

// Rule defines a condition that may trigger an intervention
type Rule struct {
	ID                string
	Type              RuleType
	Enabled           bool
	Priority          int // Lower = higher priority
	InterventionType  InterventionType
	Condition         func(state *SystemState) bool
	MessageGenerator  func(state *SystemState) string
}

// SystemState represents the current state for rule evaluation
type SystemState struct {
	Tasks              []*Task
	ActiveTask         *Task
	ActiveTimer        *Timer
	LastInterventions  []*Intervention
	UpcomingTasks      []*Task
	SessionStartTime   time.Time
	InactivityDuration time.Duration
	Timestamp          time.Time
}

// HasRecentIntervention checks if an intervention was made recently
func (s *SystemState) HasRecentIntervention(ruleType RuleType, withinLast time.Duration) bool {
	cutoff := time.Now().Add(-withinLast)
	for _, intervention := range s.LastInterventions {
		if intervention.RuleType == ruleType && intervention.CreatedAt.After(cutoff) {
			return true
		}
	}
	return false
}

// CountActiveTasks returns number of tasks in progress
func (s *SystemState) CountActiveTasks() int {
	count := 0
	for _, task := range s.Tasks {
		if task.Status == StatusInProgress {
			count++
		}
	}
	return count
}

// TopNTasks returns top N tasks sorted by effective priority
func (s *SystemState) TopNTasks(n int) []*Task {
	// Simple bubble sort for top N
	var sorted []*Task
	sorted = append(sorted, s.Tasks...)
	
	// Sort by effective priority
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-1-i; j++ {
			if sorted[j].EffectivePriority() > sorted[j+1].EffectivePriority() {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	if len(sorted) > n {
		sorted = sorted[:n]
	}
	return sorted
}
