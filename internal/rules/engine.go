package rules

import (
	"fmt"
	"time"

	"github.com/beaallombert/gotask/internal/domain"
)

// Engine evaluates rules and generates interventions
type Engine struct {
	rules []*domain.Rule
}

// NewEngine creates a new rules engine
func NewEngine() *Engine {
	engine := &Engine{
		rules: []*domain.Rule{},
	}

	// Register default rules
	engine.RegisterDefaultRules()

	return engine
}

// RegisterDefaultRules registers built-in rules
func (e *Engine) RegisterDefaultRules() {
	// Rule: Timer completed - suggest break or next task
	e.rules = append(e.rules, &domain.Rule{
		ID:       "rule-timer-completed",
		Type:     domain.RuleTypeTimerCompleted,
		Enabled:  true,
		Priority: 1,
		Condition: func(state *domain.SystemState) bool {
			// Check if a timer was recently completed
			if state.ActiveTimer == nil && len(state.LastInterventions) > 0 {
				lastIntervention := state.LastInterventions[len(state.LastInterventions)-1]
				return lastIntervention.RuleType == domain.RuleTypeTimerCompleted &&
					time.Since(lastIntervention.CreatedAt) < 30*time.Second
			}
			return false
		},
		InterventionType: domain.InterventionTypeNotification,
		MessageGenerator: func(state *domain.SystemState) string {
			if len(state.UpcomingTasks) > 0 {
				return fmt.Sprintf("Timer done! Next: %s", state.UpcomingTasks[0].Description)
			}
			return "Timer done! Take a break or start a new task."
		},
	})

	// Rule: No active task
	e.rules = append(e.rules, &domain.Rule{
		ID:       "rule-no-active-task",
		Type:     domain.RuleTypeNoActiveTask,
		Enabled:  true,
		Priority: 2,
		Condition: func(state *domain.SystemState) bool {
			// No active task and no active timer
			return state.ActiveTask == nil && state.ActiveTimer == nil && len(state.Tasks) > 0
		},
		InterventionType: domain.InterventionTypeRefocus,
		MessageGenerator: func(state *domain.SystemState) string {
			if len(state.UpcomingTasks) > 0 {
				return fmt.Sprintf("Ready to focus on: %s?", state.UpcomingTasks[0].Description)
			}
			return "No active task. Pick one to focus on."
		},
	})

	// Rule: Task overdue
	e.rules = append(e.rules, &domain.Rule{
		ID:       "rule-task-overdue",
		Type:     domain.RuleTypeTaskOverdue,
		Enabled:  true,
		Priority: 3,
		Condition: func(state *domain.SystemState) bool {
			for _, task := range state.Tasks {
				if task.IsOverdue() && task.Status != domain.StatusCompleted {
					return true
				}
			}
			return false
		},
		InterventionType: domain.InterventionTypeNotification,
		MessageGenerator: func(state *domain.SystemState) string {
			for _, task := range state.Tasks {
				if task.IsOverdue() && task.Status != domain.StatusCompleted {
					return fmt.Sprintf("Overdue: %s", task.Description)
				}
			}
			return "You have overdue tasks."
		},
	})

	// Rule: Inactivity warning
	e.rules = append(e.rules, &domain.Rule{
		ID:       "rule-inactivity-warning",
		Type:     domain.RuleTypeInactivityWarning,
		Enabled:  true,
		Priority: 4,
		Condition: func(state *domain.SystemState) bool {
			// If session started more than 2 hours ago with no activity
			return state.ActiveTask == nil &&
				state.ActiveTimer == nil &&
				time.Since(state.SessionStartTime) > 2*time.Hour
		},
		InterventionType: domain.InterventionTypeBreakReminder,
		MessageGenerator: func(state *domain.SystemState) string {
			return "You've been idle for a while. Time for a break or to pick a task?"
		},
	})
}

// Evaluate evaluates all rules against the current system state
func (e *Engine) Evaluate(state *domain.SystemState) []*domain.Intervention {
	var interventions []*domain.Intervention

	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}

		if rule.Condition(state) {
			// Check if we already have a recent intervention of this type
			if state.HasRecentIntervention(rule.Type, 5*time.Minute) {
				continue // Skip if recently triggered
			}

			intervention := &domain.Intervention{
				ID:             fmt.Sprintf("intervention-%d", time.Now().UnixNano()),
				Type:           rule.InterventionType,
				RuleType:       rule.Type,
				Message:        rule.MessageGenerator(state),
				CreatedAt:      time.Now(),
				Dismissed:      false,
				Confidence:     0.8,
			}

			// If there's an active task, associate it
			if state.ActiveTask != nil {
				taskID := state.ActiveTask.ID
				intervention.TaskID = &taskID
			}

			interventions = append(interventions, intervention)
		}
	}

	return interventions
}

// AddRule adds a custom rule to the engine
func (e *Engine) AddRule(rule *domain.Rule) {
	e.rules = append(e.rules, rule)
}

// EnableRule enables a rule by ID
func (e *Engine) EnableRule(ruleID string) {
	for _, rule := range e.rules {
		if rule.ID == ruleID {
			rule.Enabled = true
			break
		}
	}
}

// DisableRule disables a rule by ID
func (e *Engine) DisableRule(ruleID string) {
	for _, rule := range e.rules {
		if rule.ID == ruleID {
			rule.Enabled = false
			break
		}
	}
}
