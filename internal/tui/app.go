package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/beaallombert/gotask/internal/domain"
	"github.com/beaallombert/gotask/internal/rules"
	"github.com/beaallombert/gotask/internal/storage"
	"github.com/beaallombert/gotask/internal/timer"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// View represents different screens in the TUI
type View string

type tickMsg time.Time

const (
	ViewTasks View = "tasks"
	ViewFocus View = "focus"
	ViewLogs  View = "logs"
)

// Model represents the application state
type Model struct {
	view               View
	tasks              []*domain.Task
	timerManager       *timer.Manager
	rulesEngine        *rules.Engine
	inboxReader        *storage.InboxReader
	inboxWriter        *storage.InboxWriter
	logger             *storage.SQLiteLogger
	activeTask         *domain.Task
	selectedIdx        int
	selectingSubtask   bool
	selectedSubtaskIdx int
	width              int
	height             int
	editMode           bool
	moveMode           bool
	moveOriginIdx      int
	editBackup         *domain.Task
	createMode         bool
	modalInput         string
	modalError         string
	modalEditTarget    *domain.Task
	modalCreateParent  *domain.Task
	taskInput          textinput.Model
	progressBar        progress.Model
	logsViewport       viewport.Model
	logsViewportReady  bool
	logsOffset         int
	logs               string
	err                error
}

// NewModel creates a new TUI model
func NewModel(inboxPath string, dbPath string) (*Model, error) {
	reader := storage.NewInboxReader(inboxPath)
	writer := storage.NewInboxWriter(inboxPath)

	tasks, err := reader.ReadTasks()
	if err != nil {
		return nil, err
	}

	logger, err := storage.NewSQLiteLogger(dbPath)
	if err != nil {
		return nil, err
	}

	input := textinput.New()
	input.Placeholder = "Description..."
	input.CharLimit = 256
	input.Width = 48

	bar := progress.New(progress.WithDefaultGradient())
	vp := viewport.New(80, 12)

	return &Model{
		view:               ViewTasks,
		tasks:              tasks,
		timerManager:       timer.NewManager(),
		rulesEngine:        rules.NewEngine(),
		inboxReader:        reader,
		inboxWriter:        writer,
		logger:             logger,
		selectedIdx:        0,
		selectingSubtask:   false,
		selectedSubtaskIdx: -1,
		editMode:           false,
		moveMode:           false,
		moveOriginIdx:      -1,
		createMode:         false,
		modalInput:         "",
		modalError:         "",
		modalEditTarget:    nil,
		modalCreateParent:  nil,
		taskInput:          input,
		progressBar:        bar,
		logsViewport:       vp,
		logsViewportReady:  false,
		logsOffset:         0,
	}, nil
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeLogsViewport()
		return m, nil
	case tickMsg:
		return m, tickCmd()
	}

	return m, nil
}

func (m *Model) resizeLogsViewport() {
	if m.width <= 0 || m.height <= 0 {
		return
	}

	vpWidth := m.width - 4
	if vpWidth < 20 {
		vpWidth = 20
	}

	vpHeight := m.height - 8
	if vpHeight < 5 {
		vpHeight = 5
	}

	m.logsViewport.Width = vpWidth
	m.logsViewport.Height = vpHeight
	m.logsViewportReady = true
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// View renders the TUI
func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error()
	}

	var content string
	switch m.view {
	case ViewTasks:
		content = m.renderTasksView()
	case ViewFocus:
		content = m.renderFocusView()
	case ViewLogs:
		content = m.renderLogsView()
	default:
		content = "Unknown view"
	}

	if m.timerManager.GetActive() != nil {
		return m.renderTimerOverlayModal()
	}

	if m.createMode || m.editMode {
		return m.renderTaskOverlayModal()
	}

	return content
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		if stopped := m.timerManager.Stop(true); stopped != nil {
			taskID := stopped.TaskID
			_ = m.logger.LogAction("timer_stopped", &taskID, "stopped_with_escape")
			return m, nil
		}
	}

	if m.view == ViewTasks && (m.createMode || m.editMode) {
		return m.handleTasksKey(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "t", "T":
		m.view = ViewTasks
		return m, nil
	case "f", "F":
		m.view = ViewFocus
		if len(m.tasks) > 0 && m.selectedIdx >= 0 && m.selectedIdx < len(m.tasks) {
			m.activeTask = m.tasks[m.selectedIdx]
		}
		return m, nil
	case "l", "L":
		m.view = ViewLogs
		return m, nil
	case "left":
		m.view = previousView(m.view)
		return m, nil
	case "right":
		m.view = nextView(m.view)
		return m, nil
	}

	// View-specific key handling
	switch m.view {
	case ViewTasks:
		return m.handleTasksKey(msg)
	case ViewFocus:
		return m.handleFocusKey(msg)
	case ViewLogs:
		return m.handleLogsKey(msg)
	}

	return m, nil
}

