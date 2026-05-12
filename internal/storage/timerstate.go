package storage

import (
	"encoding/json"
	"os"
	"time"

	"github.com/beaallombert/gotask/internal/domain"
)

// TimerState represents the persisted timer state
type TimerState struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	PresetName string   `json:"preset_name"`
	Duration  int       `json:"duration"`
	StartedAt time.Time `json:"started_at"`
	Status    string    `json:"status"`
}

// TimerStateStore manages persistent timer state
type TimerStateStore struct {
	filePath string
}

// NewTimerStateStore creates a new timer state store
func NewTimerStateStore(filePath string) *TimerStateStore {
	return &TimerStateStore{filePath: filePath}
}

// Save saves the current timer state
func (s *TimerStateStore) Save(timer *domain.Timer) error {
	if timer == nil {
		// Delete file if timer is nil
		os.Remove(s.filePath)
		return nil
	}

	state := TimerState{
		ID:         timer.ID,
		TaskID:     timer.TaskID,
		PresetName: timer.Preset.Name,
		Duration:   timer.Preset.Duration,
		StartedAt:  timer.StartedAt,
		Status:     string(timer.Status),
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0600)
}

// Load loads the timer state from disk
func (s *TimerStateStore) Load() (*domain.Timer, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No saved timer
		}
		return nil, err
	}

	var state TimerState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	// Reconstruct timer
	timer := &domain.Timer{
		ID:        state.ID,
		TaskID:    state.TaskID,
		Preset:    domain.TimerPreset{Name: state.PresetName, Duration: state.Duration},
		StartedAt: state.StartedAt,
		Status:    domain.TimerStatus(state.Status),
	}

	return timer, nil
}

// Clear clears the timer state
func (s *TimerStateStore) Clear() error {
	return os.Remove(s.filePath)
}
