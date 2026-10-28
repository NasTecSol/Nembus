package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
)

func TestOrganizationZatcaMetadataIsExplicitAndTyped(t *testing.T) {
	for _, test := range []struct {
		metadata string
		enabled  bool
		bad      bool
	}{
		{`{}`, false, false}, {`{"zatca":{"seller":{"street":"Street"}}}`, false, false},
		{`{"zatca":{"enabled":false}}`, false, false}, {`{"zatca":{"enabled":true}}`, true, false},
		{`{"zatca":{"enabled":"true"}}`, false, true}, {`{"zatca":true}`, false, true},
	} {
		enabled, err := OrganizationZatcaEnabled(&repository.Organization{Metadata: []byte(test.metadata)})
		if enabled != test.enabled || (err != nil) != test.bad {
			t.Fatalf("%s: enabled=%v error=%v", test.metadata, enabled, err)
		}
	}
}

func TestOrganizationSwitchOverridesLegacyFlagAndPreservesIssuedDocument(t *testing.T) {
	db, uc := orderZatcaFixture(t)
	uc.cfg.Enabled = false // Legacy environment flag must not disable an opted-in organization.
	first, err := uc.EnsureSalesOrderDocument(context.Background(), db.order.ID)
	if err != nil {
		t.Fatal(err)
	}
	db.organizationMetadata = []byte(`{"zatca":{"enabled":false}}`)
	uc.cfg.Enabled = true // Likewise the environment flag must not override disabled metadata.
	if _, err := uc.EnsureSalesOrderDocument(context.Background(), db.order.ID); !errors.Is(err, ErrZatcaDisabled) {
		t.Fatal("disabled organization generated an invoice", err)
	}
	if len(db.chainEntries) != 1 {
		t.Fatal("disabling changed the document chain")
	}
	db.organizationMetadata = []byte(testZatcaSellerMetadata)
	again, err := uc.EnsureSalesOrderDocument(context.Background(), db.order.ID)
	if err != nil || again.ZatcaUuid != first.ZatcaUuid || again.SignedXml != first.SignedXml || len(db.chainEntries) != 1 {
		t.Fatal("reenabling reissued the original invoice", err)
	}
}
