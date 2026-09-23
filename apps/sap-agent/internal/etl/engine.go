package etl

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/NasTecSol/nembus-sap/contracts"
	"github.com/NasTecSol/nembus-sap/mappings"
	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap-agent/internal/etl/extractors"
	"github.com/NasTecSol/nembus-sap-agent/internal/transport"
)

// defaultDomainOrder is the canonical ordering of all migration domains.
// Dependencies flow top-to-bottom (stores before inventory, categories before products, etc.)
var defaultDomainOrder = []contracts.DomainType{
	contracts.DomainStores,
	contracts.DomainUsers,
	contracts.DomainUOM,
	contracts.DomainUOMGroups,
	contracts.DomainCategories,
	contracts.DomainBrands,
	contracts.DomainProducts,
	contracts.DomainBarcodes,
	contracts.DomainPriceLists,
	contracts.DomainInventory,
	contracts.DomainPartners,
	contracts.DomainBPAddresses,
	contracts.DomainPurchaseOrders,
	contracts.DomainGoodsReceipts,
	contracts.DomainStockMovements,
	contracts.DomainSalesOrders,
	contracts.DomainInvoices,
	contracts.DomainIncomingPayments,
}

// ProgressEvent is broadcast to all WebSocket subscribers to report pipeline progress.
type ProgressEvent struct {
	Type           string               `json:"type"` // "run_started", "step_started", "step_progress", "step_completed", "run_completed", "error"
	RunID          string               `json:"run_id"`
	Domain         contracts.DomainType `json:"domain,omitempty"`
	Status         contracts.RunStatus  `json:"status"`
	TotalDomains   int                  `json:"total_domains"`
	CompletedSteps int                  `json:"completed_steps"`
	TotalRecords   int64                `json:"total_records"`
	ProcessedCount int64                `json:"processed_count"`
	FailedCount    int64                `json:"failed_count"`
	Percentage     float64              `json:"percentage"`
	Message        string               `json:"message"`
	Timestamp      time.Time            `json:"timestamp"`
}

// Engine orchestrates the full SAP → Nembus Cloud ETL pipeline.
type Engine struct {
	cfg         *config.AgentConfig
	mssql       *db.MSSQLClient
	sqlite      *db.SQLiteStore
	cloudClient *transport.CloudClient

	listeners   map[chan ProgressEvent]struct{}
	listenersMu sync.RWMutex

	activeRunCancel context.CancelFunc
	activeRunMu     sync.Mutex
	currentRunID    string
}

func NewEngine(cfg *config.AgentConfig, mssql *db.MSSQLClient, sqlite *db.SQLiteStore, cloudClient *transport.CloudClient) *Engine {
	return &Engine{
		cfg:         cfg,
		mssql:       mssql,
		sqlite:      sqlite,
		cloudClient: cloudClient,
		listeners:   make(map[chan ProgressEvent]struct{}),
	}
}

func (e *Engine) getMSSQLClient() (*db.MSSQLClient, error) {
	if e.mssql != nil && e.mssql.DB != nil {
		if err := e.mssql.DB.Ping(); err == nil {
			return e.mssql, nil
		}
	}
	cfg := config.Get().MSSQL
	client, err := db.NewMSSQLClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("mssql database is not connected (%s:%d): %w", cfg.Host, cfg.Port, err)
	}
	e.mssql = client
	return client, nil
}

// getCloudClient returns the engine's shared CloudClient, creating one if not yet initialized.
// This avoids allocating a new http.Client (with its connection pool) on every domain step.
func (e *Engine) getCloudClient() *transport.CloudClient {
	if e.cloudClient != nil {
		return e.cloudClient
	}
	e.cloudClient = transport.NewCloudClient(config.Get().Cloud)
	return e.cloudClient
}

func (e *Engine) Subscribe() chan ProgressEvent {
	e.listenersMu.Lock()
	defer e.listenersMu.Unlock()

	ch := make(chan ProgressEvent, 100)
	e.listeners[ch] = struct{}{}
	return ch
}

func (e *Engine) Unsubscribe(ch chan ProgressEvent) {
	e.listenersMu.Lock()
	defer e.listenersMu.Unlock()

	delete(e.listeners, ch)
	close(ch)
}

func (e *Engine) broadcast(event ProgressEvent) {
	e.listenersMu.RLock()
	defer e.listenersMu.RUnlock()

	for ch := range e.listeners {
		select {
		case ch <- event:
		default:
		}
	}
}

