package tui
package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/beaallombert/gotask/internal/domain"
	"github.com/beaallombert/gotask/internal/storage"
	"github.com/beaallombert/gotask/internal/timer"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func baseModelForTest(t *testing.T) Model {
	t.Helper()

	input := textinput.New()
	input.Width = 32
	vp := viewport.New(80, 12)

	tmp := t.TempDir()
	writer := storage.NewInboxWriter(filepath.Join(tmp, "inbox.md"))
	logger, err := storage.NewSQLiteLogger(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("create sqlite logger: %v", err)
	}
	t.Cleanup(func() {
		_ = logger.Close()
	})

	return Model{
		view:              ViewTasks,
		tasks:             []*domain.Task{},
		timerManager:      timer.NewManager(),
		inboxWriter:       writer,
		logger:            logger,
		taskInput:         input,
		progressBar:       progress.New(progress.WithDefaultGradient()),
		logsViewport:      vp,
		logsViewportReady: true,
		width:             120,
		height:            40,
	}
}

func keyRunes(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func TestParsePriorityInput(t *testing.T) {
	cases := []struct {
		in      string
		expects domain.Priority
		ok      bool
	}{
		{in: "highest", expects: domain.PriorityHighest, ok: true},
		{in: "high", expects: domain.PriorityHigh, ok: true},
		{in: "medium", expects: domain.PriorityMedium, ok: true},
		{in: "low", expects: domain.PriorityLow, ok: true},
		{in: "lowest", expects: domain.PriorityLowest, ok: true},
		{in: "invalid", ok: false},
	}

	for _, tc := range cases {
		got, err := parsePriorityInput(tc.in)
		if tc.ok && err != nil {
			t.Fatalf("input %q unexpected err: %v", tc.in, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("input %q expected err, got none", tc.in)
		}
		if tc.ok && got != tc.expects {
			t.Fatalf("input %q expected %q got %q", tc.in, tc.expects, got)
		}
	}
}

func TestGetTopNTasksSortsAndFiltersCompleted(t *testing.T) {
	now := time.Now()
	later := now.Add(24 * time.Hour)

	m := baseModelForTest(t)
	m.tasks = []*domain.Task{
		{Description: "completed", Status: domain.StatusCompleted, Priority: domain.PriorityHighest, CreatedAt: now.Add(-5 * time.Hour)},
		{Description: "medium-no-date", Status: domain.StatusPaused, Priority: domain.PriorityMedium, CreatedAt: now.Add(-2 * time.Hour)},
		{Description: "high-due", Status: domain.StatusPaused, Priority: domain.PriorityHigh, DueDate: &later, CreatedAt: now.Add(-1 * time.Hour)},
		{Description: "highest", Status: domain.StatusPaused, Priority: domain.PriorityHighest, CreatedAt: now.Add(-3 * time.Hour)},
	}

	got := m.getTopNTasks(3)
	if len(got) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(got))
	}

	if got[0] != "highest" {
		t.Fatalf("expected first task highest, got %q", got[0])
	}
	if got[1] != "high-due" {
		t.Fatalf("expected second task high-due, got %q", got[1])
	}
	if got[2] != "medium-no-date" {
		t.Fatalf("expected third task medium-no-date, got %q", got[2])
	}
}

func TestMoveModeEscCancelsReorder(t *testing.T) {
	m := baseModelForTest(t)
	m.tasks = []*domain.Task{
		{Description: "A", Status: domain.StatusPaused, Priority: domain.PriorityMedium},
		{Description: "B", Status: domain.StatusPaused, Priority: domain.PriorityMedium},
	}
	m.moveMode = true
	m.moveOriginIdx = 0
	m.selectedIdx = 0
	m.moveSelectedTask(1)

	outModel, _ := m.handleTasksKey(tea.KeyMsg{Type: tea.KeyEsc})
	out := outModel.(Model)

	if out.tasks[0].Description != "A" || out.tasks[1].Description != "B" {
		t.Fatalf("expected order restored to A,B got %s,%s", out.tasks[0].Description, out.tasks[1].Description)
	}
	if out.moveMode {
		t.Fatalf("expected moveMode false after esc")
	}
}

func TestTaskModalTabNavigationAndSave(t *testing.T) {
	m := baseModelForTest(t)
	m.openCreateModal()

	// description
	m.taskInput.SetValue("Task from modal")
	outModel, _ := m.handleTasksKey(tea.KeyMsg{Type: tea.KeyTab})
	out := outModel.(Model)

	if out.currentModalField() != "priority" {
		t.Fatalf("expected field priority, got %s", out.currentModalField())
	}

	// priority
	out.taskInput.SetValue("high")
	outModel, _ = out.handleTasksKey(tea.KeyMsg{Type: tea.KeyTab})
	out = outModel.(Model)

	// due date
	out.taskInput.SetValue("2030-01-01")
	outModel, _ = out.handleTasksKey(tea.KeyMsg{Type: tea.KeyTab})
	out = outModel.(Model)

	// duration
	out.taskInput.SetValue("25")
	outModel, _ = out.handleTasksKey(tea.KeyMsg{Type: tea.KeyEnter})
	out = outModel.(Model)

	if len(out.tasks) != 1 {
		t.Fatalf("expected 1 created task, got %d", len(out.tasks))
	}
	created := out.tasks[0]
	if created.Description != "Task from modal" {
		t.Fatalf("unexpected description: %q", created.Description)
	}
	if created.Priority != domain.PriorityHigh {
		t.Fatalf("expected high priority, got %q", created.Priority)
	}
	if created.DueDate == nil || created.DueDate.Format("2006-01-02") != "2030-01-01" {
		t.Fatalf("expected due date 2030-01-01, got %+v", created.DueDate)
	}
	if created.Duration != 25 {
		t.Fatalf("expected duration 25, got %d", created.Duration)
	}
}

func TestViewOverlayPrecedenceTimerOverTaskModal(t *testing.T) {
	m := baseModelForTest(t)
	m.tasks = []*domain.Task{{ID: "t1", Description: "demo", Status: domain.StatusPaused, Priority: domain.PriorityMedium}}
	m.createMode = true
	m.modalDraft = &domain.Task{Description: "draft", Priority: domain.PriorityMedium}
	m.taskInput.SetValue("draft")

	preset := domain.DefaultPresets()[0]
	m.timerManager.Start("t1", preset)

	out := m.View()
	if !strings.Contains(out, "POMODORO IN PROGRESS") {
		t.Fatalf("expected timer overlay to have precedence")
	}
}

func TestModalInvalidDateShowsError(t *testing.T) {
	m := baseModelForTest(t)
	m.openCreateModal()
	m.taskInput.SetValue("demo")

	model, _ := m.handleTasksKey(tea.KeyMsg{Type: tea.KeyTab}) // priority
	m2 := model.(Model)
	m2.taskInput.SetValue("medium")

	model, _ = m2.handleTasksKey(tea.KeyMsg{Type: tea.KeyTab}) // due_date
	m3 := model.(Model)
	m3.taskInput.SetValue("2030-99-99")

	model, _ = m3.handleTasksKey(tea.KeyMsg{Type: tea.KeyTab})
	out := model.(Model)

	if out.modalError == "" {
		t.Fatalf("expected modalError for invalid date")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
