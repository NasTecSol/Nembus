package nembus

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	legacyconfig "github.com/NasTecSol/nembus-sap-agent/config"
	agentconfig "github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
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

func (c *HTTPClient) resolveToken(ctx context.Context, _ bool) (string, error) {
	agentCfg := agentconfig.Get()
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
	return "", fmt.Errorf("unable to obtain Nembus Cloud auth token for upstream sync (configure Migrator credentials or Cloud APIKey)")
}

func (c *HTTPClient) organizationID() int32 {
	if agentCfg := agentconfig.Get(); agentCfg != nil && agentCfg.Cloud.OrganizationID > 0 {
		return int32(agentCfg.Cloud.OrganizationID)
	}
	if c.cfg != nil && c.cfg.NembusOrganizationID > 0 {
		return c.cfg.NembusOrganizationID
	}
	return 1
}

func (c *HTTPClient) jwtOrgAndUser(token string) (int32, int32) {
	orgID := c.organizationID()
	var userID int32 = 1

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return orgID, userID
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return orgID, userID
	}
	var claims map[string]any
	if json.Unmarshal(raw, &claims) != nil {
		return orgID, userID
	}
	if v := int32(parseFlexInt64(claims["organization_id"])); v > 0 {
		orgID = v
	}
	if v := int32(parseFlexInt64(claims["user_id"])); v > 0 {
		userID = v
	}
	return orgID, userID
}

func (c *HTTPClient) doJSONRequest(ctx context.Context, token, method, path string, body any) ([]byte, int, error) {
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
	req.Header.Set("x-organization-id", fmt.Sprintf("%d", c.organizationID()))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	return respBytes, resp.StatusCode, err
}

func (c *HTTPClient) doJSON(ctx context.Context, method, path string, body any, preferServiceAccount bool) ([]byte, int, error) {
	token, err := c.resolveToken(ctx, preferServiceAccount)
	if err != nil {
		return nil, 0, err
	}
	return c.doJSONRequest(ctx, token, method, path, body)
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
	for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02T15:04:05.999999999", "2006-01-02T15:04:05", "2006-01-02 15:04:05.999999999", "2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Now()
}

func parseOptionalInt(v any) *int {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case float64:
		n := int(t)
		return &n
	case float32:
		n := int(t)
		return &n
	case int:
		n := t
		return &n
	case int64:
		n := int(t)
		return &n
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil
		}
		return &n
	default:
		return nil
	}
}

func parseOptionalInt32(v any) *int32 {
	p := parseOptionalInt(v)
	if p == nil || *p <= 0 {
		return nil
	}
	n := int32(*p)
	return &n
}

func parseFlexBool(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return strings.EqualFold(strings.TrimSpace(t), "true") || strings.TrimSpace(t) == "1"
	case float64:
		return t != 0
	case int:
		return t != 0
	default:
		return false
	}
}

func parseMetadataRaw(v any) json.RawMessage {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(t)
		if s != "" && json.Valid([]byte(s)) {
			return json.RawMessage(s)
		}
	case map[string]any:
		if b, err := json.Marshal(t); err == nil {
			return b
		}
	default:
		if b, err := json.Marshal(t); err == nil && string(b) != "null" {
			return b
		}
	}
	return nil
}

func metadataSyncState(meta json.RawMessage) (hasDocNum bool, syncState string) {
	if len(meta) == 0 {
		return false, ""
	}
	var m map[string]any
	if json.Unmarshal(meta, &m) != nil || m == nil {
		return false, ""
	}
	if docNum, ok := m["sap_doc_num"].(string); ok && strings.TrimSpace(docNum) != "" {
		hasDocNum = true
	} else if parseFlexFloat(m["sap_doc_num"]) > 0 {
		hasDocNum = true
	}
	syncState, _ = m["sap_sync_state"].(string)
	return hasDocNum, syncState
}

func (c *HTTPClient) dispatchStatusByNumber(ctx context.Context) map[string]string {
	if c.sqlite == nil {
		return make(map[string]string)
	}
	statusMap, err := c.sqlite.GetUpstreamDispatchStatusMap(ctx)
	if err != nil {
		return make(map[string]string)
	}
	return statusMap
}

