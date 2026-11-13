package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRestaurantCartHandlerValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRestaurantHandler(usecase.NewRestaurantUseCase())

	r.Use(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), middleware.RepoKey, &repository.Queries{})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	rest := r.Group("/api/restaurant")
	rest.POST("/carts/add", h.AddToRestaurantCart)
	rest.POST("/carts/new", h.CreateNewRestaurantCart)
	rest.GET("/carts", h.GetActiveRestaurantCart)
	rest.GET("/carts/:id", h.GetRestaurantCart)
	rest.PUT("/cart-items/:item_id", h.UpdateRestaurantCartItem)
	rest.POST("/carts/:id/checkout", h.ConvertRestaurantCartToOrder)

	cases := []struct {
		method string
		path   string
		body   string
		code   int
	}{
		{"POST", "/api/restaurant/carts/add", `{"store_id":1}`, http.StatusBadRequest},
		{"POST", "/api/restaurant/carts/new", `{}`, http.StatusBadRequest},
		{"GET", "/api/restaurant/carts", ``, http.StatusBadRequest},
		{"GET", "/api/restaurant/carts/not-a-uuid", ``, http.StatusBadRequest},
		{"PUT", "/api/restaurant/cart-items/not-a-uuid", `{"quantity":2}`, http.StatusBadRequest},
		{"POST", "/api/restaurant/carts/not-a-uuid/checkout", `{}`, http.StatusBadRequest},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tc.code {
			t.Fatalf("%s %s: expected status %d, got %d body=%s", tc.method, tc.path, tc.code, w.Code, w.Body.String())
		}
	}
}

