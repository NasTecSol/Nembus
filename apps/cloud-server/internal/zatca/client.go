package zatca

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ZatcaClient handles HTTP calls to ZATCA FATOORA APIs.
type ZatcaClient struct {
	baseURL    string
	httpClient *http.Client
}

// GetBaseURLForEnv returns the appropriate ZATCA API gateway base URL according to the environment.
func GetBaseURLForEnv(env string) string {
	switch env {
	case "production", "prod":
		return "https://gw-fatoora.zatca.gov.sa/e-invoicing/core"
	case "simulation", "sim":
		return "https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation"
	default:
		return "https://gw-fatoora.zatca.gov.sa/e-invoicing/developer-portal"
	}
}

// NewZatcaClient initializes a new ZATCA API client for sandbox, simulation, or production.
func NewZatcaClient(baseURL string) *ZatcaClient {
	if baseURL == "" {
		baseURL = GetBaseURLForEnv("sandbox")
	}
	return &ZatcaClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewZatcaClientForEnv returns an initialized client targeting the specified ZATCA environment.
func NewZatcaClientForEnv(env string) *ZatcaClient {
	return NewZatcaClient(GetBaseURLForEnv(env))
}

// setZatcaHeaders supplies the API version required by the FATOORA gateway.
func setZatcaHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en")
	req.Header.Set("Accept-Version", "V2")
}

// ─── Data Structures ────────────────────────────────────────────────────────

type ComplianceCSIDRequest struct {
	CSR string `json:"csr"`
}

type ComplianceCSIDResponse struct {
	RequestID           json.Number `json:"requestID"`
	IssuedCertificate   string      `json:"issuedCertificate"`
	BinarySecurityToken string      `json:"binarySecurityToken"`
	Secret              string      `json:"secret"`
	Errors              []string    `json:"errors,omitempty"`
	Warnings            []string    `json:"warnings,omitempty"`
}

type ProductionCSIDRequest struct {
	ComplianceRequestID string `json:"compliance_request_id"`
}

type ProductionCSIDResponse struct {
	IssuedCertificate   string   `json:"issuedCertificate"`
	BinarySecurityToken string   `json:"binarySecurityToken"`
	Secret              string   `json:"secret"`
	Errors              []string `json:"errors,omitempty"`
}

type InvoiceSubmissionRequest struct {
	InvoiceHash string `json:"invoiceHash"`
	UUID        string `json:"uuid"`
	Invoice     string `json:"invoice"` // Base64 encoded signed UBL XML
}

type ClearanceResponse struct {
	ClearanceStatus   string   `json:"clearanceStatus"` // "CLEARED" or "NOT_CLEARED"
	ClearedInvoice    string   `json:"clearedInvoice"`  // Base64 encoded signed XML with ZATCA stamp
	QrCode            string   `json:"qrCode"`          // ZATCA signed QR code Base64
	ValidationResults string   `json:"validationResults"`
	Errors            []string `json:"errors,omitempty"`
	Warnings          []string `json:"warnings,omitempty"`
}

type ReportingResponse struct {
	ReportingStatus   string          `json:"reportingStatus"`
	ValidationResults json.RawMessage `json:"validationResults"`
	Errors            json.RawMessage `json:"errors,omitempty"`
	Warnings          json.RawMessage `json:"warnings,omitempty"`
	RawResponse       json.RawMessage `json:"-"`
	HTTPStatus        int             `json:"-"`
}

// ─── API Methods ────────────────────────────────────────────────────────────

