package middleware

import (
	"strings"
	"testing"
)

func TestRedactSensitiveQuery(t *testing.T) {
	got := redactSensitiveQuery("token=eyJhbGciOiJIUzI1NiJ9.secret&tenant_id=ebt&limit=50")
	if strings.Contains(got, "eyJhbGci") {
		t.Fatalf("expected JWT token to be redacted, got %q", got)
	}
	if !strings.Contains(got, "token=REDACTED") {
		t.Fatalf("expected token=REDACTED in %q", got)
	}
	if !strings.Contains(got, "tenant_id=ebt") || !strings.Contains(got, "limit=50") {
		t.Fatalf("expected safe query params preserved, got %q", got)
	}

	unchanged := redactSensitiveQuery("tenant_id=ebt&approved_only=true")
	if unchanged != "tenant_id=ebt&approved_only=true" {
		t.Fatalf("expected safe query unchanged, got %q", unchanged)
	}
}
