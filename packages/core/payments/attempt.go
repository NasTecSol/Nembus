// Package payments holds the provider-neutral rules for card-terminal payment
// attempts. It is shared by the POS client (which creates attempts) and the
// cloud server (which is authoritative about what may be accepted), so both
// sides apply the same state machine and validation.
//
// Nothing here knows about any terminal vendor's wire protocol.
package payments

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Status is the lifecycle state of a payment attempt.
type Status string

const (
	StatusPending    Status = "pending"    // created locally, terminal not contacted
	StatusProcessing Status = "processing" // request may have reached the terminal
	StatusUnknown    Status = "unknown"    // outcome cannot be determined; needs manual inquiry
	StatusSucceeded  Status = "succeeded"  // terminal reported approval
	StatusDeclined   Status = "declined"   // terminal reported decline
	StatusCancelled  Status = "cancelled"  // cancelled by user/system with certainty that no charge occurred
)

// Type separates sales from refunds and reversals (distinct operations).
type Type string

const (
	TypeSale     Type = "sale"
	TypeRefund   Type = "refund"
	TypeReversal Type = "reversal"
)

// CurrencySAR is the only currency accepted for Saudi merchants.
const CurrencySAR = "SAR"

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusProcessing, StatusUnknown, StatusSucceeded, StatusDeclined, StatusCancelled:
		return true
	}
	return false
}

// IsFinal reports whether no further state change is legitimate.
func (s Status) IsFinal() bool {
	return s == StatusSucceeded || s == StatusDeclined || s == StatusCancelled
}

func (t Type) Valid() bool { return t == TypeSale || t == TypeRefund || t == TypeReversal }

// IsOpen reports a state in which another attempt for the same invoice must
// not be started (money may be in flight).
func (s Status) IsOpen() bool {
	return s == StatusPending || s == StatusProcessing || s == StatusUnknown
}

// CanTransition is the strict local state machine used by the coordinator.
func CanTransition(from, to Status) bool {
	switch from {
	case StatusPending:
		return to == StatusProcessing || to == StatusCancelled
	case StatusProcessing:
		return to == StatusSucceeded || to == StatusDeclined || to == StatusCancelled || to == StatusUnknown
	case StatusUnknown:
		// Only an explicit resolution (inquiry / manual reconcile) may close it.
		return to == StatusSucceeded || to == StatusDeclined || to == StatusCancelled
	}
	return false
}

func rank(s Status) int {
	switch s {
	case StatusPending:
		return 0
	case StatusProcessing:
		return 1
	case StatusUnknown:
		return 2
	}
	return 3 // final states
}

// Decision is the outcome of merging an incoming status into stored state.
type Decision int

const (
	// Apply the incoming state.
	Apply Decision = iota
	// Ignore the incoming state (duplicate or stale); acknowledge it.
	Ignore
	// Conflict means a final state would be replaced by a different final
	// state. Never apply; surface for investigation.
	Conflict
)

// Merge decides how the cloud treats an incoming status given what it already
// stores. It is monotonic: state never regresses, and a final state is never
// replaced. It tolerates skipped states (events may coalesce or arrive late).
func Merge(current, incoming Status) Decision {
	switch {
	case current == incoming:
		return Ignore
	case current.IsFinal():
		if incoming.IsFinal() {
			return Conflict
		}
		return Ignore // stale non-final event after a final one
	case rank(incoming) > rank(current):
		return Apply
	}
	return Ignore
}

// Attempt mirrors a payment_attempts row. JSON tags equal the column names so
// it round-trips through the sync outbox payload (row_to_json).
type Attempt struct {
	ID                uuid.UUID         `json:"id"`
	IdempotencyKey    string            `json:"idempotency_key"`
	AttemptType       Type              `json:"attempt_type"`
	OriginalAttemptID *uuid.UUID        `json:"original_attempt_id"`
	StoreID           int32             `json:"store_id"`
	PosTerminalID     *int32            `json:"pos_terminal_id"`
	CashierSessionID  *int32            `json:"cashier_session_id"`
	TransactionID     *int32            `json:"transaction_id"`
	SalesOrderID      *uuid.UUID        `json:"sales_order_id"`
	InvoiceRef        string            `json:"invoice_ref"`
	AmountMinor       int64             `json:"amount_minor"`
	Currency          string            `json:"currency"`
	Status            Status            `json:"status"`
	Provider          string            `json:"provider"`
	ProviderRefs      map[string]string `json:"provider_refs"`
	MaskedPAN         *string           `json:"masked_pan"`
	RawRedacted       json.RawMessage   `json:"raw_redacted"`
	DeviceID          *string           `json:"device_id"`
	CreatedAt         *time.Time        `json:"created_at,omitempty"`
	UpdatedAt         *time.Time        `json:"updated_at,omitempty"`
}

