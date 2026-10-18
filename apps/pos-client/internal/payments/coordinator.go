package payments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	corepayments "github.com/NasTecSol/nembus-core/payments"
	"github.com/google/uuid"
)

// InitiateSaleParams holds arguments to begin a card sale on a terminal.
type InitiateSaleParams struct {
	IdempotencyKey   string
	StoreID          int32
	PosTerminalID    *int32
	CashierSessionID *int32
	TransactionID    *int32
	SalesOrderID     *uuid.UUID
	InvoiceRef       string
	AmountMinor      int64
	Currency         string
	DeviceID         *string
}

// PaymentCoordinator coordinates payment attempts, enforcing concurrency locks,
// state transitions, and post-approval payment recording.
type PaymentCoordinator struct {
	store    Store
	provider TerminalProvider
	locks    sync.Map // key string -> *sync.Mutex
}

// NewPaymentCoordinator builds a coordinator with the given store and provider.
func NewPaymentCoordinator(store Store, provider TerminalProvider) *PaymentCoordinator {
	return &PaymentCoordinator{
		store:    store,
		provider: provider,
	}
}

func (c *PaymentCoordinator) getLock(key string) *sync.Mutex {
	val, _ := c.locks.LoadOrStore(key, &sync.Mutex{})
	return val.(*sync.Mutex)
}

// InitiateSale starts a physical card terminal payment attempt.
func (c *PaymentCoordinator) InitiateSale(ctx context.Context, p InitiateSaleParams) (*corepayments.Attempt, error) {
	if p.AmountMinor <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}
	if p.Currency == "" {
		p.Currency = corepayments.CurrencySAR
	}
	if p.Currency != corepayments.CurrencySAR {
		return nil, fmt.Errorf("currency %q not supported; expected %s", p.Currency, corepayments.CurrencySAR)
	}
	if p.InvoiceRef == "" {
		return nil, errors.New("invoice reference is required")
	}
	if p.IdempotencyKey == "" {
		return nil, errors.New("idempotency key is required")
	}
	if p.StoreID <= 0 {
		return nil, errors.New("store_id is required")
	}

	// Lock the invoice reference so concurrent attempts cannot be run for the same invoice.
	invoiceLock := c.getLock("invoice:" + p.InvoiceRef)
	invoiceLock.Lock()
	defer invoiceLock.Unlock()

	// Check if an attempt with this idempotency key already exists.
	existing, err := c.store.GetAttemptByIdempotencyKey(ctx, p.IdempotencyKey)
	if err == nil && existing != nil {
		if existing.AmountMinor != p.AmountMinor || existing.InvoiceRef != p.InvoiceRef {
			return nil, errors.New("idempotency key reused with different parameters")
		}
		return existing, nil
	} else if err != nil && !errors.Is(err, ErrAttemptNotFound) {
		return nil, fmt.Errorf("lookup idempotency key: %w", err)
	}

	// Verify no open attempt exists for this invoice.
	open, err := c.store.FindOpenAttemptForInvoice(ctx, p.InvoiceRef)
	if err == nil && open != nil && open.Status.IsOpen() {
		return open, ErrOpenAttemptExists
	} else if err != nil && !errors.Is(err, ErrAttemptNotFound) {
		return nil, fmt.Errorf("check open attempts: %w", err)
	}

	// Lock the physical terminal to prevent concurrent hardware access.
	terminalKey := "terminal:default"
	if p.PosTerminalID != nil {
		terminalKey = fmt.Sprintf("terminal:%d", *p.PosTerminalID)
	}
	termLock := c.getLock(terminalKey)
	termLock.Lock()
	defer termLock.Unlock()

	attemptID := uuid.New()
	attempt := &corepayments.Attempt{
		ID:               attemptID,
		IdempotencyKey:   p.IdempotencyKey,
		AttemptType:      corepayments.TypeSale,
		StoreID:          p.StoreID,
		PosTerminalID:    p.PosTerminalID,
		CashierSessionID: p.CashierSessionID,
		TransactionID:    p.TransactionID,
		SalesOrderID:     p.SalesOrderID,
		InvoiceRef:       p.InvoiceRef,
		AmountMinor:      p.AmountMinor,
		Currency:         p.Currency,
		Status:           corepayments.StatusPending,
		Provider:         c.provider.Name(),
		ProviderRefs:     map[string]string{},
		DeviceID:         p.DeviceID,
	}

	// Step 1: Persist pending attempt in database
	if err := c.store.CreateAttempt(ctx, attempt); err != nil {
		return nil, fmt.Errorf("record pending attempt: %w", err)
	}

	// Step 2: Transition to processing before contacting terminal
	if err := c.store.UpdateAttemptStatus(ctx, attempt.ID, corepayments.StatusProcessing, nil, nil, nil); err != nil {
		return nil, fmt.Errorf("transition to processing: %w", err)
	}
	attempt.Status = corepayments.StatusProcessing

	// Step 3: Invoke the terminal provider
	saleReq := SaleRequest{
		AttemptID:      attempt.ID,
		IdempotencyKey: attempt.IdempotencyKey,
		InvoiceRef:     attempt.InvoiceRef,
		AmountMinor:    attempt.AmountMinor,
		Currency:       attempt.Currency,
		TerminalID:     attempt.PosTerminalID,
	}

	result, termErr := c.provider.Sale(ctx, saleReq)

	// Step 4: Handle outcomes strictly based on state machine
	if termErr != nil || result.Status == corepayments.StatusUnknown {
		// Communication drop or timeout: MUST transition to UNKNOWN, NEVER auto-retry or assume declined.
		_ = c.store.UpdateAttemptStatus(ctx, attempt.ID, corepayments.StatusUnknown, result.ProviderRefs, result.MaskedPAN, result.RawRedacted)
		attempt.Status = corepayments.StatusUnknown
		attempt.ProviderRefs = result.ProviderRefs
		attempt.MaskedPAN = result.MaskedPAN
		attempt.RawRedacted = result.RawRedacted

		if termErr != nil {
			return attempt, fmt.Errorf("terminal sale status unknown: %w", termErr)
		}
		return attempt, ErrTransactionTimeout
	}

	switch result.Status {
	case corepayments.StatusSucceeded:
		if err := c.store.UpdateAttemptStatus(ctx, attempt.ID, corepayments.StatusSucceeded, result.ProviderRefs, result.MaskedPAN, result.RawRedacted); err != nil {
			return nil, fmt.Errorf("record succeeded status: %w", err)
		}
		attempt.Status = corepayments.StatusSucceeded
		attempt.ProviderRefs = result.ProviderRefs
		attempt.MaskedPAN = result.MaskedPAN
		attempt.RawRedacted = result.RawRedacted

		// Authoritatively insert pos_payments row locally and update amount_paid
		amountDecimal := corepayments.MinorToDecimal(attempt.AmountMinor)
		refNum := ""
		if rrn, ok := result.ProviderRefs["rrn"]; ok && rrn != "" {
			refNum = rrn
		} else if auth, ok := result.ProviderRefs["auth_code"]; ok && auth != "" {
			refNum = auth
		} else {
			refNum = attempt.ID.String()
		}

		meta, _ := json.Marshal(map[string]any{
			"payment_attempt_id": attempt.ID,
			"masked_pan":         attempt.MaskedPAN,
			"provider_refs":      attempt.ProviderRefs,
			"provider":           attempt.Provider,
		})

		if err := c.store.RecordApprovedPayment(ctx, attempt, amountDecimal, refNum, meta); err != nil {
			// Database record failed after terminal approved: flag for attention
			return attempt, fmt.Errorf("payment approved by terminal but failed to record locally: %w", err)
		}
		return attempt, nil

	case corepayments.StatusDeclined:
		_ = c.store.UpdateAttemptStatus(ctx, attempt.ID, corepayments.StatusDeclined, result.ProviderRefs, result.MaskedPAN, result.RawRedacted)
		attempt.Status = corepayments.StatusDeclined
		attempt.ProviderRefs = result.ProviderRefs
		attempt.MaskedPAN = result.MaskedPAN
		attempt.RawRedacted = result.RawRedacted
		return attempt, nil

	case corepayments.StatusCancelled:
		_ = c.store.UpdateAttemptStatus(ctx, attempt.ID, corepayments.StatusCancelled, result.ProviderRefs, result.MaskedPAN, result.RawRedacted)
		attempt.Status = corepayments.StatusCancelled
		return attempt, nil

	default:
		_ = c.store.UpdateAttemptStatus(ctx, attempt.ID, corepayments.StatusUnknown, result.ProviderRefs, result.MaskedPAN, result.RawRedacted)
		attempt.Status = corepayments.StatusUnknown
		return attempt, errors.New("unrecognized terminal outcome; set to unknown")
	}
}

