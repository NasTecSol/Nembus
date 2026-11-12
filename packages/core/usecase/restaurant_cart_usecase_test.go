package usecase

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.015
}

func TestComputeLineTaxAndTotals(t *testing.T) {
	t.Run("exclusive VAT with modifiers and combo adjustments", func(t *testing.T) {
		var rate pgtype.Numeric
		_ = rate.Scan("15.00")
		tc := &repository.TaxCategory{
			TaxRate:     rate,
			IsInclusive: pgtype.Bool{Bool: false, Valid: true},
		}
		// unit=20, modifier=+3, combo=+2 -> effective unit=25, qty=2 -> net=50, VAT 15%=7.50, total=57.50
		taxRate, taxAmt, subtotal, lineTotal := computeLineTaxAndTotals(2, 20, 3, 2, 0, tc)
		if !almostEqual(taxRate, 15.0) || !almostEqual(subtotal, 50.0) || !almostEqual(taxAmt, 7.50) || !almostEqual(lineTotal, 57.50) {
			t.Fatalf("unexpected exclusive calc: rate=%.2f sub=%.2f tax=%.2f total=%.2f", taxRate, subtotal, taxAmt, lineTotal)
		}
	})

	t.Run("inclusive VAT extracts tax from line_total", func(t *testing.T) {
		var rate pgtype.Numeric
		_ = rate.Scan("15.00")
		tc := &repository.TaxCategory{
			TaxRate:     rate,
			IsInclusive: pgtype.Bool{Bool: true, Valid: true},
		}
		// unit=23 -> effective 23, qty=2 -> lineTotal=46.00, tax = 46*(15/115)=6.00, subtotal=40.00
		taxRate, taxAmt, subtotal, lineTotal := computeLineTaxAndTotals(2, 23, 0, 0, 0, tc)
		if !almostEqual(taxRate, 15.0) || !almostEqual(subtotal, 40.0) || !almostEqual(taxAmt, 6.00) || !almostEqual(lineTotal, 46.00) {
			t.Fatalf("unexpected inclusive calc: rate=%.2f sub=%.2f tax=%.2f total=%.2f", taxRate, subtotal, taxAmt, lineTotal)
		}
	})

	t.Run("discount clamps at gross and zero tax when nil", func(t *testing.T) {
		taxRate, taxAmt, subtotal, lineTotal := computeLineTaxAndTotals(1, 10, 0, 0, 50, nil)
		if taxRate != 0 || taxAmt != 0 || subtotal != 0 || lineTotal != 0 {
			t.Fatalf("expected clamped zeros, got rate=%.2f sub=%.2f tax=%.2f total=%.2f", taxRate, subtotal, taxAmt, lineTotal)
		}
	})
}

func setupRestaurantMigrationDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("NEMBUS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_TEST_DATABASE_URL to run PostgreSQL restaurant cart-to-order integration tests")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	t.Cleanup(pool.Close)

	// Reset public schema cleanly inside the dedicated disposable test DB
	if _, err := pool.Exec(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
		t.Fatalf("reset public schema (use a dedicated disposable test DB): %v", err)
	}

	migrations := []string{
		filepath.Join("..", "db", "migrations", "20260918111505_initial_schema.sql"),
	}
	for _, migPath := range migrations {
		sqlBytes, err := os.ReadFile(migPath)
		if err != nil {
			t.Fatalf("read migration %s: %v", migPath, err)
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			t.Fatalf("exec migration %s: %v", migPath, err)
		}
	}
	return pool
}

type seededRestaurantFixture struct {
	OrgID        int32
	StoreID      int32
	TableID      int32
	CustomerID   int32
	MenuItemID   int32
	ModifierID   int32
	ComboItemID  int32
	ComboCompID  int32
}

