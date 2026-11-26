package sap

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/config"
	_ "github.com/microsoft/go-mssqldb"
)

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
	baseURL    string
	sessionID  string
	routeID    string
	sessionMu  sync.RWMutex
	lastLogin  time.Time
	sqlDB      *sql.DB
}

type LoginRequest struct {
	CompanyDB string `json:"CompanyDB"`
	UserName  string `json:"UserName"`
	Password  string `json:"Password"`
}

type LoginResponse struct {
	SessionID string `json:"SessionId"`
	Version   string `json:"Version"`
	Timeout   int    `json:"SessionTimeout"`
}

type ErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message struct {
			Lang  string `json:"lang"`
			Value string `json:"value"`
		} `json:"message"`
	} `json:"error"`
}

func NewClient(cfg *config.Config) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cookie jar: %w", err)
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   4 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.SAPInsecureSkipSSL,
		},
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	httpClient := &http.Client{
		Jar:       jar,
		Timeout:   15 * time.Second,
		Transport: transport,
	}

	client := &Client{
		cfg:        cfg,
		httpClient: httpClient,
		baseURL:    cfg.SAPServiceLayerURL,
	}

	if cfg.SAPMSSQLEnabled {
		connStr := cfg.MSSQLConnectionString()
		db, err := sql.Open("sqlserver", connStr)
		if err != nil {
			log.Printf("⚠️ Warning: Failed to open MSSQL direct connection: %v", err)
		} else {
			db.SetMaxOpenConns(5)
			db.SetMaxIdleConns(2)
			db.SetConnMaxLifetime(5 * time.Minute)
			client.sqlDB = db
		}
	}

	return client, nil
}

func (c *Client) EnsureLogin(ctx context.Context) error {
	c.sessionMu.RLock()
	if c.sessionID != "" && time.Since(c.lastLogin) < 25*time.Minute {
		c.sessionMu.RUnlock()
		return nil
	}
	c.sessionMu.RUnlock()

	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()

	if c.sessionID != "" && time.Since(c.lastLogin) < 25*time.Minute {
		return nil
	}

	loginPayload := LoginRequest{
		CompanyDB: c.cfg.SAPCompanyDB,
		UserName:  c.cfg.SAPUsername,
		Password:  c.cfg.SAPPassword,
	}

	body, err := json.Marshal(loginPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal login payload: %w", err)
	}

	loginURL := fmt.Sprintf("%s/Login", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send login request to SAP Service Layer: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message.Value != "" {
			return fmt.Errorf("SAP login failed (%d): %s [code %d]", resp.StatusCode, errResp.Error.Message.Value, errResp.Error.Code)
		}
		return fmt.Errorf("SAP login failed with HTTP status %d: %s", resp.StatusCode, string(respBody))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(respBody, &loginResp); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	c.sessionID = loginResp.SessionID
	c.lastLogin = time.Now()

	// Extract ROUTEID cookie if present
	if parsedURL, err := url.Parse(loginURL); err == nil {
		for _, cookie := range c.httpClient.Jar.Cookies(parsedURL) {
			if cookie.Name == "ROUTEID" {
				c.routeID = cookie.Value
			}
		}
	}

	log.Printf(" Connected to SAP Service Layer (Company: %s, Version: %s, Session: %s...)",
		c.cfg.SAPCompanyDB, loginResp.Version, c.sessionID[:min(8, len(c.sessionID))])
	return nil
}

func (c *Client) DoRequest(ctx context.Context, method, endpoint string, body any, result any) error {
	if err := c.EnsureLogin(ctx); err != nil {
		return err
	}

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to serialize request body: %w", err)
		}
		reqBody = bytes.NewBuffer(data)
	}

	reqURL := fmt.Sprintf("%s/%s", c.baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request to SAP failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read SAP response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBytes, &errResp); err == nil && errResp.Error.Message.Value != "" {
			return &APIError{
				StatusCode: resp.StatusCode,
				Code:       errResp.Error.Code,
				Message:    errResp.Error.Message.Value,
			}
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBytes),
		}
	}

	if result != nil && len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, result); err != nil {
			return fmt.Errorf("failed to deserialize response: %w", err)
		}
	}

	return nil
}

func (c *Client) TestConnection(ctx context.Context) error {
	if err := c.EnsureLogin(ctx); err != nil {
		return fmt.Errorf("Service Layer check failed: %w", err)
	}

	if c.sqlDB != nil {
		if err := c.sqlDB.PingContext(ctx); err != nil {
			return fmt.Errorf("MSSQL Direct check failed: %w", err)
		}
	}

	return nil
}

func (c *Client) GetSQLDB() *sql.DB {
	return c.sqlDB
}

