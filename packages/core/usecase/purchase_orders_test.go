package usecase

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestPurchaseOrderPartnerIDMapping(t *testing.T) {
	// Verify that repository PurchaseOrder correctly maps PartnerID to SupplierID in domain outputs
	po := repository.PurchaseOrder{
		ID:             101,
		OrganizationID: 1,
		PoNumber:       "PO-2026-0001",
		PartnerID:      42,
		StoreID:        5,
		Status:         pgtype.Text{String: "draft", Valid: true},
	}

	if po.PartnerID != 42 {
		t.Fatalf("expected po.PartnerID to be 42, got %d", po.PartnerID)
	}

	// Verify JSON serialization of PurchaseOrder uses "partner_id"
	poJSON, err := json.Marshal(po)
	if err != nil {
		t.Fatalf("failed to marshal PurchaseOrder: %v", err)
	}

	var rawMap map[string]interface{}
	if err := json.Unmarshal(poJSON, &rawMap); err != nil {
		t.Fatalf("failed to unmarshal PurchaseOrder JSON: %v", err)
	}

	if val, ok := rawMap["partner_id"]; !ok || int(val.(float64)) != 42 {
		t.Fatalf("expected JSON to contain 'partner_id': 42, got %v", rawMap["partner_id"])
	}
	if _, ok := rawMap["partners_id"]; ok {
		t.Fatalf("expected JSON to NOT contain deprecated 'partners_id'")
	}
}

func TestGetPurchaseOrderWithDetailsRowPartnerIDMapping(t *testing.T) {
	row := repository.GetPurchaseOrderWithDetailsRow{
		ID:             202,
		OrganizationID: 1,
		PoNumber:       "PO-2026-0002",
		PartnerID:      88,
		PartnerName:    pgtype.Text{String: "Acme Supplies", Valid: true},
		PartnerCode:    pgtype.Text{String: "SUP-ACME", Valid: true},
		StoreID:        3,
	}

	if row.PartnerID != 88 {
		t.Fatalf("expected row.PartnerID to be 88, got %d", row.PartnerID)
	}

	rowJSON, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("failed to marshal GetPurchaseOrderWithDetailsRow: %v", err)
	}

	var rawMap map[string]interface{}
	if err := json.Unmarshal(rowJSON, &rawMap); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if val, ok := rawMap["partner_id"]; !ok || int(val.(float64)) != 88 {
		t.Fatalf("expected JSON to contain 'partner_id': 88, got %v", rawMap["partner_id"])
	}
	if _, ok := rawMap["partners_id"]; ok {
		t.Fatalf("expected JSON to NOT contain deprecated 'partners_id'")
	}
}

func TestListPurchaseOrdersRowPartnerIDMapping(t *testing.T) {
	row := repository.ListPurchaseOrdersRow{
		ID:             303,
		OrganizationID: 1,
		PoNumber:       "PO-2026-0003",
		PartnerID:      99,
		StoreID:        2,
	}

	if row.PartnerID != 99 {
		t.Fatalf("expected row.PartnerID to be 99, got %d", row.PartnerID)
	}

	summary := PurchaseOrderSummaryOutput{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		PoNumber:       row.PoNumber,
		SupplierID:     row.PartnerID,
		StoreID:        row.StoreID,
		Status:         "approved",
	}

	if summary.SupplierID != 99 {
		t.Fatalf("expected summary.SupplierID to be 99, got %d", summary.SupplierID)
	}

	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("failed to marshal summary: %v", err)
	}

	var rawMap map[string]interface{}
	if err := json.Unmarshal(summaryJSON, &rawMap); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if val, ok := rawMap["supplier_id"]; !ok || int(val.(float64)) != 99 {
		t.Fatalf("expected JSON to contain 'supplier_id': 99, got %v", rawMap["supplier_id"])
	}
}

// findMonorepoRoot walks up from cwd until it finds apps/cloud-server (Nembus repo layout).
func findMonorepoRoot() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "apps", "cloud-server")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func TestGetPurchaseOrderWithDetailsQueryDoesNotContainPartnersID(t *testing.T) {
	root, inMonorepo := findMonorepoRoot()

	// Verify repository generated code across core and cloud-server
	repoPaths := []string{
		filepath.Join("..", "repository", "purchase_orders.sql.go"),
	}
	if inMonorepo {
		repoPaths = append(repoPaths, filepath.Join(root, "apps", "cloud-server", "internal", "repository", "purchase_orders.sql.go"))
	}

	for _, relPath := range repoPaths {
		data, err := os.ReadFile(relPath)
		if err != nil {
			t.Fatalf("file %s not found: %v", relPath, err)
		}
		content := string(data)

		// 1. Must NOT contain po.partners_id
		if strings.Contains(content, "po.partners_id") {
			t.Fatalf("expected %s to NOT contain 'po.partners_id'", relPath)
		}

		// 2. Must contain po.partner_id
		if !strings.Contains(content, "po.partner_id") {
			t.Fatalf("expected %s to contain 'po.partner_id'", relPath)
		}

		// 3. getPurchaseOrderWithDetails specifically must have po.partner_id and not po.partners_id
		idx := strings.Index(content, "const getPurchaseOrderWithDetails")
		if idx != -1 {
			startBacktick := strings.Index(content[idx:], "`")
			if startBacktick != -1 {
				rest := content[idx+startBacktick+1:]
				closingBacktick := strings.Index(rest, "`")
				if closingBacktick != -1 {
					querySnippet := rest[:closingBacktick]
					if strings.Contains(querySnippet, "partners_id") {
						t.Fatalf("expected getPurchaseOrderWithDetails in %s to not contain 'partners_id'", relPath)
					}
					if !strings.Contains(querySnippet, "po.partner_id") {
						t.Fatalf("expected getPurchaseOrderWithDetails in %s to select 'po.partner_id'", relPath)
					}
				}
			}
		}
	}

	// Also verify query source definitions
	queryPaths := []string{
		filepath.Join("..", "queries", "purchase_orders.sql"),
	}
	if inMonorepo {
		queryPaths = append(queryPaths, filepath.Join(root, "apps", "cloud-server", "queries", "purchase_orders.sql"))
	}
	for _, qPath := range queryPaths {
		data, err := os.ReadFile(qPath)
		if err != nil {
			t.Fatalf("file %s not found: %v", qPath, err)
		}
		if strings.Contains(string(data), "partners_id") {
			t.Fatalf("expected %s to NOT contain deprecated 'partners_id'", qPath)
		}
	}
}

