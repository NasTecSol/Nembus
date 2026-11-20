package pipeline

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/internal/sap"
)

func TestUpstreamPermanentErrorQuarantine(t *testing.T) {
	u := &UpstreamSync{failedRetries: make(map[string]retryEntry)}
	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	permErr := fmt.Errorf("failed to post invoice to SAP: %w", &sap.APIError{
		StatusCode: http.StatusBadRequest,
		Code:       -10,
		Message:    "Fully allocate item INV00028",
	})

	// Attempt 1 -> 1m backoff, not quarantined
	if q := u.recordRetryFailure("tx:1", now, permErr); q {
		t.Fatalf("attempt 1 should not quarantine")
	}
	skip, nextAt, quarantined := u.shouldBackoff("tx:1", now.Add(30*time.Second))
	if !skip || quarantined || !nextAt.Equal(now.Add(time.Minute)) {
		t.Fatalf("attempt 1 unexpected backoff: skip=%v nextAt=%v quarantined=%v", skip, nextAt, quarantined)
	}

	// Attempt 2 -> 5m backoff, not quarantined
	t2 := now.Add(time.Minute)
	if q := u.recordRetryFailure("tx:1", t2, permErr); q {
		t.Fatalf("attempt 2 should not quarantine")
	}
	skip, nextAt, quarantined = u.shouldBackoff("tx:1", t2.Add(time.Minute))
	if !skip || quarantined || !nextAt.Equal(t2.Add(5*time.Minute)) {
		t.Fatalf("attempt 2 unexpected backoff: skip=%v nextAt=%v quarantined=%v", skip, nextAt, quarantined)
	}

	// Attempt 3 -> quarantined for 6 hours
	t3 := t2.Add(5 * time.Minute)
	if q := u.recordRetryFailure("tx:1", t3, permErr); !q {
		t.Fatalf("attempt 3 should quarantine permanent SAP error")
	}
	skip, nextAt, quarantined = u.shouldBackoff("tx:1", t3.Add(time.Hour))
	if !skip || !quarantined || !nextAt.Equal(t3.Add(permanentQuarantineDelay)) {
		t.Fatalf("attempt 3 unexpected quarantine state: skip=%v nextAt=%v quarantined=%v", skip, nextAt, quarantined)
	}
}

func TestUpstreamTransientErrorDoesNotQuarantine(t *testing.T) {
	u := &UpstreamSync{failedRetries: make(map[string]retryEntry)}
	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	transientErr := errors.New("connection reset by peer")

	for i := 0; i < 6; i++ {
		if q := u.recordRetryFailure("tx:2", now, transientErr); q {
			t.Fatalf("transient error must never quarantine (attempt %d)", i+1)
		}
	}
	skip, nextAt, quarantined := u.shouldBackoff("tx:2", now.Add(time.Minute))
	if !skip || quarantined || !nextAt.Equal(now.Add(15*time.Minute)) {
		t.Fatalf("transient error unexpected backoff: skip=%v nextAt=%v quarantined=%v", skip, nextAt, quarantined)
	}
}

func TestUpstreamTransientErrorsDoNotConsumePermanentQuota(t *testing.T) {
	u := &UpstreamSync{failedRetries: make(map[string]retryEntry)}
	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	transientErr := errors.New("dial tcp: connection refused")
	permErr := &sap.APIError{
		StatusCode: http.StatusBadRequest,
		Code:       -10,
		Message:    "Fully allocate item INV00028",
	}

	// Two transient failures should not count toward permanent quarantine threshold.
	if q := u.recordRetryFailure("tx:3", now, transientErr); q {
		t.Fatalf("transient attempt 1 should not quarantine")
	}
	if q := u.recordRetryFailure("tx:3", now.Add(time.Minute), transientErr); q {
		t.Fatalf("transient attempt 2 should not quarantine")
	}

	// Permanent attempt 1 -> 1m backoff, not quarantined
	t3 := now.Add(3 * time.Minute)
	if q := u.recordRetryFailure("tx:3", t3, permErr); q {
		t.Fatalf("first permanent error after transient errors should not quarantine")
	}
	skip, nextAt, quarantined := u.shouldBackoff("tx:3", t3.Add(30*time.Second))
	if !skip || quarantined || !nextAt.Equal(t3.Add(time.Minute)) {
		t.Fatalf("permanent attempt 1 unexpected backoff: skip=%v nextAt=%v quarantined=%v", skip, nextAt, quarantined)
	}

	// Permanent attempt 2 -> 5m backoff, not quarantined
	t4 := t3.Add(time.Minute)
	if q := u.recordRetryFailure("tx:3", t4, permErr); q {
		t.Fatalf("second permanent error should not quarantine")
	}

	// Permanent attempt 3 -> quarantined for 6 hours
	t5 := t4.Add(5 * time.Minute)
	if q := u.recordRetryFailure("tx:3", t5, permErr); !q {
		t.Fatalf("third permanent error should quarantine")
	}
}