// Close releases the underlying MSSQL connection pool if open.
func (c *Client) Close() error {
	if c.sqlDB != nil {
		return c.sqlDB.Close()
	}
	return nil
}

// ResolveWarehouseBinAllocations resolves SAP bin-location allocations for an item
// in a bin-enabled warehouse (OWHS.BinActivat = 'Y'). When isReturn is true (Credit Memo),
// stock is received back into the warehouse default bin (or primary item bin).
func (c *Client) ResolveWarehouseBinAllocations(ctx context.Context, itemCode, warehouseCode string, requiredQty float64, isReturn bool) ([]SAPBinAllocation, error) {
	if c.sqlDB == nil || requiredQty <= 0 || itemCode == "" || warehouseCode == "" {
		return nil, nil
	}

	var binActivat string
	var dftBinAbs int
	err := c.sqlDB.QueryRowContext(ctx,
		"SELECT ISNULL(BinActivat, 'N'), ISNULL(DftBinAbs, 0) FROM OWHS WITH (NOLOCK) WHERE WhsCode = @p1",
		warehouseCode,
	).Scan(&binActivat, &dftBinAbs)
	if err != nil {
		return nil, nil
	}
	if binActivat != "Y" {
		return nil, nil
	}

	if isReturn {
		targetBin := dftBinAbs
		if targetBin <= 0 {
			_ = c.sqlDB.QueryRowContext(ctx,
				"SELECT TOP 1 BinAbs FROM OIBQ WITH (NOLOCK) WHERE ItemCode = @p1 AND WhsCode = @p2 ORDER BY OnHandQty DESC, BinAbs ASC",
				itemCode, warehouseCode,
			).Scan(&targetBin)
		}
		if targetBin <= 0 {
			_ = c.sqlDB.QueryRowContext(ctx,
				"SELECT TOP 1 AbsEntry FROM OBIN WITH (NOLOCK) WHERE WhsCode = @p1 AND ISNULL(Disabled, 'N') <> 'Y' ORDER BY AbsEntry ASC",
				warehouseCode,
			).Scan(&targetBin)
		}
		if targetBin > 0 {
			return []SAPBinAllocation{{BinAbsEntry: targetBin, Quantity: requiredQty}}, nil
		}
		return nil, nil
	}

	rows, err := c.sqlDB.QueryContext(ctx,
		`SELECT BinAbs, ISNULL(OnHandQty, 0)
		 FROM OIBQ WITH (NOLOCK)
		 WHERE ItemCode = @p1 AND WhsCode = @p2 AND ISNULL(OnHandQty, 0) > 0
		 ORDER BY CASE WHEN BinAbs = @p3 THEN 0 ELSE 1 END, OnHandQty DESC, BinAbs ASC`,
		itemCode, warehouseCode, dftBinAbs,
	)
	if err != nil {
		if dftBinAbs > 0 {
			return []SAPBinAllocation{{BinAbsEntry: dftBinAbs, Quantity: requiredQty}}, nil
		}
		return nil, nil
	}
	defer rows.Close()

	var allocs []SAPBinAllocation
	remaining := requiredQty
	for rows.Next() && remaining > 1e-9 {
		var binAbs int
		var onHand float64
		if scanErr := rows.Scan(&binAbs, &onHand); scanErr != nil || binAbs <= 0 || onHand <= 0 {
			continue
		}
		take := remaining
		if onHand < take {
			take = onHand
		}
		allocs = append(allocs, SAPBinAllocation{
			BinAbsEntry: binAbs,
			Quantity:    take,
		})
		remaining -= take
	}

	if remaining > 1e-9 && dftBinAbs > 0 {
		merged := false
		for i := range allocs {
			if allocs[i].BinAbsEntry == dftBinAbs {
				allocs[i].Quantity += remaining
				merged = true
				break
			}
		}
		if !merged {
			allocs = append(allocs, SAPBinAllocation{
				BinAbsEntry: dftBinAbs,
				Quantity:    remaining,
			})
		}
	}

	return allocs, nil
}

