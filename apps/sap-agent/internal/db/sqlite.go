package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"github.com/NasTecSol/nembus-sap/contracts"
)

type SQLiteStore struct {
	db *sql.DB
	mu sync.RWMutex
}

type MigrationRun struct {
	ID             string              `json:"id"`
	OrganizationID int                 `json:"organization_id"`
	Mode           contracts.MigrationMode `json:"mode"`
	Status         contracts.RunStatus `json:"status"`
	TotalDomains   int                 `json:"total_domains"`
	CompletedSteps int                 `json:"completed_steps"`
	TotalRecords   int64               `json:"total_records"`
	ProcessedCount int64               `json:"processed_count"`
	FailedCount    int64               `json:"failed_count"`
	StartedAt      time.Time           `json:"started_at"`
	FinishedAt     *time.Time          `json:"finished_at,omitempty"`
	ErrorMessage   string              `json:"error_message,omitempty"`
}

type MigrationStep struct {
	ID             string               `json:"id"`
	RunID          string               `json:"run_id"`
	Domain         contracts.DomainType `json:"domain"`
	Status         contracts.RunStatus  `json:"status"`
	TotalRecords   int64                `json:"total_records"`
	ProcessedCount int64                `json:"processed_count"`
	FailedCount    int64                `json:"failed_count"`
	LastWatermark  string               `json:"last_watermark,omitempty"`
	StartedAt      time.Time            `json:"started_at"`
	FinishedAt     *time.Time           `json:"finished_at,omitempty"`
	ErrorMessage   string               `json:"error_message,omitempty"`
}

type MigrationLog struct {
	ID        int64     `json:"id"`
	RunID     string    `json:"run_id"`
	Domain    string    `json:"domain,omitempty"`
	LogLevel  string    `json:"log_level"` // INFO, WARN, ERROR
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database %s: %w", dbPath, err)
	}

	// Pragmas for fast robust local transactions
	_, _ = db.Exec("PRAGMA journal_mode = WAL;")
	_, _ = db.Exec("PRAGMA busy_timeout = 5000;")
	_, _ = db.Exec("PRAGMA foreign_keys = ON;")

	store := &SQLiteStore{db: db}
	if err := store.initSchema(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize sqlite schema: %w", err)
	}

	return store, nil
}

func (s *SQLiteStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS migration_runs (
		id TEXT PRIMARY KEY,
		organization_id INTEGER NOT NULL,
		mode TEXT NOT NULL,
		status TEXT NOT NULL,
		total_domains INTEGER DEFAULT 0,
		completed_steps INTEGER DEFAULT 0,
		total_records INTEGER DEFAULT 0,
		processed_count INTEGER DEFAULT 0,
		failed_count INTEGER DEFAULT 0,
		started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		finished_at DATETIME,
		error_message TEXT
	);

	CREATE TABLE IF NOT EXISTS migration_steps (
		id TEXT PRIMARY KEY,
		run_id TEXT NOT NULL REFERENCES migration_runs(id) ON DELETE CASCADE,
		domain TEXT NOT NULL,
		status TEXT NOT NULL,
		total_records INTEGER DEFAULT 0,
		processed_count INTEGER DEFAULT 0,
		failed_count INTEGER DEFAULT 0,
		last_watermark TEXT,
		started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		finished_at DATETIME,
		error_message TEXT,
		UNIQUE(run_id, domain)
	);

	CREATE TABLE IF NOT EXISTS migration_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id TEXT NOT NULL,
		domain TEXT,
		log_level TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_logs_run_id ON migration_logs(run_id, id DESC);
	CREATE INDEX IF NOT EXISTS idx_steps_run_domain ON migration_steps(run_id, domain);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStore) CreateRun(ctx context.Context, orgID int, mode contracts.MigrationMode, totalDomains int) (*MigrationRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	run := &MigrationRun{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		Mode:           mode,
		Status:         contracts.StatusRunning,
		TotalDomains:   totalDomains,
		CompletedSteps: 0,
		TotalRecords:   0,
		ProcessedCount: 0,
		FailedCount:    0,
		StartedAt:      time.Now(),
	}

	query := `
	INSERT INTO migration_runs (id, organization_id, mode, status, total_domains, completed_steps, total_records, processed_count, failed_count, started_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err := s.db.ExecContext(ctx, query, run.ID, run.OrganizationID, string(run.Mode), string(run.Status), run.TotalDomains, run.CompletedSteps, run.TotalRecords, run.ProcessedCount, run.FailedCount, run.StartedAt)
	if err != nil {
		return nil, err
	}

	return run, nil
}

func (s *SQLiteStore) GetRun(ctx context.Context, runID string) (*MigrationRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, organization_id, mode, status, total_domains, completed_steps, total_records, processed_count, failed_count, started_at, finished_at, error_message
	FROM migration_runs WHERE id = ?;
	`
	row := s.db.QueryRowContext(ctx, query, runID)

	var run MigrationRun
	var mode, status string
	var finishedAt sql.NullTime
	var errMsg sql.NullString

	err := row.Scan(&run.ID, &run.OrganizationID, &mode, &status, &run.TotalDomains, &run.CompletedSteps, &run.TotalRecords, &run.ProcessedCount, &run.FailedCount, &run.StartedAt, &finishedAt, &errMsg)
	if err != nil {
		return nil, err
	}

	run.Mode = contracts.MigrationMode(mode)
	run.Status = contracts.RunStatus(status)
	if finishedAt.Valid {
		run.FinishedAt = &finishedAt.Time
	}
	if errMsg.Valid {
		run.ErrorMessage = errMsg.String
	}

	return &run, nil
}

