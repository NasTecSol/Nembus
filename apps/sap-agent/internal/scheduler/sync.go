package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/internal/cloudgrpc"
	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/pipeline"
	"github.com/NasTecSol/nembus-sap-agent/internal/sap"
)

type SyncScheduler struct {
	cfg      *config.AgentConfig
	cloud    *cloudgrpc.Client
	sap      *sap.Client
	mu       sync.Mutex
	running  bool
	lastDown time.Time
	lastUp   time.Time
	lastErr  string
}

func NewSyncScheduler(cfg *config.AgentConfig, cloud *cloudgrpc.Client, sapClient *sap.Client) *SyncScheduler {
	return &SyncScheduler{cfg: cfg, cloud: cloud, sap: sapClient}
}

func (s *SyncScheduler) Status() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{
		"running":            s.running,
		"last_downstream_at": s.lastDown,
		"last_upstream_at":   s.lastUp,
		"last_error":         s.lastErr,
	}
}

func (s *SyncScheduler) RunDownstream(ctx context.Context, full bool) error {
	settings := pipeline.FromAgent(s.cfg)
	down := pipeline.NewDownstreamSync(settings, s.sap, s.cloud)
	_, err := down.SyncAll(ctx, full)
	s.mu.Lock()
	s.lastDown = time.Now()
	if err != nil {
		s.lastErr = err.Error()
	}
	s.mu.Unlock()
	return err
}

func (s *SyncScheduler) RunUpstream(ctx context.Context) error {
	settings := pipeline.FromAgent(s.cfg)
	up := pipeline.NewUpstreamSync(settings, s.sap, s.cloud)
	_, err := up.ProcessOutbox(ctx)
	s.mu.Lock()
	s.lastUp = time.Now()
	if err != nil {
		s.lastErr = err.Error()
	}
	s.mu.Unlock()
	return err
}

func (s *SyncScheduler) StartBackground(ctx context.Context) {
	go func() {
		if err := s.RunDownstream(ctx, false); err != nil {
			log.Printf("[SAP-AGENT] initial downstream sync: %v", err)
		}
	}()

	downInterval := time.Duration(s.cfg.DownstreamIntervalSec) * time.Second
	if downInterval <= 0 {
		downInterval = 5 * time.Minute
	}
	upInterval := time.Duration(s.cfg.Sync.UpstreamOutboxPollMs) * time.Millisecond
	if upInterval <= 0 {
		upInterval = 3 * time.Second
	}

	downTicker := time.NewTicker(downInterval)
	upTicker := time.NewTicker(upInterval)
	defer downTicker.Stop()
	defer upTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-downTicker.C:
			if err := s.RunDownstream(ctx, false); err != nil {
				log.Printf("[SAP-AGENT] downstream sync: %v", err)
			}
		case <-upTicker.C:
			if err := s.RunUpstream(ctx); err != nil {
				log.Printf("[SAP-AGENT] upstream sync: %v", err)
			}
		}
	}
}

func (s *SyncScheduler) TriggerDownstream(ctx context.Context) {
	go func() {
		s.mu.Lock()
		s.running = true
		s.mu.Unlock()
		defer func() {
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
		}()
		if err := s.RunDownstream(ctx, true); err != nil {
			log.Printf("[SAP-AGENT] manual downstream: %v", err)
		}
	}()
}

func (s *SyncScheduler) TriggerUpstream(ctx context.Context) {
	go func() {
		s.mu.Lock()
		s.running = true
		s.mu.Unlock()
		defer func() {
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
		}()
		if err := s.RunUpstream(ctx); err != nil {
			log.Printf("[SAP-AGENT] manual upstream: %v", err)
		}
	}()
}
