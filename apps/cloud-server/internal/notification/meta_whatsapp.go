package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type WhatsAppConfig struct {
	PhoneNumberID string `json:"phone_number_id"`
	AccessToken   string `json:"access_token"`
	ApiVersion    string `json:"api_version"` // Defaults to "v19.0"
}

type WhatsAppClient struct {
	cfg        WhatsAppConfig
	httpClient *http.Client
}

func NewWhatsAppClient(cfg WhatsAppConfig) *WhatsAppClient {
	if cfg.ApiVersion == "" {
		cfg.ApiVersion = "v19.0"
	}
	return &WhatsAppClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type WhatsAppResponse struct {
	MessagingProduct string `json:"messaging_product"`
	Contacts         []struct {
		Input string `json:"input"`
		WaID  string `json:"wa_id"`
	} `json:"contacts"`
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
	Error *struct {
		Message   string `json:"message"`
		Type      string `json:"type"`
		Code      int    `json:"code"`
		FbtraceID string `json:"fbtrace_id"`
	} `json:"error,omitempty"`
}

// SendTextMessage sends a direct text message via Meta WhatsApp Cloud API
func (c *WhatsAppClient) SendTextMessage(ctx context.Context, recipientPhone, message string) (*WhatsAppResponse, error) {
	cleanPhone := cleanPhoneNumber(recipientPhone)
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", c.cfg.ApiVersion, c.cfg.PhoneNumberID)

	reqBody := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                cleanPhone,
		"type":              "text",
		"text": map[string]string{
			"preview_url": "true",
			"body":        message,
		},
	}

	return c.postMessage(ctx, url, reqBody)
}

// SendTemplateMessage sends an approved WhatsApp Template message with dynamic parameters
func (c *WhatsAppClient) SendTemplateMessage(ctx context.Context, recipientPhone, templateName, languageCode string, parameters []string) (*WhatsAppResponse, error) {
	cleanPhone := cleanPhoneNumber(recipientPhone)
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", c.cfg.ApiVersion, c.cfg.PhoneNumberID)

	var bodyParams []map[string]string
	for _, param := range parameters {
		bodyParams = append(bodyParams, map[string]string{
			"type": "text",
			"text": param,
		})
	}

	components := []map[string]interface{}{}
	if len(bodyParams) > 0 {
		components = append(components, map[string]interface{}{
			"type":       "body",
			"parameters": bodyParams,
		})
	}

	reqBody := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                cleanPhone,
		"type":              "template",
		"template": map[string]interface{}{
			"name": templateName,
			"language": map[string]string{
				"code": languageCode,
			},
			"components": components,
		},
	}

	return c.postMessage(ctx, url, reqBody)
}

// SendDocumentMessage sends a PDF document (e.g. PO or E-Receipt) with an optional caption
func (c *WhatsAppClient) SendDocumentMessage(ctx context.Context, recipientPhone, documentURL, filename, caption string) (*WhatsAppResponse, error) {
	cleanPhone := cleanPhoneNumber(recipientPhone)
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", c.cfg.ApiVersion, c.cfg.PhoneNumberID)

	reqBody := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                cleanPhone,
		"type":              "document",
		"document": map[string]string{
			"link":     documentURL,
			"caption":  caption,
			"filename": filename,
		},
	}

	return c.postMessage(ctx, url, reqBody)
}

func (c *WhatsAppClient) postMessage(ctx context.Context, url string, reqBody interface{}) (*WhatsAppResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("meta whatsapp api request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var waResp WhatsAppResponse
	if err := json.Unmarshal(respBytes, &waResp); err != nil {
		return nil, fmt.Errorf("unmarshal response (%s): %w", string(respBytes), err)
	}

	if waResp.Error != nil {
		return &waResp, fmt.Errorf("meta whatsapp error (%d): %s", waResp.Error.Code, waResp.Error.Message)
	}

	return &waResp, nil
}

func cleanPhoneNumber(phone string) string {
	cleaned := strings.TrimSpace(phone)
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.TrimPrefix(cleaned, "+")
	return cleaned
}