func (m Model) handleTasksKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.createMode || m.editMode {
		switch msg.String() {
		case "esc":
			m.closeTaskModal()
			return m, nil
		case "enter":
			description := strings.TrimSpace(m.taskInput.Value())
			if description == "" {
				m.modalError = "Description obligatoire"
				return m, nil
			}

			if m.createMode {
				if m.modalCreateParent != nil {
					m.createSubtask(m.modalCreateParent, description)
				} else {
					m.createTask(description)
				}
			} else if m.editMode && m.modalEditTarget != nil {
				m.modalEditTarget.Description = description
			}

			m.closeTaskModal()
			if err := m.persistTasks(); err != nil {
				m.err = err
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.taskInput, cmd = m.taskInput.Update(msg)
			m.modalError = ""
			return m, cmd
		}
	}

	switch msg.String() {
	case "n":
		m.openCreateModal()
		return m, nil
	case "N":
		if len(m.tasks) == 0 || m.selectedIdx < 0 || m.selectedIdx >= len(m.tasks) {
			return m, nil
		}
		m.openCreateSubtaskModal(m.tasks[m.selectedIdx])
		return m, nil
	}

	if len(m.tasks) == 0 {
		return m, nil
	}

	if m.moveMode {
		switch msg.String() {
		case "j", "down", "J":
			m.moveSelectedTask(1)
			return m, nil
		case "k", "up", "K":
			m.moveSelectedTask(-1)
			return m, nil
		case "enter":
			m.moveMode = false
			m.moveOriginIdx = -1
			if err := m.persistTasks(); err != nil {
				m.err = err
			}
			return m, nil
		case "esc":
			m.restoreMovePosition()
			m.moveMode = false
			m.moveOriginIdx = -1
			return m, nil
		}
		return m, nil
	}

	switch msg.String() {
	case "j", "down":
		if m.selectingSubtask {
			subs := m.tasks[m.selectedIdx].Subtasks
			if len(subs) > 0 && m.selectedSubtaskIdx < len(subs)-1 {
				m.selectedSubtaskIdx++
			}
			return m, nil
		}
		if m.selectedIdx < len(m.tasks)-1 {
			m.selectedIdx++
			m.selectingSubtask = false
			m.selectedSubtaskIdx = -1
		}
		return m, nil
	case "k", "up":
		if m.selectingSubtask {
			if m.selectedSubtaskIdx > 0 {
				m.selectedSubtaskIdx--
			}
			return m, nil
		}
		if m.selectedIdx > 0 {
			m.selectedIdx--
			m.selectingSubtask = false
			m.selectedSubtaskIdx = -1
		}
		return m, nil
	case "tab":
		subs := m.tasks[m.selectedIdx].Subtasks
		if len(subs) == 0 {
			return m, nil
		}
		if !m.selectingSubtask {
			m.selectingSubtask = true
			m.selectedSubtaskIdx = 0
		} else if m.selectedSubtaskIdx < len(subs)-1 {
			m.selectedSubtaskIdx++
		} else {
			m.selectingSubtask = false
			m.selectedSubtaskIdx = -1
		}
		return m, nil
	case "shift+tab":
		if !m.selectingSubtask {
			return m, nil
		}
		if m.selectedSubtaskIdx > 0 {
			m.selectedSubtaskIdx--
		} else {
			m.selectingSubtask = false
			m.selectedSubtaskIdx = -1
		}
		return m, nil
	case "J":
		if m.selectingSubtask {
			return m, nil
		}
		m.moveMode = true
		m.moveOriginIdx = m.selectedIdx
		m.moveSelectedTask(1)
		return m, nil
	case "K":
		if m.selectingSubtask {
			return m, nil
		}
		m.moveMode = true
		m.moveOriginIdx = m.selectedIdx
		m.moveSelectedTask(-1)
		return m, nil
	case "enter":
		target := m.selectedTaskTarget()
		if target == nil {
			return m, nil
		}
		m.openEditModal(target)
		return m, nil
	case "1", "2", "&", "é", "a", "z":
		target := m.selectedTaskTarget()
		if target == nil {
			return m, nil
		}
		preset, ok := presetForFocusKey(msg.String())
		if !ok {
			return m, nil
		}
		m.timerManager.Start(target.ID, preset)
		m.activeTask = target
		m.view = ViewFocus
		taskID := target.ID
		_ = m.logger.LogAction("timer_started", &taskID, preset.Name)
		return m, nil
	}

	return m, nil
}

