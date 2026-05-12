package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/beaallombert/gotask/internal/domain"
)

// SQLiteLogger manages session and intervention logging
type SQLiteLogger struct {
	db *sql.DB
}

// NewSQLiteLogger creates a new SQLite logger
func NewSQLiteLogger(filePath string) (*SQLiteLogger, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, err
	}

	logger := &SQLiteLogger{db: db}

	// Initialize schema
	if err := logger.initSchema(); err != nil {
		return nil, err
	}

	return logger, nil
}

// initSchema creates necessary tables if they don't exist
func (l *SQLiteLogger) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		task_id TEXT,
		preset_name TEXT,
		duration_minutes INTEGER,
		started_at TIMESTAMP,
		ended_at TIMESTAMP,
		status TEXT,
		aborted BOOLEAN,
		created_at TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS interventions (
		id TEXT PRIMARY KEY,
		type TEXT,
		rule_type TEXT,
		message TEXT,
		task_id TEXT,
		created_at TIMESTAMP,
		dismissed BOOLEAN,
		dismissed_at TIMESTAMP,
		confidence REAL
	);

	CREATE TABLE IF NOT EXISTS actions (
		id TEXT PRIMARY KEY,
		action_type TEXT,
		task_id TEXT,
		details TEXT,
		created_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_task_id ON sessions(task_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at);
	CREATE INDEX IF NOT EXISTS idx_interventions_created_at ON interventions(created_at);
	CREATE INDEX IF NOT EXISTS idx_actions_created_at ON actions(created_at);
	`

	_, err := l.db.Exec(schema)
	return err
}

// LogSession logs a completed timer session
func (l *SQLiteLogger) LogSession(session *domain.Session) error {
	stmt := `
	INSERT INTO sessions (id, task_id, preset_name, duration_minutes, started_at, ended_at, status, aborted, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	endedAt := session.Timer.EndedAt
	if session.Timer.Status == domain.TimerStatusActive {
		now := time.Now()
		endedAt = &now
	}

	_, err := l.db.Exec(stmt,
		session.ID,
		session.TaskID,
		session.Timer.Preset.Name,
		session.Timer.Preset.Duration,
		session.Timer.StartedAt,
		endedAt,
		session.Timer.Status,
		session.Timer.Aborted,
		session.CreatedAt,
	)

	return err
}

// LogIntervention logs a system intervention
func (l *SQLiteLogger) LogIntervention(intervention *domain.Intervention) error {
	stmt := `
	INSERT INTO interventions (id, type, rule_type, message, task_id, created_at, dismissed, dismissed_at, confidence)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	taskID := ""
	if intervention.TaskID != nil {
		taskID = *intervention.TaskID
	}

	_, err := l.db.Exec(stmt,
		intervention.ID,
		intervention.Type,
		intervention.RuleType,
		intervention.Message,
		taskID,
		intervention.CreatedAt,
		intervention.Dismissed,
		intervention.DismissedAt,
		intervention.Confidence,
	)

	return err
}

// LogAction logs a user action
func (l *SQLiteLogger) LogAction(actionType string, taskID *string, details string) error {
	stmt := `
	INSERT INTO actions (id, action_type, task_id, details, created_at)
	VALUES (?, ?, ?, ?, ?)
	`

	taskIDStr := ""
	if taskID != nil {
		taskIDStr = *taskID
	}

	_, err := l.db.Exec(stmt,
		fmt.Sprintf("action-%d", time.Now().UnixNano()),
		actionType,
		taskIDStr,
		details,
		time.Now(),
	)

	return err
}

// GetRecentSessions returns the N most recent sessions
func (l *SQLiteLogger) GetRecentSessions(limit int) ([]*domain.Session, error) {
	stmt := `
	SELECT id, task_id, preset_name, duration_minutes, started_at, ended_at, status, aborted, created_at
	FROM sessions
	ORDER BY created_at DESC
	LIMIT ?
	`

	rows, err := l.db.Query(stmt, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.Session

	for rows.Next() {
		var id, taskID, presetName, status string
		var duration int
		var startedAt, endedAt, createdAt time.Time
		var aborted bool

		if err := rows.Scan(&id, &taskID, &presetName, &duration, &startedAt, &endedAt, &status, &aborted, &createdAt); err != nil {
			return nil, err
		}

		session := &domain.Session{
			ID:     id,
			TaskID: taskID,
			Timer: &domain.Timer{
				ID:     id,
				TaskID: taskID,
				Preset: domain.TimerPreset{Name: presetName, Duration: duration},
				Status: domain.TimerStatus(status),
				Aborted: aborted,
				StartedAt: startedAt,
				EndedAt: &endedAt,
			},
			CreatedAt: createdAt,
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetRecentInterventions returns the N most recent interventions
func (l *SQLiteLogger) GetRecentInterventions(limit int) ([]*domain.Intervention, error) {
	stmt := `
	SELECT id, type, rule_type, message, task_id, created_at, dismissed, dismissed_at, confidence
	FROM interventions
	ORDER BY created_at DESC
	LIMIT ?
	`

	rows, err := l.db.Query(stmt, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interventions []*domain.Intervention

	for rows.Next() {
		var id, typeStr, ruleTypeStr, message, taskID string
		var createdAt time.Time
		var dismissed bool
		var dismissedAt sql.NullTime
		var confidence float64

		if err := rows.Scan(&id, &typeStr, &ruleTypeStr, &message, &taskID, &createdAt, &dismissed, &dismissedAt, &confidence); err != nil {
			return nil, err
		}

		var taskIDPtr *string
		if taskID != "" {
			taskIDPtr = &taskID
		}

		var dismissedAtPtr *time.Time
		if dismissedAt.Valid {
			dismissedAtPtr = &dismissedAt.Time
		}

		intervention := &domain.Intervention{
			ID:          id,
			Type:        domain.InterventionType(typeStr),
			RuleType:    domain.RuleType(ruleTypeStr),
			Message:     message,
			TaskID:      taskIDPtr,
			CreatedAt:   createdAt,
			Dismissed:   dismissed,
			DismissedAt: dismissedAtPtr,
			Confidence:  confidence,
		}

		interventions = append(interventions, intervention)
	}

	return interventions, nil
}

// GetTotalTimeOnTask returns total minutes spent on a task
func (l *SQLiteLogger) GetTotalTimeOnTask(taskID string) (int, error) {
	stmt := `
	SELECT COALESCE(SUM(duration_minutes), 0)
	FROM sessions
	WHERE task_id = ? AND status = ?
	`

	var total int
	err := l.db.QueryRow(stmt, taskID, domain.TimerStatusCompleted).Scan(&total)

	return total, err
}

// GetTimeTrackedToday returns minutes tracked today
func (l *SQLiteLogger) GetTimeTrackedToday() (int, error) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	stmt := `
	SELECT COALESCE(SUM(duration_minutes), 0)
	FROM sessions
	WHERE started_at >= ? AND started_at < ? AND status = ?
	`

	var total int
	err := l.db.QueryRow(stmt, today, tomorrow, domain.TimerStatusCompleted).Scan(&total)

	return total, err
}

// Close closes the database connection
func (l *SQLiteLogger) Close() error {
	if l.db != nil {
		return l.db.Close()
	}
	return nil
}

// GetTailLog returns the last N lines from the log file
func (l *SQLiteLogger) GetTailLog(lines int) (string, error) {
	// For now, return recent interventions and actions
	// In production, would read actual log file
	interventions, err := l.GetRecentInterventions(lines)
	if err != nil {
		return "", err
	}

	var result string
	for _, intervention := range interventions {
		result += fmt.Sprintf("[%s] %s: %s\n", intervention.CreatedAt.Format("15:04:05"), intervention.Type, intervention.Message)
	}

	return result, nil
}
