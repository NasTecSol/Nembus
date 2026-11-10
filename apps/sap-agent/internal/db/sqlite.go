package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	ID             string                  `json:"id"`
	OrganizationID int                     `json:"organization_id"`
	Mode           contracts.MigrationMode `json:"mode"`
	Status         contracts.RunStatus     `json:"status"`
	TotalDomains   int                     `json:"total_domains"`
	CompletedSteps int                     `json:"completed_steps"`
	TotalRecords   int64                   `json:"total_records"`
	ProcessedCount int64                   `json:"processed_count"`
	FailedCount    int64                   `json:"failed_count"`
	StartedAt      time.Time               `json:"started_at"`
	FinishedAt     *time.Time              `json:"finished_at,omitempty"`
	ErrorMessage   string                  `json:"error_message,omitempty"`
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

type UpstreamDispatch struct {
	ID                int64     `json:"id"`
	TransactionNumber string    `json:"transaction_number"`
	TransactionDate   time.Time `json:"transaction_date"`
	Amount            float64   `json:"amount"`
	SAPDocEntry       *int      `json:"sap_doc_entry,omitempty"`
	SAPDocNum         *int      `json:"sap_doc_num,omitempty"`
	Status            string    `json:"status"` // "posted", "held", "failed"
	ErrorMessage      string    `json:"error_message,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
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
		// Auto-heal if malformed due to dirty crash or terminated WAL write
		if strings.Contains(err.Error(), "malformed") {
			backupPath := dbPath + ".corrupt." + time.Now().Format("20060102150405")
			_ = os.Rename(dbPath, backupPath)
			_ = os.Remove(dbPath + "-wal")
			_ = os.Remove(dbPath + "-shm")
			return NewSQLiteStore(dbPath)
		}
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

	CREATE TABLE IF NOT EXISTS upstream_dispatches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		transaction_number TEXT UNIQUE NOT NULL,
		transaction_date DATETIME NOT NULL,
		amount REAL NOT NULL,
		sap_doc_entry INTEGER,
		sap_doc_num INTEGER,
		status TEXT NOT NULL,
		error_message TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_upstream_created_at ON upstream_dispatches(id DESC);

	CREATE TABLE IF NOT EXISTS downstream_watermarks (
		domain TEXT PRIMARY KEY,
		last_sync_time TEXT NOT NULL,
		records_synced INTEGER DEFAULT 0
	);

	UPDATE downstream_watermarks
	SET last_sync_time = strftime('%Y-%m-%dT00:00:00Z', 'now')
	WHERE last_sync_time LIKE '%m=+%';

	INSERT OR IGNORE INTO downstream_watermarks (domain, last_sync_time, records_synced) VALUES
		('categories',  strftime('%Y-%m-%dT00:00:00Z', 'now'), 0),
		('brands',      strftime('%Y-%m-%dT00:00:00Z', 'now'), 0),
		('products',    strftime('%Y-%m-%dT00:00:00Z', 'now'), 0),
		('barcodes',    strftime('%Y-%m-%dT00:00:00Z', 'now'), 0),
		('price_lists', strftime('%Y-%m-%dT00:00:00Z', 'now'), 0);
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
	SET status = ?,
	    error_message = ?,
	    finished_at = COALESCE(?, finished_at),
	    completed_steps = (SELECT COUNT(1) FROM migration_steps WHERE run_id = ? AND status = 'completed'),
	    total_records = (SELECT COALESCE(SUM(total_records), 0) FROM migration_steps WHERE run_id = ?),
	    processed_count = (SELECT COALESCE(SUM(processed_count), 0) FROM migration_steps WHERE run_id = ?),
	    failed_count = (SELECT COALESCE(SUM(failed_count), 0) FROM migration_steps WHERE run_id = ?)
	WHERE id = ?;
	`
	_, err := s.db.ExecContext(ctx, query, string(status), errMsg, finishedAt, runID, runID, runID, runID, runID)
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

	var (
		rows *sql.Rows
		err  error
	)
	if runID == "" {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, run_id, domain, log_level, message, created_at
			FROM migration_logs
			ORDER BY id DESC
			LIMIT ?;
		`, limit)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, run_id, domain, log_level, message, created_at
			FROM migration_logs
			WHERE run_id = ?
			ORDER BY id DESC
			LIMIT ?;
		`, runID, limit)
	}
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