func (m *Model) openCreateModal() {
	m.createMode = true
	m.editMode = false
	m.modalEditTarget = nil
	m.modalCreateParent = nil
	m.modalInput = ""
	m.modalError = ""
	m.taskInput.SetValue("")
	m.taskInput.Focus()
}

func (m *Model) openCreateSubtaskModal(parent *domain.Task) {
	m.createMode = true
	m.editMode = false
	m.modalEditTarget = nil
	m.modalCreateParent = parent
	m.modalInput = ""
	m.modalError = ""
	m.taskInput.SetValue("")
	m.taskInput.Focus()
}

func (m *Model) openEditModal(target *domain.Task) {
	m.createMode = false
	m.editMode = true
	m.modalEditTarget = target
	m.modalCreateParent = nil
	m.modalInput = target.Description
	m.modalError = ""
	m.taskInput.SetValue(target.Description)
	m.taskInput.Focus()
}

func (m *Model) closeTaskModal() {
	m.createMode = false
	m.editMode = false
	m.modalInput = ""
	m.modalError = ""
	m.modalEditTarget = nil
	m.modalCreateParent = nil
	m.taskInput.Blur()
}

func (m *Model) selectedTaskTarget() *domain.Task {
	if len(m.tasks) == 0 || m.selectedIdx < 0 || m.selectedIdx >= len(m.tasks) {
		return nil
	}

	t := m.tasks[m.selectedIdx]
	if !m.selectingSubtask {
		return t
	}

	if m.selectedSubtaskIdx < 0 || m.selectedSubtaskIdx >= len(t.Subtasks) {
		return t
	}

	return t.Subtasks[m.selectedSubtaskIdx]
}

func (m *Model) persistTasks() error {
	return m.inboxWriter.WriteTasks(m.tasks)
}

func (m *Model) createTask(description string) {
	newTask := &domain.Task{
		ID:          fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Description: description,
		Status:      domain.StatusPaused,
		Priority:    domain.PriorityMedium,
		CreatedAt:   time.Now(),
	}
	m.tasks = append(m.tasks, newTask)
	m.selectedIdx = len(m.tasks) - 1
	m.selectingSubtask = false
	m.selectedSubtaskIdx = -1
}

func (m *Model) createSubtask(parent *domain.Task, description string) {
	if parent == nil {
		return
	}

	newSubtask := &domain.Task{
		ID:          fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Description: description,
		Status:      domain.StatusPaused,
		Priority:    domain.PriorityMedium,
		CreatedAt:   time.Now(),
		ParentID:    parent.ID,
	}

	parent.Subtasks = append(parent.Subtasks, newSubtask)
	m.selectingSubtask = true
	m.selectedSubtaskIdx = len(parent.Subtasks) - 1
}

func (m *Model) moveSelectedTask(delta int) bool {
	newIdx := m.selectedIdx + delta
	if newIdx < 0 || newIdx >= len(m.tasks) {
		return false
	}

	m.tasks[m.selectedIdx], m.tasks[newIdx] = m.tasks[newIdx], m.tasks[m.selectedIdx]
	m.selectedIdx = newIdx
	return true
}

func (m *Model) restoreMovePosition() {
	if m.moveOriginIdx < 0 || m.moveOriginIdx >= len(m.tasks) || m.selectedIdx == m.moveOriginIdx {
		return
	}

	task := m.tasks[m.selectedIdx]
	m.tasks = append(m.tasks[:m.selectedIdx], m.tasks[m.selectedIdx+1:]...)

	if m.moveOriginIdx >= len(m.tasks) {
		m.tasks = append(m.tasks, task)
		m.selectedIdx = len(m.tasks) - 1
		return
	}

	m.tasks = append(m.tasks[:m.moveOriginIdx], append([]*domain.Task{task}, m.tasks[m.moveOriginIdx:]...)...)
	m.selectedIdx = m.moveOriginIdx
}

func cloneTaskForEdit(task *domain.Task) *domain.Task {
	copy := *task
	if task.DueDate != nil {
		d := *task.DueDate
		copy.DueDate = &d
	}
	if task.CompletedAt != nil {
		c := *task.CompletedAt
		copy.CompletedAt = &c
	}
	return &copy
}

