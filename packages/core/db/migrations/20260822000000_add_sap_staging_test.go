package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestSAPStagingMigrationHistoryProvisionsCanonicalObjects(t *testing.T) {
	contents, err := os.ReadFile("20260822000000_add_sap_staging.sql")
	if err != nil {
		t.Fatalf("read SAP staging migration: %v", err)
	}

	migration := strings.ToLower(string(contents))
	for _, object := range []string{
		"create schema if not exists staging",
		"create table if not exists staging.sap_migration_batches",
		"create table if not exists staging.sap_stores",
		"create table if not exists staging.sap_products",
		"create table if not exists staging.sap_inventory",
	} {
		if !strings.Contains(migration, object) {
			t.Errorf("SAP staging migration is missing canonical object %q", object)
		}
	}
}
