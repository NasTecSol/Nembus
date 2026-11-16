package etl

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap/contracts"
)

func TestResolveDownstreamSince(t *testing.T) {
	store, err := db.NewSQLiteStore(filepath.Join(t.TempDir(), "test_watermark.db"))
	if err != nil {
		t.Fatalf("failed to open sqlite store: %v", err)
	}
	defer store.Close()

	engine := &Engine{sqlite: store}
	ctx := context.Background()
	now := time.Now().UTC()

	// Without any migration: unseeded domain (inventory) falls back to start of day.
	invNoMig := engine.resolveDownstreamSince(ctx, contracts.DomainInventory, now)
	if want := now.Truncate(24 * time.Hour); !invNoMig.Equal(want) {
		t.Fatalf("expected start-of-day fallback %v, got %v", want, invNoMig)
	}

	// Simulate completing a full migration strictly after the seeded/fallback baseline.
	time.Sleep(1100 * time.Millisecond)
	run1, err := store.CreateRun(ctx, 1, contracts.MigrationModeFull, 2)
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	for _, domain := range []contracts.DomainType{contracts.DomainProducts, contracts.DomainInventory} {
		if err := store.CreateOrUpdateStep(ctx, &db.MigrationStep{
			RunID: run1.ID, Domain: domain, Status: contracts.StatusCompleted, StartedAt: time.Now(),
		}); err != nil {
			t.Fatalf("CreateOrUpdateStep(%s): %v", domain, err)
		}
	}
	if err := store.UpdateRunStatus(ctx, run1.ID, contracts.StatusCompleted, ""); err != nil {
		t.Fatalf("UpdateRunStatus: %v", err)
	}

	mig1End, err := store.GetLastCompletedFullMigrationEndTime(ctx, contracts.DomainProducts)
	if err != nil || mig1End == nil {
		t.Fatalf("expected migration end time, got %v (err: %v)", mig1End, err)
	}

	// First downstream sync after full migration uses the migration end time for both seeded and unseeded domains.
	for _, domain := range []contracts.DomainType{contracts.DomainProducts, contracts.DomainInventory} {
		got := engine.resolveDownstreamSince(ctx, domain, time.Now())
		if !got.Equal(*mig1End) {
			t.Fatalf("%s first downstream since: got %v, want migration end %v", domain, got, *mig1End)
		}
	}

	// Recording a subsequent downstream sync start time makes auto-scheduled runs resume from that time.
	sync1Start := mig1End.Add(2 * time.Minute).Truncate(time.Second)
	if err := store.SetDownstreamWatermark(ctx, string(contracts.DomainProducts), sync1Start, 3); err != nil {
		t.Fatalf("SetDownstreamWatermark: %v", err)
	}
	if got := engine.resolveDownstreamSince(ctx, contracts.DomainProducts, time.Now()); !got.Equal(sync1Start) {
		t.Fatalf("scheduled downstream since: got %v, want sync1Start %v", got, sync1Start)
	}

	// An older watermark (before the last completed full migration) is superseded by the migration end time.
	olderSync := mig1End.Add(-5 * time.Minute).Truncate(time.Second)
	if err := store.SetDownstreamWatermark(ctx, string(contracts.DomainInventory), olderSync, 1); err != nil {
		t.Fatalf("SetDownstreamWatermark older: %v", err)
	}
	if got := engine.resolveDownstreamSince(ctx, contracts.DomainInventory, time.Now()); !got.Equal(*mig1End) {
		t.Fatalf("post-migration override since: got %v, want mig1End %v", got, *mig1End)
	}
}