// ResolveDocumentSeries resolves the active SAP numbering series (NNM1.Series) for
// an ObjectCode ("13"=Invoice, "14"=Credit Memo, "24"=Incoming Payment) and posting date.
// Explicit config overrides take precedence over MSSQL NNM1 lookup.
func (c *Client) ResolveDocumentSeries(ctx context.Context, objectCode string, docDate string, bplID int) (*int, error) {
	if c.cfg != nil {
		switch objectCode {
		case "13":
			if c.cfg.SAPInvoiceSeries > 0 {
				s := c.cfg.SAPInvoiceSeries
				return &s, nil
			}
		case "14":
			if c.cfg.SAPCreditMemoSeries > 0 {
				s := c.cfg.SAPCreditMemoSeries
				return &s, nil
			}
		case "24":
			if c.cfg.SAPPaymentSeries > 0 {
				s := c.cfg.SAPPaymentSeries
				return &s, nil
			}
		}
	}

	if c.sqlDB == nil || objectCode == "" {
		return nil, nil
	}

	var series int
	// 1. Prefer an unlocked, automatic NNM1 series whose posting period (OFPR) covers docDate
	if docDate != "" {
		err := c.sqlDB.QueryRowContext(ctx, `
			SELECT TOP 1 n.Series
			FROM NNM1 n WITH (NOLOCK)
			INNER JOIN OFPR p WITH (NOLOCK) ON p.Indicator = n.Indicator
			WHERE n.ObjectCode = @p1
			  AND ISNULL(n.Locked, 'N') <> 'Y'
			  AND ISNULL(n.IsManual, 'N') <> 'Y'
			  AND (@p3 <= 0 OR ISNULL(n.BPLId, 0) = @p3 OR n.BPLId IS NULL)
			  AND p.F_RefDate <= CAST(@p2 AS DATE)
			  AND p.T_RefDate >= CAST(@p2 AS DATE)
			ORDER BY CASE WHEN n.BPLId = @p3 THEN 0 ELSE 1 END, n.Series ASC
		`, objectCode, docDate, bplID).Scan(&series)
		if err == nil && series > 0 {
			return &series, nil
		}
	}

	// 2. Fallback: only accept a period-agnostic ('Default' or empty Indicator) unlocked
	// automatic series — never pick a series bound to a different fiscal period, which
	// causes SAP error -5002 ("Series period does not match current period").
	err := c.sqlDB.QueryRowContext(ctx, `
		SELECT TOP 1 n.Series
		FROM NNM1 n WITH (NOLOCK)
		WHERE n.ObjectCode = @p1
		  AND ISNULL(n.Locked, 'N') <> 'Y'
		  AND ISNULL(n.IsManual, 'N') <> 'Y'
		  AND (n.Indicator IS NULL OR n.Indicator = '' OR n.Indicator = 'Default')
		  AND (@p2 <= 0 OR ISNULL(n.BPLId, 0) = @p2 OR n.BPLId IS NULL)
		ORDER BY CASE WHEN n.BPLId = @p2 THEN 0 ELSE 1 END, n.Series ASC
	`, objectCode, bplID).Scan(&series)
	if err == nil && series > 0 {
		return &series, nil
	}

	if docDate != "" {
		log.Printf("⚠️ [SAP-Series] No unlocked NNM1 numbering series found for ObjectCode=%s covering DocDate=%s (BPLId=%d); check SAP Administration -> System Initialization -> Document Numbering", objectCode, docDate, bplID)
	}
	return nil, nil
}

// ResolveInvoiceBaseLine returns the 0-based INV1.LineNum for an item on a base A/R Invoice.
func (c *Client) ResolveInvoiceBaseLine(ctx context.Context, baseDocEntry int, itemCode string) (*int, error) {
	if c.sqlDB == nil || baseDocEntry <= 0 || itemCode == "" {
		return nil, nil
	}
	var lineNum int
	err := c.sqlDB.QueryRowContext(ctx,
		"SELECT TOP 1 LineNum FROM INV1 WITH (NOLOCK) WHERE DocEntry = @p1 AND ItemCode = @p2 ORDER BY LineNum ASC",
		baseDocEntry, itemCode,
	).Scan(&lineNum)
	if err != nil {
		return nil, nil
	}
	return &lineNum, nil
}

// CheckItemWarehouseStock retrieves the current OnHand stock for an item in a specific warehouse.
func (c *Client) CheckItemWarehouseStock(ctx context.Context, itemCode, warehouseCode string) (float64, error) {
	if c.sqlDB != nil {
		var onHand float64
		err := c.sqlDB.QueryRowContext(ctx, "SELECT ISNULL(OnHand, 0) FROM OITW WHERE ItemCode = @p1 AND WhsCode = @p2", itemCode, warehouseCode).Scan(&onHand)
		if err == nil {
			return onHand, nil
		}
	}
	// Fallback to Service Layer Items lookup
	endpoint := fmt.Sprintf("Items('%s')?$select=ItemWarehouseInfoCollection", url.PathEscape(itemCode))
	var res struct {
		ItemWarehouseInfoCollection []struct {
			WarehouseCode string  `json:"WarehouseCode"`
			InStock       float64 `json:"InStock"`
		} `json:"ItemWarehouseInfoCollection"`
	}
	if err := c.DoRequest(ctx, http.MethodGet, endpoint, nil, &res); err != nil {
		return 0, err
	}
	for _, wh := range res.ItemWarehouseInfoCollection {
		if wh.WarehouseCode == warehouseCode {
			return wh.InStock, nil
		}
	}
	return 0, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
