package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/beaallombert/gotask/internal/domain"
	"github.com/beaallombert/gotask/internal/rules"
	"github.com/beaallombert/gotask/internal/storage"
	"github.com/beaallombert/gotask/internal/timer"
)

// App represents the CLI application
type App struct {
	inboxPath           string
	dbPath              string
	inboxReader         *storage.InboxReader
	inboxWriter         *storage.InboxWriter
	logger              *storage.SQLiteLogger
	timerManager        *timer.Manager
	timerStatePath      string
	timerStateStore     *storage.TimerStateStore
	recurrenceGenerator *rules.RecurrenceGenerator
}

// NewApp creates a new CLI app
func NewApp() (*App, error) {
	// Default paths
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	appDir := filepath.Join(home, ".gotask")
	os.MkdirAll(appDir, 0700)

	inboxPath := "inbox.md" // Current directory
	dbPath := filepath.Join(appDir, "gotask.db")
	timerStatePath := filepath.Join(appDir, "timer.state")

	reader := storage.NewInboxReader(inboxPath)
	writer := storage.NewInboxWriter(inboxPath)

	logger, err := storage.NewSQLiteLogger(dbPath)
	if err != nil {
		return nil, err
	}

	timerStateStore := storage.NewTimerStateStore(timerStatePath)

	app := &App{
		inboxPath:           inboxPath,
		dbPath:              dbPath,
		inboxReader:         reader,
		inboxWriter:         writer,
		logger:              logger,
		timerManager:        timer.NewManager(),
		timerStatePath:      timerStatePath,
		timerStateStore:     timerStateStore,
		recurrenceGenerator: rules.NewRecurrenceGenerator(),
	}

	// Load existing timer from disk
	if existingTimer, err := timerStateStore.Load(); err == nil && existingTimer != nil {
		app.timerManager.RestoreTimer(existingTimer)
	}

	return app, nil
}

// HandleInboxCommand handles inbox subcommands
func (app *App) HandleInboxCommand(args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: gotask inbox <subcommand> [options]")
		return nil
	}

	if err := app.syncRecurringTasks(); err != nil {
		return err
	}

	fs := flag.NewFlagSet("inbox", flag.ContinueOnError)

	switch args[0] {
	case "top":
		return app.handleInboxTop()

	case "add":
		fs.String("description", "", "Description of the task")
		fs.Parse(args[1:])
		desc := fs.Lookup("description").Value.String()
		if desc == "" {
			fmt.Println("Error: --description is required")
			return nil
		}
		return app.handleInboxAdd(desc)

	case "start":
		fs.Int("line", -1, "Line number of the task")
		fs.Parse(args[1:])
		lineStr := fs.Lookup("line").Value.String()
		line, _ := strconv.Atoi(lineStr)
		if line < 0 {
			fmt.Println("Error: --line is required")
			return nil
		}
		return app.handleInboxStart(line)

	case "pause":
		fs.Int("line", -1, "Line number of the task")
		fs.Parse(args[1:])
		lineStr := fs.Lookup("line").Value.String()
		line, _ := strconv.Atoi(lineStr)
		if line < 0 {
			fmt.Println("Error: --line is required")
			return nil
		}
		return app.handleInboxPause(line)

	case "complete":
		fs.Int("line", -1, "Line number of the task")
		fs.Parse(args[1:])
		lineStr := fs.Lookup("line").Value.String()
		line, _ := strconv.Atoi(lineStr)
		if line < 0 {
			fmt.Println("Error: --line is required")
			return nil
		}
		return app.handleInboxComplete(line)

	default:
		fmt.Printf("Unknown inbox subcommand: %s\n", args[0])
		return nil
	}
}

// HandleTimerCommand handles timer subcommands
func (app *App) HandleTimerCommand(args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: gotask timer <subcommand> [options]")
		return nil
	}

	switch args[0] {
	case "presets":
		return app.handleTimerPresets()

	case "start":
		fs := flag.NewFlagSet("timer start", flag.ContinueOnError)
		preset := fs.String("preset", "Pomodoro Standard", "Preset name")
		line := fs.Int("line", -1, "Task line number")
		fs.Parse(args[1:])
		return app.handleTimerStart(*preset, *line)

	case "stop":
		fs := flag.NewFlagSet("timer stop", flag.ContinueOnError)
		aborted := fs.Bool("aborted", false, "Mark as aborted")
		fs.Parse(args[1:])
		return app.handleTimerStop(*aborted)

	case "status":
		return app.handleTimerStatus()

	default:
		fmt.Printf("Unknown timer subcommand: %s\n", args[0])
		return nil
	}
}