func TestRestaurantCartHTTPFlowIntegration(t *testing.T) {
	dsn := os.Getenv("NEMBUS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_TEST_DATABASE_URL to run HTTP integration test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect db: %v", err)
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

	if _, err := pool.Exec(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
		t.Fatalf("reset public schema: %v", err)
	}
	for _, m := range []string{
		filepath.Join("..", "db", "migrations", "20260918111505_initial_schema.sql"),
	} {
		sqlBytes, err := os.ReadFile(m)
		if err != nil {
			t.Fatalf("read migration: %v", err)
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			t.Fatalf("exec migration: %v", err)
		}
	}

	var orgID, storeID, tableID, taxID, catID, itemID, modID int32
	if err := pool.QueryRow(ctx, `INSERT INTO organizations (name, code) VALUES ('HTTP Org', 'HORG') RETURNING id`).Scan(&orgID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO stores (organization_id, name, code, store_type) VALUES ($1, 'HTTP Store', 'HST', 'restaurant') RETURNING id`, orgID).Scan(&storeID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO restaurant_tables (store_id, table_number, table_name) VALUES ($1, 'T9', 'Table 9') RETURNING id`, storeID).Scan(&tableID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO tax_categories (name, code, tax_rate, is_inclusive) VALUES ('VAT 15', 'VAT15', 15.00, false) RETURNING id`).Scan(&taxID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO menu_categories (store_id, name, code) VALUES ($1, 'Mains', 'MAINS') RETURNING id`, storeID).Scan(&catID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO menu_items (store_id, menu_category_id, name, base_price, tax_category_id) VALUES ($1, $2, 'Grill Plate', 40.00, $3) RETURNING id`, storeID, catID, taxID).Scan(&itemID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO menu_item_modifiers (menu_item_id, modifier_name, modifier_type, price_adjustment) VALUES ($1, 'Garlic Sauce', 'addon', 5.00) RETURNING id`, itemID).Scan(&modID); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := repository.New(pool)
	h := NewRestaurantHandler(usecase.NewRestaurantUseCase())
	r.Use(func(c *gin.Context) {
		cCtx := context.WithValue(c.Request.Context(), middleware.RepoKey, repo)
		cCtx = context.WithValue(cCtx, middleware.PoolKey, pool)
		c.Request = c.Request.WithContext(cCtx)
		c.Next()
	})
	rest := r.Group("/api/restaurant")
	rest.POST("/carts/add", h.AddToRestaurantCart)
	rest.GET("/carts/:id", h.GetRestaurantCart)
	rest.POST("/carts/:id/checkout", h.ConvertRestaurantCartToOrder)
	rest.POST("/carts/:id/coupon", h.ApplyRestaurantCoupon)
	rest.DELETE("/carts/:id/coupon", h.RemoveRestaurantCoupon)
	rest.POST("/orders/online", h.CreateOnlineOrder)
	rest.POST("/orders/:order_id/settle", h.SettleOrder)
	rest.PATCH("/orders/:order_id/status", h.UpdateOrderStatus)
	rest.GET("/orders/:order_id/status-history", h.GetRestaurantOrderStatusHistory)
	rest.GET("/stores/:store_id/kds", h.GetKdsOrders)

	// Seed open cashier session so POS checkout & settle endpoints auto-create POS transactions
	var userID, cashierID, terminalID, sessionID int32
	if err := pool.QueryRow(ctx, `INSERT INTO users (organization_id, username, email, password_hash) VALUES ($1, 'httpcashier', 'httpcashier@example.com', 'hash') RETURNING id`, orgID).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO cashiers (user_id, store_id, cashier_code) VALUES ($1, $2, 'HTTPC-1') RETURNING id`, userID, storeID).Scan(&cashierID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO pos_terminals (store_id, terminal_code, terminal_name) VALUES ($1, 'HTTPT-1', 'HTTP Term') RETURNING id`, storeID).Scan(&terminalID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO cashier_sessions (cashier_id, pos_terminal_id, session_number, opening_balance, expected_balance, status) VALUES ($1, $2, 'HTTPS-1', 50.00, 50.00, 'open') RETURNING id`, cashierID, terminalID).Scan(&sessionID); err != nil {
		t.Fatal(err)
	}

	// Add to cart via HTTP
	addBody, _ := json.Marshal(map[string]interface{}{
		"store_id":     storeID,
		"table_id":     tableID,
		"menu_item_id": itemID,
		"quantity":     2,
		"modifier_ids": []int32{modID},
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/api/restaurant/carts/add", bytes.NewReader(addBody)))
	if w.Code != http.StatusOK {
		t.Fatalf("POST /carts/add returned %d: %s", w.Code, w.Body.String())
	}

	var addResp struct {
		Data struct {
			Cart struct {
				ID          string  `json:"id"`
				TotalAmount float64 `json:"total_amount"`
			} `json:"cart"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &addResp); err != nil {
		t.Fatal(err)
	}
	// (40 + 5) * 2 = 90 + 15% (13.50) = 103.50
	if addResp.Data.Cart.ID == "" || addResp.Data.Cart.TotalAmount < 103.49 || addResp.Data.Cart.TotalAmount > 103.51 {
		t.Fatalf("unexpected cart HTTP response: id=%s total=%.2f body=%s", addResp.Data.Cart.ID, addResp.Data.Cart.TotalAmount, w.Body.String())
	}

	// Checkout via HTTP
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest("POST", "/api/restaurant/carts/"+addResp.Data.Cart.ID+"/checkout", bytes.NewBufferString(`{"amount_paid":103.50}`)))
	if w2.Code != http.StatusCreated {
		t.Fatalf("POST /carts/:id/checkout returned %d: %s", w2.Code, w2.Body.String())
	}

	// Online order via HTTP (verifying items & totals saved)
	onlineBody, _ := json.Marshal(map[string]interface{}{
		"store_id": storeID,
		"items": []map[string]interface{}{
			{"menu_item_id": itemID, "quantity": "1", "unit_price": "40.00"},
		},
	})
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, httptest.NewRequest("POST", "/api/restaurant/orders/online", bytes.NewReader(onlineBody)))
	if w3.Code != http.StatusCreated {
		t.Fatalf("POST /orders/online returned %d: %s", w3.Code, w3.Body.String())
	}
	var onlineResp struct {
		Data struct {
			ID    int32             `json:"id"`
			Items []json.RawMessage `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w3.Body.Bytes(), &onlineResp); err != nil {
		t.Fatal(err)
	}
	if onlineResp.Data.ID <= 0 || len(onlineResp.Data.Items) != 1 {
		t.Fatalf("expected online order with 1 saved item, got id=%d items=%d body=%s", onlineResp.Data.ID, len(onlineResp.Data.Items), w3.Body.String())
	}

	// Settle the online/unpaid order via HTTP without explicit pos_transaction_id (auto-creates POS transaction)
	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, httptest.NewRequest("POST", "/api/restaurant/orders/"+strconvI32(onlineResp.Data.ID)+"/settle", nil))
	if w4.Code != http.StatusOK {
		t.Fatalf("POST /orders/:order_id/settle returned %d: %s", w4.Code, w4.Body.String())
	}
	var settleResp struct {
		Data struct {
			PosTransactionID int32 `json:"pos_transaction_id"`
			TransactionID    int32 `json:"transaction_id"`
			Order            struct {
				Status           string `json:"status"`
				PaymentStatus    string `json:"payment_status"`
				PosTransactionID *int32 `json:"pos_transaction_id"`
			} `json:"order"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w4.Body.Bytes(), &settleResp); err != nil {
		t.Fatal(err)
	}
	if settleResp.Data.PosTransactionID <= 0 || settleResp.Data.TransactionID != settleResp.Data.PosTransactionID ||
		settleResp.Data.Order.Status != "pending" || settleResp.Data.Order.PaymentStatus != "paid" {
		t.Fatalf("unexpected settle HTTP response: %+v body=%s", settleResp.Data, w4.Body.String())
	}

	// POS channel cart checkout via HTTP (auto-creates pos_transactions + pos_payments)
	posAddBody, _ := json.Marshal(map[string]interface{}{
		"store_id":     storeID,
		"channel":      "pos",
		"order_type":   "takeout",
		"menu_item_id": itemID,
		"quantity":     1,
	})
	w5 := httptest.NewRecorder()
	r.ServeHTTP(w5, httptest.NewRequest("POST", "/api/restaurant/carts/add", bytes.NewReader(posAddBody)))
	if w5.Code != http.StatusOK {
		t.Fatalf("POST /carts/add pos returned %d: %s", w5.Code, w5.Body.String())
	}
	var posAddResp struct {
		Data struct {
			Cart struct {
				ID string `json:"id"`
			} `json:"cart"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w5.Body.Bytes(), &posAddResp); err != nil {
		t.Fatal(err)
	}
	w6 := httptest.NewRecorder()
	r.ServeHTTP(w6, httptest.NewRequest("POST", "/api/restaurant/carts/"+posAddResp.Data.Cart.ID+"/checkout", bytes.NewBufferString(`{"order_source":"pos"}`)))
	if w6.Code != http.StatusCreated {
		t.Fatalf("POST /carts/:id/checkout pos returned %d: %s", w6.Code, w6.Body.String())
	}
	var posCheckoutResp struct {
		Data struct {
			PosTransactionID int32 `json:"pos_transaction_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w6.Body.Bytes(), &posCheckoutResp); err != nil {
		t.Fatal(err)
	}
	if posCheckoutResp.Data.PosTransactionID <= 0 {
		t.Fatalf("expected pos_transaction_id > 0 on POS cart checkout, body=%s", w6.Body.String())
	}

	// 50.00 initial + 103.50 (first paid POS checkout) + 46.00 (settled online order) + 46.00 (POS takeout checkout) = 245.50
	var expectedBal float64
	if err := pool.QueryRow(ctx, `SELECT expected_balance::float8 FROM cashier_sessions WHERE id = $1`, sessionID).Scan(&expectedBal); err != nil {
		t.Fatal(err)
	}
	if expectedBal < 245.49 || expectedBal > 245.51 {
		t.Fatalf("expected cashier session expected_balance 245.50, got %.2f", expectedBal)
	}

	// === HTTP tests for Coupon, Status History & KDS ===
	if _, err := pool.Exec(ctx, `
		INSERT INTO restaurant_promotions (store_id, code, name, promotion_type, applies_to, coupon_code, discount_value, is_active)
		VALUES ($1, 'HTTP10', 'HTTP 10% Off', 'percentage_discount', 'all', 'HTTP10', 10.00, true)
	`, storeID); err != nil {
		t.Fatal(err)
	}

	kdsReqCartBody, _ := json.Marshal(map[string]interface{}{
		"store_id":     storeID,
		"table_id":     tableID,
		"order_type":   "dine_in",
		"menu_item_id": itemID,
		"quantity":     1,
	})
	w7 := httptest.NewRecorder()
	r.ServeHTTP(w7, httptest.NewRequest("POST", "/api/restaurant/carts/add", bytes.NewReader(kdsReqCartBody)))
	if w7.Code != http.StatusOK {
		t.Fatalf("POST /carts/add for coupon+kds failed %d: %s", w7.Code, w7.Body.String())
	}
	var c7 struct {
		Data struct {
			Cart struct {
				ID string `json:"id"`
			} `json:"cart"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w7.Body.Bytes(), &c7)

	wCoupon := httptest.NewRecorder()
	r.ServeHTTP(wCoupon, httptest.NewRequest("POST", "/api/restaurant/carts/"+c7.Data.Cart.ID+"/coupon", bytes.NewBufferString(`{"coupon_code":"HTTP10"}`)))
	if wCoupon.Code != http.StatusOK {
		t.Fatalf("POST /carts/:id/coupon returned %d: %s", wCoupon.Code, wCoupon.Body.String())
	}
	wDelCoupon := httptest.NewRecorder()
	r.ServeHTTP(wDelCoupon, httptest.NewRequest("DELETE", "/api/restaurant/carts/"+c7.Data.Cart.ID+"/coupon", nil))
	if wDelCoupon.Code != http.StatusOK {
		t.Fatalf("DELETE /carts/:id/coupon returned %d: %s", wDelCoupon.Code, wDelCoupon.Body.String())
	}
	wCheckout7 := httptest.NewRecorder()
	r.ServeHTTP(wCheckout7, httptest.NewRequest("POST", "/api/restaurant/carts/"+c7.Data.Cart.ID+"/checkout", bytes.NewBufferString(`{"order_source":"table","order_type":"dine_in","coupon_code":"HTTP10"}`)))
	if wCheckout7.Code != http.StatusCreated {
		t.Fatalf("POST /carts/:id/checkout with coupon returned %d: %s", wCheckout7.Code, wCheckout7.Body.String())
	}
	var chk7 struct {
		Data struct {
			Order struct {
				ID         int32  `json:"id"`
				CouponCode string `json:"coupon_code"`
			} `json:"order"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wCheckout7.Body.Bytes(), &chk7)
	if chk7.Data.Order.ID <= 0 || chk7.Data.Order.CouponCode != "HTTP10" {
		t.Fatalf("expected order with coupon_code HTTP10, body=%s", wCheckout7.Body.String())
	}

	wPatchStatus := httptest.NewRecorder()
	r.ServeHTTP(wPatchStatus, httptest.NewRequest("PATCH", "/api/restaurant/orders/"+strconvI32(chk7.Data.Order.ID)+"/status", bytes.NewBufferString(`{"status":"preparing","reason":"started"}`)))
	if wPatchStatus.Code != http.StatusOK {
		t.Fatalf("PATCH /orders/:order_id/status returned %d: %s", wPatchStatus.Code, wPatchStatus.Body.String())
	}

	wHist := httptest.NewRecorder()
	r.ServeHTTP(wHist, httptest.NewRequest("GET", "/api/restaurant/orders/"+strconvI32(chk7.Data.Order.ID)+"/status-history", nil))
	if wHist.Code != http.StatusOK {
		t.Fatalf("GET /orders/:order_id/status-history returned %d: %s", wHist.Code, wHist.Body.String())
	}
	var histResp struct {
		Data []struct {
			ToStatus string `json:"to_status"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wHist.Body.Bytes(), &histResp)
	if len(histResp.Data) != 2 || histResp.Data[0].ToStatus != "pending" || histResp.Data[1].ToStatus != "preparing" {
		t.Fatalf("unexpected status history payload: %s", wHist.Body.String())
	}

	wKds := httptest.NewRecorder()
	r.ServeHTTP(wKds, httptest.NewRequest("GET", "/api/restaurant/stores/"+strconvI32(storeID)+"/kds?statuses=pending,preparing", nil))
	if wKds.Code != http.StatusOK {
		t.Fatalf("GET /stores/:store_id/kds returned %d: %s", wKds.Code, wKds.Body.String())
	}
	var kdsHTTP struct {
		Data []struct {
			ID     int32         `json:"id"`
			Status string        `json:"status"`
			Items  []interface{} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(wKds.Body.Bytes(), &kdsHTTP); err != nil {
		t.Fatal(err)
	}
	foundPrepared := false
	for _, ko := range kdsHTTP.Data {
		if ko.ID == chk7.Data.Order.ID && ko.Status == "preparing" && len(ko.Items) == 1 {
			foundPrepared = true
		}
	}
	if !foundPrepared {
		t.Fatalf("expected order %d (preparing, 1 item) in GET /kds payload: %s", chk7.Data.Order.ID, wKds.Body.String())
	}

	// Underpayment via HTTP POS checkout (5.00 vs 46.00) must return 400 Bad Request
	wUnderAdd := httptest.NewRecorder()
	r.ServeHTTP(wUnderAdd, httptest.NewRequest("POST", "/api/restaurant/carts/add", bytes.NewReader(kdsReqCartBody)))
	if wUnderAdd.Code != http.StatusOK {
		t.Fatalf("POST /carts/add for underpayment HTTP test failed %d: %s", wUnderAdd.Code, wUnderAdd.Body.String())
	}
	var cUnder struct {
		Data struct {
			Cart struct {
				ID string `json:"id"`
			} `json:"cart"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wUnderAdd.Body.Bytes(), &cUnder)

	wUnderCheckout := httptest.NewRecorder()
	r.ServeHTTP(wUnderCheckout, httptest.NewRequest("POST", "/api/restaurant/carts/"+cUnder.Data.Cart.ID+"/checkout", bytes.NewBufferString(`{"order_source":"pos","amount_paid":5}`)))
	if wUnderCheckout.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request on underpaid HTTP POS checkout, got %d: %s", wUnderCheckout.Code, wUnderCheckout.Body.String())
	}

	// Underpayment via HTTP settle endpoint must return 400 Bad Request
	wUnderSettle := httptest.NewRecorder()
	r.ServeHTTP(wUnderSettle, httptest.NewRequest("POST", "/api/restaurant/orders/"+strconvI32(chk7.Data.Order.ID)+"/settle", bytes.NewBufferString(`{"payment_method":"cash","amount_paid":5}`)))
	if wUnderSettle.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request on underpaid HTTP settle, got %d: %s", wUnderSettle.Code, wUnderSettle.Body.String())
	}
}

func strconvI32(v int32) string {
	return strconv.Itoa(int(v))
}
