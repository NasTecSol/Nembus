package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SyncService struct {
	ctx        context.Context
	masterPool *pgxpool.Pool
	cloudURL   string
	tenantSlug string
}

func NewSyncService(ctx context.Context, pool *pgxpool.Pool, cloudURL, slug string) *SyncService {
	return &SyncService{
		ctx:        ctx,
		masterPool: pool,
		cloudURL:   cloudURL,
		tenantSlug: slug,
	}
}

func (s *SyncService) Start() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.performSync()
			}
		}
	}()
}

func (s *SyncService) performSync() {
	fmt.Printf("Performing sync for tenant: %s\n", s.tenantSlug)

	// 1. Pull changes from Cloud
	// For now, we'll re-use the cloner for a full sync or implement incremental fetch
	// A real implementation would fetch deltas: /api/sync/pull?since=LAST_SYNC_TIME

	// 2. Push local changes to Cloud
	// /api/sync/push [POST] with local records updated since LAST_SYNC_TIME
}

func (s *SyncService) applyDeltaLocally(table string, action string, data []byte) {
	// Logic to upsert/delete record in local master/tenant DB
}
