package pipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/nembus"
	"github.com/NasTecSol/nembus-sap-agent/internal/sap"
)

type DownstreamSync struct {
	cfg          *config.Config
	sapClient    *sap.Client
	nembusClient *nembus.Client
	lastSyncTime time.Time
}

func NewDownstreamSync(cfg *config.Config, sapClient *sap.Client, nembusClient *nembus.Client) *DownstreamSync {
	return &DownstreamSync{
		cfg:          cfg,
		sapClient:    sapClient,
		nembusClient: nembusClient,
	}
}

type SyncStats struct {
	Categories int
	UOMs       int
	Products   int
	Barcodes   int
	Prices     int
	Duration   time.Duration
}

func (s *DownstreamSync) SyncAll(ctx context.Context, fullSync bool) (*SyncStats, error) {
	start := time.Now()
	stats := &SyncStats{}

	log.Println("⬇️ [SAP -> Nembus] Starting Downstream Master Data Sync...")

	var watermark *time.Time
	if !fullSync && !s.lastSyncTime.IsZero() {
		watermark = &s.lastSyncTime
	}

	// 1. Sync Categories / Item Groups
	categoryMap := make(map[int]int32)
	itemGroups, err := s.sapClient.FetchItemGroups(ctx)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to fetch Item Groups from SAP: %v", err)
	} else {
		for _, grp := range itemGroups {
			catCode := fmt.Sprintf("SAP_GRP_%d", grp.Number)
			if id, err := s.nembusClient.UpsertCategory(ctx, catCode, grp.Name); err == nil {
				categoryMap[grp.Number] = id
				stats.Categories++
			}
		}
		log.Printf("  Synced %d product categories", stats.Categories)
	}

	// 2. Sync Units of Measure
	uomMap := make(map[int]int32)
	uoms, err := s.sapClient.FetchUnitsOfMeasure(ctx)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to fetch UOMs from SAP: %v", err)
	} else {
		for _, u := range uoms {
			if id, err := s.nembusClient.UpsertUoM(ctx, u.Code, u.Name); err == nil {
				uomMap[u.Entry] = id
				stats.UOMs++
			}
		}
		log.Printf("  Synced %d units of measure", stats.UOMs)
	}

	// 3. Sync Products, Barcodes, and Prices in batches
	skip := 0
	top := s.cfg.BatchSize
	if top <= 0 {
		top = 100
	}

	for {
		items, hasNext, err := s.sapClient.FetchItems(ctx, watermark, top, skip)
		if err != nil {
			return stats, fmt.Errorf("failed fetching SAP items batch (skip=%d): %w", skip, err)
		}

		for _, item := range items {
			if item.ItemCode == "" {
				continue
			}

			// Map Category
			var categoryID *int32
			if cid, exists := categoryMap[item.ItemsGroupCode]; exists {
				categoryID = &cid
			}

			// Map Base UOM
			var baseUoMID *int32
			if item.DefaultSalesUoMEntry != nil {
				if uid, exists := uomMap[*item.DefaultSalesUoMEntry]; exists {
					baseUoMID = &uid
				}
			}

			// Tax category default
			taxCatID := s.cfg.NembusTaxCategoryID

			// Check if weighted / scale item
			isWeighted := false
			if item.InventoryUOM == "KG" || item.InventoryUOM == "G" {
				isWeighted = true
			}

			if s.cfg.DryRun {
				log.Printf("[DRY-RUN] Would upsert product SKU: %s (%s)", item.ItemCode, item.ItemName)
				stats.Products++
				continue
			}

			// Upsert Product Header
			costPrice := item.AvgPrice
			if costPrice <= 0 {
				costPrice = item.MovingAveragePrice
			}

			prodID, err := s.nembusClient.UpsertProduct(ctx, item.ItemCode, item.ItemName, item.ForeignName, categoryID, baseUoMID, &taxCatID, isWeighted, costPrice)
			if err != nil {
				log.Printf("❌ Failed to upsert product SKU %s: %v", item.ItemCode, err)
				continue
			}
			stats.Products++

			// Sync Primary Barcode (from item header if present)
			if item.BarCode != "" {
				if err := s.nembusClient.UpsertBarcode(ctx, prodID, item.BarCode, true); err == nil {
					stats.Barcodes++
				}
			}

			// Sync Extended Barcodes
			for _, bc := range item.ItemBarCodeCollection {
				if bc.Barcode != "" && bc.Barcode != item.BarCode {
					if err := s.nembusClient.UpsertBarcode(ctx, prodID, bc.Barcode, false); err == nil {
						stats.Barcodes++
					}
				}
			}

			// Query approval policy rules for catalog price changes
			rules, requiresApproval, _ := s.nembusClient.GetApprovalPolicyRules(ctx, "catalog", "price_change")
			priceDropRequiresApproval := true
			costIncreaseThresholdPct := 10.0
			belowCostRequiresApproval := true

			if requiresApproval && rules != nil {
				if v, ok := rules["price_drop_requires_approval"].(bool); ok {
					priceDropRequiresApproval = v
				}
				if v, ok := rules["cost_increase_threshold_pct"].(float64); ok {
					costIncreaseThresholdPct = v
				}
				if v, ok := rules["below_cost_requires_approval"].(bool); ok {
					belowCostRequiresApproval = v
				}
			}

			// Check Cost Price Governance
			currentCost, _ := s.nembusClient.GetProductCostPrice(ctx, prodID)
			if requiresApproval && currentCost > 0 && costPrice > currentCost {
				pctIncrease := ((costPrice - currentCost) / currentCost) * 100.0
				if pctIncrease >= costIncreaseThresholdPct {
					reason := fmt.Sprintf("Cost price increased by %.1f%% (from %.2f to %.2f, threshold: %.1f%%)", pctIncrease, currentCost, costPrice, costIncreaseThresholdPct)
					log.Printf("⚠️ %s for SKU %s. Routing to approval review...", reason, item.ItemCode)
					_ = s.nembusClient.SubmitPriceApprovalRequest(ctx, prodID, item.ItemCode, s.cfg.NembusPriceListID, nil, nil, nil, &currentCost, &costPrice, reason)
				}
			}

			// Sync Prices with Governance
			for _, price := range item.ItemPrices {
				if price.PriceList == int(s.cfg.NembusPriceListID) && price.Price > 0 {
					var uomEntry *int32
					if price.UoMEntry != nil {
						if uid, exists := uomMap[*price.UoMEntry]; exists {
							uomEntry = &uid
						}
					}

					oldPrice, priceExists, _ := s.nembusClient.GetActiveProductPrice(ctx, prodID, s.cfg.NembusPriceListID)

					priceReviewTriggered := false
					var priceReviewReason string

					if requiresApproval && priceExists && oldPrice > 0 {
						if priceDropRequiresApproval && price.Price < oldPrice {
							priceReviewTriggered = true
							priceReviewReason = fmt.Sprintf("Selling price dropped from %.2f to %.2f", oldPrice, price.Price)
						} else if belowCostRequiresApproval && costPrice > 0 && price.Price < costPrice {
							priceReviewTriggered = true
							priceReviewReason = fmt.Sprintf("Selling price %.2f is below cost price %.2f", price.Price, costPrice)
						}
					}

					if priceReviewTriggered {
						log.Printf("⚠️ %s for SKU %s. Holding active POS price at %.2f and routing to approval review...", priceReviewReason, item.ItemCode, oldPrice)
						_ = s.nembusClient.SubmitPriceApprovalRequest(ctx, prodID, item.ItemCode, s.cfg.NembusPriceListID, uomEntry, &oldPrice, &price.Price, &currentCost, &costPrice, priceReviewReason)
					} else {
						if err := s.nembusClient.UpsertProductPrice(ctx, prodID, s.cfg.NembusPriceListID, uomEntry, price.Price); err == nil {
							stats.Prices++
						}
					}
				}
			}

		}

		if !hasNext || len(items) < top {
			break
		}
		skip += len(items)
	}

	stats.Duration = time.Since(start)
	s.lastSyncTime = time.Now()

	log.Printf(" [SAP -> Nembus] Sync complete in %v: %d products, %d barcodes, %d prices",
		stats.Duration, stats.Products, stats.Barcodes, stats.Prices)
	return stats, nil
}