// ResolveAttempt resolves an attempt in 'unknown' status following manual inquiry or portal verification.
func (c *PaymentCoordinator) ResolveAttempt(ctx context.Context, attemptID uuid.UUID, resolution corepayments.Status, reason string, manualRef string) (*corepayments.Attempt, error) {
	attempt, err := c.store.GetAttempt(ctx, attemptID)
	if err != nil {
		return nil, fmt.Errorf("get attempt: %w", err)
	}

	if attempt.Status.IsFinal() {
		return attempt, ErrAttemptFinalized
	}

	if !corepayments.CanTransition(attempt.Status, resolution) {
		return attempt, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidResolution, attempt.Status, resolution)
	}

	refs := attempt.ProviderRefs
	if refs == nil {
		refs = make(map[string]string)
	}
	refs["resolution_reason"] = reason
	if manualRef != "" {
		refs["resolution_reference"] = manualRef
	}

	if resolution == corepayments.StatusSucceeded {
		if manualRef == "" {
			return nil, errors.New("resolution to succeeded requires a manual reference (terminal slip RRN or approval code)")
		}
		if _, ok := refs["rrn"]; !ok {
			refs["rrn"] = manualRef
		}

		if err := c.store.UpdateAttemptStatus(ctx, attempt.ID, corepayments.StatusSucceeded, refs, attempt.MaskedPAN, attempt.RawRedacted); err != nil {
			return nil, fmt.Errorf("update status to succeeded: %w", err)
		}
		attempt.Status = corepayments.StatusSucceeded
		attempt.ProviderRefs = refs

		amountDecimal := corepayments.MinorToDecimal(attempt.AmountMinor)
		meta, _ := json.Marshal(map[string]any{
			"payment_attempt_id":  attempt.ID,
			"masked_pan":          attempt.MaskedPAN,
			"provider_refs":       refs,
			"provider":            attempt.Provider,
			"resolved_manually":   true,
			"resolution_reason":   reason,
		})

		if err := c.store.RecordApprovedPayment(ctx, attempt, amountDecimal, manualRef, meta); err != nil {
			return attempt, fmt.Errorf("record approved payment: %w", err)
		}
		return attempt, nil
	}

	// Declined or Cancelled
	if err := c.store.UpdateAttemptStatus(ctx, attempt.ID, resolution, refs, attempt.MaskedPAN, attempt.RawRedacted); err != nil {
		return nil, fmt.Errorf("update status to %s: %w", resolution, err)
	}
	attempt.Status = resolution
	attempt.ProviderRefs = refs
	return attempt, nil
}
