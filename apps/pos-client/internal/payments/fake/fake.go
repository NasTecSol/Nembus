package fake

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/NasTecSol/nembus-client/internal/payments"
	corepayments "github.com/NasTecSol/nembus-core/payments"
)

// FakeTerminalProvider implements payments.TerminalProvider for unit and integration testing.
type FakeTerminalProvider struct {
	mu           sync.Mutex
	name         string
	delay        time.Duration
	nextResult   payments.SaleResult
	nextError    error
	recordedReqs []payments.SaleRequest
}

// NewFakeTerminalProvider creates a fake terminal provider.
func NewFakeTerminalProvider(name string) *FakeTerminalProvider {
	if name == "" {
		name = "fake-terminal"
	}
	return &FakeTerminalProvider{
		name: name,
		nextResult: payments.SaleResult{
			Status: corepayments.StatusSucceeded,
			ProviderRefs: map[string]string{
				"rrn":       "123456789012",
				"auth_code": "OK1234",
				"scheme":    "mada",
			},
			MaskedPAN:   strPtr("411111******1111"),
			RawRedacted: json.RawMessage(`{"response_code":"00","terminal_id":"T1001"}`),
		},
	}
}

func (f *FakeTerminalProvider) Name() string {
	return f.name
}

func (f *FakeTerminalProvider) SetDelay(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.delay = d
}

func (f *FakeTerminalProvider) SetNextResult(res payments.SaleResult, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextResult = res
	f.nextError = err
}

func (f *FakeTerminalProvider) RecordedRequests() []payments.SaleRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make([]payments.SaleRequest, len(f.recordedReqs))
	copy(cp, f.recordedReqs)
	return cp
}

func (f *FakeTerminalProvider) Sale(ctx context.Context, req payments.SaleRequest) (payments.SaleResult, error) {
	f.mu.Lock()
	delay := f.delay
	res := f.nextResult
	retErr := f.nextError
	f.recordedReqs = append(f.recordedReqs, req)
	f.mu.Unlock()

	if delay > 0 {
		select {
		case <-ctx.Done():
			return payments.SaleResult{Status: corepayments.StatusUnknown}, ctx.Err()
		case <-time.After(delay):
		}
	}

	return res, retErr
}

func strPtr(s string) *string {
	return &s
}
