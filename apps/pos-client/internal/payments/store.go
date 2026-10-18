package payments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	corepayments "github.com/NasTecSol/nembus-core/payments"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAttemptNotFound = errors.New("payment attempt not found")

// Store defines persistence operations for payment attempts and local payment outcomes.
type Store interface {
	CreateAttempt(ctx context.Context, a *corepayments.Attempt) error
	GetAttempt(ctx context.Context, id uuid.UUID) (*corepayments.Attempt, error)
	GetAttemptByIdempotencyKey(ctx context.Context, key string) (*corepayments.Attempt, error)
	FindOpenAttemptForInvoice(ctx context.Context, invoiceRef string) (*corepayments.Attempt, error)
	UpdateAttemptStatus(ctx context.Context, id uuid.UUID, newStatus corepayments.Status, providerRefs map[string]string, maskedPAN *string, rawRedacted json.RawMessage) error
	ListUnresolvedAttempts(ctx context.Context, storeID int32) ([]*corepayments.Attempt, error)
	MarkOrphanProcessingAsUnknown(ctx context.Context) (int64, error)
	RecordApprovedPayment(ctx context.Context, attempt *corepayments.Attempt, amountDecimal string, refNum string, meta []byte) error
}

// PgxStore implements Store using a PostgreSQL pgxpool.
type PgxStore struct {
	pool *pgxpool.Pool
}

// NewPgxStore creates a new PostgreSQL store for payment attempts.
func NewPgxStore(pool *pgxpool.Pool) *PgxStore {
	return &PgxStore{pool: pool}
}

// CreateAttempt persists a new attempt in the database.
func (s *PgxStore) CreateAttempt(ctx context.Context, a *corepayments.Attempt) error {
	if a == nil {
		return errors.New("attempt cannot be nil")
	}
	if err := a.Validate(); err != nil {
		return fmt.Errorf("validate attempt: %w", err)
	}

	providerRefsJSON, err := json.Marshal(a.ProviderRefs)
	if err != nil {
		providerRefsJSON = []byte("{}")
	}
	rawRedactedJSON := a.RawRedacted
	if len(rawRedactedJSON) == 0 {
		rawRedactedJSON = json.RawMessage("{}")
	}

	query := `
		INSERT INTO payment_attempts (
			id, idempotency_key, attempt_type, original_attempt_id,
			store_id, pos_terminal_id, cashier_session_id, transaction_id,
			sales_order_id, invoice_ref, amount_minor, currency,
			status, provider, provider_refs, masked_pan,
			raw_redacted, device_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10, $11, $12,
			$13, $14, $15, $16,
			$17, $18, NOW(), NOW()
		)
	`
	_, err = s.pool.Exec(ctx, query,
		a.ID, a.IdempotencyKey, string(a.AttemptType), a.OriginalAttemptID,
		a.StoreID, a.PosTerminalID, a.CashierSessionID, a.TransactionID,
		a.SalesOrderID, a.InvoiceRef, a.AmountMinor, a.Currency,
		string(a.Status), a.Provider, providerRefsJSON, a.MaskedPAN,
		rawRedactedJSON, a.DeviceID,
	)
	if err != nil {
		return fmt.Errorf("create payment attempt: %w", err)
	}
	return nil
}

// GetAttempt retrieves an attempt by ID.
func (s *PgxStore) GetAttempt(ctx context.Context, id uuid.UUID) (*corepayments.Attempt, error) {
	query := `
		SELECT
			id, idempotency_key, attempt_type, original_attempt_id,
			store_id, pos_terminal_id, cashier_session_id, transaction_id,
			sales_order_id, invoice_ref, amount_minor, currency,
			status, provider, provider_refs, masked_pan,
			raw_redacted, device_id, created_at, updated_at
		FROM payment_attempts
		WHERE id = $1
	`
	row := s.pool.QueryRow(ctx, query, id)
	return scanAttempt(row)
}

// GetAttemptByIdempotencyKey retrieves an attempt by idempotency key.
func (s *PgxStore) GetAttemptByIdempotencyKey(ctx context.Context, key string) (*corepayments.Attempt, error) {
	query := `
		SELECT
			id, idempotency_key, attempt_type, original_attempt_id,
			store_id, pos_terminal_id, cashier_session_id, transaction_id,
			sales_order_id, invoice_ref, amount_minor, currency,
			status, provider, provider_refs, masked_pan,
			raw_redacted, device_id, created_at, updated_at
		FROM payment_attempts
		WHERE idempotency_key = $1
	`
	row := s.pool.QueryRow(ctx, query, key)
	return scanAttempt(row)
}

