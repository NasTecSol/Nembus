package pipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	coreusecase "github.com/NasTecSol/nembus-core/usecase"
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

	rules, requiresApproval, policyErr := s.nembusClient.GetApprovalPolicyRules(ctx, "catalog", "price_change")
	if policyErr != nil {
		log.Printf("⚠️ Warning: could not load catalog.price_change policy: %v", policyErr)
	}
	catalogRules := coreusecase.ParseCatalogPriceRules(rules)

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

			var categoryID *int32
			if cid, exists := categoryMap[item.ItemsGroupCode]; exists {
				categoryID = &cid
			}

			var baseUoMID *int32
			if item.DefaultSalesUoMEntry != nil {
				if uid, exists := uomMap[*item.DefaultSalesUoMEntry]; exists {
					baseUoMID = &uid
				}
			}

			taxCatID := s.cfg.NembusTaxCategoryID

			isWeighted := false
			if item.InventoryUOM == "KG" || item.InventoryUOM == "G" {
				isWeighted = true
			}

			if s.cfg.DryRun {
				log.Printf("[DRY-RUN] Would upsert product SKU: %s (%s)", item.ItemCode, item.ItemName)
				stats.Products++
				continue
			}

			costPrice := item.AvgPrice
			if costPrice <= 0 {
				costPrice = item.MovingAveragePrice
			}

			currentCost, err := s.nembusClient.GetProductCostPriceBySKU(ctx, item.ItemCode)
			if err != nil {
				log.Printf("⚠️ Could not read cost for SKU %s: %v", item.ItemCode, err)
			}

			holdCost, costReason := coreusecase.EvaluateCostChange(requiresApproval, catalogRules, currentCost, costPrice)
			costToApply := costPrice
			if holdCost {
				log.Printf("⚠️ %s for SKU %s. Deferring cost update to approval review...", costReason, item.ItemCode)
				costToApply = 0
			}

			prodID, err := s.nembusClient.UpsertProduct(ctx, item.ItemCode, item.ItemName, item.ForeignName, categoryID, baseUoMID, &taxCatID, isWeighted, costToApply)
			if err != nil {
				log.Printf("❌ Failed to upsert product SKU %s: %v", item.ItemCode, err)
				continue
			}
			stats.Products++

			if holdCost {
				if err := s.nembusClient.SubmitPriceApprovalRequest(ctx, prodID, item.ItemCode, s.cfg.NembusPriceListID, nil, nil, nil, &currentCost, &costPrice, costReason); err != nil {
					log.Printf("❌ Failed to submit cost approval for SKU %s: %v", item.ItemCode, err)
				}
			}

			if item.BarCode != "" {
				if err := s.nembusClient.UpsertBarcode(ctx, prodID, item.BarCode, true); err == nil {
					stats.Barcodes++
				}
			}

			for _, bc := range item.ItemBarCodeCollection {
				if bc.Barcode != "" && bc.Barcode != item.BarCode {
					if err := s.nembusClient.UpsertBarcode(ctx, prodID, bc.Barcode, false); err == nil {
						stats.Barcodes++
					}
				}
			}

			effectiveCost := currentCost
			if costToApply > 0 {
				effectiveCost = costPrice
			}

			for _, price := range item.ItemPrices {
				if price.PriceList == int(s.cfg.NembusPriceListID) && price.Price > 0 {
					var uomEntry *int32
					if price.UoMEntry != nil {
						if uid, exists := uomMap[*price.UoMEntry]; exists {
							uomEntry = &uid
						}
					}

					oldPrice, priceExists, err := s.nembusClient.GetActiveProductPrice(ctx, prodID, s.cfg.NembusPriceListID, uomEntry)
					if err != nil {
						log.Printf("⚠️ Could not read active price for SKU %s: %v", item.ItemCode, err)
					}

					if hold, reason := coreusecase.EvaluateSellPriceChange(requiresApproval, catalogRules, oldPrice, price.Price, effectiveCost, priceExists); hold {
						log.Printf("⚠️ %s for SKU %s. Holding active price and routing to approval review...", reason, item.ItemCode)
						if err := s.nembusClient.SubmitPriceApprovalRequest(ctx, prodID, item.ItemCode, s.cfg.NembusPriceListID, uomEntry, &oldPrice, &price.Price, &currentCost, &costPrice, reason); err != nil {
							log.Printf("❌ Failed to submit price approval for SKU %s: %v", item.ItemCode, err)
						}
						continue
					}

					if err := s.nembusClient.UpsertProductPrice(ctx, prodID, s.cfg.NembusPriceListID, uomEntry, price.Price); err == nil {
						stats.Prices++
					} else {
						log.Printf("❌ Failed to upsert price for SKU %s: %v", item.ItemCode, err)
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
