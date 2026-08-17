package contracts

import (
	"time"

	"github.com/NasTecSol/nembus-sap/mappings"
)

// Migration Domain Identifiers
type DomainType string

const (
	DomainStores          DomainType = "stores"
	DomainUsers           DomainType = "users"
	DomainUOM             DomainType = "uom"
	DomainUOMGroups       DomainType = "uom_groups"
	DomainCategories      DomainType = "categories"
	DomainBrands          DomainType = "brands"
	DomainProducts        DomainType = "products"
	DomainBarcodes        DomainType = "barcodes"
	DomainInventory       DomainType = "inventory"
	DomainPartners        DomainType = "partners"
	DomainSalesOrders     DomainType = "sales_orders"
	DomainInvoices        DomainType = "invoices"
	DomainPriceLists      DomainType = "price_lists"
	DomainBPAddresses     DomainType = "bp_addresses"
)

// Migration Execution Modes
type MigrationMode string

const (
	MigrationModeFull        MigrationMode = "full"
	MigrationModeIncremental MigrationMode = "incremental"
	MigrationModeValidateOnly MigrationMode = "validate_only"
)

// Status Enum for Migration Steps and Runs
type RunStatus string

const (
	StatusPending        RunStatus = "pending"
	StatusRunning        RunStatus = "running"
	StatusPaused         RunStatus = "paused"
	StatusCompleted      RunStatus = "completed"
	StatusFailed         RunStatus = "failed"
	StatusCancelled      RunStatus = "cancelled"
	StatusPartialSuccess RunStatus = "partial_success"
	StatusReconciled     RunStatus = "reconciled"
)

// Ingestion Batch Payload dispatched from Agent to Cloud Server
type MigrationBatchPayload struct {
	BatchID        string                  `json:"batch_id"`
	RunID          string                  `json:"run_id"`
	OrganizationID int                     `json:"organization_id"`
	Domain         DomainType              `json:"domain"`
	SequenceNumber int                     `json:"sequence_number"`
	IsLastBatch    bool                    `json:"is_last_batch"`
	Timestamp      time.Time               `json:"timestamp"`
	WatermarkFrom  string                  `json:"watermark_from,omitempty"`
	WatermarkTo    string                  `json:"watermark_to,omitempty"`
	
	// Typed Payload collections (Domain-specific)
	Stores         []mappings.CanonicalStore           `json:"stores,omitempty"`
	Locations      []mappings.CanonicalStorageLocation `json:"locations,omitempty"`
	Users          []mappings.CanonicalUser            `json:"users,omitempty"`
	Cashiers       []mappings.CanonicalCashier         `json:"cashiers,omitempty"`
	UOMs           []mappings.CanonicalUOM             `json:"uom,omitempty"`
	UOMGroups      []mappings.CanonicalUOMGroup        `json:"uom_groups,omitempty"`
	Categories     []mappings.CanonicalCategory        `json:"categories,omitempty"`
	Brands         []mappings.CanonicalBrand           `json:"brands,omitempty"`
	Products       []mappings.CanonicalProduct         `json:"products,omitempty"`
	Barcodes       []mappings.CanonicalBarcode         `json:"barcodes,omitempty"`
	Inventory      []mappings.CanonicalInventoryStock  `json:"inventory,omitempty"`
	Partners       []mappings.CanonicalPartner         `json:"partners,omitempty"`
	SalesOrders    []mappings.CanonicalSalesOrder      `json:"sales_orders,omitempty"`
	Invoices       []mappings.CanonicalInvoice         `json:"invoices,omitempty"`
	PriceLists     []mappings.CanonicalPriceList       `json:"price_lists,omitempty"`
	PriceItems     []mappings.CanonicalPriceListItem   `json:"price_items,omitempty"`
	BPAddresses    []mappings.CanonicalBPAddress       `json:"bp_addresses,omitempty"`
}

func (p *MigrationBatchPayload) RecordCount() int {
	switch p.Domain {
	case DomainStores:
		return len(p.Stores) + len(p.Locations)
	case DomainUsers:
		return len(p.Users) + len(p.Cashiers)
	case DomainUOM:
		return len(p.UOMs)
	case DomainUOMGroups:
		return len(p.UOMGroups)
	case DomainCategories:
		return len(p.Categories)
	case DomainBrands:
		return len(p.Brands)
	case DomainProducts:
		return len(p.Products)
	case DomainBarcodes:
		return len(p.Barcodes)
	case DomainInventory:
		return len(p.Inventory)
	case DomainPartners:
		return len(p.Partners)
	case DomainSalesOrders:
		return len(p.SalesOrders)
	case DomainInvoices:
		return len(p.Invoices)
	case DomainPriceLists:
		return len(p.PriceLists) + len(p.PriceItems)
	case DomainBPAddresses:
		return len(p.BPAddresses)
	default:
		return 0
	}
}

// Ingestion Response from Cloud Server to Agent
type MigrationBatchResponse struct {
	Success        bool      `json:"success"`
	BatchID        string    `json:"batch_id"`
	Domain         DomainType`json:"domain"`
	RecordsStaged  int       `json:"records_staged"`
	RecordsMerged  int       `json:"records_merged"`
	RecordsFailed  int       `json:"records_failed"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	Errors         []string  `json:"errors,omitempty"`
	WatermarkSaved string    `json:"watermark_saved,omitempty"`
}

// Pre-migration SAP Discovery Information
type DiscoveryResult struct {
	CompanyName    string            `json:"company_name"`
	SAPVersion     string            `json:"sap_version"`
	PatchLevel     string            `json:"patch_level"`
	Address        string            `json:"address"`
	DatabaseName   string            `json:"database_name"`
	ConnectedAt    time.Time         `json:"connected_at"`
	TableCounts    map[string]int64  `json:"table_counts"`
	Warnings       []string          `json:"warnings"`
	IsCompatible   bool              `json:"is_compatible"`
}

// Post-migration Audit & Reconciliation DTO
type DomainReconciliation struct {
	Domain          DomainType `json:"domain"`
	SAPSourceCount  int64      `json:"sap_source_count"`
	TargetCount     int64      `json:"target_count"`
	Difference      int64      `json:"difference"`
	SAPNumericSum   float64    `json:"sap_numeric_sum,omitempty"`
	TargetNumericSum float64   `json:"target_numeric_sum,omitempty"`
	SumDifference   float64    `json:"sum_difference,omitempty"`
	Status          string     `json:"status"` // "MATCH", "MISMATCH", "WARNING"
	Notes           string     `json:"notes,omitempty"`
}

type ReconciliationReport struct {
	RunID          string                 `json:"run_id"`
	OrganizationID int                    `json:"organization_id"`
	GeneratedAt    time.Time              `json:"generated_at"`
	TotalDomains   int                    `json:"total_domains"`
	PassedDomains  int                    `json:"passed_domains"`
	FailedDomains  int                    `json:"failed_domains"`
	Domains        []DomainReconciliation `json:"domains"`
	AuditSummary   string                 `json:"audit_summary"`
}