// Validate checks structural invariants that hold for every attempt in every
// state. It does not consult any database.
func (a *Attempt) Validate() error {
	var errs []string
	if a.ID == uuid.Nil {
		errs = append(errs, "id is required")
	}
	if strings.TrimSpace(a.IdempotencyKey) == "" {
		errs = append(errs, "idempotency_key is required")
	}
	if !a.AttemptType.Valid() {
		errs = append(errs, "invalid attempt_type")
	}
	if a.AttemptType != TypeSale && a.OriginalAttemptID == nil {
		errs = append(errs, "refund/reversal must reference original_attempt_id")
	}
	if a.StoreID <= 0 {
		errs = append(errs, "store_id is required")
	}
	if strings.TrimSpace(a.InvoiceRef) == "" {
		errs = append(errs, "invoice_ref is required")
	}
	if a.AmountMinor <= 0 {
		errs = append(errs, "amount_minor must be positive")
	}
	if a.Currency != CurrencySAR {
		errs = append(errs, "currency must be SAR")
	}
	if !a.Status.Valid() {
		errs = append(errs, "invalid status")
	}
	if strings.TrimSpace(a.Provider) == "" {
		errs = append(errs, "provider is required")
	}
	if a.Status == StatusSucceeded && len(a.ProviderRefs) == 0 {
		errs = append(errs, "a succeeded attempt requires at least one provider reference")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// ImmutableMatches verifies that the fields which must never change after
// creation are identical between the stored and the incoming attempt.
func ImmutableMatches(stored, in *Attempt) error {
	switch {
	case stored.IdempotencyKey != in.IdempotencyKey:
		return errors.New("idempotency_key changed")
	case stored.AttemptType != in.AttemptType:
		return errors.New("attempt_type changed")
	case stored.StoreID != in.StoreID:
		return errors.New("store_id changed")
	case stored.InvoiceRef != in.InvoiceRef:
		return errors.New("invoice_ref changed")
	case stored.AmountMinor != in.AmountMinor:
		return errors.New("amount_minor changed")
	case stored.Currency != in.Currency:
		return errors.New("currency changed")
	case stored.Provider != in.Provider:
		return errors.New("provider changed")
	}
	return nil
}

// MinorToDecimal renders halalas as a two-decimal string without floats.
func MinorToDecimal(minor int64) string {
	sign := ""
	if minor < 0 {
		sign, minor = "-", -minor
	}
	return fmt.Sprintf("%s%d.%02d", sign, minor/100, minor%100)
}

// DecimalToMinor parses a non-negative decimal with at most two fractional
// digits ("12", "12.5", "12.50") into halalas without floats.
func DecimalToMinor(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" || strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+") {
		return 0, fmt.Errorf("invalid amount %q", s)
	}
	whole, frac, hasDot := strings.Cut(s, ".")
	if whole == "" || (hasDot && frac == "") || len(frac) > 2 {
		return 0, fmt.Errorf("invalid amount %q", s)
	}
	frac += strings.Repeat("0", 2-len(frac))
	var w, f int64
	for _, ch := range whole {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid amount %q", s)
		}
		if w > (1<<62)/10 {
			return 0, fmt.Errorf("amount %q too large", s)
		}
		w = w*10 + int64(ch-'0')
	}
	for _, ch := range frac {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid amount %q", s)
		}
		f = f*10 + int64(ch-'0')
	}
	return w*100 + f, nil
}

// Policy decides which payment rows must be backed by a succeeded attempt.
type Policy struct {
	// TerminalGateways are payment_gateway values that identify terminal
	// payments (compared case-insensitively).
	TerminalGateways []string
	// EnforcedMethods are payment_method values (e.g. "card", "mada") that
	// must be backed by an attempt even when no gateway is set. Empty = off.
	EnforcedMethods []string
}

// DefaultPolicy treats the Geidea provider as a terminal gateway.
func DefaultPolicy() Policy { return Policy{TerminalGateways: []string{ProviderGeidea}} }

// Provider identifiers stored in payment_attempts.provider and
// pos_payments.payment_gateway.
const ProviderGeidea = "geidea"

func contains(list []string, v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return false
	}
	for _, x := range list {
		if strings.ToLower(strings.TrimSpace(x)) == v {
			return true
		}
	}
	return false
}

// IsTerminalGateway reports whether a gateway name denotes a terminal payment.
func (p Policy) IsTerminalGateway(gateway string) bool { return contains(p.TerminalGateways, gateway) }

// RequiresAttempt reports whether a pos_payments row must be verified against
// a succeeded payment attempt.
func (p Policy) RequiresAttempt(method, gateway string) bool {
	return p.IsTerminalGateway(gateway) || contains(p.EnforcedMethods, method)
}

// ParseList splits a comma-separated configuration value.
func ParseList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