func (e *Engine) StartMigration(parentCtx context.Context, mode contracts.MigrationMode, selectedDomains []contracts.DomainType) (*db.MigrationRun, error) {
	e.activeRunMu.Lock()
	defer e.activeRunMu.Unlock()

	if e.activeRunCancel != nil {
		return nil, fmt.Errorf("a migration run is already currently executing")
	}

	if len(selectedDomains) == 0 {
		selectedDomains = defaultDomainOrder
	}

	run, err := e.sqlite.CreateRun(parentCtx, e.cfg.Cloud.OrganizationID, mode, len(selectedDomains))
	if err != nil {
		return nil, fmt.Errorf("failed to create migration run in sqlite: %w", err)
	}

	e.currentRunID = run.ID
	runCtx, cancel := context.WithCancel(context.Background())
	e.activeRunCancel = cancel

	e.sqlite.Log(runCtx, run.ID, "", "INFO", fmt.Sprintf("Started migration run %s in %s mode with %d domains.", run.ID, mode, len(selectedDomains)))

	go e.executePipeline(runCtx, run, selectedDomains)

	return run, nil
}

func (e *Engine) CancelMigration() error {
	e.activeRunMu.Lock()
	defer e.activeRunMu.Unlock()

	if e.activeRunCancel == nil {
		return fmt.Errorf("no migration run is active")
	}

	e.activeRunCancel()
	e.activeRunCancel = nil
	if e.currentRunID != "" {
		// Use StatusCancelled (not StatusFailed) so history shows operator intent
		_ = e.sqlite.UpdateRunStatus(context.Background(), e.currentRunID, contracts.StatusCancelled, "Cancelled by operator")
		e.sqlite.Log(context.Background(), e.currentRunID, "", "WARN", "Migration run cancelled by operator.")
		e.broadcast(ProgressEvent{
			Type:      "run_cancelled",
			RunID:     e.currentRunID,
			Status:    contracts.StatusCancelled,
			Message:   "Migration cancelled by operator",
			Timestamp: time.Now(),
		})
	}
	return nil
}

func (e *Engine) executePipeline(ctx context.Context, run *db.MigrationRun, domains []contracts.DomainType) {
	defer func() {
		e.activeRunMu.Lock()
		e.activeRunCancel = nil
		e.activeRunMu.Unlock()
	}()

	e.broadcast(ProgressEvent{
		Type:           "run_started",
		RunID:          run.ID,
		Status:         contracts.StatusRunning,
		TotalDomains:   len(domains),
		CompletedSteps: 0,
		Message:        "Migration started",
		Timestamp:      time.Now(),
	})

	var totalProcessed int64 = 0
	var totalFailed int64 = 0
	completedSteps := 0
	failedDomains := 0

	for _, domain := range domains {
		select {
		case <-ctx.Done():
			_ = e.sqlite.UpdateRunStatus(context.Background(), run.ID, contracts.StatusCancelled, "Operation cancelled")
			return
		default:
		}

		step := &db.MigrationStep{
			RunID:     run.ID,
			Domain:    domain,
			Status:    contracts.StatusRunning,
			StartedAt: time.Now(),
		}
		_ = e.sqlite.CreateOrUpdateStep(ctx, step)
		e.sqlite.Log(ctx, run.ID, string(domain), "INFO", fmt.Sprintf("Beginning extraction for domain: %s", domain))

		e.broadcast(ProgressEvent{
			Type:           "step_started",
			RunID:          run.ID,
			Domain:         domain,
			Status:         contracts.StatusRunning,
			TotalDomains:   len(domains),
			CompletedSteps: completedSteps,
			ProcessedCount: totalProcessed,
			Message:        fmt.Sprintf("Extracting %s...", domain),
			Timestamp:      time.Now(),
		})

		processed, failed, watermark, err := e.executeDomainStep(ctx, run.ID, domain)
		now := time.Now()
		step.FinishedAt = &now
		step.ProcessedCount = processed
		step.FailedCount = failed
		step.LastWatermark = watermark

		if err != nil {
			step.Status = contracts.StatusFailed
			step.ErrorMessage = err.Error()
			_ = e.sqlite.CreateOrUpdateStep(ctx, step)
			e.sqlite.Log(ctx, run.ID, string(domain), "ERROR", fmt.Sprintf("Failed step %s: %v", domain, err))

			failedDomains++
			e.broadcast(ProgressEvent{
				Type:           "error",
				RunID:          run.ID,
				Domain:         domain,
				Status:         contracts.StatusFailed,
				TotalDomains:   len(domains),
				CompletedSteps: completedSteps,
				Message:        fmt.Sprintf("Error in %s: %v — continuing to next domain", domain, err),
				Timestamp:      time.Now(),
			})
			// Continue-on-error: do NOT return; proceed to next domain
			continue
		}

		// Persist watermark for delta/incremental resumability
		step.LastWatermark = watermark
		step.Status = contracts.StatusCompleted
		_ = e.sqlite.CreateOrUpdateStep(ctx, step)
		completedSteps++
		totalProcessed += processed
		totalFailed += failed

		e.sqlite.Log(ctx, run.ID, string(domain), "INFO", fmt.Sprintf("Completed domain %s: %d records ingested.", domain, processed))
		e.broadcast(ProgressEvent{
			Type:           "step_completed",
			RunID:          run.ID,
			Domain:         domain,
			Status:         contracts.StatusCompleted,
			TotalDomains:   len(domains),
			CompletedSteps: completedSteps,
			ProcessedCount: totalProcessed,
			Percentage:     float64(completedSteps) / float64(len(domains)) * 100,
			Message:        fmt.Sprintf("Completed %s (%d records)", domain, processed),
			Timestamp:      time.Now(),
		})
	}

	// Determine final run status
	finalStatus := contracts.StatusCompleted
	if failedDomains > 0 && completedSteps == 0 {
		finalStatus = contracts.StatusFailed
	} else if failedDomains > 0 {
		finalStatus = contracts.StatusPartialSuccess
	}

	_ = e.sqlite.UpdateRunStatus(ctx, run.ID, finalStatus, "")
	e.sqlite.Log(ctx, run.ID, "", "INFO", fmt.Sprintf(
		"Migration %s finished with status %s. Processed: %d records across %d domains (%d domain(s) failed).",
		run.ID, finalStatus, totalProcessed, completedSteps, failedDomains,
	))

	e.broadcast(ProgressEvent{
		Type:           "run_completed",
		RunID:          run.ID,
		Status:         finalStatus,
		TotalDomains:   len(domains),
		CompletedSteps: completedSteps,
		ProcessedCount: totalProcessed,
		FailedCount:    totalFailed,
		Percentage:     float64(completedSteps) / float64(len(domains)) * 100,
		Message:        fmt.Sprintf("Migration finished (%s). %d/%d domains successful.", finalStatus, completedSteps, len(domains)),
		Timestamp:      time.Now(),
	})
}

