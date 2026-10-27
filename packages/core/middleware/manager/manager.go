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
		`CREATE TABLE IF NOT EXISTS rfqs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
			rfq_number VARCHAR(50) UNIQUE NOT NULL,
			vendor_name VARCHAR(255) NOT NULL DEFAULT '',
			title VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(50) NOT NULL DEFAULT 'draft',
			token VARCHAR(255) UNIQUE NOT NULL,
			token_valid_from TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			token_expires_at TIMESTAMP NOT NULL,
			invited_partner_id INTEGER REFERENCES business_partners(id) ON DELETE SET NULL,
			created_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`ALTER TABLE IF EXISTS rfqs ADD COLUMN IF NOT EXISTS vendor_name VARCHAR(255) NOT NULL DEFAULT '';`,
		`CREATE TABLE IF NOT EXISTS rfq_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			rfq_id UUID NOT NULL REFERENCES rfqs(id) ON DELETE CASCADE,
			product_id INTEGER REFERENCES products(id) ON DELETE SET NULL,
			product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
			item_name VARCHAR(255) NOT NULL,
			item_description TEXT,
			quantity DECIMAL(15,3) NOT NULL CHECK (quantity > 0),
			uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
			target_price DECIMAL(15,4),
			line_number INTEGER NOT NULL,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS rfq_responses (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			rfq_id UUID NOT NULL REFERENCES rfqs(id) ON DELETE CASCADE,
			organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			partner_id INTEGER REFERENCES business_partners(id) ON DELETE SET NULL,
			is_registered_supplier BOOLEAN NOT NULL DEFAULT false,
			supplier_name VARCHAR(255) NOT NULL,
			supplier_email VARCHAR(255),
			supplier_phone VARCHAR(50),
			response_status VARCHAR(50) NOT NULL DEFAULT 'submitted',
			total_quoted_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
			valid_until DATE,
			notes TEXT,
			metadata JSONB DEFAULT '{}',
			submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS rfq_response_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			rfq_response_id UUID NOT NULL REFERENCES rfq_responses(id) ON DELETE CASCADE,
			rfq_item_id UUID REFERENCES rfq_items(id) ON DELETE SET NULL,
			uom_id INTEGER NOT NULL REFERENCES units_of_measure(id) ON DELETE RESTRICT,
			quantity DECIMAL(15,3) NOT NULL CHECK (quantity > 0),
			unit_price DECIMAL(15,4) NOT NULL DEFAULT 0.0000,
			tax_amount DECIMAL(15,2) DEFAULT 0.00,
			line_total DECIMAL(15,2) NOT NULL DEFAULT 0.00,
			delivery_lead_time_days INTEGER,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`ALTER TABLE IF EXISTS rfq_response_items ADD COLUMN IF NOT EXISTS uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE RESTRICT;`,
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