// HandleFocusCommand handles focus subcommands
func (app *App) HandleFocusCommand(args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: gotask focus <subcommand>")
		return nil
	}

	switch args[0] {
	case "snapshot":
		return app.handleFocusSnapshot()
	default:
		fmt.Printf("Unknown focus subcommand: %s\n", args[0])
		return nil
	}
}

// Implementation of subcommands

func (app *App) handleInboxTop() error {
	tasks, err := app.inboxReader.ReadTasks()
	if err != nil {
		return err
	}

	// Filter and sort by effective priority
	var activeTasks []*domain.Task
	for _, task := range tasks {
		if task.Status != domain.StatusCompleted {
			activeTasks = append(activeTasks, task)
		}
	}

	// Sort by effective priority
	sort.Slice(activeTasks, func(i, j int) bool {
		return activeTasks[i].EffectivePriority() < activeTasks[j].EffectivePriority()
	})

	// Show top 3
	limit := 3
	if len(activeTasks) < limit {
		limit = len(activeTasks)
	}

	fmt.Println("Top 3 tasks:")
	for i := 0; i < limit; i++ {
		task := activeTasks[i]
		status := "[ ]"
		if task.Status == domain.StatusInProgress {
			status = "[/]"
		} else if task.Status == domain.StatusPaused {
			status = "[ ]"
		}
		fmt.Printf("[Line %d] %s %s\n", task.LineNumber, status, task.Description)
	}

	return nil
}

func (app *App) handleInboxAdd(description string) error {
	tasks, err := app.inboxReader.ReadTasks()
	if err != nil {
		return err
	}

	newTask := &domain.Task{
		ID:          fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Description: description,
		Status:      domain.StatusPaused,
		Priority:    domain.PriorityMedium,
		CreatedAt:   time.Now(),
		LineNumber:  len(tasks) + 1,
	}

	tasks = append(tasks, newTask)
	if err := app.inboxWriter.WriteTasks(tasks); err != nil {
		return err
	}

	fmt.Printf("✓ Task added: %s\n", description)
	return nil
}

func (app *App) handleInboxStart(lineNum int) error {
	tasks, err := app.inboxReader.ReadTasks()
	if err != nil {
		return err
	}

	// Stop any currently active task
	for _, task := range tasks {
		if task.Status == domain.StatusInProgress && task.LineNumber != lineNum {
			task.Status = domain.StatusPaused
		}
	}

	// Find and start the task
	found := false
	for _, task := range tasks {
		if task.LineNumber == lineNum {
			task.Status = domain.StatusInProgress
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("Task not found at line %d\n", lineNum)
		return nil
	}

	if err := app.inboxWriter.WriteTasks(tasks); err != nil {
		return err
	}

	fmt.Printf("▶ Task started at line %d\n", lineNum)
	return nil
}

func (app *App) handleInboxPause(lineNum int) error {
	tasks, err := app.inboxReader.ReadTasks()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.LineNumber == lineNum {
			task.Status = domain.StatusPaused
			if err := app.inboxWriter.WriteTasks(tasks); err != nil {
				return err
			}
			fmt.Printf("⏸ Task paused at line %d\n", lineNum)
			return nil
		}
	}

	fmt.Printf("Task not found at line %d\n", lineNum)
	return nil
}

func (app *App) handleInboxComplete(lineNum int) error {
	tasks, err := app.inboxReader.ReadTasks()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.LineNumber == lineNum {
			task.Status = domain.StatusCompleted
			now := time.Now()
			task.CompletedAt = &now
			if err := app.inboxWriter.WriteTasks(tasks); err != nil {
				return err
			}
			fmt.Printf("✓ Task completed at line %d\n", lineNum)
			return nil
		}
	}

	fmt.Printf("Task not found at line %d\n", lineNum)
	return nil
}

func (app *App) handleTimerPresets() error {
	presets := domain.DefaultPresets()
	fmt.Println("Available timer presets:")
	for _, p := range presets {
		fmt.Printf("  - %s: %d minutes\n", p.Name, p.Duration)
	}
	return nil
}

