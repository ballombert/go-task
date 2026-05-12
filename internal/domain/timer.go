package domain

import "time"

// TimerPreset defines a preset timer configuration
type TimerPreset struct {
	Name     string
	Duration int // Minutes
}

// DefaultPresets returns the default Pomodoro presets
func DefaultPresets() []TimerPreset {
	return []TimerPreset{
		{Name: "Pomodoro Standard", Duration: 25},
		{Name: "Pomodoro Court", Duration: 15},
		{Name: "Break Court", Duration: 5},
		{Name: "Break Long", Duration: 15},
	}
}

// Timer represents an active or historical timer session
type Timer struct {
	ID        string
	TaskID    string
	Preset    TimerPreset
	StartedAt time.Time
	EndedAt   *time.Time
	Status    TimerStatus
	Aborted   bool // true if user stopped before completion
}

// TimerStatus represents the state of a timer
type TimerStatus string

const (
	TimerStatusActive    TimerStatus = "active"
	TimerStatusCompleted TimerStatus = "completed"
	TimerStatusStopped   TimerStatus = "stopped"
)

// Elapsed returns the elapsed time in seconds
func (t *Timer) Elapsed() int64 {
	endTime := t.EndedAt
	if t.Status == TimerStatusActive {
		endTime = timePtr(time.Now())
	}
	if endTime == nil {
		return 0
	}
	return int64(endTime.Sub(t.StartedAt).Seconds())
}

// Remaining returns the remaining time in seconds
func (t *Timer) Remaining() int64 {
	totalSecs := int64(t.Preset.Duration * 60)
	elapsedSecs := t.Elapsed()
	remaining := totalSecs - elapsedSecs
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// Progress returns a value from 0.0 to 1.0
func (t *Timer) Progress() float64 {
	totalSecs := int64(t.Preset.Duration * 60)
	if totalSecs == 0 {
		return 0
	}
	return float64(t.Elapsed()) / float64(totalSecs)
}

// IsActive returns true if timer is currently running
func (t *Timer) IsActive() bool {
	return t.Status == TimerStatusActive
}

// Session represents a logged timer session for analytics
type Session struct {
	ID        string
	Timer     *Timer
	TaskID    string
	Notes     string
	CreatedAt time.Time
}

func timePtr(t time.Time) *time.Time {
	return &t
}
