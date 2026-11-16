package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	coreusecase "github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-sap/contracts"
	"github.com/NasTecSol/nembus-sap/mappings"
	"github.com/NasTecSol/nembus-sap-agent/internal/config"
)

const (
	sapServiceUsername = "sap_migrator"
	sapServicePassword = "SapAgent@2026"
)

type cloudCatalogProduct struct {
	ProductID  int32               `json:"product_id"`
	SKU        string              `json:"sku"`
	CostPrice  float64             `json:"-"`
	CostLoaded bool                `json:"-"`
	Metadata   map[string]any      `json:"metadata"`
	RawPrices  json.RawMessage     `json:"prices"`
	Prices     []cloudCatalogPrice `json:"-"`
}

type cloudCatalogPrice struct {
	ID            int32
	PriceListID   int32
	PriceListCode string
	UoMID         *int32
	UoMCode       *string
	Price         float64
	IsActive      bool
}

func parseFlexibleFloat(v any) float64 {
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
	case json.Number:
		f, _ := t.Float64()
		return f
	default:
		return 0
	}
}

func parseFlexibleInt32(v any) int32 {
	switch t := v.(type) {
	case float64:
		return int32(t)
	case int:
		return int32(t)
	case int64:
		return int32(t)
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(t), 10, 32)
		return int32(n)
	default:
		return 0
	}
}

func decodeCatalogPrices(raw json.RawMessage) []cloudCatalogPrice {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var items []map[string]any
	if err := json.Unmarshal(raw, &items); err != nil {
		var str string
		if sErr := json.Unmarshal(raw, &str); sErr == nil && str != "" {
			_ = json.Unmarshal([]byte(str), &items)
		}
	}
	var out []cloudCatalogPrice
	for _, m := range items {
		p := cloudCatalogPrice{
			ID:          parseFlexibleInt32(m["id"]),
			PriceListID: parseFlexibleInt32(m["price_list_id"]),
			Price:       parseFlexibleFloat(m["price"]),
			IsActive:    true,
		}
		if code, ok := m["price_list_code"].(string); ok {
			p.PriceListCode = code
		}
		if active, ok := m["is_active"].(bool); ok {
			p.IsActive = active
		}
		if uomID := parseFlexibleInt32(m["uom_id"]); uomID > 0 {
			p.UoMID = &uomID
		}
		if uomCode, ok := m["uom_code"].(string); ok && uomCode != "" {
			p.UoMCode = &uomCode
		}
		out = append(out, p)
	}
	return out
}

// extractLoginToken handles both {"data": "<jwt_string>"} and {"data": {"token": "<jwt_string>"}} responses.
func extractLoginToken(body []byte) string {
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
		var dataStr string
		if err := json.Unmarshal(raw.Data, &dataStr); err == nil && strings.TrimSpace(dataStr) != "" {
			return strings.TrimSpace(dataStr)
		}
		var dataObj struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal(raw.Data, &dataObj); err == nil && strings.TrimSpace(dataObj.Token) != "" {
			return strings.TrimSpace(dataObj.Token)
		}
	}
	return ""
}

func (c *CloudClient) tryLogin(ctx context.Context, baseURL, tenantID, userLogin, password string) (string, int, string, error) {
	loginURL := fmt.Sprintf("%s/api/auth/login", baseURL)
	bodyBytes, _ := json.Marshal(map[string]string{
		"user_login": userLogin,
		"password":   password,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tenant-id", tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, "", err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		if tok := extractLoginToken(respBytes); tok != "" {
			return tok, resp.StatusCode, string(respBytes), nil
		}
	}
	return "", resp.StatusCode, string(respBytes), nil
}