func (app *App) handleTimerStart(presetName string, taskLine int) error {
	// Find the preset
	presets := domain.DefaultPresets()
	var preset *domain.TimerPreset
	for i := range presets {
		if presets[i].Name == presetName {
			preset = &presets[i]
			break
		}
	}

	if preset == nil {
		fmt.Printf("Preset not found: %s\n", presetName)
		return nil
	}

	// Get task info if provided
	taskID := ""
	if taskLine > 0 {
		tasks, err := app.inboxReader.ReadTasks()
		if err == nil {
			for _, task := range tasks {
				if task.LineNumber == taskLine {
					taskID = task.ID
					break
				}
			}
		}
	}

	// Start timer
	timer := app.timerManager.Start(taskID, *preset)

	// Persist timer state
	if err := app.timerStateStore.Save(timer); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not save timer state: %v\n", err)
	}

	fmt.Printf("⏱ Timer started: %s (%d min)\n", preset.Name, preset.Duration)
	fmt.Printf("Timer ID: %s\n", timer.ID)

	// Log to database
	session := &domain.Session{
		ID:        timer.ID,
		Timer:     timer,
		TaskID:    taskID,
		CreatedAt: time.Now(),
	}
	app.logger.LogSession(session)

	return nil
}

func (app *App) handleTimerStop(aborted bool) error {
	timer := app.timerManager.Stop(aborted)
	if timer == nil {
		fmt.Println("No active timer to stop")
		return nil
	}

	// Clear persisted timer state
	if err := app.timerStateStore.Clear(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not clear timer state: %v\n", err)
	}

	remaining := timer.Remaining()
	if remaining > 0 {
		fmt.Printf("⏱ Timer stopped (aborted: %v)\n", aborted)
	} else {
		fmt.Printf("✓ Timer completed!\n")
	}

	return nil
}

func (app *App) handleTimerStatus() error {
	timer := app.timerManager.GetActive()
	if timer == nil {
		fmt.Println("No active timer")
		return nil
	}

	remaining := timer.Remaining()
	elapsed := timer.Elapsed()
	progress := timer.Progress()

	fmt.Printf("Active Timer: %s\n", timer.Preset.Name)
	fmt.Printf("Elapsed: %ds / Total: %ds\n", elapsed, int64(timer.Preset.Duration)*60)
	fmt.Printf("Remaining: %ds\n", remaining)
	fmt.Printf("Progress: %.1f%%\n", progress*100)

	return nil
}

func (app *App) handleFocusSnapshot() error {
	tasks, err := app.inboxReader.ReadTasks()
	if err != nil {
		return err
	}

	fmt.Println("=== FOCUS SNAPSHOT ===")

	// Show active timer
	if timer := app.timerManager.GetActive(); timer != nil {
		remaining := timer.Remaining()
		progress := timer.Progress()
		fmt.Printf("⏱ Active Timer: %s\n", timer.Preset.Name)
		fmt.Printf("  Remaining: %ds | Progress: %.0f%%\n\n", remaining, progress*100)
	} else {
		fmt.Println("⏱ No active timer")
	}

	// Show top 3 tasks
	var activeTasks []*domain.Task
	for _, task := range tasks {
		if task.Status != domain.StatusCompleted {
			activeTasks = append(activeTasks, task)
		}
	}

	sort.Slice(activeTasks, func(i, j int) bool {
		return activeTasks[i].EffectivePriority() < activeTasks[j].EffectivePriority()
	})

	limit := 3
	if len(activeTasks) < limit {
		limit = len(activeTasks)
	}

	fmt.Println("📋 Top 3 tasks:")
	for i := 0; i < limit; i++ {
		task := activeTasks[i]
		status := "[ ]"
		if task.Status == domain.StatusInProgress {
			status = "[/]"
		}
		fmt.Printf("%d. %s %s\n", i+1, status, task.Description)
	}

	return nil
}

// Close closes the app resources
func (app *App) Close() error {
	if app.logger != nil {
		return app.logger.Close()
	}
	return nil
}

func (app *App) syncRecurringTasks() error {
	if app.recurrenceGenerator == nil {
		return nil
	}

	tasks, err := app.inboxReader.ReadTasks()
	if err != nil {
		return err
	}

	now := time.Now()
	horizon := now.Add(14 * 24 * time.Hour)
	generated := app.recurrenceGenerator.Generate(tasks, now, horizon)
	if len(generated) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(tasks)+len(generated))
	for _, task := range tasks {
		key := recurrenceDedupKey(task)
		if key == "" {
			continue
		}
		seen[key] = struct{}{}
	}

	toAppend := make([]*domain.Task, 0, len(generated))
	for _, task := range generated {
		key := recurrenceDedupKey(task)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		toAppend = append(toAppend, task)
	}

	if len(toAppend) == 0 {
		return nil
	}

	count, err := app.inboxWriter.AppendTasksSafely(toAppend)
	if err != nil {
		return err
	}

	if count > 0 {
		fmt.Printf("↻ %d recurring task(s) generated\n", count)
	}

	return nil
}

func recurrenceDedupKey(task *domain.Task) string {
	if task == nil || task.DueDate == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(task.Description)) + "|" + string(task.Priority) + "|" + task.DueDate.Format("2006-01-02")
}
