package main

import (
	"context"
	"fmt"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap-agent/internal/etl/extractors"
)

func main() {
	cfg, err := config.LoadConfig("agent_config.json")
	if err != nil {
		fmt.Printf("Config err: %v\n", err)
		return
	}

	client, err := db.NewMSSQLClient(cfg.MSSQL)
	if err != nil {
		fmt.Printf("MSSQL client err: %v\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// 1. Stores
	stExt := extractors.NewStoresExtractor(client)
	stores, locs, err := stExt.ExtractStores(ctx)
	fmt.Printf("1. Stores: %d stores, %d locs (err: %v)\n", len(stores), len(locs), err)

	// 2. Users
	uExt := extractors.NewUsersExtractor(client)
	users, cashiers, err := uExt.ExtractUsers(ctx, "01")
	fmt.Printf("2. Users: %d users, %d cashiers (err: %v)\n", len(users), len(cashiers), err)

	// 3. Categories & Brands & Products
	catExt := extractors.NewCatalogExtractor(client)
	cats, err := catExt.ExtractCategories(ctx)
	fmt.Printf("3. Categories: %d (err: %v)\n", len(cats), err)

	brands, err := catExt.ExtractBrands(ctx)
	fmt.Printf("4. Brands: %d (err: %v)\n", len(brands), err)

	prods, err := catExt.ExtractProducts(ctx)
	fmt.Printf("5. Products: %d (err: %v)\n", len(prods), err)

	bcs, err := catExt.ExtractBarcodes(ctx)
	fmt.Printf("6. Barcodes: %d (err: %v)\n", len(bcs), err)

	// 4. Inventory
	invExt := extractors.NewInventoryExtractor(client)
	inv, err := invExt.ExtractInventory(ctx)
	fmt.Printf("7. Inventory: %d (err: %v)\n", len(inv), err)

	// 5. Partners
	ptExt := extractors.NewPartnersExtractor(client)
	pts, err := ptExt.ExtractPartners(ctx)
	fmt.Printf("8. Partners: %d (err: %v)\n", len(pts), err)

	// 6. Sales
	sExt := extractors.NewSalesExtractor(client)
	orders, err := sExt.ExtractSalesOrders(ctx, time.Time{}, time.Time{})
	fmt.Printf("9. Sales Orders: %d (err: %v)\n", len(orders), err)

	invs, err := sExt.ExtractInvoices(ctx, time.Time{}, time.Time{})
	fmt.Printf("10. Invoices: %d (err: %v)\n", len(invs), err)
}