// FindOpenAttemptForInvoice finds any pending, processing, or unknown attempt for an invoice.
func (s *PgxStore) FindOpenAttemptForInvoice(ctx context.Context, invoiceRef string) (*corepayments.Attempt, error) {
	query := `
		SELECT
			id, idempotency_key, attempt_type, original_attempt_id,
			store_id, pos_terminal_id, cashier_session_id, transaction_id,
			sales_order_id, invoice_ref, amount_minor, currency,
			status, provider, provider_refs, masked_pan,
			raw_redacted, device_id, created_at, updated_at
		FROM payment_attempts
		WHERE invoice_ref = $1 AND status IN ('pending', 'processing', 'unknown')
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := s.pool.QueryRow(ctx, query, invoiceRef)
	return scanAttempt(row)
}

// UpdateAttemptStatus updates the attempt status and metadata.
func (s *PgxStore) UpdateAttemptStatus(ctx context.Context, id uuid.UUID, newStatus corepayments.Status, providerRefs map[string]string, maskedPAN *string, rawRedacted json.RawMessage) error {
	var providerRefsJSON []byte
	if providerRefs != nil {
		providerRefsJSON, _ = json.Marshal(providerRefs)
	}

	query := `
		UPDATE payment_attempts
		SET status = $2,
		    provider_refs = CASE WHEN $3::jsonb IS NOT NULL THEN provider_refs || $3::jsonb ELSE provider_refs END,
		    masked_pan = COALESCE($4, masked_pan),
		    raw_redacted = CASE WHEN $5::jsonb IS NOT NULL THEN raw_redacted || $5::jsonb ELSE raw_redacted END,
		    updated_at = NOW()
		WHERE id = $1
	`
	tag, err := s.pool.Exec(ctx, query, id, string(newStatus), providerRefsJSON, maskedPAN, rawRedacted)
	if err != nil {
		return fmt.Errorf("update payment attempt status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAttemptNotFound
	}
	return nil
}

// ListUnresolvedAttempts lists attempts in 'unknown' status.
func (s *PgxStore) ListUnresolvedAttempts(ctx context.Context, storeID int32) ([]*corepayments.Attempt, error) {
	query := `
		SELECT
			id, idempotency_key, attempt_type, original_attempt_id,
			store_id, pos_terminal_id, cashier_session_id, transaction_id,
			sales_order_id, invoice_ref, amount_minor, currency,
			status, provider, provider_refs, masked_pan,
			raw_redacted, device_id, created_at, updated_at
		FROM payment_attempts
		WHERE status = 'unknown' AND ($1 = 0 OR store_id = $1)
		ORDER BY created_at ASC
	`
	rows, err := s.pool.Query(ctx, query, storeID)
	if err != nil {
		return nil, fmt.Errorf("list unresolved attempts: %w", err)
	}
	defer rows.Close()

	var attempts []*corepayments.Attempt
	for rows.Next() {
		a, err := scanAttemptFromRow(rows)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, a)
	}
	return attempts, rows.Err()
}

// MarkOrphanProcessingAsUnknown marks attempts left in 'processing' status (e.g. after POS crash/power loss) as 'unknown'.
func (s *PgxStore) MarkOrphanProcessingAsUnknown(ctx context.Context) (int64, error) {
	query := `
		UPDATE payment_attempts
		SET status = 'unknown',
		    updated_at = NOW()
		WHERE status = 'processing'
	`
	tag, err := s.pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("mark orphan processing attempts: %w", err)
	}
	return tag.RowsAffected(), nil
}

// RecordApprovedPayment inserts a pos_payments record and updates transaction amount_paid in a single transaction.
func (s *PgxStore) RecordApprovedPayment(ctx context.Context, attempt *corepayments.Attempt, amountDecimal string, refNum string, meta []byte) error {
	if attempt.TransactionID == nil || *attempt.TransactionID <= 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if payment for this attempt already exists
	var existingID int32
	err = tx.QueryRow(ctx, `SELECT id FROM pos_payments WHERE payment_attempt_id = $1`, attempt.ID).Scan(&existingID)
	if err == nil {
		// Already recorded
		return nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("check existing payment: %w", err)
	}

	if len(meta) == 0 {
		meta = []byte("{}")
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO pos_payments (
			transaction_id, payment_attempt_id, payment_method, payment_gateway,
			amount, payment_reference, reference_number, metadata, payment_date
		) VALUES (
			$1, $2, 'card', $3, $4::numeric, $5, $6, $7::jsonb, NOW()
		)
	`, *attempt.TransactionID, attempt.ID, attempt.Provider, amountDecimal, attempt.ID.String(), refNum, meta)
	if err != nil {
		return fmt.Errorf("insert pos_payment: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE pos_transactions
		SET amount_paid = COALESCE(amount_paid, 0) + $2::numeric,
		    updated_at = NOW()
		WHERE id = $1
	`, *attempt.TransactionID, amountDecimal)
	if err != nil {
		return fmt.Errorf("update pos_transactions amount_paid: %w", err)
	}

	return tx.Commit(ctx)
}

type scannable interface {
	Scan(dest ...any) error
}

func scanAttempt(row pgx.Row) (*corepayments.Attempt, error) {
	a, err := scanAttemptFromRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAttemptNotFound
	}
	return a, err
}

func scanAttemptFromRow(row scannable) (*corepayments.Attempt, error) {
	var (
		a                corepayments.Attempt
		attemptType      string
		status           string
		providerRefsJSON []byte
		rawRedactedJSON  []byte
		createdAt        time.Time
		updatedAt        time.Time
	)

	err := row.Scan(
		&a.ID, &a.IdempotencyKey, &attemptType, &a.OriginalAttemptID,
		&a.StoreID, &a.PosTerminalID, &a.CashierSessionID, &a.TransactionID,
		&a.SalesOrderID, &a.InvoiceRef, &a.AmountMinor, &a.Currency,
		&status, &a.Provider, &providerRefsJSON, &a.MaskedPAN,
		&rawRedactedJSON, &a.DeviceID, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	a.AttemptType = corepayments.Type(attemptType)
	a.Status = corepayments.Status(status)
	a.CreatedAt = &createdAt
	a.UpdatedAt = &updatedAt

	if len(providerRefsJSON) > 0 {
		_ = json.Unmarshal(providerRefsJSON, &a.ProviderRefs)
	}
	if len(rawRedactedJSON) > 0 {
		a.RawRedacted = rawRedactedJSON
	}

	return &a, nil
}
