package storage

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/beaallombert/gotask/internal/domain"
)

// InboxReader reads and parses inbox.md
type InboxReader struct {
	FilePath string
}

// NewInboxReader creates a new inbox reader
func NewInboxReader(filePath string) *InboxReader {
	return &InboxReader{FilePath: filePath}
}

// ReadTasks reads all tasks from inbox.md
func (r *InboxReader) ReadTasks() ([]*domain.Task, error) {
	file, err := os.Open(r.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty list if file doesn't exist yet
			return []*domain.Task{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var tasks []*domain.Task
	scanner := bufio.NewScanner(file)
	lineNum := 0
	var parentTask *domain.Task
	var parentIndent int

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		indent := countIndent(line)
		trimmed := strings.TrimSpace(line)

		if !isTaskLine(trimmed) {
			continue
		}

		task := parseTaskLine(trimmed, lineNum)
		if task == nil {
			continue
		}

		if indent == 0 {
			tasks = append(tasks, task)
			parentTask = task
			parentIndent = indent
		} else if indent > parentIndent && parentTask != nil {
			// It's a subtask
			parentTask.Subtasks = append(parentTask.Subtasks, task)
			task.ParentID = parentTask.ID
		} else {
			// Back to root level or same level
			tasks = append(tasks, task)
			parentTask = task
			parentIndent = indent
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// parseTaskLine parses a single task line
// Format: - [STATUS] Description 📅 DATE ⏫ PRIORITY ⏱ 25m
func parseTaskLine(line string, lineNum int) *domain.Task {
	// Extract status
	status := parseStatus(line)
	if status == "" {
		return nil
	}

	// Remove checkbox part
	afterCheckbox := strings.TrimPrefix(line, "- ["+strings.TrimPrefix(line, "- [")[0:1]+"]")
	afterCheckbox = strings.TrimSpace(afterCheckbox)

	// Extract duration ⏱ Xm
	duration := extractDuration(afterCheckbox)

	// Extract date 📅 YYYY-MM-DD
	dueDate := extractDate(afterCheckbox)

	// Extract priority emoji
	priority := extractPriority(afterCheckbox)

	// Remove all metadata from description
	description := cleanDescription(afterCheckbox)

	task := &domain.Task{
		ID:          hashTaskLine(line, lineNum),
		Description: description,
		Status:      status,
		Priority:    priority,
		DueDate:     dueDate,
		Duration:    duration,
		LineNumber:  lineNum,
		CreatedAt:   time.Now(),
	}

	return task
}

func parseStatus(line string) domain.Status {
	if strings.Contains(line, "- [ ]") {
		return domain.StatusPaused
	}
	if strings.Contains(line, "- [/]") || strings.Contains(line, "- [>]") {
		return domain.StatusInProgress
	}
	if strings.Contains(line, "- [x]") || strings.Contains(line, "- [X]") {
		return domain.StatusCompleted
	}
	return ""
}

func extractDuration(text string) int {
	// Match ⏱ Xm or ⏱ Xh or ⏱ Xh Ym
	re := regexp.MustCompile(`⏱\s+(\d+)([hm])(?:\s+(\d+)m)?`)
	matches := re.FindStringSubmatch(text)
	if len(matches) == 0 {
		return 0
	}

	value, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	if unit == "h" {
		value *= 60 // Convert hours to minutes
		if len(matches) > 3 && matches[3] != "" {
			addMin, _ := strconv.Atoi(matches[3])
			value += addMin
		}
	}

	return value
}

func extractDate(text string) *time.Time {
	// Match 📅 YYYY-MM-DD
	re := regexp.MustCompile(`📅\s+(\d{4})-(\d{2})-(\d{2})`)
	matches := re.FindStringSubmatch(text)
	if len(matches) == 0 {
		return nil
	}

	year, _ := strconv.Atoi(matches[1])
	month, _ := strconv.Atoi(matches[2])
	day, _ := strconv.Atoi(matches[3])

	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	return &t
}

func extractPriority(text string) domain.Priority {
	if strings.Contains(text, "🔺") {
		return domain.PriorityHighest
	}
	if strings.Contains(text, "⏫") {
		return domain.PriorityHigh
	}
	if strings.Contains(text, "🔼") {
		return domain.PriorityMedium
	}
	if strings.Contains(text, "🔽") {
		return domain.PriorityLow
	}
	if strings.Contains(text, "⏬") {
		return domain.PriorityLowest
	}
	return domain.PriorityMedium // Default
}

func cleanDescription(text string) string {
	// Remove priority emojis
	text = strings.ReplaceAll(text, "🔺", "")
	text = strings.ReplaceAll(text, "⏫", "")
	text = strings.ReplaceAll(text, "🔼", "")
	text = strings.ReplaceAll(text, "🔽", "")
	text = strings.ReplaceAll(text, "⏬", "")

	// Remove date emoji and date
	re := regexp.MustCompile(`📅\s+\d{4}-\d{2}-\d{2}`)
	text = re.ReplaceAllString(text, "")

	// Remove duration emoji and value
	re = regexp.MustCompile(`⏱\s+\d+[hm](?:\s+\d+m)?`)
	text = re.ReplaceAllString(text, "")

	// Remove completion emoji if present
	text = strings.ReplaceAll(text, "✅", "")

	// Clean up extra spaces
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	return text
}

func countIndent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " \t"))
}

func isTaskLine(line string) bool {
	return strings.HasPrefix(line, "- [")
}

func hashTaskLine(line string, lineNum int) string {
	// Simple hash: use line number as ID for now
	// In production, could use a proper hash
	return fmt.Sprintf("task-%d", lineNum)
}

// InboxWriter writes tasks back to inbox.md
type InboxWriter struct {
	FilePath string
}

// NewInboxWriter creates a new inbox writer
func NewInboxWriter(filePath string) *InboxWriter {
	return &InboxWriter{FilePath: filePath}
}

// WriteTasks writes tasks to inbox.md
func (w *InboxWriter) WriteTasks(tasks []*domain.Task) error {
	file, err := os.Create(w.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	w.writeTasks(writer, tasks, 0)
	return writer.Flush()
}

// AppendTasksSafely appends root-level tasks without rewriting existing content.
// It preserves existing file content and ensures proper newline separation.
func (w *InboxWriter) AppendTasksSafely(tasks []*domain.Task) (int, error) {
	if len(tasks) == 0 {
		return 0, nil
	}

	file, err := os.OpenFile(w.FilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	needLeadingNewline := false
	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}
	if stat.Size() > 0 {
		last := make([]byte, 1)
		if _, err := file.ReadAt(last, stat.Size()-1); err == nil && last[0] != '\n' {
			needLeadingNewline = true
		}
	}

	writer := bufio.NewWriter(file)
	if needLeadingNewline {
		if _, err := writer.WriteString("\n"); err != nil {
			return 0, err
		}
	}

	count := 0
	for _, task := range tasks {
		if task == nil {
			continue
		}
		if _, err := writer.WriteString(w.taskToLine(task, "") + "\n"); err != nil {
			return count, err
		}
		count++
	}

	if err := writer.Flush(); err != nil {
		return count, err
	}

	return count, nil
}

// AppendTaskTreeSafely appends a task and its subtasks to the target file.
// It preserves the hierarchy by writing nested subtasks with indentation.
func (w *InboxWriter) AppendTaskTreeSafely(task *domain.Task) error {
	if task == nil {
		return nil
	}

	file, err := os.OpenFile(w.FilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	needLeadingNewline := false
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if stat.Size() > 0 {
		last := make([]byte, 1)
		if _, err := file.ReadAt(last, stat.Size()-1); err == nil && last[0] != '\n' {
			needLeadingNewline = true
		}
	}

	writer := bufio.NewWriter(file)
	if needLeadingNewline {
		if _, err := writer.WriteString("\n"); err != nil {
			return err
		}
	}

	if err := w.writeTaskTree(writer, task, 0); err != nil {
		return err
	}

	return writer.Flush()
}

func (w *InboxWriter) writeTasks(writer *bufio.Writer, tasks []*domain.Task, indent int) {
	indentStr := strings.Repeat("  ", indent)
	for _, task := range tasks {
		line := w.taskToLine(task, indentStr)
		writer.WriteString(line + "\n")

		if len(task.Subtasks) > 0 {
			w.writeTasks(writer, task.Subtasks, indent+1)
		}
	}
}

func (w *InboxWriter) writeTaskTree(writer *bufio.Writer, task *domain.Task, indent int) error {
	indentStr := strings.Repeat("  ", indent)
	if _, err := writer.WriteString(w.taskToLine(task, indentStr) + "\n"); err != nil {
		return err
	}

	for _, subtask := range task.Subtasks {
		if err := w.writeTaskTree(writer, subtask, indent+1); err != nil {
			return err
		}
	}

	return nil
}

func (w *InboxWriter) taskToLine(task *domain.Task, indent string) string {
	checkbox := "- [ ]"
	if task.Status == domain.StatusInProgress {
		checkbox = "- [/]"
	} else if task.Status == domain.StatusCompleted {
		checkbox = "- [X]"
	}

	line := indent + checkbox + " " + task.Description

	// Add date if present
	if task.DueDate != nil {
		line += " 📅 " + task.DueDate.Format("2006-01-02")
	}

	// Add priority emoji
	priorityEmoji := map[domain.Priority]string{
		domain.PriorityHighest: "🔺",
		domain.PriorityHigh:    "⏫",
		domain.PriorityMedium:  "🔼",
		domain.PriorityLow:     "🔽",
		domain.PriorityLowest:  "⏬",
	}
	if emoji, ok := priorityEmoji[task.Priority]; ok {
		line += " " + emoji
	}

	// Add duration if present
	if task.Duration > 0 {
		line += " ⏱ " + formatDurationForMarkdown(task.Duration)
	}

	// Add completion date if completed
	if task.Status == domain.StatusCompleted && task.CompletedAt != nil {
		line += " ✅ " + task.CompletedAt.Format("2006-01-02")
	}

	return line
}

func formatDurationForMarkdown(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	remaining := minutes % 60
	if remaining == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, remaining)
}
