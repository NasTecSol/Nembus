package manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/NasTecSol/nembus-core/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Manager struct {
	masterRepo *repository.Queries
	pools      sync.Map
}

func NewManager(repo *repository.Queries) *Manager {
	return &Manager{masterRepo: repo}
}

func (m *Manager) GetPool(ctx context.Context, slug string) (*pgxpool.Pool, error) {
	// Check if pool is already cached
	if val, ok := m.pools.Load(slug); ok {
		return val.(*pgxpool.Pool), nil
	}

	// This calls the Master DB to get the connection string
	// GetTenantBySlug only returns active tenants (WHERE is_active = true)
	tenant, err := m.masterRepo.GetTenantBySlug(ctx, slug)
	if err != nil {
		// Provide a more helpful error message
		// The error could be: tenant not found, tenant inactive, or database error
		return nil, fmt.Errorf("tenant '%s' not found or inactive: %w (hint: check if slug matches exactly and is_active = true)", slug, err)
	}

	// Verify tenant is active (double check, though query already filters)
	if !tenant.IsActive.Bool || !tenant.IsActive.Valid {
		return nil, fmt.Errorf("tenant '%s' is not active (is_active = false or NULL)", slug)
	}

	// Create connection pool for tenant database
	pool, err := pgxpool.New(ctx, tenant.DbConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tenant database for '%s' using connection string: %w", slug, err)
	}

	// Ensure required columns / schema patches exist
	PatchDatabase(ctx, pool)

	// Cache the pool for future use
	m.pools.Store(slug, pool)
	return pool, nil
}

// PatchDatabase applies non-destructive idempotent schema patches to ensure
// compatibility across existing tenant and master databases.
func PatchDatabase(ctx context.Context, pool *pgxpool.Pool) {
	if pool == nil {
		return
	}
	queries := []string{
		`ALTER TABLE IF EXISTS transfer_requests ADD COLUMN IF NOT EXISTS is_stock_reserved BOOLEAN NOT NULL DEFAULT false;`,
		`CREATE TABLE IF NOT EXISTS payment_attempts (
			id UUID PRIMARY KEY,
			idempotency_key VARCHAR(255) UNIQUE NOT NULL,
			attempt_type VARCHAR(50) NOT NULL DEFAULT 'sale',
			original_attempt_id UUID REFERENCES payment_attempts(id) ON DELETE SET NULL,
			store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
			pos_terminal_id INTEGER REFERENCES pos_terminals(id) ON DELETE SET NULL,
			cashier_session_id INTEGER REFERENCES cashier_sessions(id) ON DELETE SET NULL,
			transaction_id INTEGER REFERENCES pos_transactions(id) ON DELETE SET NULL,
			sales_order_id UUID REFERENCES sales_orders_v2(id) ON DELETE SET NULL,
			invoice_ref VARCHAR(100) NOT NULL,
			amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
			currency VARCHAR(3) NOT NULL DEFAULT 'SAR',
			status VARCHAR(50) NOT NULL CHECK (status IN ('pending','processing','succeeded','declined','cancelled','unknown')),
			provider VARCHAR(50) NOT NULL,
			provider_refs JSONB NOT NULL DEFAULT '{}',
			masked_pan VARCHAR(50),
			raw_redacted JSONB DEFAULT '{}',
			device_id VARCHAR(100),
			reconciled_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_one_open_attempt_per_invoice ON payment_attempts(invoice_ref)
			WHERE status IN ('pending','processing','unknown');`,
		`CREATE INDEX IF NOT EXISTS idx_payment_attempts_status ON payment_attempts(status);`,
		`CREATE INDEX IF NOT EXISTS idx_payment_attempts_transaction_id ON payment_attempts(transaction_id);`,
		`ALTER TABLE IF EXISTS pos_payments ADD COLUMN IF NOT EXISTS payment_attempt_id UUID REFERENCES payment_attempts(id) ON DELETE SET NULL;`,
	}
	for _, q := range queries {
		_, _ = pool.Exec(ctx, q)
	}
}

// GetTenantDSN returns the raw connection string (DSN) for a tenant without
// creating a connection pool. This is used by the gRPC backup service to pass
// the DSN directly to pg_dump.
func (m *Manager) GetTenantDSN(ctx context.Context, slug string) (string, error) {
	tenant, err := m.masterRepo.GetTenantBySlug(ctx, slug)
	if err != nil {
		return "", fmt.Errorf("tenant '%s' not found or inactive: %w", slug, err)
	}
	if !tenant.IsActive.Bool || !tenant.IsActive.Valid {
		return "", fmt.Errorf("tenant '%s' is not active", slug)
	}
	return tenant.DbConnStr, nil
}

// EvictPool removes and closes a cached connection pool for a tenant slug.
func (m *Manager) EvictPool(slug string) {
	if val, ok := m.pools.LoadAndDelete(slug); ok {
		if pool, ok := val.(*pgxpool.Pool); ok && pool != nil {
			pool.Close()
		}
	}
}