func (s *SQLiteStore) GetLatestRun(ctx context.Context) (*MigrationRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, organization_id, mode, status, total_domains, completed_steps, total_records, processed_count, failed_count, started_at, finished_at, error_message
	FROM migration_runs ORDER BY started_at DESC LIMIT 1;
	`
	row := s.db.QueryRowContext(ctx, query)

	var run MigrationRun
	var mode, status string
	var finishedAt sql.NullTime
	var errMsg sql.NullString

	err := row.Scan(&run.ID, &run.OrganizationID, &mode, &status, &run.TotalDomains, &run.CompletedSteps, &run.TotalRecords, &run.ProcessedCount, &run.FailedCount, &run.StartedAt, &finishedAt, &errMsg)
	if err != nil {
		return nil, err
	}

	run.Mode = contracts.MigrationMode(mode)
	run.Status = contracts.RunStatus(status)
	if finishedAt.Valid {
		run.FinishedAt = &finishedAt.Time
	}
	if errMsg.Valid {
		run.ErrorMessage = errMsg.String
	}

	return &run, nil
}

// GetRuns returns migration runs ordered by start time descending, with pagination.
func (s *SQLiteStore) GetRuns(ctx context.Context, limit, offset int) ([]*MigrationRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}

	query := `
	SELECT id, organization_id, mode, status, total_domains, completed_steps, total_records, processed_count, failed_count, started_at, finished_at, error_message
	FROM migration_runs ORDER BY started_at DESC LIMIT ? OFFSET ?;
	`
	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*MigrationRun
	for rows.Next() {
		var run MigrationRun
		var mode, status string
		var finishedAt sql.NullTime
		var errMsg sql.NullString
		if err := rows.Scan(&run.ID, &run.OrganizationID, &mode, &status, &run.TotalDomains, &run.CompletedSteps, &run.TotalRecords, &run.ProcessedCount, &run.FailedCount, &run.StartedAt, &finishedAt, &errMsg); err != nil {
			return nil, err
		}
		run.Mode = contracts.MigrationMode(mode)
		run.Status = contracts.RunStatus(status)
		if finishedAt.Valid {
			run.FinishedAt = &finishedAt.Time
		}
		if errMsg.Valid {
			run.ErrorMessage = errMsg.String
		}
		runs = append(runs, &run)
	}
	return runs, nil
}

func (s *SQLiteStore) UpdateRunStatus(ctx context.Context, runID string, status contracts.RunStatus, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var finishedAt *time.Time
	switch status {
	case contracts.StatusCompleted, contracts.StatusFailed, contracts.StatusCancelled, contracts.StatusPartialSuccess:
		now := time.Now()
		finishedAt = &now
	}

	query := `
	UPDATE migration_runs 
	SET status = ?, error_message = ?, finished_at = COALESCE(?, finished_at)
	WHERE id = ?;
	`
	_, err := s.db.ExecContext(ctx, query, string(status), errMsg, finishedAt, runID)
	return err
}

func (s *SQLiteStore) CreateOrUpdateStep(ctx context.Context, step *MigrationStep) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if step.ID == "" {
		step.ID = uuid.New().String()
	}

	query := `
	INSERT INTO migration_steps (id, run_id, domain, status, total_records, processed_count, failed_count, last_watermark, started_at, finished_at, error_message)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(run_id, domain) DO UPDATE SET
		status = excluded.status,
		total_records = excluded.total_records,
		processed_count = excluded.processed_count,
		failed_count = excluded.failed_count,
		last_watermark = excluded.last_watermark,
		finished_at = excluded.finished_at,
		error_message = excluded.error_message;
	`
	_, err := s.db.ExecContext(ctx, query, step.ID, step.RunID, string(step.Domain), string(step.Status), step.TotalRecords, step.ProcessedCount, step.FailedCount, step.LastWatermark, step.StartedAt, step.FinishedAt, step.ErrorMessage)
	return err
}

func (s *SQLiteStore) GetSteps(ctx context.Context, runID string) ([]MigrationStep, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, run_id, domain, status, total_records, processed_count, failed_count, last_watermark, started_at, finished_at, error_message
	FROM migration_steps WHERE run_id = ? ORDER BY started_at ASC;
	`
	rows, err := s.db.QueryContext(ctx, query, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []MigrationStep
	for rows.Next() {
		var step MigrationStep
		var domain, status string
		var watermark, errMsg sql.NullString
		var finishedAt sql.NullTime

		if err := rows.Scan(&step.ID, &step.RunID, &domain, &status, &step.TotalRecords, &step.ProcessedCount, &step.FailedCount, &watermark, &step.StartedAt, &finishedAt, &errMsg); err != nil {
			return nil, err
		}
		step.Domain = contracts.DomainType(domain)
		step.Status = contracts.RunStatus(status)
		if watermark.Valid {
			step.LastWatermark = watermark.String
		}
		if finishedAt.Valid {
			step.FinishedAt = &finishedAt.Time
		}
		if errMsg.Valid {
			step.ErrorMessage = errMsg.String
		}
		steps = append(steps, step)
	}

	return steps, nil
}

func (s *SQLiteStore) Log(ctx context.Context, runID string, domain string, level string, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT INTO migration_logs (run_id, domain, log_level, message, created_at) VALUES (?, ?, ?, ?, ?);`
	_, _ = s.db.ExecContext(ctx, query, runID, domain, level, message, time.Now())
}

func (s *SQLiteStore) GetRecentLogs(ctx context.Context, runID string, limit int) ([]MigrationLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	query := `
	SELECT id, run_id, domain, log_level, message, created_at
	FROM migration_logs 
	WHERE run_id = ? 
	ORDER BY id DESC 
	LIMIT ?;
	`
	rows, err := s.db.QueryContext(ctx, query, runID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []MigrationLog
	for rows.Next() {
		var l MigrationLog
		var dom sql.NullString
		if err := rows.Scan(&l.ID, &l.RunID, &dom, &l.LogLevel, &l.Message, &l.CreatedAt); err != nil {
			return nil, err
		}
		if dom.Valid {
			l.Domain = dom.String
		}
		logs = append(logs, l)
	}

	// Reverse to return chronological order
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	return logs, nil
}

// GetLastCompletedStep returns the most recently completed step for a given
// domain and organization. This is used by delta/incremental runs to determine
// the watermark cursor from which extraction should resume.
func (s *SQLiteStore) GetLastCompletedStep(ctx context.Context, orgID int, domain contracts.DomainType) (*MigrationStep, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT ms.id, ms.run_id, ms.domain, ms.status, ms.total_records,
	       ms.processed_count, ms.failed_count, ms.last_watermark,
	       ms.started_at, ms.finished_at, ms.error_message
	FROM migration_steps ms
	JOIN migration_runs mr ON ms.run_id = mr.id
	WHERE mr.organization_id = ?
	  AND ms.domain = ?
	  AND ms.status = 'completed'
	  AND ms.last_watermark IS NOT NULL
	  AND ms.last_watermark != ''
	ORDER BY ms.finished_at DESC
	LIMIT 1;
	`
	row := s.db.QueryRowContext(ctx, query, orgID, string(domain))

	var step MigrationStep
	var domain2, status string
	var watermark, errMsg sql.NullString
	var finishedAt sql.NullTime

	err := row.Scan(
		&step.ID, &step.RunID, &domain2, &status,
		&step.TotalRecords, &step.ProcessedCount, &step.FailedCount,
		&watermark, &step.StartedAt, &finishedAt, &errMsg,
	)
	if err != nil {
		return nil, err
	}

	step.Domain = contracts.DomainType(domain2)
	step.Status = contracts.RunStatus(status)
	if watermark.Valid {
		step.LastWatermark = watermark.String
	}
	if finishedAt.Valid {
		step.FinishedAt = &finishedAt.Time
	}
	if errMsg.Valid {
		step.ErrorMessage = errMsg.String
	}

	return &step, nil
}

func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
