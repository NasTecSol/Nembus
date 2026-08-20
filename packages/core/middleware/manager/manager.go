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
	if m == nil || m.masterRepo == nil {
		return nil, fmt.Errorf("tenant registry unavailable")
	}

	// Revalidate the active registry row on every selection. A cached pool must
	// never keep a disabled tenant eligible for new migration requests.
	tenant, err := m.masterRepo.GetTenantBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("tenant %q not found or inactive", slug)
	}

	// Verify tenant is active (double check, though query already filters)
	if !tenant.IsActive.Bool || !tenant.IsActive.Valid {
		return nil, fmt.Errorf("tenant %q is not active", slug)
	}

	// Reuse an already cached pool only after the active registry check.
	if val, ok := m.pools.Load(slug); ok {
		return val.(*pgxpool.Pool), nil
	}

	// Create connection pool for tenant database
	pool, err := pgxpool.New(ctx, tenant.DbConnStr)
	if err != nil {
		return nil, fmt.Errorf("tenant %q database unavailable", slug)
	}

	// Cache the pool for future use
	m.pools.Store(slug, pool)
	return pool, nil
}

// GetTenantDSN returns the raw connection string (DSN) for a tenant without
// creating a connection pool. This is used by the gRPC backup service to pass
// the DSN directly to pg_dump.
func (m *Manager) GetTenantDSN(ctx context.Context, slug string) (string, error) {
	tenant, err := m.masterRepo.GetTenantBySlug(ctx, slug)
	if err != nil {
		return "", fmt.Errorf("tenant %q not found or inactive", slug)
	}
	if !tenant.IsActive.Bool || !tenant.IsActive.Valid {
		return "", fmt.Errorf("tenant %q is not active", slug)
	}
	return tenant.DbConnStr, nil
}
