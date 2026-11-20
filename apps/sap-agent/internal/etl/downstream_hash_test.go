package etl

import (
	"testing"

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
}
