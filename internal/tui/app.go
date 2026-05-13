package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/beaallombert/gotask/internal/config"
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

var modalFields = []string{"description", "priority", "due_date", "duration"}

const (
	ViewTasks View = "tasks"
	ViewFocus View = "focus"
	ViewLogs  View = "logs"
)

const (
	pomodoroPhaseWork      = "work"
	pomodoroPhaseBreak     = "break"
	pomodoroPhaseLongBreak = "long_break"
)

type pomodoroSession struct {
	profile     config.PomodoroType
	phase       string
	workCycles  int
	selectedFor *domain.Task
}

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
	modalDraft         *domain.Task
	modalFieldIdx      int
	taskInput          textinput.Model
	progressBar        progress.Model
	logsViewport       viewport.Model
	logsViewportReady  bool
	logsOffset         int
	logs               string
	pomodoroTypes      []config.PomodoroType
	pomodoroModalOpen  bool
	pomodoroSelectIdx  int
	pomodoroSession    *pomodoroSession
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
	cfg, _ := config.LoadFromFile("config.yml")

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
		modalDraft:         nil,
		modalFieldIdx:      0,
		taskInput:          input,
		progressBar:        bar,
		logsViewport:       vp,
		logsViewportReady:  false,
		logsOffset:         0,
		pomodoroTypes:      cfg.PomodoroTypes,
		pomodoroModalOpen:  false,
		pomodoroSelectIdx:  0,
		pomodoroSession:    nil,
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
		m.handlePomodoroTick()
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

	if m.pomodoroModalOpen {
		return m.renderPomodoroSelectOverlayModal()
	}

	if m.createMode || m.editMode {
		return m.renderTaskOverlayModal()
	}

	return content
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.pomodoroModalOpen {
		return m.handlePomodoroModalKey(msg)
	}

	if msg.String() == "esc" {
		if stopped := m.timerManager.Stop(true); stopped != nil {
			taskID := stopped.TaskID
			_ = m.logger.LogAction("timer_stopped", &taskID, "stopped_with_escape")
			m.pomodoroSession = nil
			return m, nil
		}
	}

	if m.view == ViewTasks && (m.createMode || m.editMode) {
		return m.handleTasksKey(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "s", "S":
		m.openPomodoroModalForCurrentTask()
		return m, nil
	case "p", "P":
		m.switchPomodoroToBreak()
		return m, nil
	case "w", "W":
		m.switchPomodoroToWork()
		return m, nil
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
		case "tab", "shift+tab":
			if err := m.saveModalInputToDraft(); err != nil {
				m.modalError = err.Error()
				return m, nil
			}
			if msg.String() == "tab" {
				m.modalFieldIdx = (m.modalFieldIdx + 1) % len(modalFields)
			} else {
				m.modalFieldIdx = (m.modalFieldIdx - 1 + len(modalFields)) % len(modalFields)
			}
			m.loadModalInputFromDraft()
			m.modalError = ""
			return m, nil
		case "enter":
			if err := m.saveModalInputToDraft(); err != nil {
				m.modalError = err.Error()
				return m, nil
			}

			if m.modalDraft == nil || strings.TrimSpace(m.modalDraft.Description) == "" {
				m.modalError = "Description obligatoire"
				return m, nil
			}

			if m.createMode {
				if m.modalCreateParent != nil {
					m.createSubtaskFromDraft(m.modalCreateParent)
				} else {
					m.createTaskFromDraft()
				}
			} else if m.editMode && m.modalEditTarget != nil {
				m.applyDraftToTask(m.modalEditTarget)
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
	m.modalFieldIdx = 0
	m.modalDraft = &domain.Task{
		Status:    domain.StatusPaused,
		Priority:  domain.PriorityMedium,
		CreatedAt: time.Now(),
	}
	m.loadModalInputFromDraft()
	m.taskInput.Focus()
}

func (m *Model) openCreateSubtaskModal(parent *domain.Task) {
	m.createMode = true
	m.editMode = false
	m.modalEditTarget = nil
	m.modalCreateParent = parent
	m.modalInput = ""
	m.modalError = ""
	m.modalFieldIdx = 0
	m.modalDraft = &domain.Task{
		Status:    domain.StatusPaused,
		Priority:  domain.PriorityMedium,
		CreatedAt: time.Now(),
	}
	m.loadModalInputFromDraft()
	m.taskInput.Focus()
}

func (m *Model) openEditModal(target *domain.Task) {
	m.createMode = false
	m.editMode = true
	m.modalEditTarget = target
	m.modalCreateParent = nil
	m.modalInput = target.Description
	m.modalError = ""
	m.modalFieldIdx = 0
	m.modalDraft = cloneTaskForEdit(target)
	m.loadModalInputFromDraft()
	m.taskInput.Focus()
}

func (m *Model) closeTaskModal() {
	m.createMode = false
	m.editMode = false
	m.modalInput = ""
	m.modalError = ""
	m.modalEditTarget = nil
	m.modalCreateParent = nil
	m.modalDraft = nil
	m.modalFieldIdx = 0
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

func (m *Model) createTaskFromDraft() {
	if m.modalDraft == nil {
		return
	}

	newTask := &domain.Task{
		ID:          fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Description: strings.TrimSpace(m.modalDraft.Description),
		Status:      domain.StatusPaused,
		Priority:    m.modalDraft.Priority,
		DueDate:     m.modalDraft.DueDate,
		Duration:    m.modalDraft.Duration,
		CreatedAt:   time.Now(),
	}
	if newTask.Priority == "" {
		newTask.Priority = domain.PriorityMedium
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

func (m *Model) createSubtaskFromDraft(parent *domain.Task) {
	if parent == nil || m.modalDraft == nil {
		return
	}

	newSubtask := &domain.Task{
		ID:          fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Description: strings.TrimSpace(m.modalDraft.Description),
		Status:      domain.StatusPaused,
		Priority:    m.modalDraft.Priority,
		DueDate:     m.modalDraft.DueDate,
		Duration:    m.modalDraft.Duration,
		CreatedAt:   time.Now(),
		ParentID:    parent.ID,
	}
	if newSubtask.Priority == "" {
		newSubtask.Priority = domain.PriorityMedium
	}

	parent.Subtasks = append(parent.Subtasks, newSubtask)
	m.selectingSubtask = true
	m.selectedSubtaskIdx = len(parent.Subtasks) - 1
}

func (m *Model) applyDraftToTask(target *domain.Task) {
	if target == nil || m.modalDraft == nil {
		return
	}
	target.Description = strings.TrimSpace(m.modalDraft.Description)
	target.Priority = m.modalDraft.Priority
	target.DueDate = m.modalDraft.DueDate
	target.Duration = m.modalDraft.Duration
}

func (m *Model) currentModalField() string {
	if m.modalFieldIdx < 0 || m.modalFieldIdx >= len(modalFields) {
		return "description"
	}
	return modalFields[m.modalFieldIdx]
}

func (m *Model) loadModalInputFromDraft() {
	if m.modalDraft == nil {
		m.taskInput.SetValue("")
		return
	}

	switch m.currentModalField() {
	case "description":
		m.taskInput.SetValue(m.modalDraft.Description)
		m.taskInput.Placeholder = "Description..."
	case "priority":
		m.taskInput.SetValue(string(m.modalDraft.Priority))
		m.taskInput.Placeholder = "highest|high|medium|low|lowest"
	case "due_date":
		if m.modalDraft.DueDate != nil {
			m.taskInput.SetValue(m.modalDraft.DueDate.Format("2006-01-02"))
		} else {
			m.taskInput.SetValue("")
		}
		m.taskInput.Placeholder = "YYYY-MM-DD (empty to clear)"
	case "duration":
		if m.modalDraft.Duration > 0 {
			m.taskInput.SetValue(strconv.Itoa(m.modalDraft.Duration))
		} else {
			m.taskInput.SetValue("")
		}
		m.taskInput.Placeholder = "minutes (ex: 25)"
	}
}

func (m *Model) saveModalInputToDraft() error {
	if m.modalDraft == nil {
		return nil
	}

	value := strings.TrimSpace(m.taskInput.Value())
	field := m.currentModalField()

	switch field {
	case "description":
		m.modalDraft.Description = value
		if value == "" {
			return fmt.Errorf("description obligatoire")
		}
	case "priority":
		priority, err := parsePriorityInput(value)
		if err != nil {
			return err
		}
		m.modalDraft.Priority = priority
	case "due_date":
		if value == "" {
			m.modalDraft.DueDate = nil
			return nil
		}
		d, err := time.Parse("2006-01-02", value)
		if err != nil {
			return fmt.Errorf("date invalide: YYYY-MM-DD")
		}
		m.modalDraft.DueDate = &d
	case "duration":
		if value == "" {
			m.modalDraft.Duration = 0
			return nil
		}
		dur, err := strconv.Atoi(value)
		if err != nil || dur < 0 {
			return fmt.Errorf("duree invalide: minutes >= 0")
		}
		m.modalDraft.Duration = dur
	}

	return nil
}

func parsePriorityInput(value string) (domain.Priority, error) {
	v := strings.ToLower(strings.TrimSpace(value))
	switch v {
	case "", "medium", "m", "normal", "🔼":
		return domain.PriorityMedium, nil
	case "highest", "h1", "critical", "🔺":
		return domain.PriorityHighest, nil
	case "high", "h", "⏫":
		return domain.PriorityHigh, nil
	case "low", "l", "🔽":
		return domain.PriorityLow, nil
	case "lowest", "l2", "⏬":
		return domain.PriorityLowest, nil
	default:
		return "", fmt.Errorf("priorite invalide: highest|high|medium|low|lowest")
	}
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

func (m *Model) openPomodoroModalForCurrentTask() {
	if m.timerManager.GetActive() != nil {
		return
	}

	target := m.currentPomodoroTargetTask()
	if target == nil || len(m.pomodoroTypes) == 0 {
		return
	}

	m.activeTask = target
	m.pomodoroModalOpen = true
	m.pomodoroSelectIdx = 0
	m.view = ViewFocus
}

func (m Model) handlePomodoroModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.pomodoroTypes) == 0 {
		m.pomodoroModalOpen = false
		return m, nil
	}

	switch msg.String() {
	case "esc":
		m.pomodoroModalOpen = false
		return m, nil
	case "j", "down":
		if m.pomodoroSelectIdx < len(m.pomodoroTypes)-1 {
			m.pomodoroSelectIdx++
		}
		return m, nil
	case "k", "up":
		if m.pomodoroSelectIdx > 0 {
			m.pomodoroSelectIdx--
		}
		return m, nil
	case "enter":
		profile := m.pomodoroTypes[m.pomodoroSelectIdx]
		m.startPomodoroProfile(profile)
		m.pomodoroModalOpen = false
		return m, nil
	}

	return m, nil
}

func (m *Model) currentPomodoroTargetTask() *domain.Task {
	if m.view == ViewTasks {
		return m.selectedTaskTarget()
	}

	if m.view == ViewFocus {
		if m.activeTask != nil {
			return m.activeTask
		}
		if len(m.tasks) > 0 && m.selectedIdx >= 0 && m.selectedIdx < len(m.tasks) {
			return m.tasks[m.selectedIdx]
		}
	}

	if len(m.tasks) > 0 && m.selectedIdx >= 0 && m.selectedIdx < len(m.tasks) {
		return m.tasks[m.selectedIdx]
	}

	return nil
}

func (m *Model) startPomodoroProfile(profile config.PomodoroType) {
	if m.activeTask == nil {
		return
	}

	m.pomodoroSession = &pomodoroSession{
		profile:     profile,
		phase:       pomodoroPhaseWork,
		workCycles:  0,
		selectedFor: m.activeTask,
	}
	m.startCurrentPomodoroPhaseTimer()
}

func (m *Model) switchPomodoroToBreak() {
	if m.timerManager.GetActive() == nil || m.pomodoroSession == nil {
		return
	}
	m.pomodoroSession.phase = pomodoroPhaseBreak
	m.startCurrentPomodoroPhaseTimer()
}

func (m *Model) switchPomodoroToWork() {
	if m.timerManager.GetActive() == nil || m.pomodoroSession == nil {
		return
	}
	m.pomodoroSession.phase = pomodoroPhaseWork
	m.startCurrentPomodoroPhaseTimer()
}

func (m *Model) startCurrentPomodoroPhaseTimer() {
	if m.pomodoroSession == nil || m.activeTask == nil {
		return
	}

	phase := m.pomodoroSession.phase
	profile := m.pomodoroSession.profile
	duration := profile.WorkDuration
	name := fmt.Sprintf("%s work", profile.Name)

	switch phase {
	case pomodoroPhaseBreak:
		duration = profile.BreakDuration
		name = fmt.Sprintf("%s break", profile.Name)
	case pomodoroPhaseLongBreak:
		duration = profile.LongBreakDuration
		name = fmt.Sprintf("%s long break", profile.Name)
	}

	preset := domain.TimerPreset{Name: name, Duration: duration}
	m.timerManager.Start(m.activeTask.ID, preset)
	taskID := m.activeTask.ID
	_ = m.logger.LogAction("timer_started", &taskID, name)
}

func (m *Model) handlePomodoroTick() {
	if m.pomodoroSession == nil {
		return
	}

	active := m.timerManager.GetActive()
	if active == nil {
		return
	}

	if active.Remaining() > 0 {
		return
	}

	finished := m.timerManager.Stop(false)
	if finished == nil {
		return
	}

	taskID := finished.TaskID
	_ = m.logger.LogAction("timer_completed", &taskID, finished.Preset.Name)
	m.advancePomodoroPhaseAndStartNext()
}

func (m *Model) advancePomodoroPhaseAndStartNext() {
	if m.pomodoroSession == nil {
		return
	}

	profile := m.pomodoroSession.profile

	switch m.pomodoroSession.phase {
	case pomodoroPhaseWork:
		m.pomodoroSession.workCycles++
		if m.pomodoroSession.workCycles >= profile.CyclesBeforeLongBreak {
			m.pomodoroSession.phase = pomodoroPhaseLongBreak
			m.pomodoroSession.workCycles = 0
		} else {
			m.pomodoroSession.phase = pomodoroPhaseBreak
		}
	case pomodoroPhaseBreak, pomodoroPhaseLongBreak:
		m.pomodoroSession.phase = pomodoroPhaseWork
	}

	m.startCurrentPomodoroPhaseTimer()
}

