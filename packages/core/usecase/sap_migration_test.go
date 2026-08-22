package usecase

import (
	"strings"
	"testing"
)

// This is a source-level regression test because SAP migration tests have no
// live database seam. It locks the actual product upsert contract: INSERT
// keeps new products incomplete, while ON CONFLICT preserves only absent
// brand/description values from the existing organization-local row.
func TestSAPProductUpsertPreservesAbsentEnrichmentValues(t *testing.T) {
	query := strings.ToLower(sapProductUpsertQuery)

	for _, fragment := range []string{
		"on conflict(organization_id, sku) do update set",
		"description = case",
		"when nullif(btrim(excluded.description), '') is not null then excluded.description",
		"else products.description",
		"brand_id = case",
		"when excluded.brand_id is not null then excluded.brand_id",
		"else products.brand_id",
		"category_id = excluded.category_id",
		"product_type = excluded.product_type",
		"is_active = excluded.is_active",
		"is_sellable = excluded.is_sellable",
		"track_inventory = excluded.track_inventory",
		"$1, $2, $3, $4",
		"(select id from brands where code = $6)",
	} {
		if !strings.Contains(query, fragment) {
			t.Errorf("SAP product upsert is missing contract fragment %q", fragment)
		}
	}

	for _, forbidden := range []string{
		"description = excluded.description,",
		"brand_id = excluded.brand_id,",
	} {
		if strings.Contains(query, forbidden) {
			t.Errorf("SAP product upsert still unconditionally overwrites %s", forbidden)
		}
	}
}
