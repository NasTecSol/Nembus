package etl

import (
	"testing"

	"github.com/NasTecSol/nembus-sap/contracts"
)

func TestMigrationIngestChunkSize(t *testing.T) {
	if got := migrationIngestChunkSize(contracts.DomainProducts, 500); got != 50 {
		t.Fatalf("products cap: got %d want 50", got)
	}
	if got := migrationIngestChunkSize(contracts.DomainProducts, 25); got != 25 {
		t.Fatalf("products respect lower config: got %d want 25", got)
	}
	if got := migrationIngestChunkSize(contracts.DomainBarcodes, 500); got != 200 {
		t.Fatalf("barcodes cap: got %d want 200", got)
	}
	if got := migrationIngestChunkSize(contracts.DomainStores, 500); got != 500 {
		t.Fatalf("stores unchanged: got %d want 500", got)
	}
}