func seedRestaurantTestData(t *testing.T, ctx context.Context, pool *pgxpool.Pool) seededRestaurantFixture {
	t.Helper()
	var fix seededRestaurantFixture

	if err := pool.QueryRow(ctx, `
		INSERT INTO organizations (name, code) VALUES ('Test Org', 'TORG') RETURNING id
	`).Scan(&fix.OrgID); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	if err := pool.QueryRow(ctx, `
		INSERT INTO stores (organization_id, name, code, store_type) VALUES ($1, 'Main Branch', 'MAIN', 'restaurant') RETURNING id
	`, fix.OrgID).Scan(&fix.StoreID); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	if err := pool.QueryRow(ctx, `
		INSERT INTO restaurant_tables (store_id, table_number, table_name, capacity) VALUES ($1, 'T1', 'Table 1', 4) RETURNING id
	`, fix.StoreID).Scan(&fix.TableID); err != nil {
		t.Fatalf("seed table: %v", err)
	}

	if err := pool.QueryRow(ctx, `
		INSERT INTO customers (organization_id, customer_code, name, email, phone)
		VALUES ($1, 'CUST-1', 'Alice', 'alice@example.com', '0500000001') RETURNING id
	`, fix.OrgID).Scan(&fix.CustomerID); err != nil {
		t.Fatalf("seed customer: %v", err)
	}

	var taxCatID int32
	if err := pool.QueryRow(ctx, `
		INSERT INTO tax_categories (name, code, tax_rate, is_inclusive) VALUES ('VAT 15%', 'VAT15', 15.00, false) RETURNING id
	`).Scan(&taxCatID); err != nil {
		t.Fatalf("seed tax category: %v", err)
	}

	var catID int32
	if err := pool.QueryRow(ctx, `
		INSERT INTO menu_categories (store_id, name, code) VALUES ($1, 'Burgers', 'BURGERS') RETURNING id
	`, fix.StoreID).Scan(&catID); err != nil {
		t.Fatalf("seed menu category: %v", err)
	}

	if err := pool.QueryRow(ctx, `
		INSERT INTO menu_items (store_id, menu_category_id, name, base_price, tax_category_id, is_available, is_active, item_type)
		VALUES ($1, $2, 'Signature Burger', 20.00, $3, true, true, 'standard') RETURNING id
	`, fix.StoreID, catID, taxCatID).Scan(&fix.MenuItemID); err != nil {
		t.Fatalf("seed menu item: %v", err)
	}

	if err := pool.QueryRow(ctx, `
		INSERT INTO menu_item_modifiers (menu_item_id, modifier_name, modifier_type, price_adjustment, is_active)
		VALUES ($1, 'Extra Cheese', 'addon', 4.00, true) RETURNING id
	`, fix.MenuItemID).Scan(&fix.ModifierID); err != nil {
		t.Fatalf("seed modifier: %v", err)
	}

	var sideID int32
	if err := pool.QueryRow(ctx, `
		INSERT INTO menu_items (store_id, menu_category_id, name, base_price, tax_category_id, is_available, is_active, item_type)
		VALUES ($1, $2, 'Fries', 8.00, $3, true, true, 'standard') RETURNING id
	`, fix.StoreID, catID, taxCatID).Scan(&sideID); err != nil {
		t.Fatalf("seed side item: %v", err)
	}

	if err := pool.QueryRow(ctx, `
		INSERT INTO menu_items (store_id, menu_category_id, name, base_price, tax_category_id, is_available, is_active, item_type)
		VALUES ($1, $2, 'Burger Meal', 30.00, $3, true, true, 'combo') RETURNING id
	`, fix.StoreID, catID, taxCatID).Scan(&fix.ComboItemID); err != nil {
		t.Fatalf("seed combo menu item: %v", err)
	}

	if err := pool.QueryRow(ctx, `
		INSERT INTO menu_item_combo_components (parent_menu_item_id, component_menu_item_id, group_name, price_adjustment)
		VALUES ($1, $2, 'Sides', 5.00) RETURNING id
	`, fix.ComboItemID, sideID).Scan(&fix.ComboCompID); err != nil {
		t.Fatalf("seed combo component: %v", err)
	}

	return fix
}

