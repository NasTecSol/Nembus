package telemetry

import (
	"testing"

	otellog "go.opentelemetry.io/otel/log"
)

func TestClassifyLogSeverity(t *testing.T) {
	tests := []struct {
		msg      string
		wantSev  otellog.Severity
		wantText string
	}{
		{"❌ Failed to post sales return RET-1", otellog.SeverityError, "ERROR"},
		{"❌ Quarantined transaction TXN-1", otellog.SeverityError, "ERROR"},
		{"⚠️ [SalesReturn RET-1] Original transaction ID 10 has no SAP DocEntry", otellog.SeverityWarn, "WARN"},
		{"[WARN] slow query detected", otellog.SeverityWarn, "WARN"},
		{"[gRPC SYNC] Cycle finished in 120ms", otellog.SeverityInfo, "INFO"},
	}

	for _, tc := range tests {
		sev, text := classifyLogSeverity(tc.msg)
		if sev != tc.wantSev || text != tc.wantText {
			t.Errorf("classifyLogSeverity(%q) = (%v, %q), want (%v, %q)", tc.msg, sev, text, tc.wantSev, tc.wantText)
		}
	}
}
