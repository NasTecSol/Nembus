package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterAccountingRoutes registers all double-entry accounting engine routes
// under /api/accounting.
func RegisterAccountingRoutes(r *gin.RouterGroup, h *handler.AccountingHandler) {
	acc := r.Group("/accounting")
	{
		// ── Fiscal Calendar ──────────────────────────────────────────
		acc.POST("/fiscal-years", h.OpenFiscalYear)
		acc.GET("/fiscal-years", h.ListFiscalYears)

		acc.GET("/posting-periods", h.ListPostingPeriods)
		acc.PUT("/posting-periods/:id/close", h.ClosePeriod)

		// ── Document Series ──────────────────────────────────────────
		acc.GET("/document-series", h.ListDocumentSeries)

		// ── Journal Entries ──────────────────────────────────────────
		acc.GET("/journal-entries", h.ListJournalEntries)
		acc.POST("/journal-entries/:id/reverse", h.ReverseJournalEntry)

		// ── Financial Reports ────────────────────────────────────────
		acc.GET("/trial-balance", h.GetTrialBalance)
		acc.GET("/profit-loss", h.GetProfitAndLoss)
		acc.GET("/balance-sheet", h.GetBalanceSheet)
		acc.GET("/ledger/:account_id", h.GetAccountLedger)

		// ── GL Posting Rules ─────────────────────────────────────────
		acc.GET("/gl-posting-rules", h.ListGLPostingRules)
		acc.PUT("/gl-posting-rules/:posting_type", h.UpsertGLPostingRule)

		// ── Profit Centers ───────────────────────────────────────────
		acc.GET("/profit-centers", h.ListProfitCenters)

		// ── WHT Entries ──────────────────────────────────────────────
		acc.GET("/wht-entries", h.ListWHTEntries)
	}
}
