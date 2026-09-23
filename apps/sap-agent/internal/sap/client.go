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
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.SAPInsecureSkipSSL,
		},
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
	}

	httpClient := &http.Client{
		Jar:       jar,
		Timeout:   60 * time.Second,
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
			return fmt.Errorf("SAP API error (%d): %s [code %d]", resp.StatusCode, errResp.Error.Message.Value, errResp.Error.Code)
		}
		return fmt.Errorf("SAP request returned HTTP %d: %s", resp.StatusCode, string(respBytes))
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