func (s *SQLiteStore) RecordUpstreamDispatch(ctx context.Context, txNum string, txDate time.Time, amount float64, docEntry, docNum *int, status, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	INSERT INTO upstream_dispatches (transaction_number, transaction_date, amount, sap_doc_entry, sap_doc_num, status, error_message, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(transaction_number) DO UPDATE SET
		sap_doc_entry = excluded.sap_doc_entry,
		sap_doc_num = excluded.sap_doc_num,
		status = excluded.status,
		error_message = excluded.error_message,
		created_at = CURRENT_TIMESTAMP;
	`
	_, err := s.db.ExecContext(ctx, query, txNum, txDate, amount, docEntry, docNum, status, errMsg)
	return err
}

func (s *SQLiteStore) GetUpstreamDispatches(ctx context.Context, limit int) ([]UpstreamDispatch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}

	query := `
	SELECT id, transaction_number, transaction_date, amount, sap_doc_entry, sap_doc_num, status, error_message, created_at
	FROM upstream_dispatches
	ORDER BY id DESC
	LIMIT ?;
	`
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []UpstreamDispatch
	for rows.Next() {
		var d UpstreamDispatch
		var docEntry, docNum sql.NullInt64
		var errMsg sql.NullString
		if err := rows.Scan(&d.ID, &d.TransactionNumber, &d.TransactionDate, &d.Amount, &docEntry, &docNum, &d.Status, &errMsg, &d.CreatedAt); err != nil {
			return nil, err
		}
		if docEntry.Valid {
			v := int(docEntry.Int64)
			d.SAPDocEntry = &v
		}
		if docNum.Valid {
			v := int(docNum.Int64)
			d.SAPDocNum = &v
		}
		if errMsg.Valid {
			d.ErrorMessage = errMsg.String
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

func (s *SQLiteStore) GetUpstreamStats(ctx context.Context) (posted int, held int, errors int, lastFlush *time.Time, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRowContext(ctx, `
		SELECT 
			COALESCE(SUM(CASE WHEN status = 'posted' THEN 1 ELSE 0 END), 0) AS posted_count,
			COALESCE(SUM(CASE WHEN status = 'held' THEN 1 ELSE 0 END), 0) AS held_count,
			COALESCE(SUM(CASE WHEN status = 'failed' OR status = 'error' THEN 1 ELSE 0 END), 0) AS error_count,
			MAX(created_at) AS last_flush
		FROM upstream_dispatches;
	`)

	var lf sql.NullTime
	err = row.Scan(&posted, &held, &errors, &lf)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	if lf.Valid {
		lastFlush = &lf.Time
	}
	return posted, held, errors, lastFlush, nil
}

// ClearLogs deletes all entries in migration_logs.
func (s *SQLiteStore) ClearLogs(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, "DELETE FROM migration_logs;")
	return err
}

// GetDownstreamWatermark retrieves the last successful downstream sync timestamp for a domain.
func (s *SQLiteStore) GetDownstreamWatermark(ctx context.Context, domain string) (*time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var raw string
	err := s.db.QueryRowContext(ctx, "SELECT last_sync_time FROM downstream_watermarks WHERE domain = ?;", domain).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02"} {
		if t, parseErr := time.Parse(layout, raw); parseErr == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("unrecognized watermark timestamp format %q", raw)
}

// SetDownstreamWatermark records the last successful sync timestamp and records count for a domain.
func (s *SQLiteStore) SetDownstreamWatermark(ctx context.Context, domain string, t time.Time, count int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	formatted := t.UTC().Format(time.RFC3339)
	query := `
	INSERT INTO downstream_watermarks (domain, last_sync_time, records_synced)
	VALUES (?, ?, ?)
	ON CONFLICT(domain) DO UPDATE SET
		last_sync_time = excluded.last_sync_time,
		records_synced = excluded.records_synced;
	`
	_, err := s.db.ExecContext(ctx, query, domain, formatted, count)
	return err
}

func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
