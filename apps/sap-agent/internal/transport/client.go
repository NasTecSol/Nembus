package transport

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap/contracts"
)

const (
	maxRetries    = 3
	retryBaseWait = time.Second
)

// CloudClient handles all outbound HTTP communication with the Nembus cloud server.
type CloudClient struct {
	cfg        config.CloudConfig
	httpClient *http.Client
}

func NewCloudClient(cfg config.CloudConfig) *CloudClient {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	return &CloudClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// isRetryableStatus returns true for HTTP status codes that indicate a
// transient failure which may succeed on a subsequent attempt.
func isRetryableStatus(code int) bool {
	return code == 429 || code == 500 || code == 502 || code == 503 || code == 504
}

// SendBatch compresses and posts a migration batch payload to the cloud server.
// Deprecated: prefer SendBatchWithRetry for production migrations.
func (c *CloudClient) SendBatch(ctx context.Context, payload *contracts.MigrationBatchPayload) (*contracts.MigrationBatchResponse, error) {
	return c.sendOnce(ctx, payload, uuid.New().String())
}

// SendBatchWithRetry posts a migration batch with automatic retry on transient
// failures. It makes up to maxRetries attempts with exponential backoff
// (1s → 2s → 4s). 4xx client errors are never retried.
func (c *CloudClient) SendBatchWithRetry(ctx context.Context, payload *contracts.MigrationBatchPayload) (*contracts.MigrationBatchResponse, error) {
	requestID := uuid.New().String()
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			wait := retryBaseWait * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}

		resp, retryable, err := c.sendOnceClassified(ctx, payload, requestID)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !retryable {
			return nil, lastErr
		}
		// retryable error — loop to next attempt
	}

	return nil, fmt.Errorf("batch send failed after %d attempts: %w", maxRetries, lastErr)
}

// sendOnce performs a single HTTP POST attempt and returns the parsed response.
func (c *CloudClient) sendOnce(ctx context.Context, payload *contracts.MigrationBatchPayload, requestID string) (*contracts.MigrationBatchResponse, error) {
	resp, _, err := c.sendOnceClassified(ctx, payload, requestID)
	return resp, err
}

// sendOnceClassified performs a single HTTP POST attempt.
// Returns (response, isRetryable, error).
func (c *CloudClient) sendOnceClassified(ctx context.Context, payload *contracts.MigrationBatchPayload, requestID string) (*contracts.MigrationBatchResponse, bool, error) {
	if strings.TrimSpace(c.cfg.M2MToken) == "" {
		return nil, false, errors.New("cloud migration requires a tenant/org-bound M2M token")
	}
	if strings.TrimSpace(c.cfg.TenantSlug) == "" {
		return nil, false, errors.New("cloud migration requires an explicit tenant slug")
	}
	if c.cfg.OrganizationID <= 0 {
		return nil, false, errors.New("cloud migration requires an explicit positive organization ID")
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, false, fmt.Errorf("failed to marshal batch payload: %w", err)
	}

	// Gzip compress the JSON payload
	var compressed bytes.Buffer
	gz := gzip.NewWriter(&compressed)
	if _, err := gz.Write(jsonData); err != nil {
		return nil, false, fmt.Errorf("failed to gzip batch payload: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, false, fmt.Errorf("failed to finalize gzip stream: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/migration/batch", c.cfg.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &compressed)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("X-Request-ID", requestID)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.cfg.M2MToken))
	req.Header.Set("x-tenant-id", c.cfg.TenantSlug)
	req.Header.Set("x-organization-id", strconv.Itoa(c.cfg.OrganizationID))

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		// Network/timeout errors are always retryable
		return nil, true, fmt.Errorf("failed to dispatch request to cloud server %s: %w", url, err)
	}
	defer httpResp.Body.Close()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, true, fmt.Errorf("failed to read cloud response: %w", err)
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		retryable := isRetryableStatus(httpResp.StatusCode)
		return nil, retryable, fmt.Errorf("cloud server returned error status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}

	var batchResp contracts.MigrationBatchResponse
	if err := json.Unmarshal(bodyBytes, &batchResp); err != nil {
		return nil, false, fmt.Errorf("failed to decode cloud response JSON: %w (raw: %s)", err, string(bodyBytes))
	}

	return &batchResp, false, nil
}

// PingCloud tests connectivity and authentication against the cloud server.
func (c *CloudClient) PingCloud(ctx context.Context) (bool, string, error) {
	url := fmt.Sprintf("%s/health", c.cfg.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("cloud endpoint unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Sprintf("Status %d", resp.StatusCode), nil
	}
	return true, "Connected", nil
}
