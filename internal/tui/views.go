package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/beaallombert/gotask/internal/domain"
	"github.com/charmbracelet/lipgloss"
)

// Colors and styles
var (
	// Status colors
	pausedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	activeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42")) // Green
	completedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	strikethrough  = lipgloss.NewStyle().Strikethrough(true).Foreground(lipgloss.Color("240"))

	// Priority colors
	priorityHighestStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
	priorityHighStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Orange
	priorityMediumStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
	priorityLowStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("51"))  // Cyan
	priorityLowestStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("31"))  // Blue

	// UI styles
	titleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	helpStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	selectedStyle   = lipgloss.NewStyle().Background(lipgloss.Color("237")).Foreground(lipgloss.Color("226"))
	headerStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Underline(true)
	borderStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	modalStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("39")).Padding(0, 1)
	errorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	timerModalStyle = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("42")).Padding(1, 2)
	timerTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	bigTimerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226"))
)

func (m Model) renderTasksView() string {
	var output strings.Builder

	// Header
	output.WriteString(titleStyle.Render("📋 TASKS") + "\n")
	output.WriteString(borderStyle.Render(strings.Repeat("─", m.width)) + "\n\n")

	if len(m.tasks) == 0 {
		output.WriteString("No tasks in inbox.\n")
		output.WriteString(helpStyle.Render("Press 'n' to create a new task, 'q' to quit\n"))
		return output.String()
	}

	// Show tasks
	for i, task := range m.tasks {
		isSelected := i == m.selectedIdx && !m.selectingSubtask

		line := m.formatTaskLine(task, isSelected)
		output.WriteString(line + "\n")

		// Show subtasks if expanded
		if i == m.selectedIdx && len(task.Subtasks) > 0 {
			for j, subtask := range task.Subtasks {
				subSelected := m.selectingSubtask && j == m.selectedSubtaskIdx
				subLine := "  └─ " + m.formatTaskLine(subtask, subSelected)
				output.WriteString(subLine + "\n")
			}
		}
	}

	output.WriteString("\n" + borderStyle.Render(strings.Repeat("─", m.width)) + "\n")

	// Footer with instructions
	footer := ""
	if m.moveMode {
		footer = helpStyle.Render("MOVE MODE | j/k: move | Enter: confirm | Esc: cancel")
	} else if m.createMode || m.editMode {
		footer = helpStyle.Render("TASK MODAL | saisir texte | Enter: save | Esc: cancel")
	} else {
		footer = helpStyle.Render("j/k: navigate | Tab/Shift+Tab: subtask | Enter: edit | n: new | Shift+N: subtask | s: start pomodoro | p: break | w: work | Shift+J/K: move | q: quit")
	}
	output.WriteString(footer + "\n")

	return output.String()
}

func (m Model) renderTaskModal() string {
	title := "Edition tache"
	if m.createMode {
		title = "Nouvelle tache"
		if m.modalCreateParent != nil {
			title = "Nouvelle sous-tache"
		}
	}

	field := m.currentModalField()
	fieldHeader := fmt.Sprintf("Champ: %s (Tab/Shift+Tab)", field)

	descriptionVal := ""
	priorityVal := "medium"
	dueDateVal := "-"
	durationVal := "0"
	if m.modalDraft != nil {
		descriptionVal = strings.TrimSpace(m.modalDraft.Description)
		if descriptionVal == "" {
			descriptionVal = "-"
		}
		if m.modalDraft.Priority != "" {
			priorityVal = string(m.modalDraft.Priority)
		}
		if m.modalDraft.DueDate != nil {
			dueDateVal = m.modalDraft.DueDate.Format("2006-01-02")
		}
		if m.modalDraft.Duration > 0 {
			durationVal = fmt.Sprintf("%d", m.modalDraft.Duration)
		}
	}

	fieldLabel := func(name, value string) string {
		prefix := "  "
		if field == name {
			prefix = "> "
		}
		return fmt.Sprintf("%s%-12s %s", prefix, name+":", value)
	}

	fieldsOverview := strings.Join([]string{
		fieldLabel("description", descriptionVal),
		fieldLabel("priority", priorityVal),
		fieldLabel("due_date", dueDateVal),
		fieldLabel("duration", durationVal+" min"),
	}, "\n")

	content := title + "\n" +
		fieldsOverview + "\n\n" +
		headerStyle.Render(fieldHeader) + "\n" +
		m.taskInput.View() + "\n"

	if m.modalError != "" {
		content += errorStyle.Render(m.modalError) + "\n"
	}

	content += helpStyle.Render(modalFieldHint(field))
	return modalStyle.Render(content)
}

