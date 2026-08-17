package db_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-sap/contracts"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
)

func TestSQLiteStoreLifecycle(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent_db_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test_agent.db")
	store, err := db.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// 1. Create Migration Run
	run, err := store.CreateRun(ctx, 1, contracts.MigrationModeFull, 5)
	if err != nil {
		t.Fatalf("failed to create run: %v", err)
	}
	if run.ID == "" {
		t.Errorf("expected non-empty run ID")
	}

	// 2. Add Steps
	step := &db.MigrationStep{
		RunID:          run.ID,
		Domain:         contracts.DomainStores,
		Status:         contracts.StatusRunning,
		TotalRecords:   10,
		ProcessedCount: 10,
		FailedCount:    0,
		StartedAt:      time.Now(),
	}
	if err := store.CreateOrUpdateStep(ctx, step); err != nil {
		t.Fatalf("failed to create step: %v", err)
	}

	steps, err := store.GetSteps(ctx, run.ID)
	if err != nil || len(steps) != 1 {
		t.Fatalf("expected 1 step, got %d (err: %v)", len(steps), err)
	}
	if steps[0].Domain != contracts.DomainStores {
		t.Errorf("expected domain 'stores', got %s", steps[0].Domain)
	}

	// 3. Add Logs
	store.Log(ctx, run.ID, "stores", "INFO", "Stores extraction finished.")
	logs, err := store.GetRecentLogs(ctx, run.ID, 10)
	if err != nil || len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d (err: %v)", len(logs), err)
	}
	if logs[0].Message != "Stores extraction finished." {
		t.Errorf("unexpected log message: %s", logs[0].Message)
	}

	// 4. Update Run Status
	if err := store.UpdateRunStatus(ctx, run.ID, contracts.StatusCompleted, ""); err != nil {
		t.Fatalf("failed to update run status: %v", err)
	}

	latest, err := store.GetLatestRun(ctx)
	if err != nil || latest.Status != contracts.StatusCompleted {
		t.Errorf("expected status 'completed', got %v", latest.Status)
	}
}