func (c *HTTPClient) approvedEntityNumbers(ctx context.Context, entityType string) map[string]bool {
	approved := make(map[string]bool)
	path := "/api/approvals?domain=sap_invoice&status=approved&limit=100"
	bodyBytes, status, err := c.doJSON(ctx, http.MethodGet, path, nil, false)
	if err != nil || status != http.StatusOK {
		return approved
	}
	type approvalItem struct {
		EntityType string          `json:"entity_type"`
		EntityID   string          `json:"entity_id"`
		Status     string          `json:"status"`
		Payload    json.RawMessage `json:"payload"`
	}
	var items []approvalItem
	if err := json.Unmarshal(bodyBytes, &items); err != nil {
		var env struct {
			Data []approvalItem `json:"data"`
		}
		if err := json.Unmarshal(bodyBytes, &env); err != nil {
			return approved
		}
		items = env.Data
	}
	for _, item := range items {
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

func (c *HTTPClient) parseTransactionRow(row map[string]any, fallbackStoreID int32) (POSTransaction, bool) {
	txNum, _ := row["transaction_number"].(string)
	if txNum == "" || strings.HasPrefix(txNum, "SIM-POS-") {
		return POSTransaction{}, false
	}
	st, _ := row["status"].(string)
	if st != "" && st != "completed" {
		return POSTransaction{}, false
	}
	metaRaw := parseMetadataRaw(row["metadata"])
	hasDocNum, _ := metadataSyncState(metaRaw)
	if hasDocNum {
		return POSTransaction{}, false
	}

	storeID := int32(parseFlexInt64(row["store_id"]))
	if storeID <= 0 {
		storeID = fallbackStoreID
	}
	cashierID := int32(parseFlexInt64(row["cashier_id"]))
	if cashierID <= 0 {
		cashierID = 1
	}
	sessionID := int32(parseFlexInt64(row["cashier_session_id"]))
	if sessionID <= 0 {
		sessionID = 1
	}
	custSAPCode, _ := row["customer_sap_code"].(string)
	total := parseFlexFloat(row["total_amount"])
	amountPaid := parseFlexFloat(row["amount_paid"])

	return POSTransaction{
		ID:                parseFlexInt64(row["id"]),
		StoreID:           storeID,
		CashierID:         cashierID,
		CashierSessionID:  sessionID,
		CustomerID:        parseOptionalInt32(row["customer_id"]),
		CustomerSAPCode:   strings.TrimSpace(custSAPCode),
		TransactionNumber: txNum,
		TransactionDate:   parseFlexTime(row["transaction_date"]),
		Subtotal:          parseFlexFloat(row["subtotal"]),
		DiscountAmount:    parseFlexFloat(row["discount_amount"]),
		TaxAmount:         parseFlexFloat(row["tax_amount"]),
		TotalAmount:       total,
		AmountPaid:        amountPaid,
		Status:            "completed",
		Metadata:          metaRaw,
	}, true
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
		return nil, fmt.Errorf("fetch store %d transactions failed (status %d)", storeID, status)
	}

	var env struct {
		Data struct {
			Data []map[string]any `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &env); err != nil {
		return nil, err
	}

	var out []POSTransaction
	for _, row := range env.Data.Data {
		tx, ok := c.parseTransactionRow(row, storeID)
		if !ok {
			continue
		}
		out = append(out, tx)
	}
	return out, nil
}

func (c *HTTPClient) fetchUnpostedTransactionsEndpoint(ctx context.Context, limit int, approvedOnly bool) ([]POSTransaction, bool, error) {
	if limit <= 0 {
		limit = 50
	}
	storeID := c.cfg.NembusStoreID
	if storeID <= 0 {
		storeID = 1
	}
	path := fmt.Sprintf("/api/pos/stores/%d/transactions/unposted?limit=%d", storeID, limit)
	if approvedOnly {
		path += "&approved_only=true"
	}
	bodyBytes, status, err := c.doJSON(ctx, http.MethodGet, path, nil, false)
	if err != nil {
		return nil, false, err
	}
	if status == http.StatusNotFound {
		return nil, true, nil
	}
	if status != http.StatusOK {
		return nil, false, fmt.Errorf("fetch unposted transactions failed (status %d)", status)
	}
	var env struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &env); err != nil {
		return nil, false, err
	}
	var out []POSTransaction
	for _, row := range env.Data {
		tx, ok := c.parseTransactionRow(row, storeID)
		if !ok {
			continue
		}
		out = append(out, tx)
	}
	return out, false, nil
}

func (c *HTTPClient) FetchApprovedPendingPOSTransactions(ctx context.Context, limit int) ([]POSTransaction, error) {
	rows, notFound, err := c.fetchUnpostedTransactionsEndpoint(ctx, limit, true)
	if err != nil {
		return nil, err
	}
	dispatched := c.dispatchStatusByNumber(ctx)
	if !notFound {
		var list []POSTransaction
		for _, tx := range rows {
			if dispatched[tx.TransactionNumber] == "posted" {
				continue
			}
			tx.Metadata = json.RawMessage(`{"sap_sync_state":"approved_pending_upload"}`)
			list = append(list, tx)
		}
		return list, nil
	}

	all, err := c.fetchStoreTransactionsRaw(ctx, limit)
	if err != nil || len(all) == 0 {
		return nil, err
	}
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
	rows, notFound, err := c.fetchUnpostedTransactionsEndpoint(ctx, limit, false)
	if err != nil {
		return nil, err
	}
	dispatched := c.dispatchStatusByNumber(ctx)

	if !notFound {
		var list []POSTransaction
		for _, tx := range rows {
			st := dispatched[tx.TransactionNumber]
			if st == "posted" || st == "held" || st == "skipped" {
				continue
			}
			list = append(list, tx)
		}
		return list, nil
	}

	all, err := c.fetchStoreTransactionsRaw(ctx, limit)
	if err != nil || len(all) == 0 {
		return nil, err
	}

	var list []POSTransaction
	for _, tx := range all {
		st := dispatched[tx.TransactionNumber]
		if st == "posted" || st == "held" || st == "skipped" {
			continue
		}
		_, syncState := metadataSyncState(tx.Metadata)
		switch syncState {
		case "held_stock", "adjustment_created_pending_force", "approved_pending_upload", "synced":
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
		lineID := parseFlexInt64(row["line_id"])
		productID := parseFlexInt64(row["product_id"])
		qty := parseFlexFloat(row["quantity"])
		if lineID == 0 && productID == 0 && sku == "" && qty == 0 {
			continue
		}
		name, _ := row["product_name"].(string)
		unitPrice := parseFlexFloat(row["unit_price"])
		disc := parseFlexFloat(row["discount_amount_2"])
		lineTotal := parseFlexFloat(row["line_total"])
		if lineTotal <= 0 {
			lineTotal = (qty * unitPrice) - disc
		}
		lineTax := parseFlexFloat(row["line_tax_amount"])
		lineSubtotal := parseFlexFloat(row["line_subtotal"])
		if lineSubtotal <= 0 {
			if lineTax > 0 && lineTotal > lineTax {
				lineSubtotal = lineTotal - lineTax
			} else {
				lineSubtotal = lineTotal
			}
		}
		convFactor := parseFlexFloat(row["uom_conversion_factor"])
		if convFactor <= 0 {
			convFactor = 1.0
		}
		isBaseUoM := true
		if _, hasIsBase := row["is_base_uom"]; hasIsBase {
			isBaseUoM = parseFlexBool(row["is_base_uom"])
		}
		sapTaxCode, _ := row["sap_tax_code"].(string)

		lines = append(lines, POSTransactionLine{
			ID:                  lineID,
			TransactionID:       transactionID,
			ProductID:           int32(productID),
			ProductSKU:          sku,
			ProductName:         name,
			Quantity:            qty,
			UoMID:               parseOptionalInt32(row["uom_id"]),
			UnitPrice:           unitPrice,
			DiscountAmount:      disc,
			TaxAmount:           lineTax,
			Subtotal:            lineSubtotal,
			LineTotal:           lineTotal,
			LineNumber:          idx + 1,
			SAPUoMEntry:         parseOptionalInt(row["sap_uom_entry"]),
			SAPUgpEntry:         parseOptionalInt(row["sap_ugp_entry"]),
			UoMConversionFactor: convFactor,
			IsBaseUoM:           isBaseUoM,
			SAPTaxCode:          strings.TrimSpace(sapTaxCode),
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

func (c *HTTPClient) markEntitySyncedToSAP(ctx context.Context, path string, sapDocEntry, sapDocNum int) error {
	body := map[string]int{
		"sap_doc_entry": sapDocEntry,
		"sap_doc_num":   sapDocNum,
	}
	respBytes, status, err := c.doJSON(ctx, http.MethodPost, path, body, false)
	if err != nil {
		return err
	}
	if status == http.StatusNotFound {
		log.Printf("⚠️ Cloud SAP sync endpoint %s returned 404 (cloud server may need update)", path)
		return nil
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("mark synced %s failed (status %d): %s", path, status, strings.TrimSpace(string(respBytes)))
	}
	return nil
}

func (c *HTTPClient) MarkTransactionPostedToSAP(ctx context.Context, transactionID int64, sapDocEntry, sapDocNum int) error {
	return c.markEntitySyncedToSAP(ctx, fmt.Sprintf("/api/pos/transactions/%d/sap-sync", transactionID), sapDocEntry, sapDocNum)
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
		if meta, ok := env.Data["metadata"].(map[string]any); ok {
			if code, ok := meta["sap_card_code"].(string); ok && strings.TrimSpace(code) != "" {
				return strings.TrimSpace(code), nil
			}
		}
	}
	return c.cfg.NembusDefaultCustomer, nil
}

func (c *HTTPClient) FetchUnpostedSalesReturns(ctx context.Context, limit int) ([]SalesReturn, error) {
	if limit <= 0 {
		limit = 50
	}
	storeID := c.cfg.NembusStoreID
	if storeID <= 0 {
		storeID = 1
	}
	path := fmt.Sprintf("/api/pos/returns/unposted?store_id=%d&limit=%d", storeID, limit)
	bodyBytes, status, err := c.doJSON(ctx, http.MethodGet, path, nil, false)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, nil
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("fetch unposted sales returns failed (status %d)", status)
	}
	var env struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &env); err != nil {
		return nil, err
	}

	dispatched := c.dispatchStatusByNumber(ctx)
	var list []SalesReturn
	for _, row := range env.Data {
		retNum, _ := row["return_number"].(string)
		if retNum == "" {
			continue
		}
		if st := dispatched[retNum]; st == "posted" || st == "skipped" {
			continue
		}
		metaRaw := parseMetadataRaw(row["metadata"])
		hasDocNum, _ := metadataSyncState(metaRaw)
		if hasDocNum {
			continue
		}
		custSAPCode, _ := row["customer_sap_code"].(string)
		var reasonPtr *string
		if reason, ok := row["return_reason"].(string); ok && strings.TrimSpace(reason) != "" {
			reasonPtr = &reason
		}
		list = append(list, SalesReturn{
			ID:                    parseFlexInt64(row["id"]),
			ReturnNumber:          retNum,
			StoreID:               int32(parseFlexInt64(row["store_id"])),
			CashierID:             parseOptionalInt32(row["cashier_id"]),
			CustomerID:            parseOptionalInt32(row["customer_id"]),
			CustomerSAPCode:       strings.TrimSpace(custSAPCode),
			OriginalTransactionID: parseOptionalInt32(row["original_transaction_id"]),
			OriginalSAPDocEntry:   int(parseFlexInt64(row["original_sap_doc_entry"])),
			ReturnDate:            parseFlexTime(row["return_date"]),
			ReturnReason:          reasonPtr,
			Status:                "completed",
			Subtotal:              parseFlexFloat(row["subtotal"]),
			TaxAmount:             parseFlexFloat(row["tax_amount"]),
			TotalRefundAmount:     parseFlexFloat(row["total_refund_amount"]),
			Metadata:              metaRaw,
		})
	}
	return list, nil
}

func (c *HTTPClient) FetchSalesReturnLines(ctx context.Context, returnID int64) ([]SalesReturnLine, error) {
	path := fmt.Sprintf("/api/pos/returns/%d/lines", returnID)
	bodyBytes, status, err := c.doJSON(ctx, http.MethodGet, path, nil, false)
	if err != nil || status != http.StatusOK {
		return nil, fmt.Errorf("fetch sales return %d lines failed (status %d)", returnID, status)
	}
	var env struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &env); err != nil {
		return nil, err
	}
	var lines []SalesReturnLine
	for idx, row := range env.Data {
		sku, _ := row["product_sku"].(string)
		name, _ := row["product_name"].(string)
		convFactor := parseFlexFloat(row["uom_conversion_factor"])
		if convFactor <= 0 {
			convFactor = 1.0
		}
		isBaseUoM := true
		if _, hasIsBase := row["is_base_uom"]; hasIsBase {
			isBaseUoM = parseFlexBool(row["is_base_uom"])
		}
		sapTaxCode, _ := row["sap_tax_code"].(string)
		lineNum := int(parseFlexInt64(row["line_number"]))
		if lineNum <= 0 {
			lineNum = idx + 1
		}
		lines = append(lines, SalesReturnLine{
			ID:                  parseFlexInt64(row["id"]),
			ReturnID:            returnID,
			ProductID:           int32(parseFlexInt64(row["product_id"])),
			ProductSKU:          sku,
			ProductName:         name,
			OriginalLineID:      parseOptionalInt32(row["original_line_id"]),
			Quantity:            parseFlexFloat(row["quantity"]),
			UnitPrice:           parseFlexFloat(row["unit_price"]),
			RefundAmount:        parseFlexFloat(row["refund_amount"]),
			ReturnToStock:       parseFlexBool(row["return_to_stock"]),
			LineNumber:          lineNum,
			SAPUoMEntry:         parseOptionalInt(row["sap_uom_entry"]),
			SAPUgpEntry:         parseOptionalInt(row["sap_ugp_entry"]),
			UoMConversionFactor: convFactor,
			IsBaseUoM:           isBaseUoM,
			SAPTaxCode:          strings.TrimSpace(sapTaxCode),
		})
	}
	return lines, nil
}

func (c *HTTPClient) MarkSalesReturnPostedToSAP(ctx context.Context, returnID int64, sapDocEntry, sapDocNum int) error {
	return c.markEntitySyncedToSAP(ctx, fmt.Sprintf("/api/pos/returns/%d/sap-sync", returnID), sapDocEntry, sapDocNum)
}

func (c *HTTPClient) FetchOriginalSAPDocEntry(ctx context.Context, originalTxID int32) (int, error) {
	path := fmt.Sprintf("/api/pos/transactions/%d", originalTxID)
	bodyBytes, status, err := c.doJSON(ctx, http.MethodGet, path, nil, false)
	if err != nil || status != http.StatusOK {
		return 0, nil
	}
	var env struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &env); err != nil || env.Data == nil {
		return 0, nil
	}
	metaRaw := parseMetadataRaw(env.Data["metadata"])
	if len(metaRaw) == 0 {
		return 0, nil
	}
	var m map[string]any
	if json.Unmarshal(metaRaw, &m) == nil {
		return int(parseFlexInt64(m["sap_doc_entry"])), nil
	}
	return 0, nil
}

// GetApprovalPolicyRules retrieves the JSON rules for an enabled approval policy.
func (c *HTTPClient) GetApprovalPolicyRules(ctx context.Context, domain, action string) (map[string]any, bool, error) {
	path := fmt.Sprintf("/api/approvals/policies/%s?action=%s", url.PathEscape(domain), url.QueryEscape(action))
	bodyBytes, status, err := c.doJSON(ctx, http.MethodGet, path, nil, false)
	if err != nil {
		return nil, false, err
	}
	if status == http.StatusNotFound {
		return nil, false, nil
	}
	if status != http.StatusOK {
		return nil, false, fmt.Errorf("approval policy lookup returned status %d", status)
	}
	var policyResp struct {
		IsEnabled        bool           `json:"is_enabled"`
		RequiresApproval bool           `json:"requires_approval"`
		Rules            map[string]any `json:"rules"`
	}
	if err := json.Unmarshal(bodyBytes, &policyResp); err != nil {
		return nil, false, err
	}
	if !policyResp.IsEnabled || !policyResp.RequiresApproval {
		return nil, false, nil
	}
	return policyResp.Rules, true, nil
}

func (c *HTTPClient) submitStockHoldApproval(ctx context.Context, entityType, entityID, summary string, payload map[string]any) error {
	token, err := c.resolveToken(ctx, true)
	if err != nil {
		return err
	}
	orgID, userID := c.jwtOrgAndUser(token)
	reqBody := map[string]any{
		"organization_id": orgID,
		"domain":          "sap_invoice",
		"action":          "stock_hold",
		"entity_type":     entityType,
		"entity_id":       entityID,
		"requested_by":    userID,
		"summary":         summary,
		"payload":         payload,
	}
	respBytes, status, err := c.doJSONRequest(ctx, token, http.MethodPost, "/api/approvals/submit", reqBody)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("submit stock hold approval failed (status %d): %s", status, strings.TrimSpace(string(respBytes)))
	}
	return nil
}

func (c *HTTPClient) HoldTransactionForStockReview(ctx context.Context, txID int64, txNum string, storeID int32, whCode string, totalAmount float64, shortageItems []ShortageItem, summary string) error {
	payload := map[string]any{
		"transaction_id":     txID,
		"transaction_number": txNum,
		"store_id":           storeID,
		"warehouse_code":     whCode,
		"total_amount":       totalAmount,
		"shortage_items":     shortageItems,
		"reason":             summary,
	}
	return c.submitStockHoldApproval(ctx, "pos_transaction", fmt.Sprintf("%d", txID), summary, payload)
}

func parseSalesOrderRow(row map[string]any) (SalesOrderV2, bool) {
	idStr, _ := row["id"].(string)
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		return SalesOrderV2{}, false
	}
	orderNum, _ := row["order_number"].(string)
	if orderNum == "" || strings.HasPrefix(orderNum, "SO-SAP-") {
		return SalesOrderV2{}, false
	}
	salesChannel, _ := row["sales_channel"].(string)
	if strings.EqualFold(strings.TrimSpace(salesChannel), "pos") {
		return SalesOrderV2{}, false
	}
	orderStatus, _ := row["order_status"].(string)
	fulfillStatus, _ := row["fulfillment_status"].(string)
	if orderStatus != "confirmed" && orderStatus != "fulfilled" && fulfillStatus != "fulfilled" {
		return SalesOrderV2{}, false
	}

	metaRaw := parseMetadataRaw(row["metadata"])
	hasDocNum, _ := metadataSyncState(metaRaw)
	if hasDocNum {
		return SalesOrderV2{}, false
	}

	custSAPCode, _ := row["customer_sap_code"].(string)
	payStatus, _ := row["payment_status"].(string)

	return SalesOrderV2{
		ID:              orderID,
		OrderNumber:     orderNum,
		StoreID:         int32(parseFlexInt64(row["store_id"])),
		CustomerID:      parseOptionalInt32(row["customer_id"]),
		CustomerSAPCode: strings.TrimSpace(custSAPCode),
		OrderDate:       parseFlexTime(row["order_date"]),
		TotalAmount:     parseFlexFloat(row["total_amount"]),
		PaymentStatus:   payStatus,
		Metadata:        metaRaw,
	}, true
}

func (c *HTTPClient) fetchSalesOrdersRaw(ctx context.Context) ([]SalesOrderV2, error) {
	path := fmt.Sprintf("/api/orders?organization_id=%d", c.cfg.NembusOrganizationID)
	bodyBytes, status, err := c.doJSON(ctx, http.MethodGet, path, nil, false)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("fetch orders failed (status %d)", status)
	}
	var env struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &env); err != nil {
		return nil, err
	}

	var list []SalesOrderV2
	for _, row := range env.Data {
		if order, ok := parseSalesOrderRow(row); ok {
			list = append(list, order)
		}
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
	if limit <= 0 {
		limit = 50
	}
	path := fmt.Sprintf("/api/orders/unposted-sap?organization_id=%d&limit=%d", c.cfg.NembusOrganizationID, limit)
	bodyBytes, status, err := c.doJSON(ctx, http.MethodGet, path, nil, false)
	if err != nil {
		return nil, err
	}

	var all []SalesOrderV2
	if status == http.StatusNotFound {
		all, err = c.fetchSalesOrdersRaw(ctx)
		if err != nil || len(all) == 0 {
			return nil, err
		}
	} else if status != http.StatusOK {
		return nil, fmt.Errorf("fetch unposted sales orders failed (status %d)", status)
	} else {
		var env struct {
			Data []map[string]any `json:"data"`
		}
		if err := json.Unmarshal(bodyBytes, &env); err != nil {
			return nil, err
		}
		for _, row := range env.Data {
			if order, ok := parseSalesOrderRow(row); ok {
				all = append(all, order)
			}
		}
	}

	dispatched := c.dispatchStatusByNumber(ctx)

	var list []SalesOrderV2
	for _, o := range all {
		st := dispatched[o.OrderNumber]
		if st == "posted" || st == "held" || st == "skipped" {
			continue
		}
		_, syncState := metadataSyncState(o.Metadata)
		switch syncState {
		case "held_stock", "adjustment_created_pending_force", "approved_pending_upload", "synced":
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
		if sku == "" {
			sku, _ = row["product_sku_current"].(string)
		}
		name, _ := row["product_name"].(string)
		if name == "" {
			name, _ = row["product_name_current"].(string)
		}
		convFactor := parseFlexFloat(row["uom_conversion_factor"])
		if convFactor <= 0 {
			convFactor = 1.0
		}
		isBaseUoM := true
		if _, hasIsBase := row["is_base_uom"]; hasIsBase {
			isBaseUoM = parseFlexBool(row["is_base_uom"])
		}
		sapTaxCode, _ := row["sap_tax_code"].(string)

		lines = append(lines, SalesOrderLineV2{
			ProductID:           int32(parseFlexInt64(row["product_id"])),
			ProductSKU:          sku,
			ProductName:         name,
			Quantity:            parseFlexFloat(row["quantity_ordered"]),
			UoMID:               parseOptionalInt32(row["uom_id"]),
			UnitPrice:           parseFlexFloat(row["unit_price"]),
			TaxAmount:           parseFlexFloat(row["tax_amount"]),
			LineTotal:           parseFlexFloat(row["line_total"]),
			LineNumber:          idx + 1,
			SAPUoMEntry:         parseOptionalInt(row["sap_uom_entry"]),
			SAPUgpEntry:         parseOptionalInt(row["sap_ugp_entry"]),
			UoMConversionFactor: convFactor,
			IsBaseUoM:           isBaseUoM,
			SAPTaxCode:          strings.TrimSpace(sapTaxCode),
		})
	}
	return lines, nil
}

func (c *HTTPClient) MarkSalesOrderPostedToSAP(ctx context.Context, orderID uuid.UUID, sapDocEntry, sapDocNum int) error {
	if err := c.markEntitySyncedToSAP(ctx, fmt.Sprintf("/api/orders/%s/sap-sync", orderID.String()), sapDocEntry, sapDocNum); err != nil {
		return err
	}
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
	payload := map[string]any{
		"sales_order_id": orderID.String(),
		"order_number":   orderNum,
		"store_id":       storeID,
		"warehouse_code": whCode,
		"total_amount":   totalAmount,
		"shortage_items": shortageItems,
		"reason":         summary,
	}
	if err := c.submitStockHoldApproval(ctx, "sales_order_v2", orderID.String(), summary, payload); err != nil {
		return err
	}
	if c.sqlite != nil {
		_ = c.sqlite.RecordUpstreamDispatch(ctx, orderNum, time.Now(), totalAmount, nil, nil, "held", summary)
	}
	return nil
}
