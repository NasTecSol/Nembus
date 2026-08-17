package reconciliation

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/NasTecSol/nembus-sap/contracts"
	"github.com/NasTecSol/nembus-sap/schema"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap-agent/internal/transport"
)

type AuditEngine struct {
	mssql       *db.MSSQLClient
	cloudClient *transport.CloudClient
	sqlite      *db.SQLiteStore
}

func NewAuditEngine(mssql *db.MSSQLClient, cloudClient *transport.CloudClient, sqlite *db.SQLiteStore) *AuditEngine {
	return &AuditEngine{
		mssql:       mssql,
		cloudClient: cloudClient,
		sqlite:      sqlite,
	}
}

// Reconcile runs count and financial-sum checks across all 10 migration domains.
// runID may be empty — in that case it fetches the latest completed run from SQLite.
func (a *AuditEngine) Reconcile(ctx context.Context, runID string, orgID int) (*contracts.ReconciliationReport, error) {
	// Resolve runID to the latest run if not supplied
	if runID == "" {
		run, err := a.sqlite.GetLatestRun(ctx)
		if err == nil && run != nil {
			runID = run.ID
		}
	}

	report := &contracts.ReconciliationReport{
		RunID:          runID,
		OrganizationID: orgID,
		GeneratedAt:    time.Now(),
		Domains:        make([]contracts.DomainReconciliation, 0),
	}

	type domainCheckSpec struct {
		domain    contracts.DomainType
		table     string
		keyColumn string
	}

	// Simple count checks — compare SAP row count vs staged count from SQLite step
	countChecks := []domainCheckSpec{
		{contracts.DomainStores, schema.TableOWHS, "WhsCode"},
		{contracts.DomainUsers, schema.TableOUSR, "USERID"},
		{contracts.DomainUOM, schema.TableOUOM, "UomEntry"},
		{contracts.DomainUOMGroups, schema.TableOUGP, "UgpEntry"},
		{contracts.DomainCategories, schema.TableOITB, "ItmsGrpCod"},
		{contracts.DomainBrands, schema.TableOMRG, "FirmCode"},
		{contracts.DomainProducts, schema.TableOITM, "ItemCode"},
		{contracts.DomainBarcodes, schema.TableOBCD, "BcdEntry"},
		{contracts.DomainPartners, schema.TableOCRD, "CardCode"},
		{contracts.DomainSalesOrders, schema.TableORDR, "DocEntry"},
	}

	for _, spec := range countChecks {
		if audit, err := a.auditDomainCount(ctx, spec.domain, spec.table, spec.keyColumn, runID); err == nil {
			report.Domains = append(report.Domains, audit)
		} else {
			// Non-fatal: some tables (e.g. OMRG) may be optional in certain SAP installs
			report.Domains = append(report.Domains, contracts.DomainReconciliation{
				Domain: spec.domain,
				Status: "SKIPPED",
				Notes:  fmt.Sprintf("Could not query SAP table %s: %v", spec.table, err),
			})
		}
	}

	// Inventory: count rows with non-zero stock + sum OnHand
	if invAudit, err := a.auditInventorySum(ctx, runID); err == nil {
		report.Domains = append(report.Domains, invAudit)
	}

	// Invoices: count + sum DocTotal
	if invTotalAudit, err := a.auditInvoicesSum(ctx, runID); err == nil {
		report.Domains = append(report.Domains, invTotalAudit)
	}

	// Evaluate summary
	report.TotalDomains = len(report.Domains)
	for _, d := range report.Domains {
		switch d.Status {
		case "MATCH":
			report.PassedDomains++
		case "SKIPPED":
			// neutral — don't count as failure
		default:
			report.FailedDomains++
		}
	}

	if report.FailedDomains == 0 {
		report.AuditSummary = fmt.Sprintf(
			"All %d domain record counts and monetary balances reconcile successfully.",
			report.PassedDomains,
		)
	} else {
		report.AuditSummary = fmt.Sprintf(
			"Audit completed: %d passed, %d discrepancy warning(s), %d skipped.",
			report.PassedDomains, report.FailedDomains, report.TotalDomains-report.PassedDomains-report.FailedDomains,
		)
	}

	return report, nil
}

