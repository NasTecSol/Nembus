package etl

import (
	"testing"

	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap/contracts"
	"github.com/NasTecSol/nembus-sap/mappings"
)

func TestDownstreamPayloadHashDedup(t *testing.T) {
	e := &Engine{lastDownstreamHash: make(map[contracts.DomainType]string)}

	uomsV1 := []mappings.CanonicalUOM{
		{Code: "EA", Name: "Each"},
		{Code: "BOX", Name: "Box"},
	}
	h1, ok := computePayloadHash(uomsV1)
	if !ok || h1 == "" {
		t.Fatalf("expected valid payload hash")
	}
	if e.isDownstreamPayloadUnchanged(contracts.DomainUOM, h1) {
		t.Fatalf("first run should not be marked unchanged")
	}

	e.setDownstreamPayloadHash(contracts.DomainUOM, h1)
	if !e.isDownstreamPayloadUnchanged(contracts.DomainUOM, h1) {
		t.Fatalf("identical payload should be marked unchanged")
	}

	uomsV2 := []mappings.CanonicalUOM{
		{Code: "EA", Name: "Each"},
		{Code: "BOX", Name: "Carton"},
	}
	h2, ok := computePayloadHash(uomsV2)
	if !ok || h2 == h1 {
		t.Fatalf("modified payload should produce a different hash")
	}
	if e.isDownstreamPayloadUnchanged(contracts.DomainUOM, h2) {
		t.Fatalf("modified payload should not be marked unchanged")
	}

	// Updating cloud config must invalidate cached hashes so payloads are sent to the new destination.
	e.setDownstreamPayloadHash(contracts.DomainUOM, h2)
	e.UpdateCloudConfig(config.CloudConfig{OrganizationID: 2})
	if e.isDownstreamPayloadUnchanged(contracts.DomainUOM, h2) {
		t.Fatalf("expected UpdateCloudConfig to clear downstream payload hash cache")
	}
}

func TestIsBatchFullySucceeded(t *testing.T) {
	if isBatchFullySucceeded(nil) {
		t.Fatalf("nil response must not be treated as fully succeeded")
	}
	if isBatchFullySucceeded(&contracts.MigrationBatchResponse{Success: false, RecordsFailed: 0}) {
		t.Fatalf("Success=false must not be treated as fully succeeded")
	}
	if isBatchFullySucceeded(&contracts.MigrationBatchResponse{Success: true, RecordsFailed: 2}) {
		t.Fatalf("RecordsFailed>0 must not be treated as fully succeeded")
	}
	if !isBatchFullySucceeded(&contracts.MigrationBatchResponse{Success: true, RecordsStaged: 5, RecordsFailed: 0}) {
		t.Fatalf("Success=true with RecordsFailed=0 should be treated as fully succeeded")
	}
}
