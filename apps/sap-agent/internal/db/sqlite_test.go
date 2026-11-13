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

	// 5. Test Downstream Watermarks
	// Verify seeded baseline watermark exists for 'products'
	wm, err := store.GetDownstreamWatermark(ctx, "products")
	if err != nil || wm == nil {
		t.Fatalf("expected seeded baseline watermark for products, got %v (err: %v)", wm, err)
	}

	// Verify unseeded domain returns nil
	unseeded, err := store.GetDownstreamWatermark(ctx, "custom_domain")
	if err != nil || unseeded != nil {
		t.Fatalf("expected nil watermark for unseeded domain, got %v (err: %v)", unseeded, err)
	}

	// Pass raw time.Now() with monotonic clock suffix (m=+...) to ensure RFC3339 formatting strips it cleanly
	now := time.Now()
	if err := store.SetDownstreamWatermark(ctx, "products", now, 42); err != nil {
		t.Fatalf("failed to set downstream watermark: %v", err)
	}

	wm, err = store.GetDownstreamWatermark(ctx, "products")
	if err != nil || wm == nil {
		t.Fatalf("expected watermark, got nil (err: %v)", err)
	}
	if wm.Unix() != now.Unix() {
		t.Errorf("watermark mismatch: got %v, want %v", wm.Unix(), now.Unix())
	}
}

func TestGetLastCompletedFullMigrationEndTime(t *testing.T) {
	store, err := db.NewSQLiteStore(filepath.Join(t.TempDir(), "test_mig_end.db"))
	if err != nil {
		t.Fatalf("failed to open sqlite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// 1. No runs yet -> nil
	if got, err := store.GetLastCompletedFullMigrationEndTime(ctx, contracts.DomainInventory); err != nil || got != nil {
		t.Fatalf("expected nil with no runs, got %v (err: %v)", got, err)
	}

	// 2. Completed incremental run should be ignored
	incRun, err := store.CreateRun(ctx, 1, contracts.MigrationModeIncremental, 1)
	if err != nil {
		t.Fatalf("CreateRun incremental: %v", err)
	}
	if err := store.CreateOrUpdateStep(ctx, &db.MigrationStep{
		RunID: incRun.ID, Domain: contracts.DomainInventory, Status: contracts.StatusCompleted, StartedAt: time.Now(),
	}); err != nil {
		t.Fatalf("CreateOrUpdateStep incremental: %v", err)
	}
	if err := store.UpdateRunStatus(ctx, incRun.ID, contracts.StatusCompleted, ""); err != nil {
		t.Fatalf("UpdateRunStatus incremental: %v", err)
	}
	if got, err := store.GetLastCompletedFullMigrationEndTime(ctx, contracts.DomainInventory); err != nil || got != nil {
		t.Fatalf("expected nil for incremental-only run, got %v (err: %v)", got, err)
	}

	// 3. Partial-success full run counts only for domains whose step completed
	fullRun, err := store.CreateRun(ctx, 1, contracts.MigrationModeFull, 2)
	if err != nil {
		t.Fatalf("CreateRun full: %v", err)
	}
	if err := store.CreateOrUpdateStep(ctx, &db.MigrationStep{
		RunID: fullRun.ID, Domain: contracts.DomainInventory, Status: contracts.StatusCompleted, StartedAt: time.Now(),
	}); err != nil {
		t.Fatalf("CreateOrUpdateStep inventory: %v", err)
	}
	if err := store.CreateOrUpdateStep(ctx, &db.MigrationStep{
		RunID: fullRun.ID, Domain: contracts.DomainProducts, Status: contracts.StatusFailed, StartedAt: time.Now(),
	}); err != nil {
		t.Fatalf("CreateOrUpdateStep products: %v", err)
	}
	if err := store.UpdateRunStatus(ctx, fullRun.ID, contracts.StatusPartialSuccess, ""); err != nil {
		t.Fatalf("UpdateRunStatus full: %v", err)
	}

	invEnd, err := store.GetLastCompletedFullMigrationEndTime(ctx, contracts.DomainInventory)
	if err != nil || invEnd == nil {
		t.Fatalf("expected completed full migration end time for inventory, got %v (err: %v)", invEnd, err)
	}
	if got, err := store.GetLastCompletedFullMigrationEndTime(ctx, contracts.DomainProducts); err != nil || got != nil {
		t.Fatalf("expected nil for failed products step, got %v (err: %v)", got, err)
	}

	// 4. GetLatestCompletedFullMigrationEndTime & upstream watermarks
	latestFullEnd, err := store.GetLatestCompletedFullMigrationEndTime(ctx)
	if err != nil || latestFullEnd == nil {
		t.Fatalf("expected latest full migration end time, got %v (err: %v)", latestFullEnd, err)
	}
	if wm, err := store.GetUpstreamWatermark(ctx, "upstream"); err != nil || wm != nil {
		t.Fatalf("expected nil initial upstream watermark, got %v (err: %v)", wm, err)
	}
	ts := time.Now().UTC().Truncate(time.Second)
	if err := store.SetUpstreamWatermark(ctx, "upstream", ts, 5); err != nil {
		t.Fatalf("SetUpstreamWatermark: %v", err)
	}
	if wm, err := store.GetUpstreamWatermark(ctx, "upstream"); err != nil || wm == nil || !wm.Equal(ts) {
		t.Fatalf("expected upstream watermark %v, got %v (err: %v)", ts, wm, err)
	}
}
