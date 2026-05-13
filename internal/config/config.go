package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// PomodoroType describes one pomodoro profile loaded from config.yml.
type PomodoroType struct {
	Name                  string
	WorkDuration          int
	BreakDuration         int
	LongBreakDuration     int
	CyclesBeforeLongBreak int
}

// AppConfig is the normalized application configuration.
type AppConfig struct {
	InboxPath     string
	PomodoroTypes []PomodoroType
}

type rawPomodoroDurations struct {
	WorkDuration          int `yaml:"work_duration"`
	BreakDuration         int `yaml:"break_duration"`
	LongBreakDuration     int `yaml:"long_break_duration"`
	CyclesBeforeLongBreak int `yaml:"cycles_before_long_break"`
}

type rawConfig struct {
	PomodoroType []map[string]rawPomodoroDurations `yaml:"pomodoro_type"`
	InboxPath    string                             `yaml:"inbox_path"`
}

// Default returns a sane default configuration.
func Default() AppConfig {
	return AppConfig{
		InboxPath: "inbox.md",
		PomodoroTypes: []PomodoroType{
			{Name: "classique", WorkDuration: 25, BreakDuration: 5, LongBreakDuration: 15, CyclesBeforeLongBreak: 4},
			{Name: "court", WorkDuration: 15, BreakDuration: 3, LongBreakDuration: 10, CyclesBeforeLongBreak: 4},
			{Name: "long", WorkDuration: 50, BreakDuration: 10, LongBreakDuration: 20, CyclesBeforeLongBreak: 2},
		},
	}
}

// LoadFromFile loads config from YAML file. If file does not exist, defaults are returned.
func LoadFromFile(path string) (AppConfig, error) {
	cfg := Default()

	rawBytes, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return AppConfig{}, err
	}

	var raw rawConfig
	if err := yaml.Unmarshal(rawBytes, &raw); err != nil {
		return AppConfig{}, err
	}

	if raw.InboxPath != "" {
		cfg.InboxPath = raw.InboxPath
	}

	if len(raw.PomodoroType) > 0 {
		types := make([]PomodoroType, 0, len(raw.PomodoroType))
		for _, entry := range raw.PomodoroType {
			for name, durations := range entry {
				t := PomodoroType{
					Name:                  name,
					WorkDuration:          clampMinutes(durations.WorkDuration, 25),
					BreakDuration:         clampMinutes(durations.BreakDuration, 5),
					LongBreakDuration:     clampMinutes(durations.LongBreakDuration, 15),
					CyclesBeforeLongBreak: clampCycles(durations.CyclesBeforeLongBreak, 4),
				}
				types = append(types, t)
			}
		}
		if len(types) > 0 {
			cfg.PomodoroTypes = types
		}
	}

	return cfg, nil
}

func clampMinutes(v int, fallback int) int {
	if v <= 0 {
		return fallback
	}
	return v
}

func clampCycles(v int, fallback int) int {
	if v <= 0 {
		return fallback
	}
	return v
}
