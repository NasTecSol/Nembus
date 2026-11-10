package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/jackc/pgx/v5"
)

// ZReport is calculated on request; Finalized describes session closure, not an
// immutable saved report. Money is returned as decimal strings.
type ZReport struct {
	ReportType         string                `json:"report_type"`
	Finalized          bool                  `json:"finalized"`
	GeneratedAt        string                `json:"generated_at"`
	CalculationVersion int                   `json:"calculation_version"`
	Currency           *string               `json:"currency"`
	Store              ZReportStore          `json:"store"`
	Terminal           ZReportIdentity       `json:"terminal"`
	Cashier            ZReportIdentity       `json:"cashier"`
	Session            ZReportSession        `json:"session"`
	Sales              ZReportSales          `json:"sales"`
	Voids              ZReportVoids          `json:"voids"`
	Returns            ZReportReturns        `json:"returns"`
	Net                ZReportNet            `json:"net"`
	Payments           []ZReportPayment      `json:"payments"`
	PaymentTotals      ZReportPaymentTotals  `json:"payment_totals"`
	Drawer             ZReportDrawer         `json:"drawer"`
	Reconciliation     ZReportReconciliation `json:"reconciliation"`
}

type ZReportStore struct {
	ID       int32   `json:"id"`
	Name     string  `json:"name"`
	Timezone *string `json:"timezone"`
}

type ZReportIdentity struct {
	ID   int32   `json:"id"`
	Code string  `json:"code"`
	Name *string `json:"name"`
}

type ZReportSession struct {
	ID       int32   `json:"id"`
	Number   string  `json:"number"`
	Status   *string `json:"status"`
	OpenedAt string  `json:"opened_at"`
	ClosedAt *string `json:"closed_at"`
}

type ZReportSales struct {
	TransactionCount           int64  `json:"transaction_count"`
	BeforeDiscountExcludingTax string `json:"before_discount_excluding_tax"`
	Discounts                  string `json:"discounts"`
	AfterDiscountExcludingTax  string `json:"after_discount_excluding_tax"`
	Tax                        string `json:"tax"`
	TotalIncludingTax          string `json:"total_including_tax"`
}

type ZReportVoids struct {
	TransactionCount  int64  `json:"transaction_count"`
	TotalIncludingTax string `json:"total_including_tax"`
}

type ZReportReturns struct {
	ReturnCount   int64  `json:"return_count"`
	ExcludingTax  string `json:"excluding_tax"`
	Tax           string `json:"tax"`
	TotalRefunded string `json:"total_refunded"`
}

type ZReportNet struct {
	SalesExcludingTax string `json:"sales_excluding_tax"`
	Tax               string `json:"tax"`
	TotalIncludingTax string `json:"total_including_tax"`
}

type ZReportPayment struct {
	Method       string `json:"method"`
	PaymentCount int64  `json:"payment_count"`
	Received     string `json:"received"`
	ChangeGiven  string `json:"change_given"`
	Refunded     string `json:"refunded"`
	Net          string `json:"net"`
}

type ZReportPaymentTotals struct {
	Received    string `json:"received"`
	ChangeGiven string `json:"change_given"`
	Refunded    string `json:"refunded"`
	Net         string `json:"net"`
}

type ZReportDrawer struct {
	OpeningCash             string  `json:"opening_cash"`
	CashReceived            string  `json:"cash_received"`
	ChangeGiven             string  `json:"change_given"`
	CashRefunded            string  `json:"cash_refunded"`
	ManualMovementsIncluded bool    `json:"manual_movements_included"`
	CalculatedExpectedCash  string  `json:"calculated_expected_cash"`
	RecordedExpectedCash    *string `json:"recorded_expected_cash"`
	CountedClosingCash      *string `json:"counted_closing_cash"`
	Variance                *string `json:"variance"`
}

type ZReportWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ZReportReconciliation struct {
	SalesPaymentDifference string           `json:"sales_payment_difference"`
	ExpectedCashDifference *string          `json:"expected_cash_difference"`
	Warnings               []ZReportWarning `json:"warnings"`
}

type ZReportRepository interface {
	GetCashierSessionZReport(context.Context, repository.GetCashierSessionZReportParams) (json.RawMessage, error)
}

// CalculateCashierSessionZReport performs no writes. The SQL statement enforces
// organization and user store access and reads all sections in one snapshot.
func CalculateCashierSessionZReport(ctx context.Context, repo ZReportRepository, filters repository.GetCashierSessionZReportParams) *repository.Response {
	if filters.OrganizationID <= 0 || filters.UserID <= 0 {
		return utils.NewResponse(utils.CodeUnauthorized, "authenticated organization and user are required", nil)
	}
	if filters.CashierSessionID <= 0 || filters.StoreID < 0 || filters.PosTerminalID < 0 || filters.CashierID < 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid session or report filters", nil)
	}
	if repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	payload, err := repo.GetCashierSessionZReport(ctx, filters)
	if errors.Is(err, pgx.ErrNoRows) {
		return utils.NewResponse(utils.CodeNotFound, "session not found, inaccessible, or filters do not match", nil)
	}
	if err != nil {
		log.Printf("calculate cashier session Z-report: %v", err)
		return utils.NewResponse(utils.CodeError, "failed to calculate session Z-report", nil)
	}
	var report ZReport
	if err := json.Unmarshal(payload, &report); err != nil {
		log.Printf("decode cashier session Z-report: %v", err)
		return utils.NewResponse(utils.CodeError, "failed to decode session Z-report", nil)
	}
	return utils.NewResponse(utils.CodeOK, "session Z-report calculated successfully", report)
}
