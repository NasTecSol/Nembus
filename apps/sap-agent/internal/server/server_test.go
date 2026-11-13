package server

import (
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap/contracts"
)

func TestDownstreamMasterDomainsIncludesInventory(t *testing.T) {
	for _, want := range []contracts.DomainType{contracts.DomainUOM, contracts.DomainUOMGroups, contracts.DomainInventory} {
		if !slices.Contains(DownstreamMasterDomains, want) {
			t.Fatalf("expected DownstreamMasterDomains to include %q, got %v", want, DownstreamMasterDomains)
		}
	}
	uomIdx := slices.Index(DownstreamMasterDomains, contracts.DomainUOM)
	groupsIdx := slices.Index(DownstreamMasterDomains, contracts.DomainUOMGroups)
	productsIdx := slices.Index(DownstreamMasterDomains, contracts.DomainProducts)
	if !(uomIdx < groupsIdx && groupsIdx < productsIdx) {
		t.Fatalf("expected uom < uom_groups < products ordering, got %v", DownstreamMasterDomains)
	}
}

func TestBuildPipeConfigUsesConfiguredStoreID(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "agent_config.json")
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	s := &Server{}
	cfg.NembusStoreID = 42
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	if got := s.buildPipeConfig().NembusStoreID; got != 42 {
		t.Fatalf("configured NembusStoreID: got %d want 42", got)
	}

	cfg.NembusStoreID = 0
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	t.Setenv("NEMBUS_STORE_ID", "7")
	if got := s.buildPipeConfig().NembusStoreID; got != 7 {
		t.Fatalf("env fallback NembusStoreID: got %d want 7", got)
	}
}

func TestNextScheduledRunTime(t *testing.T) {
	runStart := time.Date(2026, 11, 13, 7, 0, 0, 0, time.UTC)
	interval := 5 * time.Minute

	// Run finishes in 28s -> next run is 07:05:00 (not 07:05:28)
	now := runStart.Add(28 * time.Second)
	want := time.Date(2026, 11, 13, 7, 5, 0, 0, time.UTC)
	if got := nextScheduledRunTime(runStart, now, interval); !got.Equal(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}

	// Run takes longer than interval -> next run clamps to now
	overrun := runStart.Add(6 * time.Minute)
	if got := nextScheduledRunTime(runStart, overrun, interval); !got.Equal(overrun) {
		t.Fatalf("expected %v on overrun, got %v", overrun, got)
	}
}