func TestRestaurantCartToOrderIntegration(t *testing.T) {
	pool := setupRestaurantMigrationDB(t)
	ctx := context.Background()
	fix := seedRestaurantTestData(t, ctx, pool)

	repo := repository.New(pool)
	uc := NewRestaurantCartUseCase()
	uc.SetRepository(repo)
	uc.SetPool(pool)

	// 1. Add item with modifier to table cart (qty=1: 20 + 4 = 24 net, VAT 15%=3.60, total=27.60)
	addResp := uc.AddToCart(ctx, RestaurantAddToCartInput{
		StoreID:     fix.StoreID,
		TableID:     &fix.TableID,
		MenuItemID:  fix.MenuItemID,
		Quantity:    1,
		ModifierIDs: []int32{fix.ModifierID},
	})
	if addResp.StatusCode != utils.CodeOK {
		t.Fatalf("AddToCart failed: %d %s", addResp.StatusCode, addResp.Message)
	}
	payload1 := addResp.Data.(map[string]interface{})
	cartOut1 := payload1["cart"].(RestaurantCartOutput)
	if !almostEqual(cartOut1.Subtotal, 24.00) || !almostEqual(cartOut1.TaxAmount, 3.60) || !almostEqual(cartOut1.TotalAmount, 27.60) {
		t.Fatalf("unexpected cart totals after add 1: sub=%.2f tax=%.2f tot=%.2f", cartOut1.Subtotal, cartOut1.TaxAmount, cartOut1.TotalAmount)
	}

	// 2. Add exact same item + modifier again -> merges into qty=2 (48 net, VAT=7.20, total=55.20)
	addResp2 := uc.AddToCart(ctx, RestaurantAddToCartInput{
		CartID:      &cartOut1.ID,
		MenuItemID:  fix.MenuItemID,
		Quantity:    1,
		ModifierIDs: []int32{fix.ModifierID},
	})
	if addResp2.StatusCode != utils.CodeOK {
		t.Fatalf("AddToCart merge failed: %d %s", addResp2.StatusCode, addResp2.Message)
	}
	payload2 := addResp2.Data.(map[string]interface{})
	items2 := payload2["items"].([]RestaurantCartItemOutput)
	if len(items2) != 1 || !almostEqual(items2[0].Quantity, 2.0) {
		t.Fatalf("expected 1 merged item with qty 2, got len=%d", len(items2))
	}

	// 3. Add same menu_item WITHOUT modifier -> creates a SECOND distinct line (20 net, VAT=3.00, total=23.00 -> cart total=78.20)
	addResp3 := uc.AddToCart(ctx, RestaurantAddToCartInput{
		CartID:     &cartOut1.ID,
		MenuItemID: fix.MenuItemID,
		Quantity:   1,
		Notes:      "No pickles",
	})
	if addResp3.StatusCode != utils.CodeOK {
		t.Fatalf("AddToCart distinct failed: %d %s", addResp3.StatusCode, addResp3.Message)
	}
	payload3 := addResp3.Data.(map[string]interface{})
	cartOut3 := payload3["cart"].(RestaurantCartOutput)
	items3 := payload3["items"].([]RestaurantCartItemOutput)
	if len(items3) != 2 {
		t.Fatalf("expected 2 distinct cart lines, got %d", len(items3))
	}
	if !almostEqual(cartOut3.Subtotal, 68.00) || !almostEqual(cartOut3.TaxAmount, 10.20) || !almostEqual(cartOut3.TotalAmount, 78.20) {
		t.Fatalf("unexpected cart totals after add 3: sub=%.2f tax=%.2f tot=%.2f", cartOut3.Subtotal, cartOut3.TaxAmount, cartOut3.TotalAmount)
	}

	// 4. Add combo item (30 + 5 = 35 net, VAT 5.25, total 40.25 -> cart total=118.45)
	addCombo := uc.AddToCart(ctx, RestaurantAddToCartInput{
		CartID:            &cartOut1.ID,
		MenuItemID:        fix.ComboItemID,
		Quantity:          1,
		ComboComponentIDs: []int32{fix.ComboCompID},
	})
	if addCombo.StatusCode != utils.CodeOK {
		t.Fatalf("AddToCart combo failed: %d %s", addCombo.StatusCode, addCombo.Message)
	}
	payload4 := addCombo.Data.(map[string]interface{})
	cartOut4 := payload4["cart"].(RestaurantCartOutput)
	if !almostEqual(cartOut4.TotalAmount, 118.45) {
		t.Fatalf("expected combo cart total 118.45, got %.2f", cartOut4.TotalAmount)
	}

	// 5. Checkout (ConvertToOrder)
	checkoutResp := uc.ConvertToOrder(ctx, RestaurantCheckoutInput{
		CartID:      cartOut1.ID,
		CustomerID:  &fix.CustomerID,
		AmountPaid:  120.00,
		OrderSource: "pos",
		OrderType:   "dine_in",
	})
	if checkoutResp.StatusCode != utils.CodeCreated {
		t.Fatalf("ConvertToOrder failed: %d %s", checkoutResp.StatusCode, checkoutResp.Message)
	}
	checkoutData := checkoutResp.Data.(map[string]interface{})
	order := checkoutData["order"].(repository.RestaurantOrder)
	orderItems := checkoutData["items"].([]repository.RestaurantOrderItem)
	convertedCart := checkoutData["cart"].(RestaurantCartOutput)

	if len(orderItems) != 3 {
		t.Fatalf("expected 3 restaurant_order_items, got %d", len(orderItems))
	}
	if order.PaymentStatus != "paid" || !almostEqual(numericToFloat(order.ChangeGiven), 1.55) {
		t.Fatalf("expected payment_status=paid & change=1.55, got %s %.2f", order.PaymentStatus, numericToFloat(order.ChangeGiven))
	}
	if convertedCart.CartStatus != "converted" || convertedCart.ConvertedToOrderID == nil || *convertedCart.ConvertedToOrderID != order.ID {
		t.Fatalf("cart not properly marked converted: status=%s order_id=%v", convertedCart.CartStatus, convertedCart.ConvertedToOrderID)
	}
	if !order.SourceCartID.Valid || uuid.UUID(order.SourceCartID.Bytes) != cartOut1.ID {
		t.Fatalf("order.source_cart_id mismatch")
	}

	// Double checkout must fail
	dupResp := uc.ConvertToOrder(ctx, RestaurantCheckoutInput{CartID: cartOut1.ID})
	if dupResp.StatusCode != utils.CodeBadReq {
		t.Fatalf("expected 400 on double checkout, got %d", dupResp.StatusCode)
	}

	// 6. Test ReopenOrderToCart for unpaid order
	directResp := uc.CreateOrderWithItems(ctx, RestaurantDirectOrderInput{
		StoreID:     fix.StoreID,
		TableID:     &fix.TableID,
		CustomerID:  &fix.CustomerID,
		OrderSource: "online",
		OrderType:   "delivery",
		Items: []RestaurantLineItemInput{
			{
				MenuItemID:  fix.MenuItemID,
				Quantity:    2,
				ModifierIDs: []int32{fix.ModifierID},
			},
		},
	})
	if directResp.StatusCode != utils.CodeCreated {
		t.Fatalf("CreateOrderWithItems failed: %d %s", directResp.StatusCode, directResp.Message)
	}
	directOrder := directResp.Data.(RestaurantOrderWithItems)
	if len(directOrder.Items) != 1 || !almostEqual(numericToFloat(directOrder.TotalAmount), 55.20) {
		t.Fatalf("expected direct order total 55.20 with 1 item, got total=%.2f items=%d", numericToFloat(directOrder.TotalAmount), len(directOrder.Items))
	}

	// Add second item directly to order
	addOrderItemResp := uc.AddOrderItem(ctx, directOrder.ID, RestaurantLineItemInput{
		MenuItemID: fix.MenuItemID,
		Quantity:   1,
	})
	if addOrderItemResp.StatusCode != utils.CodeCreated {
		t.Fatalf("AddOrderItem failed: %d %s", addOrderItemResp.StatusCode, addOrderItemResp.Message)
	}
	addOIData := addOrderItemResp.Data.(map[string]interface{})
	updatedOrd := addOIData["order"].(repository.RestaurantOrder)
	if !almostEqual(numericToFloat(updatedOrd.TotalAmount), 78.20) {
		t.Fatalf("expected updated direct order total 78.20, got %.2f", numericToFloat(updatedOrd.TotalAmount))
	}

	// Reopen unpaid order back to cart
	reopenResp := uc.ReopenOrderToCart(ctx, directOrder.ID)
	if reopenResp.StatusCode != utils.CodeOK {
		t.Fatalf("ReopenOrderToCart failed: %d %s", reopenResp.StatusCode, reopenResp.Message)
	}
	reopenData := reopenResp.Data.(map[string]interface{})
	reopenedCart := reopenData["cart"].(RestaurantCartOutput)
	reopenedItems := reopenData["items"].([]RestaurantCartItemOutput)
	if reopenedCart.CartStatus != "active" || len(reopenedItems) != 2 || !almostEqual(reopenedCart.TotalAmount, 78.20) {
		t.Fatalf("unexpected reopened cart state: status=%s items=%d total=%.2f", reopenedCart.CartStatus, len(reopenedItems), reopenedCart.TotalAmount)
	}
	voidedOrder, err := repo.GetRestaurantOrder(ctx, directOrder.ID)
	if err != nil || voidedOrder.Status != "voided" {
		t.Fatalf("expected reopened order to be voided, got status=%s err=%v", voidedOrder.Status, err)
	}

	// 7. Guest cart merge into customer cart
	guestAdd := uc.AddToCart(ctx, RestaurantAddToCartInput{
		StoreID:         fix.StoreID,
		GuestIdentifier: fmt.Sprintf("guest-%s", strings.ToLower(reopenedCart.ID.String()[:8])),
		MenuItemID:      fix.MenuItemID,
		Quantity:        1,
	})
	if guestAdd.StatusCode != utils.CodeOK {
		t.Fatalf("guest AddToCart failed: %d %s", guestAdd.StatusCode, guestAdd.Message)
	}
	guestID := *guestAdd.Data.(map[string]interface{})["cart"].(RestaurantCartOutput).GuestIdentifier
	mergeResp := uc.MergeGuestCart(ctx, fix.StoreID, guestID, fix.CustomerID)
	if mergeResp.StatusCode != utils.CodeOK {
		t.Fatalf("MergeGuestCart failed: %d %s", mergeResp.StatusCode, mergeResp.Message)
	}
	mergedCart := mergeResp.Data.(map[string]interface{})["cart"].(RestaurantCartOutput)
	if mergedCart.CustomerID == nil || *mergedCart.CustomerID != fix.CustomerID {
		t.Fatalf("expected merged cart customer_id=%d, got %v", fix.CustomerID, mergedCart.CustomerID)
	}

	// 8. POS Transaction + POS Payment + Cashier Session Balance on POS Checkout & SettleOrder
	var userID, cashierID, terminalID, sessionID int32
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (organization_id, username, email, password_hash)
		VALUES ($1, 'poscashier', 'poscashier@example.com', 'hash') RETURNING id
	`, fix.OrgID).Scan(&userID); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO cashiers (user_id, store_id, cashier_code) VALUES ($1, $2, 'CASH-01') RETURNING id
	`, userID, fix.StoreID).Scan(&cashierID); err != nil {
		t.Fatalf("seed cashier: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO pos_terminals (store_id, terminal_code, terminal_name) VALUES ($1, 'TERM-01', 'Counter 1') RETURNING id
	`, fix.StoreID).Scan(&terminalID); err != nil {
		t.Fatalf("seed pos_terminal: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO cashier_sessions (cashier_id, pos_terminal_id, session_number, opening_balance, expected_balance, status)
		VALUES ($1, $2, 'SESS-01', 100.00, 100.00, 'open') RETURNING id
	`, cashierID, terminalID).Scan(&sessionID); err != nil {
		t.Fatalf("seed cashier_session: %v", err)
	}

	posAdd := uc.AddToCart(ctx, RestaurantAddToCartInput{
		StoreID:    fix.StoreID,
		OrderType:  "takeout",
		Channel:    "pos",
		MenuItemID: fix.MenuItemID,
		Quantity:   1, // 20.00 + 15% VAT = 23.00
	})
	if posAdd.StatusCode != utils.CodeOK {
		t.Fatalf("pos AddToCart failed: %d %s", posAdd.StatusCode, posAdd.Message)
	}
	posCart := posAdd.Data.(map[string]interface{})["cart"].(RestaurantCartOutput)

	posCheckout := uc.ConvertToOrder(ctx, RestaurantCheckoutInput{
		CartID:        posCart.ID,
		OrderSource:   "pos",
		OrderType:     "takeout",
		PaymentMethod: "cash",
	})
	if posCheckout.StatusCode != utils.CodeCreated {
		t.Fatalf("pos ConvertToOrder failed: %d %s", posCheckout.StatusCode, posCheckout.Message)
	}
	posOut := posCheckout.Data.(map[string]interface{})
	txnIDVal, ok := posOut["pos_transaction_id"].(int32)
	if !ok || txnIDVal <= 0 {
		t.Fatalf("expected pos_transaction_id > 0 in ConvertToOrder output, got %v", posOut["pos_transaction_id"])
	}
	posOrd := posOut["order"].(repository.RestaurantOrder)
	if !posOrd.PosTransactionID.Valid || posOrd.PosTransactionID.Int32 != txnIDVal || posOrd.PaymentStatus != "paid" {
		t.Fatalf("unexpected linked pos order: pos_txn_id=%v payment_status=%s", posOrd.PosTransactionID, posOrd.PaymentStatus)
	}

	var txnType string
	var txnTotal float64
	if err := pool.QueryRow(ctx, `SELECT transaction_type, total_amount::float8 FROM pos_transactions WHERE id = $1`, txnIDVal).Scan(&txnType, &txnTotal); err != nil {
		t.Fatalf("query pos_transactions: %v", err)
	}
	if txnType != "restaurant_sale" || !almostEqual(txnTotal, 23.00) {
		t.Fatalf("expected restaurant_sale txn 23.00, got type=%s total=%.2f", txnType, txnTotal)
	}

	var payCount int
	var payAmount float64
	if err := pool.QueryRow(ctx, `SELECT COUNT(*), COALESCE(SUM(amount), 0)::float8 FROM pos_payments WHERE transaction_id = $1 AND payment_method = 'cash'`, txnIDVal).Scan(&payCount, &payAmount); err != nil {
		t.Fatalf("query pos_payments: %v", err)
	}
	if payCount != 1 || !almostEqual(payAmount, 23.00) {
		t.Fatalf("expected 1 cash payment of 23.00, got count=%d amount=%.2f", payCount, payAmount)
	}

	var expectedBal float64
	if err := pool.QueryRow(ctx, `SELECT expected_balance::float8 FROM cashier_sessions WHERE id = $1`, sessionID).Scan(&expectedBal); err != nil {
		t.Fatalf("query cashier_sessions expected_balance: %v", err)
	}
	if !almostEqual(expectedBal, 123.00) {
		t.Fatalf("expected cashier session balance 123.00, got %.2f", expectedBal)
	}

	// Now create an unpaid dine-in order and settle it via SettleOrderWithPosTransaction
	unpaidResp := uc.CreateOrderWithItems(ctx, RestaurantDirectOrderInput{
		StoreID:     fix.StoreID,
		TableID:     &fix.TableID,
		OrderSource: "table",
		OrderType:   "dine_in",
		Items: []RestaurantLineItemInput{
			{MenuItemID: fix.MenuItemID, Quantity: 2}, // 40.00 + 6.00 = 46.00
		},
	})
	if unpaidResp.StatusCode != utils.CodeCreated {
		t.Fatalf("CreateOrderWithItems unpaid failed: %d %s", unpaidResp.StatusCode, unpaidResp.Message)
	}
	unpaidOrd := unpaidResp.Data.(RestaurantOrderWithItems)

	settleResp := uc.SettleOrderWithPosTransaction(ctx, unpaidOrd.ID, RestaurantSettleOrderInput{
		PaymentMethod: "cash",
	})
	if settleResp.StatusCode != utils.CodeOK {
		t.Fatalf("SettleOrderWithPosTransaction failed: %d %s", settleResp.StatusCode, settleResp.Message)
	}
	settleOut := settleResp.Data.(map[string]interface{})
	settleTxnID := settleOut["pos_transaction_id"].(int32)
	settledOrder := settleOut["order"].(*repository.RestaurantOrder)
	if settledOrder.Status != "paid" || settledOrder.PaymentStatus != "paid" || !settledOrder.PosTransactionID.Valid || settledOrder.PosTransactionID.Int32 != settleTxnID {
		t.Fatalf("unexpected settled order: status=%s payment=%s txn=%v", settledOrder.Status, settledOrder.PaymentStatus, settledOrder.PosTransactionID)
	}
	if err := pool.QueryRow(ctx, `SELECT expected_balance::float8 FROM cashier_sessions WHERE id = $1`, sessionID).Scan(&expectedBal); err != nil {
		t.Fatalf("query cashier_sessions expected_balance after settle: %v", err)
	}
	if !almostEqual(expectedBal, 169.00) {
		t.Fatalf("expected cashier session balance 169.00 after settle, got %.2f", expectedBal)
	}
}

