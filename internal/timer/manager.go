package timer

import (
	"sync"
	"time"

	"github.com/beaallombert/gotask/internal/domain"
)

// Manager manages the active timer and timer sessions
type Manager struct {
	activeTimer *domain.Timer
	mu          sync.RWMutex
}

// NewManager creates a new timer manager
func NewManager() *Manager {
	return &Manager{
		activeTimer: nil,
	}
}

// Start starts a new timer
func (m *Manager) Start(taskID string, preset domain.TimerPreset) *domain.Timer {
	m.mu.Lock()
	defer m.mu.Unlock()

	timer := &domain.Timer{
		ID:        generateTimerID(),
		TaskID:    taskID,
		Preset:    preset,
		StartedAt: time.Now(),
		Status:    domain.TimerStatusActive,
	}

	m.activeTimer = timer
	return timer
}

// Stop stops the active timer
func (m *Manager) Stop(aborted bool) *domain.Timer {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.activeTimer == nil {
		return nil
	}

	now := time.Now()
	m.activeTimer.EndedAt = &now
	m.activeTimer.Aborted = aborted

	if !aborted {
		m.activeTimer.Status = domain.TimerStatusCompleted
	} else {
		m.activeTimer.Status = domain.TimerStatusStopped
	}

	timer := m.activeTimer
	m.activeTimer = nil

	return timer
}

// GetActive returns the currently active timer
func (m *Manager) GetActive() *domain.Timer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.activeTimer
}

// HasActive returns true if there is an active timer
func (m *Manager) HasActive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.activeTimer != nil
}

// GetProgress returns the progress of the active timer (0.0 to 1.0)
func (m *Manager) GetProgress() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeTimer == nil {
		return 0
	}

	return m.activeTimer.Progress()
}

// GetRemainingSeconds returns remaining seconds for active timer
func (m *Manager) GetRemainingSeconds() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeTimer == nil {
		return 0
	}

	return m.activeTimer.Remaining()
}

// RestoreTimer restores a timer from disk
func (m *Manager) RestoreTimer(timer *domain.Timer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.activeTimer = timer
}

func generateTimerID() string {
	return "timer-" + time.Now().Format("20060102150405")
}