func (m Model) renderTaskOverlayModal() string {
	modal := m.renderTaskModal()
	if m.width <= 0 || m.height <= 0 {
		return modal
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

func (m Model) renderFocusView() string {
	var output strings.Builder

	output.WriteString(titleStyle.Render("🎯 FOCUS") + "\n")
	output.WriteString(borderStyle.Render(strings.Repeat("─", m.width)) + "\n\n")

	// Show active timer
	if activeTimer := m.timerManager.GetActive(); activeTimer != nil {
		remaining := activeTimer.Remaining()
		output.WriteString(m.formatActiveTimer(remaining))
	} else {
		output.WriteString("⏱ No active timer\n\n")
	}

	// Show top 3 tasks
	output.WriteString(headerStyle.Render("Focus candidate (j/k + Enter):") + "\n")
	if len(m.tasks) > 0 {
		for i := 0; i < 3 && i < len(m.tasks); i++ {
			prefix := "  "
			if i == m.selectedIdx {
				prefix = "> "
			}
			line := fmt.Sprintf("%s[%s] %s", prefix, m.tasks[i].Status, m.tasks[i].Description)
			output.WriteString(line + "\n")
		}

		if m.activeTask != nil {
			output.WriteString("\n" + activeStyle.Render("Selected focus: "+m.activeTask.Description) + "\n")
		}
	} else {
		output.WriteString("No tasks available.\n")
	}

	output.WriteString("\n" + borderStyle.Render(strings.Repeat("─", m.width)) + "\n")
	footer := helpStyle.Render("j/k: choose task | Enter: select | s: start pomodoro | p: break | w: work | t/f/l or ←/→ | q: quit")
	output.WriteString(footer + "\n")

	return output.String()
}

func (m Model) renderLogsView() string {
	var output strings.Builder

	output.WriteString(titleStyle.Render("📝 LOGS") + "\n")
	output.WriteString(borderStyle.Render(strings.Repeat("─", m.width)) + "\n\n")
	m.resizeLogsViewport()

	// Show recent sessions
	var logsContent strings.Builder
	if sessions, err := m.logger.GetRecentSessions(100); err == nil && len(sessions) > 0 {
		logsContent.WriteString(headerStyle.Render("Recent sessions:") + "\n")
		for _, session := range sessions {
			line := fmt.Sprintf("  • %s (%d min) - %s\n",
				session.Timer.Preset.Name,
				session.Timer.Preset.Duration,
				session.CreatedAt.Format("15:04"))
			logsContent.WriteString(line)
		}
	} else {
		logsContent.WriteString("No sessions logged yet.\n")
	}

	m.logsViewport.SetContent(logsContent.String())
	output.WriteString(m.logsViewport.View())

	output.WriteString("\n" + borderStyle.Render(strings.Repeat("─", m.width)) + "\n")
	footer := helpStyle.Render("j/k or ↑/↓: scroll | PgUp/PgDn: page | t/f/l or ←/→ | q: quit")
	output.WriteString(footer + "\n")

	return output.String()
}

func (m Model) formatTaskLine(task *domain.Task, selected bool) string {
	var line string
	status := "[ ]"
	if task.Status == domain.StatusInProgress {
		status = "[>]"
	} else if task.Status == domain.StatusCompleted {
		status = "[x]"
	}

	line = status + " " + task.Description

	if selected {
		return selectedStyle.Render(line)
	}
	return line
}

func (m Model) formatActiveTimer(remaining int64) string {
	// Format remaining time as MM:SS
	minutes := remaining / 60
	seconds := remaining % 60
	timerStr := fmt.Sprintf("%02d:%02d", minutes, seconds)
	return fmt.Sprintf("⏱ Active Timer: %s (running)\n  Remaining: %s\n\n", activeStyle.Render(timerStr), timerStr)
}

func (m Model) getTopNTasks(n int) []string {
	if len(m.tasks) == 0 || n <= 0 {
		return []string{}
	}

	candidates := make([]*domain.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		if task == nil || task.Status == domain.StatusCompleted {
			continue
		}
		candidates = append(candidates, task)
	}

	if len(candidates) == 0 {
		return []string{}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		a := candidates[i]
		b := candidates[j]

		if a.EffectivePriority() != b.EffectivePriority() {
			return a.EffectivePriority() < b.EffectivePriority()
		}

		if a.DueDate != nil && b.DueDate != nil && !a.DueDate.Equal(*b.DueDate) {
			return a.DueDate.Before(*b.DueDate)
		}
		if a.DueDate != nil && b.DueDate == nil {
			return true
		}
		if a.DueDate == nil && b.DueDate != nil {
			return false
		}

		return a.CreatedAt.Before(b.CreatedAt)
	})

	var result []string
	for i := 0; i < n && i < len(candidates); i++ {
		result = append(result, candidates[i].Description)
	}
	return result
}

func modalFieldHint(field string) string {
	switch field {
	case "description":
		return "Description de la tache. Enter: sauver, Esc: annuler"
	case "priority":
		return "Priorite: highest|high|medium|low|lowest"
	case "due_date":
		return "Echeance: YYYY-MM-DD (vide pour effacer)"
	case "duration":
		return "Duree en minutes (ex: 25)"
	default:
		return "Enter: sauver | Esc: annuler"
	}
}

func (m Model) renderTimerOverlayModal() string {
	activeTimer := m.timerManager.GetActive()
	if activeTimer == nil {
		return ""
	}

	remaining := activeTimer.Remaining()
	minutes := remaining / 60
	seconds := remaining % 60
	timeStr := fmt.Sprintf("%02d:%02d", minutes, seconds)

	progress := activeTimer.Progress()
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	progressBarWidth := 46
	if m.width > 70 {
		progressBarWidth = m.width / 2
		if progressBarWidth > 70 {
			progressBarWidth = 70
		}
	}
	if progressBarWidth < 20 {
		progressBarWidth = 20
	}

	m.progressBar.Width = progressBarWidth
	bar := m.progressBar.ViewAs(progress)
	percent := fmt.Sprintf("%3d%%", int(progress*100))

	taskLabel := "No selected task"
	if m.activeTask != nil && m.activeTask.Description != "" {
		taskLabel = m.activeTask.Description
	}

	bigTime := bigTimerStyle.Render(renderBigTimer(timeStr))
	content := strings.Join([]string{
		timerTitleStyle.Render("POMODORO IN PROGRESS"),
		"",
		bigTime,
		"",
		"Task: " + taskLabel,
		"Preset: " + activeTimer.Preset.Name,
		"",
		bar,
		"Progress: " + percent,
		"",
		helpStyle.Render("Esc: stop timer | p: break | w: work | q: quit"),
	}, "\n")

	modal := timerModalStyle.Render(content)
	if m.width <= 0 || m.height <= 0 {
		return modal
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

func renderBigTimer(value string) string {
	glyphs := map[rune][5]string{
		'0': {" ### ", "#   #", "#   #", "#   #", " ### "},
		'1': {"  #  ", " ##  ", "  #  ", "  #  ", " ### "},
		'2': {" ### ", "#   #", "   # ", "  #  ", "#####"},
		'3': {"#### ", "    #", " ### ", "    #", "#### "},
		'4': {"#   #", "#   #", "#####", "    #", "    #"},
		'5': {"#####", "#    ", "#### ", "    #", "#### "},
		'6': {" ### ", "#    ", "#### ", "#   #", " ### "},
		'7': {"#####", "   # ", "  #  ", " #   ", "#    "},
		'8': {" ### ", "#   #", " ### ", "#   #", " ### "},
		'9': {" ### ", "#   #", " ####", "    #", " ### "},
		':': {"     ", "  #  ", "     ", "  #  ", "     "},
	}

	lines := make([]string, 5)
	for _, ch := range value {
		glyph, ok := glyphs[ch]
		if !ok {
			glyph = [5]string{"     ", "     ", "     ", "     ", "     "}
		}
		for i := 0; i < 5; i++ {
			if lines[i] != "" {
				lines[i] += "  "
			}
			lines[i] += glyph[i]
		}
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderPomodoroSelectOverlayModal() string {
	modal := m.renderPomodoroSelectModal()
	if m.width <= 0 || m.height <= 0 {
		return modal
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

func (m Model) renderPomodoroSelectModal() string {
	if len(m.pomodoroTypes) == 0 {
		return modalStyle.Render("Aucun type de pomodoro configure")
	}

	var rows []string
	for i, p := range m.pomodoroTypes {
		prefix := "  "
		if i == m.pomodoroSelectIdx {
			prefix = "> "
		}
		row := fmt.Sprintf("%s%s  (work:%d break:%d long:%d cycles:%d)", prefix, p.Name, p.WorkDuration, p.BreakDuration, p.LongBreakDuration, p.CyclesBeforeLongBreak)
		if i == m.pomodoroSelectIdx {
			row = selectedStyle.Render(row)
		}
		rows = append(rows, row)
	}

	content := strings.Join([]string{
		headerStyle.Render("Choisir un type de Pomodoro"),
		"",
		strings.Join(rows, "\n"),
		"",
		helpStyle.Render("j/k ou haut/bas: naviguer | Enter: valider | Esc: annuler"),
	}, "\n")

	return modalStyle.Render(content)
}
