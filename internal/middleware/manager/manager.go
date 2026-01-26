package manager

import (
	"context"
	"fmt"
	"sync"

	"NEMBUS/internal/repository"

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

	// Cache the pool for future use
	m.pools.Store(slug, pool)
	return pool, nil
}
