package syncpb

import (
	"testing"
)

func TestPullEntityTypesMasterDataOrder(t *testing.T) {
	types := PullEntityTypes()

	expectedMasterData := []string{
		"units_of_measure",
		"product_categories",
		"payment_terms",
		"users",
		"roles",
		"permissions",
		"role_permissions",
		"user_roles",
		"cashiers",
		"business_partners",
		"partner_addresses",
	}

	if len(types) < len(expectedMasterData) {
		t.Fatalf("PullEntityTypes() returned %d entities, expected at least %d", len(types), len(expectedMasterData))
	}

	for i, expected := range expectedMasterData {
		if types[i] != expected {
			t.Errorf("expected master data entity at index %d to be %q, got %q", i, expected, types[i])
		}
	}

	indexMap := make(map[string]int, len(types))
	for i, entity := range types {
		indexMap[entity] = i
	}

	// Verify foreign key dependency hierarchy:
	// 1. units_of_measure must precede products
	if indexMap["units_of_measure"] >= indexMap["products"] {
		t.Errorf("units_of_measure (%d) must precede products (%d)", indexMap["units_of_measure"], indexMap["products"])
	}
	// 2. product_categories must precede products
	if indexMap["product_categories"] >= indexMap["products"] {
		t.Errorf("product_categories (%d) must precede products (%d)", indexMap["product_categories"], indexMap["products"])
	}
	// 3. payment_terms must precede business_partners
	if indexMap["payment_terms"] >= indexMap["business_partners"] {
		t.Errorf("payment_terms (%d) must precede business_partners (%d)", indexMap["payment_terms"], indexMap["business_partners"])
	}
	// 4. users must precede cashiers (RBAC entities between users and cashiers)
	if indexMap["users"] >= indexMap["cashiers"] {
		t.Errorf("users (%d) must precede cashiers (%d)", indexMap["users"], indexMap["cashiers"])
	}
	if indexMap["user_roles"] >= indexMap["cashiers"] {
		t.Errorf("user_roles (%d) must precede cashiers (%d)", indexMap["user_roles"], indexMap["cashiers"])
	}
	if indexMap["cashiers"] >= indexMap["user_biometric_templates"] {
		t.Errorf("cashiers (%d) must precede user_biometric_templates (%d)", indexMap["cashiers"], indexMap["user_biometric_templates"])
	}
	// 5. users must precede business_partners (sales_rep_user_id)
	if indexMap["users"] >= indexMap["business_partners"] {
		t.Errorf("users (%d) must precede business_partners (%d)", indexMap["users"], indexMap["business_partners"])
	}
	// 6. business_partners must precede partner_addresses
	if indexMap["business_partners"] >= indexMap["partner_addresses"] {
		t.Errorf("business_partners (%d) must precede partner_addresses (%d)", indexMap["business_partners"], indexMap["partner_addresses"])
	}

	// Verify IsPullEntity
	for _, entity := range expectedMasterData {
		if !IsPullEntity(entity) {
			t.Errorf("expected IsPullEntity(%q) to be true", entity)
		}
	}

	if IsPullEntity("unknown_nonexistent_table") {
		t.Errorf("expected IsPullEntity(unknown_nonexistent_table) to be false")
	}

	// Verify slice mutation isolation
	types[0] = "corrupted_entity"
	typesSecondCall := PullEntityTypes()
	if typesSecondCall[0] == "corrupted_entity" {
		t.Errorf("PullEntityTypes() returned a mutable reference to shared slice")
	}
}
