package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap-agent/internal/etl/extractors"
	"github.com/NasTecSol/nembus-sap-agent/internal/transport"
	"github.com/NasTecSol/nembus-sap/contracts"
	"github.com/NasTecSol/nembus-sap/mappings"
)

var sourceItemCodeMetadataKeys = []string{
	"sap_item_code",
	"source_item_code",
}

type matchedProduct struct {
	product        mappings.CanonicalProduct
	metadataKey    string
	sourceItemCode string
}

func main() {
	configPath := flag.String("config", "", "Path to the explicitly supplied SAP Agent configuration file")
	itemCode := flag.String("item", "", "SAP ItemCode to extract and optionally send")
	send := flag.Bool("send", false, "Send the single matched product to Cloud")
	flag.Parse()

	if err := run(context.Background(), *configPath, *itemCode, *send); err != nil {
		fmt.Fprintln(os.Stderr, "controlled product runner failed:", err)
		os.Exit(1)
	}
}

func run(parent context.Context, configPath, requestedItemCode string, send bool) error {
	configPath = strings.TrimSpace(configPath)
	requestedItemCode = strings.TrimSpace(requestedItemCode)
	if configPath == "" {
		return errors.New("-config must be nonblank")
	}
	if requestedItemCode == "" {
		return errors.New("-item must be nonblank")
	}

	configInfo, err := os.Stat(configPath)
	if err != nil {
		return fmt.Errorf("cannot access explicitly supplied config: %w", err)
	}
	if configInfo.IsDir() {
		return errors.New("-config must name a config file")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load supplied config: %w", err)
	}
	if strings.TrimSpace(cfg.Cloud.TenantSlug) == "" {
		return errors.New("cloud tenant slug must be nonblank")
	}
	if cfg.Cloud.OrganizationID <= 0 {
		return errors.New("cloud organization ID must be positive")
	}
	if strings.TrimSpace(cfg.Cloud.M2MToken) == "" {
		return errors.New("cloud M2M token must be nonblank")
	}

	ctx, stop := signal.NotifyContext(parent, os.Interrupt)
	defer stop()

	mssqlClient, err := db.NewMSSQLClient(cfg.MSSQL)
	if err != nil {
		return fmt.Errorf("failed to connect to SAP SQL Server: %w", err)
	}
	defer mssqlClient.Close()

	products, err := extractors.NewCatalogExtractor(mssqlClient).ExtractProducts(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract SAP products: %w", err)
	}

	match, err := findExactlyOneProduct(products, requestedItemCode)
	if err != nil {
		return err
	}

	if !send {
		printSummary(match)
		return nil
	}

	payload := &contracts.MigrationBatchPayload{
		BatchID:        uuid.New().String(),
		RunID:          uuid.New().String(),
		OrganizationID: cfg.Cloud.OrganizationID,
		Domain:         contracts.DomainProducts,
		SequenceNumber: 0,
		IsLastBatch:    true,
		Timestamp:      time.Now(),
		Products:       []mappings.CanonicalProduct{match.product},
	}
	if payload.RecordCount() != 1 || len(payload.Products) != 1 {
		return errors.New("safety check failed: payload does not contain exactly one product")
	}

	response, err := transport.NewCloudClient(cfg.Cloud).SendBatchWithRetry(ctx, payload)
	if err != nil {
		return fmt.Errorf("failed to send controlled product batch: %w", err)
	}
	if response == nil {
		return errors.New("cloud returned an empty migration response")
	}

	fmt.Printf("RecordsStaged: %d\n", response.RecordsStaged)
	fmt.Printf("RecordsFailed: %d\n", response.RecordsFailed)
	for _, responseError := range response.Errors {
		fmt.Printf("Error: %q\n", responseError)
	}
	return nil
}

func findExactlyOneProduct(products []mappings.CanonicalProduct, requestedItemCode string) (matchedProduct, error) {
	var match matchedProduct
	matchCount := 0

	for _, product := range products {
		metadataKey, sourceItemCode, present, err := authoritativeSourceItemCode(product)
		if err != nil {
			return matchedProduct{}, err
		}
		if !present || sourceItemCode != requestedItemCode {
			continue
		}
		if strings.TrimSpace(product.SKU) != sourceItemCode {
			return matchedProduct{}, fmt.Errorf("canonical SKU/source metadata mismatch for source ItemCode %q", sourceItemCode)
		}

		matchCount++
		match = matchedProduct{
			product:        product,
			metadataKey:    metadataKey,
			sourceItemCode: sourceItemCode,
		}
	}

	switch matchCount {
	case 0:
		return matchedProduct{}, fmt.Errorf("no canonical product matched source ItemCode %q", requestedItemCode)
	case 1:
		return match, nil
	default:
		return matchedProduct{}, fmt.Errorf("expected exactly one canonical product for source ItemCode %q, found %d", requestedItemCode, matchCount)
	}
}

func authoritativeSourceItemCode(product mappings.CanonicalProduct) (string, string, bool, error) {
	var selectedKey string
	var selectedCode string

	for _, key := range sourceItemCodeMetadataKeys {
		value, exists := product.Metadata[key]
		if !exists || value == nil {
			continue
		}
		code, ok := value.(string)
		if !ok {
			return "", "", false, fmt.Errorf("canonical product metadata %q is not a string", key)
		}
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if selectedCode != "" && selectedCode != code {
			return "", "", false, fmt.Errorf("canonical product has conflicting source ItemCode metadata")
		}
		selectedKey = key
		selectedCode = code
	}

	if selectedCode == "" {
		return "", "", false, nil
	}
	return selectedKey, selectedCode, true, nil
}

func printSummary(match matchedProduct) {
	product := match.product
	fmt.Printf("source ItemCode: %q\n", match.sourceItemCode)
	fmt.Printf("canonical SKU: %q\n", product.SKU)
	fmt.Printf("name: %q\n", product.Name)
	fmt.Printf("description present: %t\n", strings.TrimSpace(product.Description) != "")
	fmt.Printf("category code: %q\n", product.CategoryCode)
	fmt.Printf("brand code: %q\n", product.BrandCode)
	fmt.Printf("base UoM code: %q\n", product.BaseUOMCode)
	fmt.Printf("sales UoM: %q\n", product.SalesUOMCode)
	fmt.Printf("purchase UoM: %q\n", product.PurchaseUOMCode)
	fmt.Printf("UoM group: %q\n", product.UOMGroupCode)
	fmt.Printf("product_type: %q\n", product.ProductType)
	fmt.Printf("active: %t\n", product.IsActive)
	fmt.Printf("sellable: %t\n", product.IsSellable)
	fmt.Printf("purchasable: %t\n", product.IsPurchasable)
	fmt.Printf("track inventory: %t\n", product.TrackInventory)
	fmt.Printf("serialized: %t\n", product.IsSerialized)
	fmt.Printf("batch managed: %t\n", product.IsBatchManaged)
}
