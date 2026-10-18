package payments_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/NasTecSol/nembus-client/internal/payments"
	"github.com/NasTecSol/nembus-client/internal/payments/fake"
	corepayments "github.com/NasTecSol/nembus-core/payments"
	"github.com/google/uuid"
)

// memStore is an in-memory implementation of payments.Store for testing.
type memStore struct {
	mu           sync.Mutex
	attempts     map[uuid.UUID]*corepayments.Attempt
	idempIndex   map[string]uuid.UUID
	openIndex    map[string]uuid.UUID
	recordedPmts []recordedPayment
}

type recordedPayment struct {
	AttemptID     uuid.UUID
	AmountDecimal string
	RefNum        string
	Metadata      []byte
}

func newMemStore() *memStore {
	return &memStore{
		attempts:   make(map[uuid.UUID]*corepayments.Attempt),
		idempIndex: make(map[string]uuid.UUID),
		openIndex:  make(map[string]uuid.UUID),
	}
}

func (m *memStore) CreateAttempt(ctx context.Context, a *corepayments.Attempt) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.attempts[a.ID]; exists {
		return errors.New("attempt already exists")
	}
	cp := *a
	m.attempts[a.ID] = &cp
	m.idempIndex[a.IdempotencyKey] = a.ID
	if a.Status.IsOpen() {
		m.openIndex[a.InvoiceRef] = a.ID
	}
	return nil
}

func (m *memStore) GetAttempt(ctx context.Context, id uuid.UUID) (*corepayments.Attempt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.attempts[id]
	if !ok {
		return nil, payments.ErrAttemptNotFound
	}
	cp := *a
	return &cp, nil
}

func (m *memStore) GetAttemptByIdempotencyKey(ctx context.Context, key string) (*corepayments.Attempt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.idempIndex[key]
	if !ok {
		return nil, payments.ErrAttemptNotFound
	}
	cp := *m.attempts[id]
	return &cp, nil
}

func (m *memStore) FindOpenAttemptForInvoice(ctx context.Context, invoiceRef string) (*corepayments.Attempt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.openIndex[invoiceRef]
	if !ok {
		return nil, payments.ErrAttemptNotFound
	}
	a := m.attempts[id]
	if !a.Status.IsOpen() {
		delete(m.openIndex, invoiceRef)
		return nil, payments.ErrAttemptNotFound
	}
	cp := *a
	return &cp, nil
}

func (m *memStore) UpdateAttemptStatus(ctx context.Context, id uuid.UUID, newStatus corepayments.Status, providerRefs map[string]string, maskedPAN *string, rawRedacted json.RawMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.attempts[id]
	if !ok {
		return payments.ErrAttemptNotFound
	}
	a.Status = newStatus
	if providerRefs != nil {
		if a.ProviderRefs == nil {
			a.ProviderRefs = make(map[string]string)
		}
		for k, v := range providerRefs {
			a.ProviderRefs[k] = v
		}
	}
	if maskedPAN != nil {
		a.MaskedPAN = maskedPAN
	}
	if len(rawRedacted) > 0 {
		a.RawRedacted = rawRedacted
	}

	if !newStatus.IsOpen() {
		delete(m.openIndex, a.InvoiceRef)
	} else {
		m.openIndex[a.InvoiceRef] = a.ID
	}
	return nil
}