// executeDomainStep runs the extract-transform-load pipeline for a single domain.
// Returns (processed, failed, watermark, error).
// watermark is a string cursor (e.g. last DocDate, last ItemCode) enabling delta runs.
func (e *Engine) executeDomainStep(ctx context.Context, runID string, domain contracts.DomainType) (int64, int64, string, error) {
	mssqlClient, err := e.getMSSQLClient()
	if err != nil {
		return 0, 0, "", err
	}
	cloudClient := e.getCloudClient()
	cloudCfg := config.Get().Cloud

	batchSize := e.cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 500
	}

	// defaultStoreCode is used to associate cashiers with a primary store.
	// Configurable via agent_config.json; falls back to "01".
	defaultStoreCode := e.cfg.DefaultStoreCode
	if defaultStoreCode == "" {
		defaultStoreCode = "01"
	}

	switch domain {
	case contracts.DomainStores:
		ext := extractors.NewStoresExtractor(mssqlClient)
		stores, locs, err := ext.ExtractStores(ctx)
		if err != nil {
			return 0, 0, "", err
		}
		payload := &contracts.MigrationBatchPayload{
			BatchID:        uuid.New().String(),
			RunID:          runID,
			OrganizationID: cloudCfg.OrganizationID,
			Domain:         domain,
			Stores:         stores,
			Locations:      locs,
			IsLastBatch:    true,
			Timestamp:      time.Now(),
		}
		resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
		if err != nil {
			return 0, 0, "", err
		}
		return int64(resp.RecordsStaged), int64(resp.RecordsFailed), "", nil

	case contracts.DomainUsers:
		ext := extractors.NewUsersExtractor(mssqlClient)
		cashierDefaults := mappings.CashierDefaults{
			DefaultStoreCode: defaultStoreCode,
			DrawerLimit:      e.cfg.CashierDrawerLimit,
			DiscountLimit:    e.cfg.CashierDiscountLimit,
		}
		users, cashiers, err := ext.ExtractUsers(ctx, cashierDefaults)
		if err != nil {
			return 0, 0, "", err
		}
		payload := &contracts.MigrationBatchPayload{
			BatchID:        uuid.New().String(),
			RunID:          runID,
			OrganizationID: cloudCfg.OrganizationID,
			Domain:         domain,
			Users:          users,
			Cashiers:       cashiers,
			IsLastBatch:    true,
			Timestamp:      time.Now(),
		}
		resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
		if err != nil {
			return 0, 0, "", err
		}
		return int64(resp.RecordsStaged), int64(resp.RecordsFailed), "", nil

	case contracts.DomainUOM:
		ext := extractors.NewUOMExtractor(mssqlClient)
		uoms, err := ext.ExtractUOMs(ctx)
		if err != nil {
			return 0, 0, "", err
		}
		if len(uoms) == 0 {
			return 0, 0, "", nil
		}
		payload := &contracts.MigrationBatchPayload{
			BatchID:        uuid.New().String(),
			RunID:          runID,
			OrganizationID: cloudCfg.OrganizationID,
			Domain:         domain,
			UOMs:           uoms,
			IsLastBatch:    true,
			Timestamp:      time.Now(),
		}
		resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
		if err != nil {
			return 0, 0, "", err
		}
		return int64(resp.RecordsStaged), int64(resp.RecordsFailed), "", nil

	case contracts.DomainUOMGroups:
		ext := extractors.NewUOMExtractor(mssqlClient)
		groups, err := ext.ExtractUOMGroups(ctx)
		if err != nil {
			return 0, 0, "", err
		}
		if len(groups) == 0 {
			return 0, 0, "", nil
		}
		payload := &contracts.MigrationBatchPayload{
			BatchID:        uuid.New().String(),
			RunID:          runID,
			OrganizationID: cloudCfg.OrganizationID,
			Domain:         domain,
			UOMGroups:      groups,
			IsLastBatch:    true,
			Timestamp:      time.Now(),
		}
		resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
		if err != nil {
			return 0, 0, "", err
		}
		return int64(resp.RecordsStaged), int64(resp.RecordsFailed), "", nil

	case contracts.DomainCategories:
		ext := extractors.NewCatalogExtractor(mssqlClient)
		cats, err := ext.ExtractCategories(ctx)
		if err != nil {
			return 0, 0, "", err
		}
		payload := &contracts.MigrationBatchPayload{
			BatchID:        uuid.New().String(),
			RunID:          runID,
			OrganizationID: cloudCfg.OrganizationID,
			Domain:         domain,
			Categories:     cats,
			IsLastBatch:    true,
			Timestamp:      time.Now(),
		}
		resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
		if err != nil {
			return 0, 0, "", err
		}
		return int64(resp.RecordsStaged), int64(resp.RecordsFailed), "", nil

	case contracts.DomainBrands:
		ext := extractors.NewCatalogExtractor(mssqlClient)
		brands, err := ext.ExtractBrands(ctx)
		if err != nil {
			return 0, 0, "", err
		}
		payload := &contracts.MigrationBatchPayload{
			BatchID:        uuid.New().String(),
			RunID:          runID,
			OrganizationID: cloudCfg.OrganizationID,
			Domain:         domain,
			Brands:         brands,
			IsLastBatch:    true,
			Timestamp:      time.Now(),
		}
		resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
		if err != nil {
			return 0, 0, "", err
		}
		return int64(resp.RecordsStaged), int64(resp.RecordsFailed), "", nil

	case contracts.DomainProducts:
		ext := extractors.NewCatalogExtractor(mssqlClient)
		products, err := ext.ExtractProducts(ctx)
		if err != nil {
			return 0, 0, "", err
		}
		var totalStaged int64
		var lastWatermark string
		for i := 0; i < len(products); i += batchSize {
			end := i + batchSize
			if end > len(products) {
				end = len(products)
			}
			chunk := products[i:end]
			isLast := end == len(products)
			payload := &contracts.MigrationBatchPayload{
				BatchID:        uuid.New().String(),
				RunID:          runID,
				OrganizationID: cloudCfg.OrganizationID,
				Domain:         domain,
				SequenceNumber: i / batchSize,
				Products:       chunk,
				IsLastBatch:    isLast,
				Timestamp:      time.Now(),
			}
			resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
			if err != nil {
				return totalStaged, 0, lastWatermark, err
			}
			totalStaged += int64(resp.RecordsStaged)
			if isLast && len(chunk) > 0 {
				lastWatermark = chunk[len(chunk)-1].SKU
			}
			e.broadcastProgress(runID, domain, totalStaged, int64(len(products)))
		}
		return totalStaged, 0, lastWatermark, nil

	case contracts.DomainBarcodes:
		ext := extractors.NewCatalogExtractor(mssqlClient)
		barcodes, err := ext.ExtractBarcodes(ctx)
		if err != nil {
			return 0, 0, "", err
		}
		if len(barcodes) == 0 {
			return 0, 0, "", nil
		}
		payload := &contracts.MigrationBatchPayload{
			BatchID:        uuid.New().String(),
			RunID:          runID,
			OrganizationID: cloudCfg.OrganizationID,
			Domain:         domain,
			Barcodes:       barcodes,
			IsLastBatch:    true,
			Timestamp:      time.Now(),
		}
		resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
		if err != nil {
			return 0, 0, "", err
		}
		return int64(resp.RecordsStaged), int64(resp.RecordsFailed), "", nil

	case contracts.DomainInventory:
		ext := extractors.NewInventoryExtractor(mssqlClient)
		inv, err := ext.ExtractInventory(ctx)
		if err != nil {
			return 0, 0, "", err
		}
		var totalStaged int64
		var lastWatermark string
		for i := 0; i < len(inv); i += batchSize {
			end := i + batchSize
			if end > len(inv) {
				end = len(inv)
			}
			chunk := inv[i:end]
			isLast := end == len(inv)
			payload := &contracts.MigrationBatchPayload{
				BatchID:        uuid.New().String(),
				RunID:          runID,
				OrganizationID: cloudCfg.OrganizationID,
				Domain:         domain,
				SequenceNumber: i / batchSize,
				Inventory:      chunk,
				IsLastBatch:    isLast,
				Timestamp:      time.Now(),
			}
			resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
			if err != nil {
				return totalStaged, 0, lastWatermark, err
			}
			totalStaged += int64(resp.RecordsStaged)
			if isLast && len(chunk) > 0 {
				last := chunk[len(chunk)-1]
				lastWatermark = fmt.Sprintf("%s:%s", last.ProductSKU, last.StoreCode)
			}
			e.broadcastProgress(runID, domain, totalStaged, int64(len(inv)))
		}
		return totalStaged, 0, lastWatermark, nil

	case contracts.DomainPartners:
		ext := extractors.NewPartnersExtractor(mssqlClient)
		partners, err := ext.ExtractPartners(ctx)
		if err != nil {
			return 0, 0, "", err
		}
		payload := &contracts.MigrationBatchPayload{
			BatchID:        uuid.New().String(),
			RunID:          runID,
			OrganizationID: cloudCfg.OrganizationID,
			Domain:         domain,
			Partners:       partners,
			IsLastBatch:    true,
			Timestamp:      time.Now(),
		}
		resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
		if err != nil {
			return 0, 0, "", err
		}
		return int64(resp.RecordsStaged), int64(resp.RecordsFailed), "", nil

	case contracts.DomainPurchaseOrders:
		ext := extractors.NewPurchasingExtractor(mssqlClient)
		var totalStaged int64
		var lastWatermark string
		var seqNum int
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		for _, win := range dateWindows(start, time.Now().AddDate(0, 3, 0), 3) {
			select {
			case <-ctx.Done():
				return totalStaged, 0, lastWatermark, ctx.Err()
			default:
			}
			orders, err := ext.ExtractPurchaseOrders(ctx, win[0], win[1])
			if err != nil {
				return totalStaged, 0, lastWatermark, fmt.Errorf("purchase_orders window %s–%s: %w", win[0].Format("2006-01"), win[1].Format("2006-01"), err)
			}
			if len(orders) == 0 {
				continue
			}
			for i := 0; i < len(orders); i += batchSize {
				end := i + batchSize
				if end > len(orders) {
					end = len(orders)
				}
				chunk := orders[i:end]
				payload := &contracts.MigrationBatchPayload{
					BatchID:        uuid.New().String(),
					RunID:          runID,
					OrganizationID: cloudCfg.OrganizationID,
					Domain:         domain,
					SequenceNumber: seqNum,
					PurchaseOrders: chunk,
					IsLastBatch:    false,
					Timestamp:      time.Now(),
				}
				seqNum++
				resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
				if err != nil {
					return totalStaged, 0, lastWatermark, err
				}
				totalStaged += int64(resp.RecordsStaged)
				lastWatermark = chunk[len(chunk)-1].PODate.Format(time.RFC3339)
				e.broadcastProgress(runID, domain, totalStaged, -1)
			}
		}
		return totalStaged, 0, lastWatermark, nil

	case contracts.DomainGoodsReceipts, "goods_receipts":
		ext := extractors.NewPurchasingExtractor(mssqlClient)
		var totalStaged int64
		var lastWatermark string
		var seqNum int
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		for _, win := range dateWindows(start, time.Now(), 3) {
			select {
			case <-ctx.Done():
				return totalStaged, 0, lastWatermark, ctx.Err()
			default:
			}
			receipts, err := ext.ExtractGoodsReceipts(ctx, win[0], win[1])
			if err != nil {
				return totalStaged, 0, lastWatermark, fmt.Errorf("goods_receipts window %s–%s: %w", win[0].Format("2006-01"), win[1].Format("2006-01"), err)
			}
			if len(receipts) == 0 {
				continue
			}
			for i := 0; i < len(receipts); i += batchSize {
				end := i + batchSize
				if end > len(receipts) {
					end = len(receipts)
				}
				chunk := receipts[i:end]
				payload := &contracts.MigrationBatchPayload{
					BatchID:        uuid.New().String(),
					RunID:          runID,
					OrganizationID: cloudCfg.OrganizationID,
					Domain:         domain,
					SequenceNumber: seqNum,
					GoodsReceipts:  chunk,
					IsLastBatch:    false,
					Timestamp:      time.Now(),
				}
				seqNum++
				resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
				if err != nil {
					return totalStaged, 0, lastWatermark, err
				}
				totalStaged += int64(resp.RecordsStaged)
				lastWatermark = chunk[len(chunk)-1].ReceiptDate.Format(time.RFC3339)
				e.broadcastProgress(runID, domain, totalStaged, -1)
			}
		}
		return totalStaged, 0, lastWatermark, nil

	case contracts.DomainStockMovements:
		ext := extractors.NewStockMovementExtractor(mssqlClient)
		var totalStaged int64
		var lastWatermark string
		var seqNum int
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		for _, win := range dateWindows(start, time.Now(), 1) {
			select {
			case <-ctx.Done():
				return totalStaged, 0, lastWatermark, ctx.Err()
			default:
			}
			moves, err := ext.ExtractStockMovements(ctx, win[0], win[1])
			if err != nil {
				return totalStaged, 0, lastWatermark, fmt.Errorf("stock_movements window %s–%s: %w", win[0].Format("2006-01"), win[1].Format("2006-01"), err)
			}
			if len(moves) == 0 {
				continue
			}
			for i := 0; i < len(moves); i += batchSize {
				end := i + batchSize
				if end > len(moves) {
					end = len(moves)
				}
				chunk := moves[i:end]
				payload := &contracts.MigrationBatchPayload{
					BatchID:        uuid.New().String(),
					RunID:          runID,
					OrganizationID: cloudCfg.OrganizationID,
					Domain:         domain,
					SequenceNumber: seqNum,
					StockMovements: chunk,
					IsLastBatch:    false,
					Timestamp:      time.Now(),
				}
				seqNum++
				resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
				if err != nil {
					return totalStaged, 0, lastWatermark, err
				}
				totalStaged += int64(resp.RecordsStaged)
				lastWatermark = chunk[len(chunk)-1].MovementDate.Format(time.RFC3339)
				e.broadcastProgress(runID, domain, totalStaged, -1)
			}
		}
		return totalStaged, 0, lastWatermark, nil

	case contracts.DomainSalesOrders:
		ext := extractors.NewSalesExtractor(mssqlClient)
		var totalStaged int64
		var lastWatermark string
		var seqNum int
		// Stream in 3-month windows to avoid loading the entire history into memory.
		for _, win := range dateWindows(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), time.Now(), 3) {
			select {
			case <-ctx.Done():
				return totalStaged, 0, lastWatermark, ctx.Err()
			default:
			}
			orders, err := ext.ExtractSalesOrders(ctx, win[0], win[1])
			if err != nil {
				return totalStaged, 0, lastWatermark, fmt.Errorf("sales_orders window %s–%s: %w", win[0].Format("2006-01"), win[1].Format("2006-01"), err)
			}
			if len(orders) == 0 {
				continue
			}
			for i := 0; i < len(orders); i += batchSize {
				end := i + batchSize
				if end > len(orders) {
					end = len(orders)
				}
				chunk := orders[i:end]
				payload := &contracts.MigrationBatchPayload{
					BatchID:        uuid.New().String(),
					RunID:          runID,
					OrganizationID: cloudCfg.OrganizationID,
					Domain:         domain,
					SequenceNumber: seqNum,
					SalesOrders:    chunk,
					IsLastBatch:    false, // windows may continue
					Timestamp:      time.Now(),
				}
				seqNum++
				resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
				if err != nil {
					return totalStaged, 0, lastWatermark, err
				}
				totalStaged += int64(resp.RecordsStaged)
				lastWatermark = chunk[len(chunk)-1].OrderDate.Format(time.RFC3339)
				e.broadcastProgress(runID, domain, totalStaged, -1)
			}
		}
		return totalStaged, 0, lastWatermark, nil

	case contracts.DomainInvoices:
		ext := extractors.NewSalesExtractor(mssqlClient)
		var totalStaged int64
		var lastWatermark string
		var seqNum int
		start := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		if e.cfg != nil && e.cfg.InvoiceStartDate != "" {
			if parsed, pErr := time.Parse("2006-01-02", e.cfg.InvoiceStartDate); pErr == nil {
				start = parsed
				log.Printf("[ETL] Resuming invoices from configured start date: %s", start.Format("2006-01-02"))
			}
		}
		// Stream in 3-month windows to bound memory usage.
		// Each window extracts headers+lines for that period only, sends batches,
		// then releases memory before the next window starts.
		for _, win := range dateWindows(start, time.Now(), 3) {
			select {
			case <-ctx.Done():
				return totalStaged, 0, lastWatermark, ctx.Err()
			default:
			}
			var invoices []mappings.CanonicalInvoice
			var extErr error
			for attempt := 1; attempt <= 5; attempt++ {
				invoices, extErr = ext.ExtractInvoices(ctx, win[0], win[1])
				if extErr == nil {
					break
				}
				log.Printf("[ETL] Warning: invoices window %s–%s attempt %d/5 failed: %v. Retrying in %ds...",
					win[0].Format("2006-01"), win[1].Format("2006-01"), attempt, extErr, attempt*10)
				select {
				case <-ctx.Done():
					return totalStaged, 0, lastWatermark, ctx.Err()
				case <-time.After(time.Duration(attempt*10) * time.Second):
				}
				// Refresh MSSQL client connection in case of network reset
				if refreshed, rErr := e.getMSSQLClient(); rErr == nil {
					ext = extractors.NewSalesExtractor(refreshed)
				}
			}
			if extErr != nil {
				return totalStaged, 0, lastWatermark, fmt.Errorf("invoices window %s–%s: %w", win[0].Format("2006-01"), win[1].Format("2006-01"), extErr)
			}
			if len(invoices) == 0 {
				continue
			}
			for i := 0; i < len(invoices); i += batchSize {
				end := i + batchSize
				if end > len(invoices) {
					end = len(invoices)
				}
				chunk := invoices[i:end]
				payload := &contracts.MigrationBatchPayload{
					BatchID:        uuid.New().String(),
					RunID:          runID,
					OrganizationID: cloudCfg.OrganizationID,
					Domain:         domain,
					SequenceNumber: seqNum,
					Invoices:       chunk,
					IsLastBatch:    false, // windows may continue
					Timestamp:      time.Now(),
				}
				seqNum++
				resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
				if err != nil {
					return totalStaged, 0, lastWatermark, err
				}
				totalStaged += int64(resp.RecordsStaged)
				lastWatermark = chunk[len(chunk)-1].InvoiceDate.Format(time.RFC3339)
				e.broadcastProgress(runID, domain, totalStaged, -1)
			}
		}
		return totalStaged, 0, lastWatermark, nil

	case contracts.DomainIncomingPayments:
		ext := extractors.NewPaymentExtractor(mssqlClient)
		var totalStaged int64
		var lastWatermark string
		var seqNum int
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		for _, win := range dateWindows(start, time.Now(), 3) {
			select {
			case <-ctx.Done():
				return totalStaged, 0, lastWatermark, ctx.Err()
			default:
			}
			payments, err := ext.ExtractIncomingPayments(ctx, win[0], win[1])
			if err != nil {
				return totalStaged, 0, lastWatermark, fmt.Errorf("incoming_payments window %s–%s: %w", win[0].Format("2006-01"), win[1].Format("2006-01"), err)
			}
			if len(payments) == 0 {
				continue
			}
			for i := 0; i < len(payments); i += batchSize {
				end := i + batchSize
				if end > len(payments) {
					end = len(payments)
				}
				chunk := payments[i:end]
				payload := &contracts.MigrationBatchPayload{
					BatchID:          uuid.New().String(),
					RunID:            runID,
					OrganizationID:   cloudCfg.OrganizationID,
					Domain:           domain,
					SequenceNumber:   seqNum,
					IncomingPayments: chunk,
					IsLastBatch:      false,
					Timestamp:        time.Now(),
				}
				seqNum++
				resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
				if err != nil {
					return totalStaged, 0, lastWatermark, err
				}
				totalStaged += int64(resp.RecordsStaged)
				lastWatermark = chunk[len(chunk)-1].PaymentDate.Format(time.RFC3339)
				e.broadcastProgress(runID, domain, totalStaged, -1)
			}
		}
		return totalStaged, 0, lastWatermark, nil

	case contracts.DomainPriceLists:
		ext := extractors.NewCatalogExtractor(mssqlClient)
		priceLists, err := ext.ExtractPriceLists(ctx)
		if err != nil {
			return 0, 0, "", err
		}
		priceItems, err := ext.ExtractPriceListItems(ctx)
		if err != nil {
			return 0, 0, "", err
		}
		var totalStaged int64
		// Send price list headers first (usually small, single batch)
		if len(priceLists) > 0 {
			headerPayload := &contracts.MigrationBatchPayload{
				BatchID:        uuid.New().String(),
				RunID:          runID,
				OrganizationID: cloudCfg.OrganizationID,
				Domain:         domain,
				SequenceNumber: 0,
				PriceLists:     priceLists,
				IsLastBatch:    len(priceItems) == 0,
				Timestamp:      time.Now(),
			}
			resp, err := cloudClient.SendBatchWithRetry(ctx, headerPayload)
			if err != nil {
				return 0, 0, "", err
			}
			totalStaged += int64(resp.RecordsStaged)
		}
		// Send price items in chunks
		for i := 0; i < len(priceItems); i += batchSize {
			end := i + batchSize
			if end > len(priceItems) {
				end = len(priceItems)
			}
			chunk := priceItems[i:end]
			payload := &contracts.MigrationBatchPayload{
				BatchID:        uuid.New().String(),
				RunID:          runID,
				OrganizationID: cloudCfg.OrganizationID,
				Domain:         domain,
				SequenceNumber: (i / batchSize) + 1,
				PriceItems:     chunk,
				IsLastBatch:    end == len(priceItems),
				Timestamp:      time.Now(),
			}
			resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
			if err != nil {
				return totalStaged, 0, "", err
			}
			totalStaged += int64(resp.RecordsStaged)
			e.broadcastProgress(runID, domain, totalStaged, int64(len(priceItems)))
		}
		return totalStaged, 0, "", nil

	case contracts.DomainBPAddresses:
		ext := extractors.NewPartnersExtractor(mssqlClient)
		addresses, err := ext.ExtractBPAddresses(ctx)
		if err != nil {
			return 0, 0, "", err
		}
		var totalStaged int64
		for i := 0; i < len(addresses); i += batchSize {
			end := i + batchSize
			if end > len(addresses) {
				end = len(addresses)
			}
			chunk := addresses[i:end]
			payload := &contracts.MigrationBatchPayload{
				BatchID:        uuid.New().String(),
				RunID:          runID,
				OrganizationID: cloudCfg.OrganizationID,
				Domain:         domain,
				SequenceNumber: i / batchSize,
				BPAddresses:    chunk,
				IsLastBatch:    end == len(addresses),
				Timestamp:      time.Now(),
			}
			resp, err := cloudClient.SendBatchWithRetry(ctx, payload)
			if err != nil {
				return totalStaged, 0, "", err
			}
			totalStaged += int64(resp.RecordsStaged)
			e.broadcastProgress(runID, domain, totalStaged, int64(len(addresses)))
		}
		return totalStaged, 0, "", nil
	}

	return 0, 0, "", fmt.Errorf("unsupported migration domain: %s", domain)
}