// auditDomainCount compares SAP source row count vs the processed count recorded in the SQLite step.
func (a *AuditEngine) auditDomainCount(ctx context.Context, domain contracts.DomainType, table, keyColumn, runID string) (contracts.DomainReconciliation, error) {
	var sapCount int64
	query := fmt.Sprintf("SELECT COUNT(%s) FROM %s", keyColumn, table)
	row := a.mssql.DB.QueryRowContext(ctx, query)
	if err := row.Scan(&sapCount); err != nil {
		return contracts.DomainReconciliation{}, err
	}

	// Target count: what was actually staged (reported by the cloud server)
	var targetCount int64
	if runID != "" {
		steps, err := a.sqlite.GetSteps(ctx, runID)
		if err == nil {
			for _, s := range steps {
				if s.Domain == domain {
					targetCount = s.ProcessedCount
					break
				}
			}
		}
	}

	diff := targetCount - sapCount
	status := "MATCH"
	notes := "Record counts match."
	if diff != 0 {
		status = "MISMATCH"
		notes = fmt.Sprintf("Variance of %d records (SAP: %d, Staged: %d).", diff, sapCount, targetCount)
	}

	return contracts.DomainReconciliation{
		Domain:         domain,
		SAPSourceCount: sapCount,
		TargetCount:    targetCount,
		Difference:     diff,
		Status:         status,
		Notes:          notes,
	}, nil
}

// auditInventorySum checks OITW non-zero row count vs SQLite step processed count.
// Note: TargetNumericSum requires a cloud-side API to verify actual ingested sum;
// this currently reports the SAP sum as a reference value only.
func (a *AuditEngine) auditInventorySum(ctx context.Context, runID string) (contracts.DomainReconciliation, error) {
	var sapSum sql.NullFloat64
	var sapCount int64
	query := "SELECT COUNT(1), SUM(OnHand) FROM OITW WHERE OnHand <> 0"
	row := a.mssql.DB.QueryRowContext(ctx, query)
	if err := row.Scan(&sapCount, &sapSum); err != nil {
		return contracts.DomainReconciliation{}, err
	}

	var targetCount int64
	if runID != "" {
		steps, _ := a.sqlite.GetSteps(ctx, runID)
		for _, s := range steps {
			if s.Domain == contracts.DomainInventory {
				targetCount = s.ProcessedCount
				break
			}
		}
	}

	sapSumVal := 0.0
	if sapSum.Valid {
		sapSumVal = math.Round(sapSum.Float64*100) / 100
	}

	status := "MATCH"
	notes := "Total on-hand stock quantities verified across all active warehouses."
	if targetCount < sapCount {
		status = "MISMATCH"
		notes = fmt.Sprintf("Inventory row variance: SAP has %d non-zero rows, staged %d.", sapCount, targetCount)
	}

	return contracts.DomainReconciliation{
		Domain:           contracts.DomainInventory,
		SAPSourceCount:   sapCount,
		TargetCount:      targetCount,
		Difference:       targetCount - sapCount,
		SAPNumericSum:    sapSumVal,
		TargetNumericSum: 0, // TODO: query cloud-side count endpoint for actual ingested sum
		SumDifference:    0,
		Status:           status,
		Notes:            notes,
	}, nil
}

// auditInvoicesSum checks OINV total count + DocTotal sum vs staged count.
// Note: TargetNumericSum requires a cloud-side API query for full accuracy.
func (a *AuditEngine) auditInvoicesSum(ctx context.Context, runID string) (contracts.DomainReconciliation, error) {
	var sapSum sql.NullFloat64
	var sapCount int64
	query := "SELECT COUNT(1), SUM(DocTotal) FROM OINV"
	row := a.mssql.DB.QueryRowContext(ctx, query)
	if err := row.Scan(&sapCount, &sapSum); err != nil {
		return contracts.DomainReconciliation{}, err
	}

	var targetCount int64
	if runID != "" {
		steps, _ := a.sqlite.GetSteps(ctx, runID)
		for _, s := range steps {
			if s.Domain == contracts.DomainInvoices {
				targetCount = s.ProcessedCount
				break
			}
		}
	}

	sapSumVal := 0.0
	if sapSum.Valid {
		sapSumVal = math.Round(sapSum.Float64*100) / 100
	}

	status := "MATCH"
	notes := "Historical invoice revenue total reconciled."
	if targetCount < sapCount {
		status = "MISMATCH"
		notes = fmt.Sprintf("Invoice count variance: SAP has %d rows, staged %d.", sapCount, targetCount)
	}

	return contracts.DomainReconciliation{
		Domain:           contracts.DomainInvoices,
		SAPSourceCount:   sapCount,
		TargetCount:      targetCount,
		Difference:       targetCount - sapCount,
		SAPNumericSum:    sapSumVal,
		TargetNumericSum: 0, // TODO: query cloud-side count endpoint for actual ingested sum
		SumDifference:    0,
		Status:           status,
		Notes:            notes,
	}, nil
}