// RequestComplianceCSID calls POST /compliance with a 1-hour OTP to obtain a temporary Compliance CSID.
func (c *ZatcaClient) RequestComplianceCSID(ctx context.Context, csrPEM string, otp string) (*ComplianceCSIDResponse, error) {
	csrBase64 := base64.StdEncoding.EncodeToString([]byte(csrPEM))
	reqBody := ComplianceCSIDRequest{CSR: csrBase64}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal compliance request: %w", err)
	}

	url := fmt.Sprintf("%s/compliance", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	setZatcaHeaders(req)
	req.Header.Set("OTP", otp)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("compliance API call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("compliance API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var res ComplianceCSIDResponse
	if err := json.Unmarshal(respBody, &res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &res, nil
}

// SubmitComplianceInvoice submits a test invoice to /compliance/invoices for mandatory compliance testing.
func (c *ZatcaClient) SubmitComplianceInvoice(ctx context.Context, ccsidBinaryToken string, secret string, xmlHash string, docUUID string, signedXML string) (*ReportingResponse, error) {
	reqBody := InvoiceSubmissionRequest{
		InvoiceHash: xmlHash,
		UUID:        docUUID,
		Invoice:     base64.StdEncoding.EncodeToString([]byte(signedXML)),
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/compliance/invoices", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	setZatcaHeaders(req)
	req.SetBasicAuth(ccsidBinaryToken, secret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeInvoiceResponse(resp, "compliance check")
}

// RequestProductionCSID calls POST /production/csids to exchange a passed Compliance CSID for a Production CSID.
func (c *ZatcaClient) RequestProductionCSID(ctx context.Context, ccsidBinaryToken string, secret string, complianceRequestID string) (*ProductionCSIDResponse, error) {
	reqBody := ProductionCSIDRequest{ComplianceRequestID: complianceRequestID}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/production/csids", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	setZatcaHeaders(req)
	req.SetBasicAuth(ccsidBinaryToken, secret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("production CSID API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var res ProductionCSIDResponse
	if err := json.Unmarshal(respBody, &res); err != nil {
		return nil, fmt.Errorf("failed to decode production CSID response: %w", err)
	}

	return &res, nil
}

// ClearStandardInvoice submits a B2B Standard Tax Invoice synchronously to /invoices/clearance/single.
func (c *ZatcaClient) ClearStandardInvoice(ctx context.Context, pcsidBinaryToken string, secret string, xmlHash string, docUUID string, signedXML string) (*ClearanceResponse, error) {
	reqBody := InvoiceSubmissionRequest{
		InvoiceHash: xmlHash,
		UUID:        docUUID,
		Invoice:     base64.StdEncoding.EncodeToString([]byte(signedXML)),
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/invoices/clearance/single", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	setZatcaHeaders(req)
	req.Header.Set("Clearance-Status", "1")
	req.SetBasicAuth(pcsidBinaryToken, secret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var res ClearanceResponse
	if err := json.Unmarshal(respBody, &res); err != nil {
		return nil, fmt.Errorf("failed to decode clearance response: %w", err)
	}

	return &res, nil
}

// ReportSimplifiedInvoice submits a B2C Simplified Tax Invoice asynchronously to /invoices/reporting/single.
func (c *ZatcaClient) ReportSimplifiedInvoice(ctx context.Context, pcsidBinaryToken string, secret string, xmlHash string, docUUID string, signedXML string) (*ReportingResponse, error) {
	reqBody := InvoiceSubmissionRequest{
		InvoiceHash: xmlHash,
		UUID:        docUUID,
		Invoice:     base64.StdEncoding.EncodeToString([]byte(signedXML)),
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/invoices/reporting/single", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	setZatcaHeaders(req)
	req.Header.Set("Clearance-Status", "0")
	req.SetBasicAuth(pcsidBinaryToken, secret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeInvoiceResponse(resp, "reporting")
}

func decodeInvoiceResponse(resp *http.Response, operation string) (*ReportingResponse, error) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s response: %w", operation, err)
	}
	res := ReportingResponse{RawResponse: json.RawMessage(respBody), HTTPStatus: resp.StatusCode}
	// Gateways can return an empty or HTML body for authentication failures.
	// Preserve that response, but report the HTTP failure before decoding JSON.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if !json.Valid(respBody) {
			res.RawResponse, _ = json.Marshal(map[string]any{"http_status": resp.StatusCode, "body": string(respBody)})
		} else {
			_ = json.Unmarshal(respBody, &res)
		}
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			if operation == "reporting" {
				return &res, fmt.Errorf("ZATCA authentication failed (HTTP %d): reporting requires the production CSID and its matching production secret from the same ZATCA environment", resp.StatusCode)
			}
			return &res, fmt.Errorf("ZATCA %s authentication failed (HTTP %d): verify the compliance CSID, matching secret and environment", operation, resp.StatusCode)
		}
		return &res, fmt.Errorf("ZATCA %s API error (HTTP %d)", operation, resp.StatusCode)
	}
	if err := json.Unmarshal(respBody, &res); err != nil {
		res.RawResponse, _ = json.Marshal(map[string]any{"http_status": resp.StatusCode, "body": string(respBody)})
		return &res, fmt.Errorf("failed to decode %s response (HTTP %d): %w", operation, resp.StatusCode, err)
	}

	return &res, nil
}

// RenewProductionCSID calls POST /production/csids/renewal to exchange a current CSID for a renewed Production CSID.
func (c *ZatcaClient) RenewProductionCSID(ctx context.Context, currentPCSID string, secret string, csrPEM string) (*ProductionCSIDResponse, error) {
	csrBase64 := base64.StdEncoding.EncodeToString([]byte(csrPEM))
	reqBody := ComplianceCSIDRequest{CSR: csrBase64}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CSID renewal request: %w", err)
	}

	url := fmt.Sprintf("%s/production/csids/renewal", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create CSID renewal request: %w", err)
	}

	setZatcaHeaders(req)
	req.SetBasicAuth(currentPCSID, secret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CSID renewal API call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("CSID renewal API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var res ProductionCSIDResponse
	if err := json.Unmarshal(respBody, &res); err != nil {
		return nil, fmt.Errorf("failed to decode CSID renewal response: %w", err)
	}

	return &res, nil
}