// broadcastProgress sends a step_progress event during chunked batch processing.
// Pass total = -1 when the total is unknown (streaming window mode).
func (e *Engine) broadcastProgress(runID string, domain contracts.DomainType, processed, total int64) {
	pct := 0.0
	if total > 0 {
		pct = float64(processed) / float64(total) * 100
	}
	msg := fmt.Sprintf("Processing %s: %d records ingested", domain, processed)
	if total > 0 {
		msg = fmt.Sprintf("Processing %s: %d/%d records", domain, processed, total)
	}
	e.broadcast(ProgressEvent{
		Type:           "step_progress",
		RunID:          runID,
		Domain:         domain,
		Status:         contracts.StatusRunning,
		ProcessedCount: processed,
		TotalRecords:   total,
		Percentage:     pct,
		Message:        msg,
		Timestamp:      time.Now(),
	})
}

// dateWindows splits [start, end] into consecutive N-month windows.
// Returns a slice of [windowStart, windowEnd] pairs covering the full range.
func dateWindows(start, end time.Time, monthsPerWindow int) [][2]time.Time {
	var windows [][2]time.Time
	win := start
	for win.Before(end) {
		next := win.AddDate(0, monthsPerWindow, 0)
		if next.After(end) {
			next = end
		}
		windows = append(windows, [2]time.Time{win, next})
		win = next
	}
	return windows
}
