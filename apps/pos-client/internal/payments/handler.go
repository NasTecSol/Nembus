package payments

import (
	"net/http"
	"strconv"

	corepayments "github.com/NasTecSol/nembus-core/payments"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler exposes payment coordinator functionality over HTTP / REST.
type Handler struct {
	coordinator *PaymentCoordinator
	store       Store
}

// NewHandler creates a new payments HTTP handler.
func NewHandler(coordinator *PaymentCoordinator, store Store) *Handler {
	return &Handler{
		coordinator: coordinator,
		store:       store,
	}
}

// RegisterRoutes registers terminal payment endpoints on a Gin router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	terminal := rg.Group("/payments/terminal")
	{
		terminal.POST("/sale", h.InitiateSale)
		terminal.GET("/attempts/:id", h.GetAttempt)
		terminal.GET("/unresolved", h.ListUnresolved)
		terminal.POST("/resolve", h.ResolveAttempt)
	}
}

// InitiateSaleRequest is the JSON payload for starting a terminal payment.
type InitiateSaleRequest struct {
	IdempotencyKey   string     `json:"idempotency_key" binding:"required"`
	StoreID          int32      `json:"store_id" binding:"required"`
	PosTerminalID    *int32     `json:"pos_terminal_id,omitempty"`
	CashierSessionID *int32     `json:"cashier_session_id,omitempty"`
	TransactionID    *int32     `json:"transaction_id,omitempty"`
	SalesOrderID     *uuid.UUID `json:"sales_order_id,omitempty"`
	InvoiceRef       string     `json:"invoice_ref" binding:"required"`
	AmountMinor      int64      `json:"amount_minor" binding:"required"`
	Currency         string     `json:"currency"`
	DeviceID         *string    `json:"device_id,omitempty"`
}

// InitiateSale handles POST /api/payments/terminal/sale
func (h *Handler) InitiateSale(c *gin.Context) {
	var req InitiateSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	attempt, err := h.coordinator.InitiateSale(c.Request.Context(), InitiateSaleParams{
		IdempotencyKey:   req.IdempotencyKey,
		StoreID:          req.StoreID,
		PosTerminalID:    req.PosTerminalID,
		CashierSessionID: req.CashierSessionID,
		TransactionID:    req.TransactionID,
		SalesOrderID:     req.SalesOrderID,
		InvoiceRef:       req.InvoiceRef,
		AmountMinor:      req.AmountMinor,
		Currency:         req.Currency,
		DeviceID:         req.DeviceID,
	})

	if err != nil {
		status := http.StatusInternalServerError
		if errorsIs(err, ErrOpenAttemptExists) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{
			"error":   err.Error(),
			"attempt": attempt,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"attempt": attempt,
	})
}

// GetAttempt handles GET /api/payments/terminal/attempts/:id
func (h *Handler) GetAttempt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attempt id"})
		return
	}

	attempt, err := h.store.GetAttempt(c.Request.Context(), id)
	if err != nil {
		if errorsIs(err, ErrAttemptNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "attempt not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"attempt": attempt})
}

// ListUnresolved handles GET /api/payments/terminal/unresolved
func (h *Handler) ListUnresolved(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	var storeID int32
	if storeIDStr != "" {
		if val, err := strconv.Atoi(storeIDStr); err == nil {
			storeID = int32(val)
		}
	}

	attempts, err := h.store.ListUnresolvedAttempts(c.Request.Context(), storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"attempts": attempts})
}

// ResolveAttemptRequest is the JSON payload for resolving an unknown attempt.
type ResolveAttemptRequest struct {
	AttemptID  string             `json:"attempt_id" binding:"required"`
	Resolution corepayments.Status `json:"resolution" binding:"required"`
	Reason     string             `json:"reason" binding:"required"`
	ManualRef  string             `json:"manual_reference"`
}

// ResolveAttempt handles POST /api/payments/terminal/resolve
func (h *Handler) ResolveAttempt(c *gin.Context) {
	var req ResolveAttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(req.AttemptID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attempt id"})
		return
	}

	attempt, err := h.coordinator.ResolveAttempt(c.Request.Context(), id, req.Resolution, req.Reason, req.ManualRef)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "attempt": attempt})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "resolved", "attempt": attempt})
}

func errorsIs(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		if x, ok := err.(interface{ Unwrap() error }); ok {
			err = x.Unwrap()
		} else {
			return false
		}
	}
	return false
}