// provisionSAPServiceUser ensures the dedicated sap_migrator service user exists in the tenant's
// users table via the /api/v1/migration/batch endpoint so the agent can always mint a valid JWT
// and submit approval requests under a non-admin identity (allowing admin to approve them).
func (c *CloudClient) provisionSAPServiceUser(ctx context.Context, orgID int) error {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(sapServicePassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	payload := &contracts.MigrationBatchPayload{
		BatchID:        uuid.New().String(),
		RunID:          "downstream-auth",
		OrganizationID: orgID,
		Domain:         contracts.DomainUsers,
		SequenceNumber: 1,
		IsLastBatch:    true,
		Timestamp:      time.Now(),
		Users: []mappings.CanonicalUser{
			{
				Username:     sapServiceUsername,
				Email:        "sap_migrator@nembus.local",
				FirstName:    "SAP",
				LastName:     "Sync Agent",
				EmployeeCode: "SAP-AGENT",
				PasswordHash: string(hashBytes),
				IsActive:     true,
				Metadata: map[string]interface{}{
					"system_service_account": true,
				},
			},
		},
	}
	_, err = c.SendBatchWithRetry(ctx, payload)
	return err
}

// acquireCloudJWT obtains a Bearer JWT for calling /api/approvals and /api/products endpoints.
// It prefers the dedicated sap_migrator service account (provisioning it automatically if needed)
// so that submitted approval requests can be approved by admin in the Nembus UI.
func (c *CloudClient) acquireCloudJWT(ctx context.Context) (string, string, int, error) {
	appCfg := config.Get()
	baseURL := strings.TrimRight(c.BaseURL(), "/")
	if baseURL == "" {
		baseURL = strings.TrimRight(appCfg.Cloud.BaseURL, "/")
	}
	orgID := appCfg.Cloud.OrganizationID
	if orgID <= 0 {
		orgID = 1
	}
	tenantID := strings.TrimSpace(appCfg.TenantSlug)
	if tenantID == "" {
		tenantID = fmt.Sprintf("%d", orgID)
	}

	// 1. Try the dedicated sap_migrator service user first
	if tok, _, _, err := c.tryLogin(ctx, baseURL, tenantID, sapServiceUsername, sapServicePassword); err == nil && tok != "" {
		return tok, tenantID, orgID, nil
	}

	// 2. Provision sap_migrator via DomainUsers batch and log in
	if err := c.provisionSAPServiceUser(ctx, orgID); err == nil {
		if tok, _, _, lErr := c.tryLogin(ctx, baseURL, tenantID, sapServiceUsername, sapServicePassword); lErr == nil && tok != "" {
			return tok, tenantID, orgID, nil
		}
	}

	// 3. Fallback to configured migrator_user / migrator_password
	userLogin := strings.TrimSpace(appCfg.MigratorUser)
	if userLogin == "" {
		userLogin = "admin"
	}
	password := appCfg.MigratorPassword
	if password == "" {
		password = "Migrator"
	}
	tok, statusCode, respBody, err := c.tryLogin(ctx, baseURL, tenantID, userLogin, password)
	if err != nil {
		return "", tenantID, orgID, err
	}
	if tok != "" {
		return tok, tenantID, orgID, nil
	}

	if strings.HasPrefix(appCfg.Cloud.APIKey, "eyJ") {
		return appCfg.Cloud.APIKey, tenantID, orgID, nil
	}
	return "", tenantID, orgID, fmt.Errorf("cloud login failed (status %d): %s", statusCode, respBody)
}

func (c *CloudClient) fetchCatalogPricePolicy(ctx context.Context, baseURL, jwt, tenantID string) (bool, coreusecase.CatalogPriceRules, error) {
	policyURL := fmt.Sprintf("%s/api/approvals/policies/catalog?action=price_change", baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, policyURL, nil)
	if err != nil {
		return false, coreusecase.CatalogPriceRules{}, err
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("x-tenant-id", tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, coreusecase.CatalogPriceRules{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, coreusecase.CatalogPriceRules{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, coreusecase.CatalogPriceRules{}, fmt.Errorf("approval policy lookup returned status %d", resp.StatusCode)
	}

	var policyResp struct {
		IsEnabled        bool           `json:"is_enabled"`
		RequiresApproval bool           `json:"requires_approval"`
		Rules            map[string]any `json:"rules"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&policyResp); err != nil {
		return false, coreusecase.CatalogPriceRules{}, err
	}
	if !policyResp.IsEnabled || !policyResp.RequiresApproval {
		return false, coreusecase.CatalogPriceRules{}, nil
	}
	return true, coreusecase.ParseCatalogPriceRules(policyResp.Rules), nil
}

func (c *CloudClient) fetchProductCostPrice(ctx context.Context, baseURL, jwt, tenantID string, productID int32) float64 {
	costURL := fmt.Sprintf("%s/api/product-prices/product/%d/with-pricing", baseURL, productID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, costURL, nil)
	if err != nil {
		return 0
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("x-tenant-id", tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0
	}
	var envelope struct {
		Data struct {
			CostPrice any `json:"cost_price"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return 0
	}
	return parseFlexibleFloat(envelope.Data.CostPrice)
}

func (c *CloudClient) lookupCloudProductBySKU(ctx context.Context, baseURL, jwt, tenantID string, orgID int, sku string) (*cloudCatalogProduct, error) {
	searchURL := fmt.Sprintf("%s/api/products/master-catalog/search?organization_id=%d&q=%s&limit=20",
		baseURL, orgID, url.QueryEscape(sku))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("x-tenant-id", tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("product search returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var envelope struct {
		Data struct {
			Data []cloudCatalogProduct `json:"data"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}
	for i := range envelope.Data.Data {
		if strings.EqualFold(strings.TrimSpace(envelope.Data.Data[i].SKU), strings.TrimSpace(sku)) {
			prod := &envelope.Data.Data[i]
			prod.Prices = decodeCatalogPrices(prod.RawPrices)
			return prod, nil
		}
	}
	return nil, nil
}

func (c *CloudClient) submitCatalogApprovalRequest(ctx context.Context, baseURL, jwt, tenantID string, orgID int, entityID, summary string, payload coreusecase.CatalogPriceChangePayload) error {
	// Check if there is already a pending approval request for this entityID so we can replace it cleanly
	listURL := fmt.Sprintf("%s/api/approvals?domain=catalog&status=pending&limit=100", baseURL)
	if listReq, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil); err == nil {
		listReq.Header.Set("Authorization", "Bearer "+jwt)
		listReq.Header.Set("x-tenant-id", tenantID)
		if listResp, err := c.httpClient.Do(listReq); err == nil {
			var pending []struct {
				ID       int64  `json:"id"`
				EntityID string `json:"entity_id"`
			}
			_ = json.NewDecoder(listResp.Body).Decode(&pending)
			listResp.Body.Close()
			for _, p := range pending {
				if p.EntityID == entityID && p.ID > 0 {
					// Cancel or delete the previous pending request from sap_migrator before submitting the updated one
					cancelURL := fmt.Sprintf("%s/api/approvals/%d/cancel", baseURL, p.ID)
					if cReq, cErr := http.NewRequestWithContext(ctx, http.MethodPost, cancelURL, nil); cErr == nil {
						cReq.Header.Set("Authorization", "Bearer "+jwt)
						cReq.Header.Set("x-tenant-id", tenantID)
						if cResp, doErr := c.httpClient.Do(cReq); doErr == nil {
							cResp.Body.Close()
						}
					}
				}
			}
		}
	}

	rawPayload, _ := json.Marshal(payload)
	reqBody, _ := json.Marshal(map[string]any{
		"organization_id": orgID,
		"domain":          "catalog",
		"action":          "price_change",
		"entity_type":     "product",
		"entity_id":       entityID,
		"requested_by":    1,
		"summary":         summary,
		"payload":         json.RawMessage(rawPayload),
	})

	submitURL := fmt.Sprintf("%s/api/approvals/submit", baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, submitURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("x-tenant-id", tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("submit approval returned status %d: %s", resp.StatusCode, string(msg))
	}
	return nil
}

// FilterDownstreamPriceItems evaluates changed SAP price items against the cloud server's
// catalog.price_change approval policy. Items that require approval are submitted to
// /api/approvals/submit and withheld from the direct migration batch so the active price in
// Nembus remains unchanged until approved.
func (c *CloudClient) FilterDownstreamPriceItems(ctx context.Context, items []mappings.CanonicalPriceListItem) (directItems []mappings.CanonicalPriceListItem, heldNotices []string) {
	if len(items) == 0 {
		return items, nil
	}

	jwt, tenantID, orgID, err := c.acquireCloudJWT(ctx)
	if err != nil || jwt == "" {
		heldNotices = append(heldNotices, fmt.Sprintf("held %d price change(s): could not authenticate with cloud for approval check: %v", len(items), err))
		return nil, heldNotices
	}

	baseURL := strings.TrimRight(c.BaseURL(), "/")
	if baseURL == "" {
		baseURL = strings.TrimRight(config.Get().Cloud.BaseURL, "/")
	}

	requiresApproval, rules, err := c.fetchCatalogPricePolicy(ctx, baseURL, jwt, tenantID)
	if err != nil || !requiresApproval {
		return items, nil
	}

	productCache := make(map[string]*cloudCatalogProduct)

	for _, item := range items {
		if item.Price <= 0 {
			continue
		}
		skuKey := strings.ToUpper(strings.TrimSpace(item.ProductSKU))
		prod, cached := productCache[skuKey]
		if !cached {
			var lookupErr error
			prod, lookupErr = c.lookupCloudProductBySKU(ctx, baseURL, jwt, tenantID, orgID, item.ProductSKU)
			if lookupErr != nil {
				heldNotices = append(heldNotices, fmt.Sprintf("SKU %s lookup error: %v", item.ProductSKU, lookupErr))
				continue
			}
			productCache[skuKey] = prod
		}

		// Brand-new product not yet in Nembus: apply initial price directly
		if prod == nil || prod.ProductID <= 0 {
			directItems = append(directItems, item)
			continue
		}

		var matchedPrice *cloudCatalogPrice
		for i := range prod.Prices {
			p := &prod.Prices[i]
			if !p.IsActive {
				continue
			}
			if !strings.EqualFold(strings.TrimSpace(p.PriceListCode), strings.TrimSpace(item.PriceListCode)) {
				continue
			}
			itemUoM := strings.TrimSpace(item.UOMCode)
			priceUoM := ""
			if p.UoMCode != nil {
				priceUoM = strings.TrimSpace(*p.UoMCode)
			}
			if itemUoM == "" || priceUoM == "" || strings.EqualFold(itemUoM, priceUoM) {
				matchedPrice = p
				break
			}
		}

		// Fallback if there is only 1 active price on the product in Nembus (e.g. default RETAIL_SAR / PL-1)
		if matchedPrice == nil && len(prod.Prices) == 1 && prod.Prices[0].IsActive {
			matchedPrice = &prod.Prices[0]
		}

		// No active price for this price list yet: apply initial price directly
		if matchedPrice == nil || matchedPrice.Price <= 0 {
			directItems = append(directItems, item)
			continue
		}

		oldPrice := matchedPrice.Price
		if math.Abs(item.Price-oldPrice) <= 0.0001 {
			// Price is unchanged in Nembus; skip unnecessary write
			continue
		}

		var costPrice float64
		if rules.BelowCostRequiresApproval {
			if !prod.CostLoaded {
				prod.CostPrice = c.fetchProductCostPrice(ctx, baseURL, jwt, tenantID, prod.ProductID)
				prod.CostLoaded = true
			}
			costPrice = prod.CostPrice
		}

		hold, reason := coreusecase.EvaluateSellPriceChange(requiresApproval, rules, oldPrice, item.Price, costPrice, true)
		if !hold {
			directItems = append(directItems, item)
			continue
		}

		oldP := oldPrice
		newP := item.Price
		var oldCostPtr *float64
		if costPrice > 0 {
			cVal := costPrice
			oldCostPtr = &cVal
		}
		entityID := fmt.Sprintf("%d:pl:%d", prod.ProductID, matchedPrice.PriceListID)
		if matchedPrice.UoMID != nil {
			entityID = fmt.Sprintf("%d:pl:%d:uom:%d", prod.ProductID, matchedPrice.PriceListID, *matchedPrice.UoMID)
		}
		summary := fmt.Sprintf("Price revision for SKU %s (%s): %s", item.ProductSKU, item.PriceListCode, reason)

		err := c.submitCatalogApprovalRequest(ctx, baseURL, jwt, tenantID, orgID, entityID, summary, coreusecase.CatalogPriceChangePayload{
			ProductID:    prod.ProductID,
			ProductSKU:   item.ProductSKU,
			PriceListID:  matchedPrice.PriceListID,
			UoMID:        matchedPrice.UoMID,
			OldPrice:     &oldP,
			NewPrice:     &newP,
			OldCostPrice: oldCostPtr,
			Reason:       reason,
		})
		if err != nil {
			// If approval submission failed, still hold back direct overwrite and log the error
			heldNotices = append(heldNotices, fmt.Sprintf("SKU %s (%s) price %.2f -> %.2f held (approval submit error: %v)", item.ProductSKU, item.PriceListCode, oldPrice, item.Price, err))
			continue
		}

		heldNotices = append(heldNotices, fmt.Sprintf("SKU %s (%s): held active price %.2f and created approval request for new price %.2f", item.ProductSKU, item.PriceListCode, oldPrice, item.Price))
	}

	return directItems, heldNotices
}
