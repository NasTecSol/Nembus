package nembus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	legacyconfig "github.com/NasTecSol/nembus-sap-agent/config"
	agentconfig "github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
)

const (
	upstreamServiceUsername = "sap_migrator"
	upstreamServicePassword = "SapAgent@2026"
)

// HTTPClient implements pipeline.NembusClient using the Nembus Cloud REST API
// and local SQLite outbox state when a direct PostgreSQL connection is not configured.
type HTTPClient struct {
	cfg        *legacyconfig.Config
	sqlite     *db.SQLiteStore
	httpClient *http.Client
}

func NewHTTPClient(cfg *legacyconfig.Config, sqlite *db.SQLiteStore) *HTTPClient {
	return &HTTPClient{
		cfg:    cfg,
		sqlite: sqlite,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *HTTPClient) baseURL() string {
	agentCfg := agentconfig.Get()
	if agentCfg != nil && strings.TrimSpace(agentCfg.Cloud.BaseURL) != "" {
		return strings.TrimRight(strings.TrimSpace(agentCfg.Cloud.BaseURL), "/")
	}
	return strings.TrimRight(strings.TrimSpace(c.cfg.NembusAPIURL), "/")
}

func (c *HTTPClient) tenantHeader() string {
	agentCfg := agentconfig.Get()
	if agentCfg != nil && strings.TrimSpace(agentCfg.TenantSlug) != "" {
		return strings.TrimSpace(agentCfg.TenantSlug)
	}
	return fmt.Sprintf("%d", c.cfg.NembusOrganizationID)
}

func extractAuthToken(body []byte) string {
	var raw struct {
		Data  json.RawMessage `json:"data"`
		Token string          `json:"token"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return ""
	}
	if raw.Token != "" {
		return raw.Token
	}
	if len(raw.Data) > 0 {
		var s string
		if err := json.Unmarshal(raw.Data, &s); err == nil && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
		var obj struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal(raw.Data, &obj); err == nil && strings.TrimSpace(obj.Token) != "" {
			return strings.TrimSpace(obj.Token)
		}
	}
	return ""
}

func (c *HTTPClient) login(ctx context.Context, userLogin, password string) (string, error) {
	urlStr := fmt.Sprintf("%s/api/auth/login", c.baseURL())
	payload, _ := json.Marshal(map[string]string{
		"user_login": userLogin,
		"password":   password,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tenant-id", c.tenantHeader())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		if tok := extractAuthToken(respBytes); tok != "" {
			return tok, nil
		}
	}
	return "", fmt.Errorf("login status %d", resp.StatusCode)
}

func (c *HTTPClient) resolveToken(ctx context.Context, preferServiceAccount bool) (string, error) {
	agentCfg := agentconfig.Get()
	if preferServiceAccount {
		if tok, err := c.login(ctx, upstreamServiceUsername, upstreamServicePassword); err == nil && tok != "" {
			return tok, nil
		}
	}
	if agentCfg != nil {
		if agentCfg.MigratorUser != "" && agentCfg.MigratorPassword != "" {
			if tok, err := c.login(ctx, agentCfg.MigratorUser, agentCfg.MigratorPassword); err == nil && tok != "" {
				return tok, nil
			}
		}
		tok := strings.TrimSpace(agentCfg.MigratorToken)
		if tok == "" {
			tok = strings.TrimSpace(agentCfg.Cloud.APIKey)
		}
		if strings.Count(tok, ".") == 2 {
			return tok, nil
		}
	}
	if tok, err := c.login(ctx, upstreamServiceUsername, upstreamServicePassword); err == nil && tok != "" {
		return tok, nil
	}
	return "", fmt.Errorf("unable to obtain Nembus Cloud auth token for upstream sync")
}

func (c *HTTPClient) doJSON(ctx context.Context, method, path string, body any, preferServiceAccount bool) ([]byte, int, error) {
	token, err := c.resolveToken(ctx, preferServiceAccount)
	if err != nil {
		return nil, 0, err
	}

	var reqReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reqReader = bytes.NewReader(b)
	}

	reqURL := fmt.Sprintf("%s%s", c.baseURL(), path)
	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqReader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("x-tenant-id", c.tenantHeader())
	req.Header.Set("x-organization-id", fmt.Sprintf("%d", c.cfg.NembusOrganizationID))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	return respBytes, resp.StatusCode, err
}

func parseFlexFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f
	default:
		return 0
	}
}

func parseFlexInt64(v any) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int:
		return int64(t)
	case int64:
		return t
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		return n
	default:
		return 0
	}
}

func parseFlexTime(v any) time.Time {
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return time.Now()
	}
	s = strings.TrimSpace(s)
	for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Now()
}

func (c *HTTPClient) dispatchStatusByNumber(ctx context.Context) map[string]string {
	res := make(map[string]string)
	if c.sqlite == nil {
		return res
	}
	dispatches, err := c.sqlite.GetUpstreamDispatches(ctx, 500)
	if err != nil {
		return res
	}
	for _, d := range dispatches {
		res[d.TransactionNumber] = d.Status
	}
	return res
}

func (c *HTTPClient) approvedEntityNumbers(ctx context.Context, entityType string) map[string]bool {
	approved := make(map[string]bool)
	path := "/api/approvals?domain=sap_invoice&status=approved&limit=100"
	bodyBytes, status, err := c.doJSON(ctx, http.MethodGet, path, nil, false)
	if err != nil || status != http.StatusOK {
		return approved
	}
	var env struct {
		Data []struct {
			EntityType string          `json:"entity_type"`
			EntityID   string          `json:"entity_id"`
			Status     string          `json:"status"`
			Payload    json.RawMessage `json:"payload"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &env); err != nil {
		return approved
	}
	for _, item := range env.Data {
		if item.EntityType != entityType || item.Status != "approved" {
			continue
		}
		approved[item.EntityID] = true
		var p map[string]any
		if json.Unmarshal(item.Payload, &p) == nil {
			if num, ok := p["transaction_number"].(string); ok && num != "" {
				approved[num] = true
			}
			if num, ok := p["order_number"].(string); ok && num != "" {
				approved[num] = true
			}
		}
	}
	return approved
}

func (c *HTTPClient) fetchStoreTransactionsRaw(ctx context.Context, limit int) ([]POSTransaction, error) {
	if limit <= 0 {
		limit = 50
	}
	storeID := c.cfg.NembusStoreID
	if storeID <= 0 {
		storeID = 1
	}
	path := fmt.Sprintf("/api/pos/stores/%d/transactions?page=1&limit=%d", storeID, limit)
	bodyBytes, status, err := c.doJSON(ctx, http.MethodGet, path, nil, false)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, nil
	}

	var env struct {
		Data struct {
			Data []map[string]any `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &env); err != nil {
		return nil, nil
	}

	var out []POSTransaction
	for _, row := range env.Data.Data {
		txNum, _ := row["transaction_number"].(string)
		if txNum == "" || strings.HasPrefix(txNum, "SIM-POS-") {
			// Skip historical SAP invoices migrated into Nembus POS
			continue
		}
		st, _ := row["status"].(string)
		if st != "" && st != "completed" {
			continue
		}
		total := parseFlexFloat(row["total_amount"])
		tx := POSTransaction{
			ID:                parseFlexInt64(row["id"]),
			StoreID:           storeID,
			CashierID:         1,
			CashierSessionID:  1,
			TransactionNumber: txNum,
			TransactionDate:   parseFlexTime(row["transaction_date"]),
			TotalAmount:       total,
			AmountPaid:        total,
			Status:            "completed",
		}
		out = append(out, tx)
	}
	return out, nil
}

func (c *HTTPClient) FetchApprovedPendingPOSTransactions(ctx context.Context, limit int) ([]POSTransaction, error) {
	all, err := c.fetchStoreTransactionsRaw(ctx, limit)
	if err != nil || len(all) == 0 {
		return nil, err
	}
	dispatched := c.dispatchStatusByNumber(ctx)
	approved := c.approvedEntityNumbers(ctx, "pos_transaction")

	var list []POSTransaction
	for _, tx := range all {
		st := dispatched[tx.TransactionNumber]
		if st == "posted" {
			continue
		}
		if st == "held" && (approved[tx.TransactionNumber] || approved[fmt.Sprintf("%d", tx.ID)]) {
			tx.Metadata = json.RawMessage(`{"sap_sync_state":"approved_pending_upload"}`)
			list = append(list, tx)
		}
	}
	return list, nil
}

func (c *HTTPClient) FetchUnpostedPOSTransactions(ctx context.Context, limit int) ([]POSTransaction, error) {
	all, err := c.fetchStoreTransactionsRaw(ctx, limit)
	if err != nil || len(all) == 0 {
		return nil, err
	}
	dispatched := c.dispatchStatusByNumber(ctx)

	var list []POSTransaction
	for _, tx := range all {
		st := dispatched[tx.TransactionNumber]
		if st == "posted" || st == "held" {
			continue
		}
		list = append(list, tx)
	}
	return list, nil
}

func (c *HTTPClient) FetchTransactionLines(ctx context.Context, transactionID int64) ([]POSTransactionLine, error) {
	path := fmt.Sprintf("/api/pos/transactions/%d/full", transactionID)
	bodyBytes, status, err := c.doJSON(ctx, http.MethodGet, path, nil, false)
	if err != nil || status != http.StatusOK {
		return nil, fmt.Errorf("fetch transaction %d lines failed (status %d)", transactionID, status)
	}

	var env struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &env); err != nil {
		return nil, err
	}

	var lines []POSTransactionLine
	for idx, row := range env.Data {
		sku, _ := row["sku"].(string)
		if sku == "" {
			sku, _ = row["product_sku"].(string)
		}
		name, _ := row["product_name"].(string)
		qty := parseFlexFloat(row["quantity"])
		unitPrice := parseFlexFloat(row["unit_price"])
		disc := parseFlexFloat(row["discount_amount_2"])
		lineTotal := parseFlexFloat(row["line_total"])
		if lineTotal <= 0 {
			lineTotal = (qty * unitPrice) - disc
		}
		lines = append(lines, POSTransactionLine{
			ID:             parseFlexInt64(row["line_id"]),
			TransactionID:  transactionID,
			ProductID:      int32(parseFlexInt64(row["product_id"])),
			ProductSKU:     sku,
			ProductName:    name,
			Quantity:       qty,
			UnitPrice:      unitPrice,
			DiscountAmount: disc,
			Subtotal:       lineTotal,
			LineTotal:      lineTotal,
			LineNumber:     idx + 1,
		})
	}
	return lines, nil
}

func (c *HTTPClient) FetchTransactionPayments(ctx context.Context, transactionID int64) ([]POSPayment, error) {
	path := fmt.Sprintf("/api/pos/transactions/%d/payments", transactionID)
	bodyBytes, status, err := c.doJSON(ctx, http.MethodGet, path, nil, false)
	if err != nil || status != http.StatusOK {
		return nil, nil
	}
	var env struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &env); err != nil {
		return nil, nil
	}
	var payments []POSPayment
	for _, row := range env.Data {
		method, _ := row["payment_method"].(string)
		if method == "" {
			method = "cash"
		}
		var refPtr *string
		if ref, ok := row["reference_number"].(string); ok && ref != "" {
			refPtr = &ref
		}
		payments = append(payments, POSPayment{
			ID:            parseFlexInt64(row["id"]),
			TransactionID: transactionID,
			PaymentMethod: method,
			Amount:        parseFlexFloat(row["amount"]),
			Reference:     refPtr,
			PaymentDate:   parseFlexTime(row["payment_date"]),
		})
	}
	return payments, nil
}

func (c *HTTPClient) MarkTransactionPostedToSAP(ctx context.Context, transactionID int64, sapDocEntry, sapDocNum int) error {
	return nil
}

func (c *HTTPClient) FetchCustomerSAPCode(ctx context.Context, customerID int32) (string, error) {
	path := fmt.Sprintf("/api/customers/%d", customerID)
	bodyBytes, status, err := c.doJSON(ctx, http.MethodGet, path, nil, false)
	if err != nil || status != http.StatusOK {
		return c.cfg.NembusDefaultCustomer, nil
	}
	var env struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &env); err == nil && env.Data != nil {
		if code, ok := env.Data["customer_code"].(string); ok && strings.TrimSpace(code) != "" {
			return strings.TrimSpace(code), nil
		}
	}
	return c.cfg.NembusDefaultCustomer, nil
}

func (c *HTTPClient) FetchUnpostedSalesReturns(ctx context.Context, limit int) ([]SalesReturn, error) {
	return nil, nil
}

func (c *HTTPClient) FetchSalesReturnLines(ctx context.Context, returnID int64) ([]SalesReturnLine, error) {
	return nil, nil
}

func (c *HTTPClient) MarkSalesReturnPostedToSAP(ctx context.Context, returnID int64, sapDocEntry, sapDocNum int) error {
	return nil
}

func (c *HTTPClient) FetchOriginalSAPDocEntry(ctx context.Context, originalTxID int32) (int, error) {
	return 0, nil
}

func (c *HTTPClient) ensureStockHoldPolicy(ctx context.Context) {
	payload := map[string]any{
		"domain":              "sap_invoice",
		"action":              "stock_hold",
		"is_enabled":          true,
		"requires_approval":   true,
		"required_permission": "sap:invoice:override",
		"rules":               map[string]any{},
	}
	_, _, _ = c.doJSON(ctx, http.MethodPut, "/api/approvals/policy", payload, false)
}

func (c *HTTPClient) HoldTransactionForStockReview(ctx context.Context, txID int64, txNum string, storeID int32, whCode string, totalAmount float64, shortageItems []ShortageItem, summary string) error {
	c.ensureStockHoldPolicy(ctx)
	reqBody := map[string]any{
		"domain":              "sap_invoice",
		"action":              "stock_hold",
		"entity_type":         "pos_transaction",
		"entity_id":           fmt.Sprintf("%d", txID),
		"required_permission": "sap:invoice:override",
		"summary":             summary,
		"payload": map[string]any{
			"transaction_id":     txID,
			"transaction_number": txNum,
			"store_id":           storeID,
			"warehouse_code":     whCode,
			"total_amount":       totalAmount,
			"shortage_items":     shortageItems,
			"reason":             summary,
		},
	}
	_, _, err := c.doJSON(ctx, http.MethodPost, "/api/approvals/submit", reqBody, true)
	return err
}

func (c *HTTPClient) fetchSalesOrdersRaw(ctx context.Context) ([]SalesOrderV2, error) {
	path := fmt.Sprintf("/api/orders?organization_id=%d", c.cfg.NembusOrganizationID)
	bodyBytes, status, err := c.doJSON(ctx, http.MethodGet, path, nil, false)
	if err != nil || status != http.StatusOK {
		return nil, nil
	}
	var env struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &env); err != nil {
		return nil, nil
	}

	var list []SalesOrderV2
	for _, row := range env.Data {
		idStr, _ := row["id"].(string)
		orderID, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		orderNum, _ := row["order_number"].(string)
		if orderNum == "" || strings.HasPrefix(orderNum, "SO-SAP-") {
			continue
		}
		orderStatus, _ := row["order_status"].(string)
		fulfillStatus, _ := row["fulfillment_status"].(string)
		if orderStatus != "confirmed" && orderStatus != "fulfilled" && fulfillStatus != "fulfilled" {
			continue
		}

		var metaRaw json.RawMessage
		if m, ok := row["metadata"]; ok && m != nil {
			if b, err := json.Marshal(m); err == nil {
				metaRaw = b
				var mMap map[string]any
				if json.Unmarshal(b, &mMap) == nil {
					if docNum, _ := mMap["sap_doc_num"].(string); docNum != "" {
						continue
					}
					if docNumFloat := parseFlexFloat(mMap["sap_doc_num"]); docNumFloat > 0 {
						continue
					}
				}
			}
		}

		var custPtr *int32
		if cid := int32(parseFlexInt64(row["customer_id"])); cid > 0 {
			custPtr = &cid
		}
		payStatus, _ := row["payment_status"].(string)

		list = append(list, SalesOrderV2{
			ID:            orderID,
			OrderNumber:   orderNum,
			StoreID:       int32(parseFlexInt64(row["store_id"])),
			CustomerID:    custPtr,
			OrderDate:     parseFlexTime(row["order_date"]),
			TotalAmount:   parseFlexFloat(row["total_amount"]),
			PaymentStatus: payStatus,
			Metadata:      metaRaw,
		})
	}
	return list, nil
}

func (c *HTTPClient) FetchApprovedPendingSalesOrdersV2(ctx context.Context, limit int) ([]SalesOrderV2, error) {
	all, err := c.fetchSalesOrdersRaw(ctx)
	if err != nil || len(all) == 0 {
		return nil, err
	}
	dispatched := c.dispatchStatusByNumber(ctx)
	approved := c.approvedEntityNumbers(ctx, "sales_order_v2")

	var list []SalesOrderV2
	for _, o := range all {
		st := dispatched[o.OrderNumber]
		if st == "posted" {
			continue
		}
		if st == "held" && (approved[o.OrderNumber] || approved[o.ID.String()]) {
			o.Metadata = json.RawMessage(`{"sap_sync_state":"approved_pending_upload"}`)
			list = append(list, o)
		}
	}
	return list, nil
}

func (c *HTTPClient) FetchUnpostedSalesOrdersV2(ctx context.Context, limit int) ([]SalesOrderV2, error) {
	all, err := c.fetchSalesOrdersRaw(ctx)
	if err != nil || len(all) == 0 {
		return nil, err
	}
	dispatched := c.dispatchStatusByNumber(ctx)

	var list []SalesOrderV2
	for _, o := range all {
		st := dispatched[o.OrderNumber]
		if st == "posted" || st == "held" {
			continue
		}
		list = append(list, o)
	}
	return list, nil
}

func (c *HTTPClient) FetchSalesOrderLinesV2(ctx context.Context, orderID uuid.UUID) ([]SalesOrderLineV2, error) {
	path := fmt.Sprintf("/api/orders/%s/lines", orderID.String())
	bodyBytes, status, err := c.doJSON(ctx, http.MethodGet, path, nil, false)
	if err != nil || status != http.StatusOK {
		return nil, fmt.Errorf("fetch order %s lines failed (status %d)", orderID, status)
	}
	var env struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &env); err != nil {
		return nil, err
	}
	var lines []SalesOrderLineV2
	for idx, row := range env.Data {
		sku, _ := row["product_sku"].(string)
		name, _ := row["product_name"].(string)
		lines = append(lines, SalesOrderLineV2{
			ProductID:   int32(parseFlexInt64(row["product_id"])),
			ProductSKU:  sku,
			ProductName: name,
			Quantity:    parseFlexFloat(row["quantity_ordered"]),
			UnitPrice:   parseFlexFloat(row["unit_price"]),
			TaxAmount:   parseFlexFloat(row["tax_amount"]),
			LineTotal:   parseFlexFloat(row["line_total"]),
			LineNumber:  idx + 1,
		})
	}
	return lines, nil
}

func (c *HTTPClient) MarkSalesOrderPostedToSAP(ctx context.Context, orderID uuid.UUID, sapDocEntry, sapDocNum int) error {
	if c.sqlite != nil {
		orders, _ := c.fetchSalesOrdersRaw(ctx)
		for _, o := range orders {
			if o.ID == orderID {
				_ = c.sqlite.RecordUpstreamDispatch(ctx, o.OrderNumber, o.OrderDate, o.TotalAmount, &sapDocEntry, &sapDocNum, "posted", "")
				break
			}
		}
	}
	return nil
}

func (c *HTTPClient) HoldSalesOrderForStockReview(ctx context.Context, orderID uuid.UUID, orderNum string, storeID int32, whCode string, totalAmount float64, shortageItems []ShortageItem, summary string) error {
	c.ensureStockHoldPolicy(ctx)
	if c.sqlite != nil {
		_ = c.sqlite.RecordUpstreamDispatch(ctx, orderNum, time.Now(), totalAmount, nil, nil, "held", summary)
	}
	reqBody := map[string]any{
		"domain":              "sap_invoice",
		"action":              "stock_hold",
		"entity_type":         "sales_order_v2",
		"entity_id":           orderID.String(),
		"required_permission": "sap:invoice:override",
		"summary":             summary,
		"payload": map[string]any{
			"sales_order_id": orderID.String(),
			"order_number":   orderNum,
			"store_id":       storeID,
			"warehouse_code": whCode,
			"total_amount":   totalAmount,
			"shortage_items": shortageItems,
			"reason":         summary,
		},
	}
	_, _, err := c.doJSON(ctx, http.MethodPost, "/api/approvals/submit", reqBody, true)
	return err
}