func (m *Model) restoreEditBackup() {
	if m.editBackup == nil || m.selectedIdx < 0 || m.selectedIdx >= len(m.tasks) {
		return
	}
	backup := cloneTaskForEdit(m.editBackup)
	m.tasks[m.selectedIdx] = backup
}

func (m *Model) cycleSelectedTaskStatus(direction int) {
	task := m.tasks[m.selectedIdx]
	statuses := []domain.Status{domain.StatusPaused, domain.StatusInProgress, domain.StatusCompleted}
	idx := 0
	for i, s := range statuses {
		if task.Status == s {
			idx = i
			break
		}
	}

	if direction > 0 {
		idx = (idx + 1) % len(statuses)
	} else {
		idx = (idx - 1 + len(statuses)) % len(statuses)
	}

	task.Status = statuses[idx]
	if task.Status == domain.StatusCompleted {
		now := time.Now()
		task.CompletedAt = &now
	} else {
		task.CompletedAt = nil
	}
}

func (m *Model) toggleSelectedTaskStatus() {
	task := m.tasks[m.selectedIdx]
	now := time.Now()

	switch task.Status {
	case domain.StatusPaused:
		task.Status = domain.StatusInProgress
		task.CompletedAt = nil
	case domain.StatusInProgress:
		task.Status = domain.StatusCompleted
		task.CompletedAt = &now
	case domain.StatusCompleted:
		task.Status = domain.StatusPaused
		task.CompletedAt = nil
	default:
		task.Status = domain.StatusPaused
		task.CompletedAt = nil
	}
}

func (m Model) handleFocusKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.tasks) > 0 && (m.activeTask == nil || m.activeTask.ID == "") {
		if m.selectedIdx >= 0 && m.selectedIdx < len(m.tasks) {
			m.activeTask = m.tasks[m.selectedIdx]
		}
	}

	switch msg.String() {
	case "j", "down":
		if len(m.tasks) > 0 && m.selectedIdx < len(m.tasks)-1 {
			m.selectedIdx++
			m.activeTask = m.tasks[m.selectedIdx]
		}
		return m, nil
	case "k", "up":
		if len(m.tasks) > 0 && m.selectedIdx > 0 {
			m.selectedIdx--
			m.activeTask = m.tasks[m.selectedIdx]
		}
		return m, nil
	case "enter":
		if len(m.tasks) > 0 && m.selectedIdx >= 0 && m.selectedIdx < len(m.tasks) {
			m.activeTask = m.tasks[m.selectedIdx]
		}
		return m, nil
	case "1", "2", "&", "é", "a", "z":
		if len(m.tasks) == 0 {
			return m, nil
		}

		if m.activeTask == nil {
			if m.selectedIdx < 0 || m.selectedIdx >= len(m.tasks) {
				m.selectedIdx = 0
			}
			m.activeTask = m.tasks[m.selectedIdx]
		}

		preset, ok := presetForFocusKey(msg.String())
		if !ok {
			return m, nil
		}

		m.timerManager.Start(m.activeTask.ID, preset)
		m.view = ViewFocus
		taskID := m.activeTask.ID
		_ = m.logger.LogAction("timer_started", &taskID, preset.Name)
		return m, nil
	}

	return m, nil
}

func (m Model) handleLogsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.logsViewport.LineDown(1)
		return m, nil
	case "k", "up":
		m.logsViewport.LineUp(1)
		return m, nil
	case "pgdown":
		m.logsViewport.ViewDown()
		return m, nil
	case "pgup":
		m.logsViewport.ViewUp()
		return m, nil
	}

	return m, nil
}

func nextView(v View) View {
	switch v {
	case ViewTasks:
		return ViewFocus
	case ViewFocus:
		return ViewLogs
	default:
		return ViewTasks
	}
}

func previousView(v View) View {
	switch v {
	case ViewTasks:
		return ViewLogs
	case ViewFocus:
		return ViewTasks
	default:
		return ViewFocus
	}
}

func presetForFocusKey(key string) (domain.TimerPreset, bool) {
	presets := domain.DefaultPresets()
	switch key {
	case "1", "&", "a":
		if len(presets) > 0 {
			return presets[0], true
		}
	case "2", "é", "z":
		if len(presets) > 1 {
			return presets[1], true
		}
	}

	return domain.TimerPreset{}, false
}