func (m *memStore) ListUnresolvedAttempts(ctx context.Context, storeID int32) ([]*corepayments.Attempt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*corepayments.Attempt
	for _, a := range m.attempts {
		if a.Status == corepayments.StatusUnknown && (storeID == 0 || a.StoreID == storeID) {
			cp := *a
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *memStore) MarkOrphanProcessingAsUnknown(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var count int64
	for _, a := range m.attempts {
		if a.Status == corepayments.StatusProcessing {
			a.Status = corepayments.StatusUnknown
			count++
		}
	}
	return count, nil
}

func (m *memStore) RecordApprovedPayment(ctx context.Context, attempt *corepayments.Attempt, amountDecimal string, refNum string, meta []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordedPmts = append(m.recordedPmts, recordedPayment{
		AttemptID:     attempt.ID,
		AmountDecimal: amountDecimal,
		RefNum:        refNum,
		Metadata:      meta,
	})
	return nil
}

func TestInitiateSale_Succeeded(t *testing.T) {
	store := newMemStore()
	provider := fake.NewFakeTerminalProvider("geidea")
	coordinator := payments.NewPaymentCoordinator(store, provider)

	txnID := int32(101)
	termID := int32(5)
	params := payments.InitiateSaleParams{
		IdempotencyKey: "sale-key-001",
		StoreID:        1,
		PosTerminalID:  &termID,
		TransactionID:  &txnID,
		InvoiceRef:     "INV-2026-0001",
		AmountMinor:    15000, // 150.00 SAR
		Currency:       "SAR",
	}

	attempt, err := coordinator.InitiateSale(context.Background(), params)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if attempt.Status != corepayments.StatusSucceeded {
		t.Fatalf("expected status succeeded, got %s", attempt.Status)
	}
	if attempt.AmountMinor != 15000 {
		t.Fatalf("expected amount 15000, got %d", attempt.AmountMinor)
	}
	if attempt.ProviderRefs["rrn"] != "123456789012" {
		t.Fatalf("missing expected RRN: %+v", attempt.ProviderRefs)
	}

	// Verify payment recorded
	if len(store.recordedPmts) != 1 {
		t.Fatalf("expected 1 recorded payment, got %d", len(store.recordedPmts))
	}
	rec := store.recordedPmts[0]
	if rec.AmountDecimal != "150.00" {
		t.Fatalf("expected amount 150.00, got %s", rec.AmountDecimal)
	}
	if rec.RefNum != "123456789012" {
		t.Fatalf("expected ref 123456789012, got %s", rec.RefNum)
	}
}

func TestInitiateSale_Declined(t *testing.T) {
	store := newMemStore()
	provider := fake.NewFakeTerminalProvider("geidea")
	provider.SetNextResult(payments.SaleResult{
		Status:       corepayments.StatusDeclined,
		ErrorMessage: "Insufficient funds",
		ProviderRefs: map[string]string{"response_code": "51"},
	}, nil)

	coordinator := payments.NewPaymentCoordinator(store, provider)
	params := payments.InitiateSaleParams{
		IdempotencyKey: "sale-key-002",
		StoreID:        1,
		InvoiceRef:     "INV-2026-0002",
		AmountMinor:    5000,
		Currency:       "SAR",
	}

	attempt, err := coordinator.InitiateSale(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected coordinator error: %v", err)
	}
	if attempt.Status != corepayments.StatusDeclined {
		t.Fatalf("expected status declined, got %s", attempt.Status)
	}
	if len(store.recordedPmts) != 0 {
		t.Fatalf("declined payment must not be recorded as pos_payments")
	}
}

func TestInitiateSale_TimeoutYieldsUnknown(t *testing.T) {
	store := newMemStore()
	provider := fake.NewFakeTerminalProvider("geidea")
	provider.SetNextResult(payments.SaleResult{Status: corepayments.StatusUnknown}, payments.ErrTransactionTimeout)

	coordinator := payments.NewPaymentCoordinator(store, provider)
	params := payments.InitiateSaleParams{
		IdempotencyKey: "sale-key-003",
		StoreID:        1,
		InvoiceRef:     "INV-2026-0003",
		AmountMinor:    7500,
		Currency:       "SAR",
	}

	attempt, err := coordinator.InitiateSale(context.Background(), params)
	if err == nil {
		t.Fatal("expected error on timeout, got nil")
	}
	if attempt == nil {
		t.Fatal("attempt must be returned even on timeout so caller has tracking id")
	}
	if attempt.Status != corepayments.StatusUnknown {
		t.Fatalf("timeout must transition to UNKNOWN, got %s", attempt.Status)
	}
	if len(store.recordedPmts) != 0 {
		t.Fatal("unknown payment must not be recorded as successful payment")
	}
}

func TestInitiateSale_BlocksConcurrentOpenAttempt(t *testing.T) {
	store := newMemStore()
	provider := fake.NewFakeTerminalProvider("geidea")
	// simulate a hanging attempt that leaves status in unknown
	provider.SetNextResult(payments.SaleResult{Status: corepayments.StatusUnknown}, payments.ErrCommunication)

	coordinator := payments.NewPaymentCoordinator(store, provider)
	params1 := payments.InitiateSaleParams{
		IdempotencyKey: "key-1",
		StoreID:        1,
		InvoiceRef:     "INV-2026-0004",
		AmountMinor:    1000,
		Currency:       "SAR",
	}

	_, _ = coordinator.InitiateSale(context.Background(), params1)

	// Second attempt for same invoice must be blocked
	params2 := payments.InitiateSaleParams{
		IdempotencyKey: "key-2",
		StoreID:        1,
		InvoiceRef:     "INV-2026-0004",
		AmountMinor:    1000,
		Currency:       "SAR",
	}

	_, err := coordinator.InitiateSale(context.Background(), params2)
	if !errors.Is(err, payments.ErrOpenAttemptExists) {
		t.Fatalf("expected ErrOpenAttemptExists, got %v", err)
	}
}

func TestInitiateSale_IdempotentRetry(t *testing.T) {
	store := newMemStore()
	provider := fake.NewFakeTerminalProvider("geidea")
	coordinator := payments.NewPaymentCoordinator(store, provider)

	params := payments.InitiateSaleParams{
		IdempotencyKey: "idemp-001",
		StoreID:        1,
		InvoiceRef:     "INV-2026-0005",
		AmountMinor:    2000,
		Currency:       "SAR",
	}

	att1, err := coordinator.InitiateSale(context.Background(), params)
	if err != nil {
		t.Fatalf("first attempt error: %v", err)
	}

	att2, err := coordinator.InitiateSale(context.Background(), params)
	if err != nil {
		t.Fatalf("idempotent replay error: %v", err)
	}

	if att1.ID != att2.ID {
		t.Fatalf("expected same attempt ID, got %s and %s", att1.ID, att2.ID)
	}
}

func TestResolveAttempt_UnknownToSucceeded(t *testing.T) {
	store := newMemStore()
	provider := fake.NewFakeTerminalProvider("geidea")
	provider.SetNextResult(payments.SaleResult{Status: corepayments.StatusUnknown}, payments.ErrTransactionTimeout)

	coordinator := payments.NewPaymentCoordinator(store, provider)
	txnID := int32(42)
	params := payments.InitiateSaleParams{
		IdempotencyKey: "key-res-1",
		StoreID:        1,
		TransactionID:  &txnID,
		InvoiceRef:     "INV-2026-0006",
		AmountMinor:    3000,
		Currency:       "SAR",
	}

	att, _ := coordinator.InitiateSale(context.Background(), params)
	if att.Status != corepayments.StatusUnknown {
		t.Fatalf("expected unknown, got %s", att.Status)
	}

	// Resolve without reference should fail
	_, err := coordinator.ResolveAttempt(context.Background(), att.ID, corepayments.StatusSucceeded, "Cashier checked paper roll", "")
	if err == nil {
		t.Fatal("expected error when resolving to succeeded without manual ref")
	}

	// Resolve with terminal slip RRN
	resolved, err := coordinator.ResolveAttempt(context.Background(), att.ID, corepayments.StatusSucceeded, "Verified paper slip", "RRN-998877")
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	if resolved.Status != corepayments.StatusSucceeded {
		t.Fatalf("expected succeeded, got %s", resolved.Status)
	}
	if resolved.ProviderRefs["rrn"] != "RRN-998877" {
		t.Fatalf("missing RRN in resolved attempt: %+v", resolved.ProviderRefs)
	}

	if len(store.recordedPmts) != 1 {
		t.Fatalf("expected 1 recorded payment after manual resolution, got %d", len(store.recordedPmts))
	}
	if store.recordedPmts[0].RefNum != "RRN-998877" {
		t.Fatalf("expected ref RRN-998877, got %s", store.recordedPmts[0].RefNum)
	}
}

func TestStartupRecovery(t *testing.T) {
	store := newMemStore()
	orphanID := uuid.New()
	_ = store.CreateAttempt(context.Background(), &corepayments.Attempt{
		ID:             orphanID,
		IdempotencyKey: "orphan-1",
		AttemptType:    corepayments.TypeSale,
		StoreID:        1,
		InvoiceRef:     "INV-ORPHAN-1",
		AmountMinor:    1000,
		Currency:       "SAR",
		Status:         corepayments.StatusProcessing,
		Provider:       "geidea",
	})

	count, err := payments.StartupRecovery(context.Background(), store)
	if err != nil {
		t.Fatalf("recovery error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 orphan recovered, got %d", count)
	}

	att, _ := store.GetAttempt(context.Background(), orphanID)
	if att.Status != corepayments.StatusUnknown {
		t.Fatalf("expected orphan to be marked unknown, got %s", att.Status)
	}
}