func TestRestaurantPromotionsRecipesHistoryAndKdsIntegration(t *testing.T) {
	pool := setupRestaurantMigrationDB(t)
	ctx := context.Background()
	fix := seedRestaurantTestData(t, ctx, pool)

	repo := repository.New(pool)
	cartUC := NewRestaurantCartUseCase()
	cartUC.SetRepository(repo)
	cartUC.SetPool(pool)

	restUC := NewRestaurantUseCase()
	restUC.SetRepository(repo)

	// === 1. Seed Promotions ===
	var promoID int32
	if err := pool.QueryRow(ctx, `
		INSERT INTO restaurant_promotions (
			store_id, code, name, promotion_type, applies_to,
			target_menu_item_ids, coupon_code, discount_value, is_active
		) VALUES ($1, 'PROMO-BURGER20', '20% Off Burger', 'percentage_discount', 'menu_item', $2, 'BURGER20', 20.00, true)
		RETURNING id
	`, fix.StoreID, []int32{fix.MenuItemID}).Scan(&promoID); err != nil {
		t.Fatalf("seed restaurant promotion: %v", err)
	}

	// === 2. Seed Ingredients, Inventory Stock & Recipes (burger + combo side) ===
	var uomID int32
	if err := pool.QueryRow(ctx, `INSERT INTO units_of_measure (code, name, uom_type) VALUES ('PCS', 'Piece', 'count') RETURNING id`).Scan(&uomID); err != nil {
		t.Fatalf("seed uom: %v", err)
	}
	var prodCatID int32
	if err := pool.QueryRow(ctx, `INSERT INTO product_categories (name, code) VALUES ('Ingredients', 'ING') RETURNING id`).Scan(&prodCatID); err != nil {
		t.Fatalf("seed product category: %v", err)
	}
	seedIngredient := func(sku, name string, stock float64) int32 {
		var prodID int32
		if err := pool.QueryRow(ctx, `
			INSERT INTO products (organization_id, sku, name, category_id, base_uom_id)
			VALUES ($1, $2, $3, $4, $5) RETURNING id
		`, fix.OrgID, sku, name, prodCatID, uomID).Scan(&prodID); err != nil {
			t.Fatalf("seed product %s: %v", sku, err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO inventory_stock (product_id, store_id, quantity_on_hand, quantity_available)
			VALUES ($1, $2, $3, $3)
		`, prodID, fix.StoreID, stock); err != nil {
			t.Fatalf("seed inventory_stock %s: %v", sku, err)
		}
		return prodID
	}

	bunProdID := seedIngredient("ING-BUN", "Burger Bun", 50.0)
	pattyProdID := seedIngredient("ING-PATTY", "Beef Patty", 50.0)
	friesProdID := seedIngredient("ING-FRIES", "Frozen Fries Portion", 50.0)

	var burgerRecipeID, friesRecipeID int32
	if err := pool.QueryRow(ctx, `
		INSERT INTO recipes (organization_id, recipe_code, recipe_name, yield_quantity, yield_uom_id, is_active)
		VALUES ($1, 'REC-BURGER', 'Signature Burger Recipe', 1, $2, true) RETURNING id
	`, fix.OrgID, uomID).Scan(&burgerRecipeID); err != nil {
		t.Fatalf("seed burger recipe: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO recipe_ingredients (recipe_id, product_id, quantity, uom_id, line_number)
		VALUES ($1, $2, 1, $4, 1), ($1, $3, 2, $4, 2)
	`, burgerRecipeID, bunProdID, pattyProdID, uomID); err != nil {
		t.Fatalf("seed burger ingredients: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE menu_items SET recipe_id = $1 WHERE id = $2`, burgerRecipeID, fix.MenuItemID); err != nil {
		t.Fatalf("link burger recipe: %v", err)
	}

	var sideMenuItemID int32
	if err := pool.QueryRow(ctx, `SELECT component_menu_item_id FROM menu_item_combo_components WHERE id = $1`, fix.ComboCompID).Scan(&sideMenuItemID); err != nil {
		t.Fatalf("lookup side menu item: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO recipes (organization_id, recipe_code, recipe_name, yield_quantity, yield_uom_id, is_active)
		VALUES ($1, 'REC-FRIES', 'Fries Recipe', 1, $2, true) RETURNING id
	`, fix.OrgID, uomID).Scan(&friesRecipeID); err != nil {
		t.Fatalf("seed fries recipe: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO recipe_ingredients (recipe_id, product_id, quantity, uom_id, line_number)
		VALUES ($1, $2, 3, $3, 1)
	`, friesRecipeID, friesProdID, uomID); err != nil {
		t.Fatalf("seed fries ingredient: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE menu_items SET recipe_id = $1 WHERE id = $2`, friesRecipeID, sideMenuItemID); err != nil {
		t.Fatalf("link fries recipe: %v", err)
	}

	// === 3. Apply & Remove Coupon on Cart ===
	addResp := cartUC.AddToCart(ctx, RestaurantAddToCartInput{
		StoreID:    fix.StoreID,
		TableID:    &fix.TableID,
		CustomerID: &fix.CustomerID,
		OrderType:  "dine_in",
		MenuItemID: fix.MenuItemID,
		Quantity:   1, // 20.00 + 15% VAT = 23.00
	})
	if addResp.StatusCode != utils.CodeOK {
		t.Fatalf("AddToCart failed: %d %s", addResp.StatusCode, addResp.Message)
	}
	cart := addResp.Data.(map[string]interface{})["cart"].(RestaurantCartOutput)

	applyResp := cartUC.ApplyCouponToCart(ctx, cart.ID, "burger20") // case-insensitive
	if applyResp.StatusCode != utils.CodeOK {
		t.Fatalf("ApplyCouponToCart failed: %d %s", applyResp.StatusCode, applyResp.Message)
	}
	appliedCart := applyResp.Data.(map[string]interface{})["cart"].(RestaurantCartOutput)
	if appliedCart.CouponCode == nil || *appliedCart.CouponCode != "burger20" || !almostEqual(appliedCart.DiscountAmount, 4.00) || !almostEqual(appliedCart.TaxAmount, 2.40) || !almostEqual(appliedCart.TotalAmount, 18.40) {
		t.Fatalf("unexpected coupon cart totals: coupon=%v disc=%.2f tax=%.2f total=%.2f",
			appliedCart.CouponCode, appliedCart.DiscountAmount, appliedCart.TaxAmount, appliedCart.TotalAmount)
	}

	removeResp := cartUC.RemoveCouponFromCart(ctx, cart.ID)
	if removeResp.StatusCode != utils.CodeOK {
		t.Fatalf("RemoveCouponFromCart failed: %d %s", removeResp.StatusCode, removeResp.Message)
	}
	removedCart := removeResp.Data.(map[string]interface{})["cart"].(RestaurantCartOutput)
	if (removedCart.CouponCode != nil && *removedCart.CouponCode != "") || !almostEqual(removedCart.DiscountAmount, 0.00) || !almostEqual(removedCart.TotalAmount, 23.00) {
		t.Fatalf("expected coupon cleared: coupon=%v disc=%.2f total=%.2f",
			removedCart.CouponCode, removedCart.DiscountAmount, removedCart.TotalAmount)
	}

	// Add 1 more burger (total 2 burgers) and 1 combo meal, then checkout WITH coupon
	if resp := cartUC.AddToCart(ctx, RestaurantAddToCartInput{
		CartID:     &cart.ID,
		StoreID:    fix.StoreID,
		MenuItemID: fix.MenuItemID,
		Quantity:   1,
	}); resp.StatusCode != utils.CodeOK {
		t.Fatalf("AddToCart second burger failed: %d %s", resp.StatusCode, resp.Message)
	}
	if resp := cartUC.AddToCart(ctx, RestaurantAddToCartInput{
		CartID:            &cart.ID,
		StoreID:           fix.StoreID,
		MenuItemID:        fix.ComboItemID,
		Quantity:          1,
		ComboComponentIDs: []int32{fix.ComboCompID},
	}); resp.StatusCode != utils.CodeOK {
		t.Fatalf("AddToCart combo failed: %d %s", resp.StatusCode, resp.Message)
	}

	checkoutResp := cartUC.ConvertToOrder(ctx, RestaurantCheckoutInput{
		CartID:      cart.ID,
		OrderSource: "table",
		OrderType:   "dine_in",
		CouponCode:  "BURGER20",
	})
	if checkoutResp.StatusCode != utils.CodeCreated {
		t.Fatalf("ConvertToOrder with coupon failed: %d %s", checkoutResp.StatusCode, checkoutResp.Message)
	}
	ord := checkoutResp.Data.(map[string]interface{})["order"].(repository.RestaurantOrder)
	// BURGER20 applies 20% ONLY to MenuItemID (2 * 20.00 = 40.00 -> 8.00 discount), not ComboItemID!
	if !ord.CouponCode.Valid || ord.CouponCode.String != "BURGER20" || !almostEqual(numericToFloat(ord.DiscountAmount), 8.00) {
		t.Fatalf("expected order coupon BURGER20 and discount 8.00, got coupon=%v disc=%.2f", ord.CouponCode, numericToFloat(ord.DiscountAmount))
	}

	var usageCount int32
	if err := pool.QueryRow(ctx, `SELECT usage_count FROM restaurant_promotions WHERE id = $1`, promoID).Scan(&usageCount); err != nil {
		t.Fatalf("query promo usage_count: %v", err)
	}
	if usageCount != 1 {
		t.Fatalf("expected promotion usage_count 1, got %d", usageCount)
	}

	// === 4. Verify Recipe Ingredient Stock Deduction (2 burgers = 2 buns, 4 patties; 1 combo side = 3 fries) ===
	stockOf := func(prodID int32) float64 {
		var qty float64
		if err := pool.QueryRow(ctx, `SELECT quantity_on_hand::float8 FROM inventory_stock WHERE store_id = $1 AND product_id = $2`, fix.StoreID, prodID).Scan(&qty); err != nil {
			t.Fatalf("query stock %d: %v", prodID, err)
		}
		return qty
	}
	if !almostEqual(stockOf(bunProdID), 48.0) || !almostEqual(stockOf(pattyProdID), 46.0) || !almostEqual(stockOf(friesProdID), 47.0) {
		t.Fatalf("unexpected deducted stock: buns=%.2f patties=%.2f fries=%.2f", stockOf(bunProdID), stockOf(pattyProdID), stockOf(friesProdID))
	}

	var deductMovements int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM stock_movements WHERE reference_type = 'restaurant_order' AND reference_id = $1 AND movement_type = 'recipe_deduction'`, ord.ID).Scan(&deductMovements); err != nil {
		t.Fatal(err)
	}
	if deductMovements != 3 {
		t.Fatalf("expected 3 recipe_deduction stock_movements, got %d", deductMovements)
	}

	// === 5. KDS View & GetKdsOrders (verify previously broken view + endpoint data) ===
	viewRows, err := repo.ListActiveRestaurantOrdersView(ctx, fix.StoreID)
	if err != nil {
		t.Fatalf("ListActiveRestaurantOrdersView failed (minutes_since_ordered scan bug): %v", err)
	}
	if len(viewRows) != 1 || viewRows[0].OrderID != ord.ID {
		t.Fatalf("expected 1 order in vw_active_restaurant_orders, got %+v", viewRows)
	}
	kdsResp := restUC.GetKdsOrders(ctx, fix.StoreID, []string{"pending", "preparing"})
	if kdsResp.StatusCode != utils.CodeOK {
		t.Fatalf("GetKdsOrders failed: %d %s", kdsResp.StatusCode, kdsResp.Message)
	}
	kdsOrders := kdsResp.Data.([]KdsOrderOutput)
	if len(kdsOrders) != 1 || kdsOrders[0].ID != ord.ID || len(kdsOrders[0].Items) != 2 {
		t.Fatalf("unexpected KDS orders output: %+v", kdsOrders)
	}

	// === 6. Order Status History + Recipe Reversal on Void ===
	if resp := cartUC.UpdateOrderStatusWithHistory(ctx, ord.ID, RestaurantOrderStatusUpdateInput{
		Status: "preparing",
		Reason: "kitchen_started",
		Notes:  "Grill started",
	}); resp.StatusCode != utils.CodeOK {
		t.Fatalf("UpdateOrderStatusWithHistory(preparing) failed: %d %s", resp.StatusCode, resp.Message)
	}
	if resp := cartUC.UpdateOrderStatusWithHistory(ctx, ord.ID, RestaurantOrderStatusUpdateInput{
		Status: "ready",
		Reason: "plated",
	}); resp.StatusCode != utils.CodeOK {
		t.Fatalf("UpdateOrderStatusWithHistory(ready) failed: %d %s", resp.StatusCode, resp.Message)
	}
	// Voiding reverses recipe stock idempotently
	if resp := cartUC.UpdateOrderStatusWithHistory(ctx, ord.ID, RestaurantOrderStatusUpdateInput{
		Status: "voided",
		Reason: "customer_cancelled",
	}); resp.StatusCode != utils.CodeOK {
		t.Fatalf("UpdateOrderStatusWithHistory(voided) failed: %d %s", resp.StatusCode, resp.Message)
	}
	if !almostEqual(stockOf(bunProdID), 50.0) || !almostEqual(stockOf(pattyProdID), 50.0) || !almostEqual(stockOf(friesProdID), 50.0) {
		t.Fatalf("expected restored stock 50/50/50 after void, got buns=%.2f patties=%.2f fries=%.2f", stockOf(bunProdID), stockOf(pattyProdID), stockOf(friesProdID))
	}

	histResp := cartUC.GetOrderStatusHistory(ctx, ord.ID)
	if histResp.StatusCode != utils.CodeOK {
		t.Fatalf("GetOrderStatusHistory failed: %d %s", histResp.StatusCode, histResp.Message)
	}
	histRows := histResp.Data.([]repository.ListRestaurantOrderStatusHistoryRow)
	if len(histRows) != 4 {
		t.Fatalf("expected 4 status history entries, got %d", len(histRows))
	}
	expectedTransitions := []string{"pending", "preparing", "ready", "voided"}
	for i, exp := range expectedTransitions {
		if histRows[i].ToStatus != exp {
			t.Fatalf("history[%d]: expected to_status=%s got %s", i, exp, histRows[i].ToStatus)
		}
	}

	// Voided order should now be excluded from KDS
	kdsAfterVoid := restUC.GetKdsOrders(ctx, fix.StoreID, []string{"pending", "preparing", "ready"})
	if len(kdsAfterVoid.Data.([]KdsOrderOutput)) != 0 {
		t.Fatalf("expected 0 KDS orders after void")
	}
}
