package tui

import (
	"github.com/charmbracelet/bubbletea"
)

// Start launches the interactive TUI
func Start(inboxPath string, dbPath string) error {
	model, err := NewModel(inboxPath, dbPath)
	if err != nil {
		return err
	}

	p := tea.NewProgram(model, tea.WithAltScreen())

	_, err = p.Run()
	model.Close()

	return err
}

// Close closes TUI resources
func (m *Model) Close() error {
	if m.logger != nil {
		return m.logger.Close()
	}
	return nil
}
