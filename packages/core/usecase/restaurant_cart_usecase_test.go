package usecase

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

	lockConn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire test db lock conn: %v", err)
	}
	if _, err := lockConn.Exec(ctx, `SELECT pg_advisory_lock(918273645)`); err != nil {
		lockConn.Release()
		t.Fatalf("acquire test db advisory lock: %v", err)
	}
	t.Cleanup(func() {
		_, _ = lockConn.Exec(context.Background(), `SELECT pg_advisory_unlock(918273645)`)
		lockConn.Release()
	})

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

func TestRestaurantUnderpaymentRejectedIntegration(t *testing.T) {
	pool := setupRestaurantMigrationDB(t)
	ctx := context.Background()
	fix := seedRestaurantTestData(t, ctx, pool)

	var userID, cashierID, terminalID, sessionID int32
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (organization_id, username, email, password_hash)
		VALUES ($1, 'underpaycashier', 'underpaycashier@example.com', 'hash') RETURNING id
	`, fix.OrgID).Scan(&userID); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO cashiers (user_id, store_id, cashier_code) VALUES ($1, $2, 'CASH-UP') RETURNING id
	`, userID, fix.StoreID).Scan(&cashierID); err != nil {
		t.Fatalf("seed cashier: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO pos_terminals (store_id, terminal_code, terminal_name) VALUES ($1, 'TERM-UP', 'Counter UP') RETURNING id
	`, fix.StoreID).Scan(&terminalID); err != nil {
		t.Fatalf("seed pos_terminal: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO cashier_sessions (cashier_id, pos_terminal_id, session_number, opening_balance, expected_balance, status)
		VALUES ($1, $2, 'SESS-UP', 100.00, 100.00, 'open') RETURNING id
	`, cashierID, terminalID).Scan(&sessionID); err != nil {
		t.Fatalf("seed cashier_session: %v", err)
	}

	repo := repository.New(pool)
	uc := NewRestaurantCartUseCase()
	uc.SetRepository(repo)
	uc.SetPool(pool)

	// 1. Paying 5 against total 23 via POS checkout must fail with 400 and create no order/transaction.
	addPos := uc.AddToCart(ctx, RestaurantAddToCartInput{
		StoreID:    fix.StoreID,
		OrderType:  "takeout",
		Channel:    "pos",
		MenuItemID: fix.MenuItemID,
		Quantity:   1, // 20.00 + 15% VAT = 23.00
	})
	if addPos.StatusCode != utils.CodeOK {
		t.Fatalf("AddToCart failed: %d %s", addPos.StatusCode, addPos.Message)
	}
	posCart := addPos.Data.(map[string]interface{})["cart"].(RestaurantCartOutput)

	underpayCheckout := uc.ConvertToOrder(ctx, RestaurantCheckoutInput{
		CartID:        posCart.ID,
		OrderSource:   "pos",
		OrderType:     "takeout",
		PaymentMethod: "cash",
		AmountPaid:    5.00,
	})
	if underpayCheckout.StatusCode != utils.CodeBadReq {
		t.Fatalf("expected 400 on POS checkout underpayment (5 vs 23), got %d (%s)", underpayCheckout.StatusCode, underpayCheckout.Message)
	}

	var txnCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM pos_transactions`).Scan(&txnCount); err != nil {
		t.Fatalf("count pos_transactions: %v", err)
	}
	if txnCount != 0 {
		t.Fatalf("expected 0 pos_transactions after rejected checkout, got %d", txnCount)
	}

	var expectedBal float64
	if err := pool.QueryRow(ctx, `SELECT expected_balance::float8 FROM cashier_sessions WHERE id = $1`, sessionID).Scan(&expectedBal); err != nil {
		t.Fatalf("query cashier_sessions expected_balance: %v", err)
	}
	if !almostEqual(expectedBal, 100.00) {
		t.Fatalf("expected cashier session expected_balance 100.00 unchanged, got %.2f", expectedBal)
	}

	cartStillActive := uc.GetCart(ctx, posCart.ID)
	if cartStillActive.StatusCode != utils.CodeOK || cartStillActive.Data.(map[string]interface{})["cart"].(RestaurantCartOutput).CartStatus != "active" {
		t.Fatalf("expected cart to remain active after failed underpayment checkout")
	}

	// Paying full 23.00 on the same cart succeeds.
	fullCheckout := uc.ConvertToOrder(ctx, RestaurantCheckoutInput{
		CartID:        posCart.ID,
		OrderSource:   "pos",
		OrderType:     "takeout",
		PaymentMethod: "cash",
		AmountPaid:    23.00,
	})
	if fullCheckout.StatusCode != utils.CodeCreated {
		t.Fatalf("expected 201 on full POS checkout, got %d (%s)", fullCheckout.StatusCode, fullCheckout.Message)
	}

	// 2. Non-POS checkout with partial payment (5 against 23) remains partially_paid without POS txn,
	//    and settling without full payment must be rejected while full settlement succeeds.
	addOnline := uc.AddToCart(ctx, RestaurantAddToCartInput{
		StoreID:    fix.StoreID,
		OrderType:  "delivery",
		Channel:    "online",
		MenuItemID: fix.MenuItemID,
		Quantity:   1, // 23.00
	})
	if addOnline.StatusCode != utils.CodeOK {
		t.Fatalf("AddToCart online failed: %d %s", addOnline.StatusCode, addOnline.Message)
	}
	onlineCart := addOnline.Data.(map[string]interface{})["cart"].(RestaurantCartOutput)

	// Explicitly claiming payment_status=paid with underpayment must fail even on non-POS checkout.
	fakePaidResp := uc.ConvertToOrder(ctx, RestaurantCheckoutInput{
		CartID:        onlineCart.ID,
		OrderSource:   "online",
		OrderType:     "delivery",
		PaymentStatus: "paid",
		AmountPaid:    5.00,
	})
	if fakePaidResp.StatusCode != utils.CodeBadReq {
		t.Fatalf("expected 400 when payment_status=paid with underpayment 5 vs 23, got %d (%s)", fakePaidResp.StatusCode, fakePaidResp.Message)
	}

	// Normal partial payment (5 of 23) succeeds as partially_paid with no POS transaction.
	partialResp := uc.ConvertToOrder(ctx, RestaurantCheckoutInput{
		CartID:      onlineCart.ID,
		OrderSource: "online",
		OrderType:   "delivery",
		AmountPaid:  5.00,
	})
	if partialResp.StatusCode != utils.CodeCreated {
		t.Fatalf("expected 201 for partially_paid online checkout, got %d (%s)", partialResp.StatusCode, partialResp.Message)
	}
	partialOrd := partialResp.Data.(map[string]interface{})["order"].(repository.RestaurantOrder)
	if partialOrd.PaymentStatus != "partially_paid" || partialOrd.PosTransactionID.Valid {
		t.Fatalf("expected partially_paid order with nil pos_transaction_id, got status=%s txn=%v", partialOrd.PaymentStatus, partialOrd.PosTransactionID)
	}

	// Settling that partially paid order without supplying full amount_paid must fail (5 < 23).
	badSettleDefault := uc.SettleOrderWithPosTransaction(ctx, partialOrd.ID, RestaurantSettleOrderInput{
		PaymentMethod: "cash",
	})
	if badSettleDefault.StatusCode != utils.CodeBadReq {
		t.Fatalf("expected 400 settling partially_paid order without full amount_paid, got %d (%s)", badSettleDefault.StatusCode, badSettleDefault.Message)
	}

	// Settling with explicit underpayment (5) or negative payment (-5) must fail.
	underAmt := 5.00
	badSettleExplicit := uc.SettleOrderWithPosTransaction(ctx, partialOrd.ID, RestaurantSettleOrderInput{
		PaymentMethod: "cash",
		AmountPaid:    &underAmt,
	})
	if badSettleExplicit.StatusCode != utils.CodeBadReq {
		t.Fatalf("expected 400 on explicit underpayment settle, got %d (%s)", badSettleExplicit.StatusCode, badSettleExplicit.Message)
	}

	negAmt := -5.00
	badSettleNeg := uc.SettleOrderWithPosTransaction(ctx, partialOrd.ID, RestaurantSettleOrderInput{
		PaymentMethod: "cash",
		AmountPaid:    &negAmt,
	})
	if badSettleNeg.StatusCode != utils.CodeBadReq {
		t.Fatalf("expected 400 on negative amount_paid settle, got %d (%s)", badSettleNeg.StatusCode, badSettleNeg.Message)
	}

	// Order remains partially_paid until settled with full >= 23.00.
	reloadedOrd, err := repo.GetRestaurantOrder(ctx, partialOrd.ID)
	if err != nil || reloadedOrd.PaymentStatus != "partially_paid" || reloadedOrd.PosTransactionID.Valid {
		t.Fatalf("expected order to remain partially_paid without pos_transaction_id, got payment_status=%s txn=%v err=%v", reloadedOrd.PaymentStatus, reloadedOrd.PosTransactionID, err)
	}

	// Settling with 25.00 against 23.00 total succeeds cleanly with 2.00 change.
	fullAmt := 25.00
	goodSettle := uc.SettleOrderWithPosTransaction(ctx, partialOrd.ID, RestaurantSettleOrderInput{
		PaymentMethod: "cash",
		AmountPaid:    &fullAmt,
	})
	if goodSettle.StatusCode != utils.CodeOK {
		t.Fatalf("expected 200 on full settle (25 vs 23), got %d (%s)", goodSettle.StatusCode, goodSettle.Message)
	}
	settledOrd := goodSettle.Data.(map[string]interface{})["order"].(*repository.RestaurantOrder)
	if settledOrd.PaymentStatus != "paid" || !almostEqual(numericToFloat(settledOrd.ChangeGiven), 2.00) {
		t.Fatalf("expected settled paid order with 2.00 change, got status=%s change=%.2f", settledOrd.PaymentStatus, numericToFloat(settledOrd.ChangeGiven))
	}
}

func TestRestaurantCartConcurrencyLocksIntegration(t *testing.T) {
	pool := setupRestaurantMigrationDB(t)
	ctx := context.Background()
	fix := seedRestaurantTestData(t, ctx, pool)

	repo := repository.New(pool)
	uc := NewRestaurantCartUseCase()
	uc.SetRepository(repo)
	uc.SetPool(pool)

	// 1. Seed a cart with qty=1 of MenuItemID (20.00 + 15% VAT = 23.00)
	initResp := uc.AddToCart(ctx, RestaurantAddToCartInput{
		StoreID:    fix.StoreID,
		OrderType:  "takeaway",
		Channel:    "online",
		MenuItemID: fix.MenuItemID,
		Quantity:   1,
	})
	if initResp.StatusCode != utils.CodeOK {
		t.Fatalf("initial AddToCart failed: %d %s", initResp.StatusCode, initResp.Message)
	}
	cart := initResp.Data.(map[string]interface{})["cart"].(RestaurantCartOutput)

	// Fire 8 simultaneous AddToCart requests adding +1 to the exact same cart line.
	const workers = 8
	startCh := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startCh
			resp := uc.AddToCart(ctx, RestaurantAddToCartInput{
				CartID:     &cart.ID,
				MenuItemID: fix.MenuItemID,
				Quantity:   1,
			})
			if resp.StatusCode != utils.CodeOK {
				errs <- fmt.Errorf("AddToCart returned %d: %s", resp.StatusCode, resp.Message)
			}
		}()
	}
	close(startCh)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}

	items, err := repo.ListRestaurantCartItems(ctx, cart.ID)
	if err != nil {
		t.Fatalf("ListRestaurantCartItems failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 merged cart line, got %d", len(items))
	}
	if !almostEqual(numericToFloat(items[0].Quantity), 9.0) {
		t.Fatalf("expected serialized merged quantity 9.0, got %.3f (lost update!)", numericToFloat(items[0].Quantity))
	}
	lockedCart, err := repo.GetRestaurantCartByID(ctx, cart.ID)
	if err != nil {
		t.Fatalf("GetRestaurantCartByID failed: %v", err)
	}
	if !almostEqual(numericToFloat(lockedCart.TotalAmount), 9.0*23.00) {
		t.Fatalf("expected cart total %.2f, got %.2f", 9.0*23.00, numericToFloat(lockedCart.TotalAmount))
	}

	// 2. Concurrent checkout + AddToCart / double checkout must serialize cleanly:
	//    only 1 order is created, and order total_amount always matches sum of its order lines.
	for round := 0; round < 4; round++ {
		roundCartResp := uc.AddToCart(ctx, RestaurantAddToCartInput{
			StoreID:    fix.StoreID,
			OrderType:  "takeaway",
			Channel:    "online",
			MenuItemID: fix.MenuItemID,
			Quantity:   1,
		})
		if roundCartResp.StatusCode != utils.CodeOK {
			t.Fatalf("round %d AddToCart failed: %d %s", round, roundCartResp.StatusCode, roundCartResp.Message)
		}
		roundCart := roundCartResp.Data.(map[string]interface{})["cart"].(RestaurantCartOutput)

		barrier := make(chan struct{})
		var rwg sync.WaitGroup
		checkoutCodes := make(chan int, 2)
		addResps := make(chan *repository.Response, 1)

		for i := 0; i < 2; i++ {
			rwg.Add(1)
			go func() {
				defer rwg.Done()
				<-barrier
				r := uc.ConvertToOrder(ctx, RestaurantCheckoutInput{
					CartID:      roundCart.ID,
					OrderSource: "online",
					OrderType:   "takeaway",
				})
				checkoutCodes <- r.StatusCode
			}()
		}
		rwg.Add(1)
		go func() {
			defer rwg.Done()
			<-barrier
			addResps <- uc.AddToCart(ctx, RestaurantAddToCartInput{
				CartID:     &roundCart.ID,
				MenuItemID: fix.MenuItemID,
				Quantity:   1,
			})
		}()

		close(barrier)
		rwg.Wait()
		close(checkoutCodes)
		close(addResps)

		createdCount := 0
		badReqCount := 0
		for code := range checkoutCodes {
			switch code {
			case utils.CodeCreated:
				createdCount++
			case utils.CodeBadReq:
				badReqCount++
			default:
				t.Fatalf("round %d unexpected checkout status code %d", round, code)
			}
		}
		if createdCount != 1 || badReqCount != 1 {
			t.Fatalf("round %d expected 1 created and 1 bad_req checkout, got created=%d bad_req=%d", round, createdCount, badReqCount)
		}

		addResp := <-addResps
		if addResp.StatusCode != utils.CodeOK && addResp.StatusCode != utils.CodeBadReq {
			t.Fatalf("round %d unexpected concurrent AddToCart status %d: %s", round, addResp.StatusCode, addResp.Message)
		}

		var orderID int32
		var orderTotal, itemsSum float64
		if err := pool.QueryRow(ctx, `
			SELECT o.id, o.total_amount::float8, COALESCE(SUM(oi.line_total), 0)::float8
			FROM restaurant_orders o
			JOIN restaurant_order_items oi ON oi.order_id = o.id
			WHERE o.source_cart_id = $1
			GROUP BY o.id, o.total_amount
		`, roundCart.ID).Scan(&orderID, &orderTotal, &itemsSum); err != nil {
			t.Fatalf("round %d query created order & lines: %v", round, err)
		}
		if !almostEqual(orderTotal, itemsSum) {
			t.Fatalf("round %d order %d total %.2f != sum(order_items.line_total) %.2f", round, orderID, orderTotal, itemsSum)
		}
		if addResp.StatusCode == utils.CodeOK && !almostEqual(orderTotal, 46.00) {
			t.Fatalf("round %d concurrent AddToCart committed first, expected order total 46.00, got %.2f", round, orderTotal)
		}
		if addResp.StatusCode == utils.CodeBadReq && !almostEqual(orderTotal, 23.00) {
			t.Fatalf("round %d concurrent AddToCart rejected after checkout, expected order total 23.00, got %.2f", round, orderTotal)
		}
	}
}

func TestRestaurantConcurrentVoidOrderRestoresStockOnceIntegration(t *testing.T) {
	pool := setupRestaurantMigrationDB(t)
	ctx := context.Background()
	fix := seedRestaurantTestData(t, ctx, pool)

	var uomID, prodCatID, bunProdID, pattyProdID, burgerRecipeID int32
	if err := pool.QueryRow(ctx, `INSERT INTO units_of_measure (code, name, uom_type) VALUES ('PCS', 'Piece', 'count') RETURNING id`).Scan(&uomID); err != nil {
		t.Fatalf("seed uom: %v", err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO product_categories (name, code) VALUES ('Ingredients', 'ING-VOID') RETURNING id`).Scan(&prodCatID); err != nil {
		t.Fatalf("seed product category: %v", err)
	}
	seedIng := func(sku, name string, stock float64) int32 {
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
	bunProdID = seedIng("VOID-BUN", "Burger Bun", 50.0)
	pattyProdID = seedIng("VOID-PATTY", "Beef Patty", 50.0)

	if err := pool.QueryRow(ctx, `
		INSERT INTO recipes (organization_id, recipe_code, recipe_name, yield_quantity, yield_uom_id, is_active)
		VALUES ($1, 'REC-VOID', 'Void Race Burger', 1, $2, true) RETURNING id
	`, fix.OrgID, uomID).Scan(&burgerRecipeID); err != nil {
		t.Fatalf("seed recipe: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO recipe_ingredients (recipe_id, product_id, quantity, uom_id, line_number)
		VALUES ($1, $2, 1, $4, 1), ($1, $3, 2, $4, 2)
	`, burgerRecipeID, bunProdID, pattyProdID, uomID); err != nil {
		t.Fatalf("seed recipe ingredients: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE menu_items SET recipe_id = $1 WHERE id = $2`, burgerRecipeID, fix.MenuItemID); err != nil {
		t.Fatalf("link recipe: %v", err)
	}

	repo := repository.New(pool)
	uc := NewRestaurantCartUseCase()
	uc.SetRepository(repo)
	uc.SetPool(pool)

	ordResp := uc.CreateOrderWithItems(ctx, RestaurantDirectOrderInput{
		StoreID:     fix.StoreID,
		OrderSource: "online",
		OrderType:   "takeaway",
		Items: []RestaurantLineItemInput{
			{MenuItemID: fix.MenuItemID, Quantity: 2}, // consumes 2 buns and 4 patties -> 48 & 46
		},
	})
	if ordResp.StatusCode != utils.CodeCreated {
		t.Fatalf("CreateOrderWithItems failed: %d %s", ordResp.StatusCode, ordResp.Message)
	}
	ord := ordResp.Data.(RestaurantOrderWithItems)

	stockOf := func(prodID int32) float64 {
		var s float64
		if err := pool.QueryRow(ctx, `SELECT quantity_on_hand::float8 FROM inventory_stock WHERE store_id = $1 AND product_id = $2`, fix.StoreID, prodID).Scan(&s); err != nil {
			t.Fatalf("query stock %d: %v", prodID, err)
		}
		return s
	}
	if !almostEqual(stockOf(bunProdID), 48.0) || !almostEqual(stockOf(pattyProdID), 46.0) {
		t.Fatalf("expected stock 48 buns / 46 patties after order, got %.2f / %.2f", stockOf(bunProdID), stockOf(pattyProdID))
	}

	// Fire 6 simultaneous void status updates against the exact same order.
	const voidWorkers = 6
	barrier := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, voidWorkers)
	for i := 0; i < voidWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-barrier
			resp := uc.UpdateOrderStatusWithHistory(ctx, ord.ID, RestaurantOrderStatusUpdateInput{
				Status: "voided",
				Reason: "concurrent_void",
			})
			if resp.StatusCode != utils.CodeOK {
				errs <- fmt.Errorf("UpdateOrderStatusWithHistory returned %d: %s", resp.StatusCode, resp.Message)
			}
		}()
	}
	close(barrier)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}

	if !almostEqual(stockOf(bunProdID), 50.0) || !almostEqual(stockOf(pattyProdID), 50.0) {
		t.Fatalf("expected stock restored ONCE to 50.0 / 50.0, got buns=%.2f patties=%.2f", stockOf(bunProdID), stockOf(pattyProdID))
	}

	var reversalCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM stock_movements
		WHERE reference_id = $1 AND movement_type = 'recipe_reversal'
	`, ord.ID).Scan(&reversalCount); err != nil {
		t.Fatalf("count recipe_reversal movements: %v", err)
	}
	if reversalCount != 2 { // 1 for buns, 1 for patties
		t.Fatalf("expected 2 recipe_reversal movements (1 per ingredient), got %d", reversalCount)
	}
}

func TestRestaurantReopenOrderWithSourceCartPreservesPostCheckoutChangesIntegration(t *testing.T) {
	pool := setupRestaurantMigrationDB(t)
	ctx := context.Background()
	fix := seedRestaurantTestData(t, ctx, pool)

	repo := repository.New(pool)
	uc := NewRestaurantCartUseCase()
	uc.SetRepository(repo)
	uc.SetPool(pool)

	// 1. Create cart with 1 burger (23.00) and check out (creates order with SourceCartID set)
	addResp := uc.AddToCart(ctx, RestaurantAddToCartInput{
		StoreID:    fix.StoreID,
		TableID:    &fix.TableID,
		OrderType:  "dine_in",
		Channel:    "table",
		MenuItemID: fix.MenuItemID,
		Quantity:   1,
	})
	if addResp.StatusCode != utils.CodeOK {
		t.Fatalf("AddToCart failed: %d %s", addResp.StatusCode, addResp.Message)
	}
	cart := addResp.Data.(map[string]interface{})["cart"].(RestaurantCartOutput)

	chkResp := uc.ConvertToOrder(ctx, RestaurantCheckoutInput{
		CartID:      cart.ID,
		OrderSource: "table",
		OrderType:   "dine_in",
	})
	if chkResp.StatusCode != utils.CodeCreated {
		t.Fatalf("ConvertToOrder failed: %d %s", chkResp.StatusCode, chkResp.Message)
	}
	ord := chkResp.Data.(map[string]interface{})["order"].(repository.RestaurantOrder)
	if !ord.SourceCartID.Valid || uuid.UUID(ord.SourceCartID.Bytes) != cart.ID {
		t.Fatalf("expected SourceCartID=%s on order", cart.ID)
	}

	// 2. Waiter adds a second item directly to the order after checkout (total becomes 46.00)
	addItemResp := uc.AddOrderItem(ctx, ord.ID, RestaurantLineItemInput{
		MenuItemID: fix.MenuItemID,
		Quantity:   1,
	})
	if addItemResp.StatusCode != utils.CodeCreated {
		t.Fatalf("AddOrderItem failed: %d %s", addItemResp.StatusCode, addItemResp.Message)
	}

	// 3. Reopen the order back into its source cart: must rebuild items from current order lines (2 items, 46.00)
	reopenResp := uc.ReopenOrderToCart(ctx, ord.ID)
	if reopenResp.StatusCode != utils.CodeOK {
		t.Fatalf("ReopenOrderToCart failed: %d %s", reopenResp.StatusCode, reopenResp.Message)
	}
	reopenData := reopenResp.Data.(map[string]interface{})
	reopenedCart := reopenData["cart"].(RestaurantCartOutput)
	reopenedItems := reopenData["items"].([]RestaurantCartItemOutput)

	if reopenedCart.ID != cart.ID {
		t.Fatalf("expected reopened source cart ID %s, got %s", cart.ID, reopenedCart.ID)
	}
	if reopenedCart.CartStatus != "active" || len(reopenedItems) != 2 || !almostEqual(reopenedCart.TotalAmount, 46.00) {
		t.Fatalf("expected active reopened cart with 2 items and total 46.00, got status=%s items=%d total=%.2f",
			reopenedCart.CartStatus, len(reopenedItems), reopenedCart.TotalAmount)
	}

	voidedOrd, err := repo.GetRestaurantOrder(ctx, ord.ID)
	if err != nil || voidedOrd.Status != "voided" {
		t.Fatalf("expected original order to be voided, got status=%s err=%v", voidedOrd.Status, err)
	}
}

func TestRestaurantCouponBookkeepingPreservedOnMergeAndEditIntegration(t *testing.T) {
	pool := setupRestaurantMigrationDB(t)
	ctx := context.Background()
	fix := seedRestaurantTestData(t, ctx, pool)

	if _, err := pool.Exec(ctx, `
		INSERT INTO restaurant_promotions (
			store_id, name, code, promotion_type, discount_value,
			applies_to, target_menu_item_ids, coupon_code, is_active
		) VALUES ($1, 'Keep Coupon 20%', 'KEEP20', 'percentage_discount', 20.00, 'menu_item', $2, 'KEEP20', true)
	`, fix.StoreID, []int32{fix.MenuItemID}); err != nil {
		t.Fatalf("seed restaurant promotion: %v", err)
	}

	repo := repository.New(pool)
	uc := NewRestaurantCartUseCase()
	uc.SetRepository(repo)
	uc.SetPool(pool)

	// 1. Add 1 burger (20.00 + 15% VAT = 23.00) and apply coupon KEEP20 (discount = 4.00)
	add1 := uc.AddToCart(ctx, RestaurantAddToCartInput{
		StoreID:    fix.StoreID,
		OrderType:  "takeaway",
		Channel:    "pos",
		MenuItemID: fix.MenuItemID,
		Quantity:   1,
	})
	if add1.StatusCode != utils.CodeOK {
		t.Fatalf("initial AddToCart failed: %d %s", add1.StatusCode, add1.Message)
	}
	cart := add1.Data.(map[string]interface{})["cart"].(RestaurantCartOutput)

	if applyResp := uc.ApplyCouponToCart(ctx, cart.ID, "KEEP20"); applyResp.StatusCode != utils.CodeOK {
		t.Fatalf("ApplyCouponToCart failed: %d %s", applyResp.StatusCode, applyResp.Message)
	}

	// 2. Add the same burger again with new metadata -> triggers insertOrMergeCartItem
	add2 := uc.AddToCart(ctx, RestaurantAddToCartInput{
		CartID:     &cart.ID,
		MenuItemID: fix.MenuItemID,
		Quantity:   1,
		Metadata:   []byte(`{"source":"second_tap"}`),
	})
	if add2.StatusCode != utils.CodeOK {
		t.Fatalf("merge AddToCart failed: %d %s", add2.StatusCode, add2.Message)
	}
	mergedItem := add2.Data.(map[string]interface{})["item"].(RestaurantCartItemOutput)

	// 3. Edit quantity and modifiers on the same line with new metadata -> triggers UpdateCartItemQuantity
	upd1 := uc.UpdateCartItemQuantity(ctx, RestaurantUpdateCartItemInput{
		ItemID:   mergedItem.ID,
		Quantity: 3,
		Metadata: []byte(`{"edited":"qty_only"}`),
	})
	if upd1.StatusCode != utils.CodeOK {
		t.Fatalf("UpdateCartItemQuantity (qty-only) failed: %d %s", upd1.StatusCode, upd1.Message)
	}
	upd2 := uc.UpdateCartItemQuantity(ctx, RestaurantUpdateCartItemInput{
		ItemID:             mergedItem.ID,
		Quantity:           3,
		HasModifiersUpdate: true,
		ModifierIDs:        []int32{},
		Metadata:           []byte(`{"edited":"with_modifiers"}`),
	})
	if upd2.StatusCode != utils.CodeOK {
		t.Fatalf("UpdateCartItemQuantity (with modifiers) failed: %d %s", upd2.StatusCode, upd2.Message)
	}

	// 4. Remove coupon -> the old 4.00 discount must be completely undone (discount_amount=0, total=69.00)
	remResp := uc.RemoveCouponFromCart(ctx, cart.ID)
	if remResp.StatusCode != utils.CodeOK {
		t.Fatalf("RemoveCouponFromCart failed: %d %s", remResp.StatusCode, remResp.Message)
	}
	remData := remResp.Data.(map[string]interface{})
	clearedCart := remData["cart"].(RestaurantCartOutput)
	clearedItems := remData["items"].([]RestaurantCartItemOutput)

	if !almostEqual(clearedCart.DiscountAmount, 0.00) || !almostEqual(clearedCart.TotalAmount, 69.00) {
		t.Fatalf("expected discount 0.00 and total 69.00 after RemoveCouponFromCart, got disc=%.2f total=%.2f",
			clearedCart.DiscountAmount, clearedCart.TotalAmount)
	}
	if len(clearedItems) != 1 || !almostEqual(clearedItems[0].DiscountAmount, 0.00) || !almostEqual(clearedItems[0].LineTotal, 69.00) {
		t.Fatalf("unexpected cleared item state: items=%+v", clearedItems)
	}
}

func TestRestaurantCheckoutRevalidatesSavedCouponIntegration(t *testing.T) {
	pool := setupRestaurantMigrationDB(t)
	ctx := context.Background()
	fix := seedRestaurantTestData(t, ctx, pool)

	if _, err := pool.Exec(ctx, `
		INSERT INTO restaurant_promotions (
			store_id, name, code, promotion_type, discount_value,
			applies_to, coupon_code, is_active, valid_to
		) VALUES ($1, 'Expiring Coupon', 'EXP10', 'fixed_discount', 5.00, 'all', 'EXP10', true, NOW() + INTERVAL '1 hour')
	`, fix.StoreID); err != nil {
		t.Fatalf("seed promotion: %v", err)
	}

	repo := repository.New(pool)
	uc := NewRestaurantCartUseCase()
	uc.SetRepository(repo)
	uc.SetPool(pool)

	addResp := uc.AddToCart(ctx, RestaurantAddToCartInput{
		StoreID:    fix.StoreID,
		OrderType:  "takeaway",
		Channel:    "pos",
		MenuItemID: fix.MenuItemID,
		Quantity:   1,
	})
	if addResp.StatusCode != utils.CodeOK {
		t.Fatalf("AddToCart failed: %d %s", addResp.StatusCode, addResp.Message)
	}
	cart := addResp.Data.(map[string]interface{})["cart"].(RestaurantCartOutput)

	if applyResp := uc.ApplyCouponToCart(ctx, cart.ID, "EXP10"); applyResp.StatusCode != utils.CodeOK {
		t.Fatalf("ApplyCouponToCart failed: %d %s", applyResp.StatusCode, applyResp.Message)
	}

	// Expire the coupon after it was saved on the cart
	if _, err := pool.Exec(ctx, `UPDATE restaurant_promotions SET valid_to = NOW() - INTERVAL '1 minute' WHERE coupon_code = 'EXP10'`); err != nil {
		t.Fatalf("expire promotion: %v", err)
	}

	// Checkout without passing CouponCode must still re-validate the saved cart coupon and reject
	chkResp := uc.ConvertToOrder(ctx, RestaurantCheckoutInput{
		CartID:      cart.ID,
		OrderSource: "pos",
		OrderType:   "takeaway",
	})
	if chkResp.StatusCode != utils.CodeNotFound && chkResp.StatusCode != utils.CodeBadReq {
		t.Fatalf("expected checkout to reject expired saved coupon with 404/400, got %d (%s)", chkResp.StatusCode, chkResp.Message)
	}
}

func TestRestaurantRecipeDeductionRestorationAndVariantsIntegration(t *testing.T) {
	pool := setupRestaurantMigrationDB(t)
	ctx := context.Background()
	fix := seedRestaurantTestData(t, ctx, pool)

	var ingProductID, variantID, recipeID int32
	if err := pool.QueryRow(ctx, `
		INSERT INTO products (organization_id, sku, name, is_active)
		VALUES ($1, 'ING-PAT', 'Patty', true)
		RETURNING id
	`, fix.OrgID).Scan(&ingProductID); err != nil {
		t.Fatalf("insert ingredient product: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO product_variants (product_id, variant_sku, variant_name, variant_attributes, is_active)
		VALUES ($1, 'ING-PAT-V1', 'Patty V1', '{}'::jsonb, true)
		RETURNING id
	`, ingProductID).Scan(&variantID); err != nil {
		t.Fatalf("insert variant: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO recipes (organization_id, recipe_code, recipe_name, yield_quantity, is_active)
		VALUES ($1, 'RCP-BRG-1', 'Burger Recipe', 1, true)
		RETURNING id
	`, fix.OrgID).Scan(&recipeID); err != nil {
		t.Fatalf("insert recipe: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE menu_items SET recipe_id = $1 WHERE id = $2`, recipeID, fix.MenuItemID); err != nil {
		t.Fatalf("link recipe: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO recipe_ingredients (recipe_id, product_id, product_variant_id, quantity, is_optional)
		VALUES ($1, $2, NULL, 5, false)
	`, recipeID, ingProductID); err != nil {
		t.Fatalf("insert recipe ingredient: %v", err)
	}

	repo := repository.New(pool)
	uc := NewRestaurantCartUseCase()
	uc.SetRepository(repo)
	uc.SetPool(pool)

	// 1. Missing inventory_stock row: checkout must fail and record no movement
	addResp := uc.AddToCart(ctx, RestaurantAddToCartInput{
		StoreID:    fix.StoreID,
		OrderType:  "dine_in",
		Channel:    "pos",
		MenuItemID: fix.MenuItemID,
		Quantity:   1,
	})
	if addResp.StatusCode != utils.CodeOK {
		t.Fatalf("AddToCart failed: %d %s", addResp.StatusCode, addResp.Message)
	}
	cart := addResp.Data.(map[string]interface{})["cart"].(RestaurantCartOutput)

	failChk := uc.ConvertToOrder(ctx, RestaurantCheckoutInput{
		CartID:      cart.ID,
		OrderSource: "pos",
		OrderType:   "dine_in",
	})
	if failChk.StatusCode == utils.CodeCreated {
		t.Fatalf("expected checkout to fail when ingredient inventory_stock row is missing")
	}
	var movCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM stock_movements WHERE product_id = $1`, ingProductID).Scan(&movCount); err != nil {
		t.Fatalf("count movements: %v", err)
	}
	if movCount != 0 {
		t.Fatalf("expected 0 stock movements after failed checkout, got %d", movCount)
	}

	// 2. Seed variant stock (10/10) BEFORE null-variant stock (2/2):
	// null-variant recipe must ONLY touch the NULL variant row, and clamping available to 0 then restoring 5 must leave available=2 (not 5).
	if _, err := pool.Exec(ctx, `
		INSERT INTO inventory_stock (store_id, product_id, product_variant_id, quantity_on_hand, quantity_allocated, quantity_available)
		VALUES ($1, $2, $3, 10, 0, 10), ($1, $2, NULL, 2, 0, 2)
	`, fix.StoreID, ingProductID, variantID); err != nil {
		t.Fatalf("seed inventory_stock: %v", err)
	}

	okChk := uc.ConvertToOrder(ctx, RestaurantCheckoutInput{
		CartID:      cart.ID,
		OrderSource: "pos",
		OrderType:   "dine_in",
	})
	if okChk.StatusCode != utils.CodeCreated {
		t.Fatalf("ConvertToOrder failed: %d %s", okChk.StatusCode, okChk.Message)
	}
	ord := okChk.Data.(map[string]interface{})["order"].(repository.RestaurantOrder)

	var nullOnHand, nullAvail, varOnHand, varAvail float64
	if err := pool.QueryRow(ctx, `SELECT quantity_on_hand, quantity_available FROM inventory_stock WHERE store_id=$1 AND product_id=$2 AND product_variant_id IS NULL`, fix.StoreID, ingProductID).Scan(&nullOnHand, &nullAvail); err != nil {
		t.Fatalf("query null stock: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT quantity_on_hand, quantity_available FROM inventory_stock WHERE store_id=$1 AND product_id=$2 AND product_variant_id = $3`, fix.StoreID, ingProductID, variantID).Scan(&varOnHand, &varAvail); err != nil {
		t.Fatalf("query variant stock: %v", err)
	}
	if !almostEqual(nullOnHand, -3) || !almostEqual(nullAvail, 0) {
		t.Fatalf("expected null-variant stock (-3, 0) after deducting 5 from 2, got (%.2f, %.2f)", nullOnHand, nullAvail)
	}
	if !almostEqual(varOnHand, 10) || !almostEqual(varAvail, 10) {
		t.Fatalf("expected variant stock untouched (10, 10), got (%.2f, %.2f)", varOnHand, varAvail)
	}

	// 3. Status-only UpdateOrderWithItems to voided must restore stock to (2, 2), NOT (2, 5)
	voidResp := uc.UpdateOrderWithItems(ctx, ord.ID, RestaurantDirectOrderInput{
		Status: "voided",
	}, false)
	if voidResp.StatusCode != utils.CodeOK {
		t.Fatalf("UpdateOrderWithItems void failed: %d %s", voidResp.StatusCode, voidResp.Message)
	}

	if err := pool.QueryRow(ctx, `SELECT quantity_on_hand, quantity_available FROM inventory_stock WHERE store_id=$1 AND product_id=$2 AND product_variant_id IS NULL`, fix.StoreID, ingProductID).Scan(&nullOnHand, &nullAvail); err != nil {
		t.Fatalf("query restored null stock: %v", err)
	}
	if !almostEqual(nullOnHand, 2) || !almostEqual(nullAvail, 2) {
		t.Fatalf("expected restored null-variant stock (2, 2), got (%.2f, %.2f)", nullOnHand, nullAvail)
	}
}

func TestRestaurantCartItemPriceMergeOrderTypeModifiersAndDiscountCapsIntegration(t *testing.T) {
	pool := setupRestaurantMigrationDB(t)
	ctx := context.Background()
	fix := seedRestaurantTestData(t, ctx, pool)

	var inactiveModID int32
	if err := pool.QueryRow(ctx, `
		INSERT INTO menu_item_modifiers (menu_item_id, modifier_name, modifier_type, price_adjustment, is_active)
		VALUES ($1, 'Disabled Cheese', 'addon', 2.00, false)
		RETURNING id
	`, fix.MenuItemID).Scan(&inactiveModID); err != nil {
		t.Fatalf("insert inactive modifier: %v", err)
	}

	repo := repository.New(pool)
	uc := NewRestaurantCartUseCase()
	uc.SetRepository(repo)
	uc.SetPool(pool)

	// 1. Disabled modifier via Modifiers slice must be rejected
	badModResp := uc.AddToCart(ctx, RestaurantAddToCartInput{
		StoreID:    fix.StoreID,
		OrderType:  "takeaway",
		Channel:    "pos",
		MenuItemID: fix.MenuItemID,
		Quantity:   1,
		Modifiers: []RestaurantModifierSnapshotItem{
			{ModifierID: inactiveModID},
		},
	})
	if badModResp.StatusCode != utils.CodeBadReq {
		t.Fatalf("expected 400 for inactive modifier in Modifiers path, got %d (%s)", badModResp.StatusCode, badModResp.Message)
	}

	// 2. Discount larger than line gross (unit=10, discount=50) must be capped at 10 on saved item
	unit10 := 10.00
	add1 := uc.AddToCart(ctx, RestaurantAddToCartInput{
		StoreID:        fix.StoreID,
		OrderType:      "takeaway",
		Channel:        "pos",
		MenuItemID:     fix.MenuItemID,
		Quantity:       1,
		UnitPrice:      &unit10,
		DiscountAmount: 50.00,
	})
	if add1.StatusCode != utils.CodeOK {
		t.Fatalf("AddToCart(10) failed: %d %s", add1.StatusCode, add1.Message)
	}
	cart := add1.Data.(map[string]interface{})["cart"].(RestaurantCartOutput)
	item1 := add1.Data.(map[string]interface{})["item"].(RestaurantCartItemOutput)
	if !almostEqual(item1.DiscountAmount, 10.00) || !almostEqual(item1.LineTotal, 0.00) {
		t.Fatalf("expected capped discount 10.00 and lineTotal 0.00, got disc=%.2f total=%.2f", item1.DiscountAmount, item1.LineTotal)
	}

	// UpdateCartItemQuantity (qty=2, gross=20, discount=50) must cap saved discount at 20.00
	disc50 := 50.00
	updQty := uc.UpdateCartItemQuantity(ctx, RestaurantUpdateCartItemInput{
		ItemID:         item1.ID,
		Quantity:       2,
		DiscountAmount: &disc50,
	})
	if updQty.StatusCode != utils.CodeOK {
		t.Fatalf("UpdateCartItemQuantity failed: %d %s", updQty.StatusCode, updQty.Message)
	}
	updItem := updQty.Data.(map[string]interface{})["item"].(RestaurantCartItemOutput)
	if !almostEqual(updItem.DiscountAmount, 20.00) {
		t.Fatalf("expected capped discount 20.00 after qty update, got %.2f", updItem.DiscountAmount)
	}

	// 3. Adding same item at different unit price (20 vs 10) must NOT merge into the 10 line
	unit20 := 20.00
	add2 := uc.AddToCart(ctx, RestaurantAddToCartInput{
		CartID:     &cart.ID,
		MenuItemID: fix.MenuItemID,
		Quantity:   1,
		UnitPrice:  &unit20,
	})
	if add2.StatusCode != utils.CodeOK {
		t.Fatalf("AddToCart(20) failed: %d %s", add2.StatusCode, add2.Message)
	}
	itemsAfterMerge := add2.Data.(map[string]interface{})["items"].([]RestaurantCartItemOutput)
	if len(itemsAfterMerge) != 2 {
		t.Fatalf("expected 2 separate lines for differently priced items, got %d", len(itemsAfterMerge))
	}

	// 4. Updating only cart notes (OrderType omitted / empty string) must preserve "takeaway"
	updCartResp := uc.UpdateCart(ctx, repository.UpdateRestaurantCartParams{
		ID:    cart.ID,
		Notes: pgtype.Text{String: "extra napkins", Valid: true},
	})
	if updCartResp.StatusCode != utils.CodeOK {
		t.Fatalf("UpdateCart failed: %d %s", updCartResp.StatusCode, updCartResp.Message)
	}
	updatedCart := updCartResp.Data.(map[string]interface{})["cart"].(RestaurantCartOutput)
	if updatedCart.OrderType != "takeaway" {
		t.Fatalf("expected order_type to remain 'takeaway', got %q", updatedCart.OrderType)
	}
}
