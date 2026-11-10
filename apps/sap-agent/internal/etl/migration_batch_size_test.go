package etl

import (
	"testing"

	"github.com/NasTecSol/nembus-sap/contracts"
)

func TestMigrationIngestChunkSize(t *testing.T) {
	if got := migrationIngestChunkSize(contracts.DomainProducts, 500); got != 25 {
		t.Fatalf("products cap: got %d want 25", got)
	}
	if got := migrationIngestChunkSize(contracts.DomainProducts, 10); got != 10 {
		t.Fatalf("products respect lower config: got %d want 10", got)
	}
	if got := migrationIngestChunkSize(contracts.DomainBarcodes, 500); got != 100 {
		t.Fatalf("barcodes cap: got %d want 100", got)
	}
	if got := migrationIngestChunkSize(contracts.DomainUOMGroups, 500); got != 5 {
		t.Fatalf("uom_groups cap: got %d want 5", got)
	}
	if got := migrationIngestChunkSize(contracts.DomainStores, 500); got != 500 {
		t.Fatalf("stores unchanged: got %d want 500", got)
	}
}
