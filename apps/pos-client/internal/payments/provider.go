package payments

import (
	"context"
	"encoding/json"
	"errors"

	corepayments "github.com/NasTecSol/nembus-core/payments"
	"github.com/google/uuid"
)

var (
	// ErrTerminalBusy indicates the physical terminal is currently busy with another operation.
	ErrTerminalBusy = errors.New("terminal is busy with another transaction")
	// ErrCommunication indicates a network, serial, or bridge communication failure.
	ErrCommunication = errors.New("terminal communication failure")
	// ErrTransactionTimeout indicates terminal did not reply within the allowed timeout.
	ErrTransactionTimeout = errors.New("terminal transaction timed out")
	// ErrOpenAttemptExists indicates an in-flight payment attempt is already active for this invoice.
	ErrOpenAttemptExists = errors.New("an open payment attempt already exists for this invoice")
	// ErrAttemptFinalized indicates the attempt is already in a final state and cannot be modified.
	ErrAttemptFinalized = errors.New("payment attempt is already in a final state")
	// ErrInvalidResolution indicates an invalid transition for resolving an unknown attempt.
	ErrInvalidResolution = errors.New("invalid attempt resolution")
)

// SaleRequest contains parameters sent to a terminal provider to execute a card payment.
type SaleRequest struct {
	AttemptID      uuid.UUID
	IdempotencyKey string
	InvoiceRef     string
	AmountMinor    int64
	Currency       string
	TerminalID     *int32
}

// SaleResult contains the response from a terminal provider.
type SaleResult struct {
	Status       corepayments.Status // StatusSucceeded, StatusDeclined, or StatusUnknown
	ProviderRefs map[string]string
	MaskedPAN    *string
	RawRedacted  json.RawMessage
	ErrorMessage string
}

// TerminalProvider defines the pluggable contract for card terminal integrations (Geidea, mock, etc.).
type TerminalProvider interface {
	// Name returns the provider identifier (e.g. "geidea", "mock").
	Name() string
	// Sale initiates a synchronous or asynchronous card sale on the physical terminal.
	// If communication fails or times out while money might be in flight, Sale MUST return
	// StatusUnknown so that the transaction is never assumed to have failed.
	Sale(ctx context.Context, req SaleRequest) (SaleResult, error)
}
