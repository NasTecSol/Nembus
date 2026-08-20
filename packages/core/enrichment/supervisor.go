package enrichment

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// TenantRegistration is the control-plane identity needed by the supervisor.
// It intentionally contains no tenant-local business identifiers or database
// credentials.
type TenantRegistration struct {
	Slug   string
	Active bool
}

// TenantRegistry enumerates control-plane tenants. Implementations must not
// return tenant business rows; active state is re-evaluated on every cycle.
type TenantRegistry interface {
	ListActiveTenants(context.Context) ([]TenantRegistration, error)
}

// TenantWorkerFactory binds one worker to one tenant-local repository. The
// supervisor never retains the result between tenants or cycles.
type TenantWorkerFactory func(context.Context, TenantRegistration) (*EnrichmentWorker, error)

// TenantEnrichmentSupervisor discovers active tenants on every cycle and runs
// one bounded, sequential worker pass per tenant. Tenant setup and execution
// failures are isolated so one tenant cannot stop another.
type TenantEnrichmentSupervisor struct {
	registry TenantRegistry
	factory  TenantWorkerFactory
	interval time.Duration
	logger   *log.Logger
}

func NewTenantEnrichmentSupervisor(registry TenantRegistry, factory TenantWorkerFactory, config EnrichmentExecutionConfig, logger *log.Logger) *TenantEnrichmentSupervisor {
	if logger == nil {
		logger = log.Default()
	}
	if config.Interval <= 0 {
		config.Interval = 30 * time.Second
	}
	return &TenantEnrichmentSupervisor{
		registry: registry,
		factory:  factory,
		interval: config.Interval,
		logger:   logger,
	}
}

// Start returns immediately and stops when ctx is cancelled. Enumeration is
// repeated rather than snapshotted at startup so tenant activation and
// disablement take effect without a cloud-server restart.
func (s *TenantEnrichmentSupervisor) Start(ctx context.Context) {
	if s == nil || s.registry == nil || s.factory == nil {
		return
	}
	go func() {
		if err := s.RunOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Printf("product enrichment supervisor cycle failed: error_class=%s", supervisorErrorClass(err))
		}
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.RunOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
					s.logger.Printf("product enrichment supervisor cycle failed: error_class=%s", supervisorErrorClass(err))
				}
			}
		}
	}()
}

// RunOnce is exported for deterministic tests and operational probes. A
// tenant-level setup or worker error is logged and processing continues.
func (s *TenantEnrichmentSupervisor) RunOnce(ctx context.Context) error {
	if s == nil || s.registry == nil || s.factory == nil {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	tenants, err := s.registry.ListActiveTenants(ctx)
	if err != nil {
		return fmt.Errorf("list active tenants: %w", err)
	}
	for _, tenant := range tenants {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !tenant.Active {
			s.logger.Printf("product enrichment tenant=%q skipped=inactive", tenant.Slug)
			continue
		}
		if strings.TrimSpace(tenant.Slug) == "" {
			s.logger.Printf("product enrichment tenant skipped=invalid_slug")
			continue
		}

		worker, err := s.factory(ctx, tenant)
		if err != nil {
			s.logger.Printf("product enrichment tenant=%q setup_failed error_class=%s", tenant.Slug, supervisorErrorClass(err))
			continue
		}
		if worker == nil {
			s.logger.Printf("product enrichment tenant=%q setup_failed error_class=worker_not_configured", tenant.Slug)
			continue
		}
		if err := worker.RunOnce(ctx); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				if ctx.Err() != nil {
					return ctx.Err()
				}
			}
			s.logger.Printf("product enrichment tenant=%q worker_failed error_class=%s", tenant.Slug, supervisorErrorClass(err))
		}
	}
	return nil
}

func supervisorErrorClass(err error) string {
	if err == nil {
		return ""
	}
	if class := ProviderErrorClassOf(err); class != "" {
		return string(class)
	}
	if class := ResponseErrorClassOf(err); class != "" {
		return string(class)
	}
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "deadline_exceeded"
	}
	return "tenant_worker_error"
}
