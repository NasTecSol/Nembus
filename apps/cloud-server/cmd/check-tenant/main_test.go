package main

import (
	"strings"
	"testing"
)

func TestSanitizeConnectionStringURL(t *testing.T) {
	const dsn = "postgres://testuser:testpassword@db.example:5432/testdb"

	sanitized := sanitizeConnectionString(dsn)
	if sanitized != "host=db.example port=5432 credentials=<redacted>" {
		t.Fatalf("unexpected sanitized DSN: %s", sanitized)
	}
	for _, secret := range []string{"testuser", "testpassword", dsn} {
		if strings.Contains(sanitized, secret) {
			t.Fatalf("sanitized DSN contains sensitive input")
		}
	}
}

func TestSanitizeConnectionStringEscapedPassword(t *testing.T) {
	const dsn = "postgresql://testuser:test%40password%3F@db.example:5432/testdb"

	sanitized := sanitizeConnectionString(dsn)
	if strings.Contains(sanitized, "testuser") || strings.Contains(sanitized, "test%40password") || strings.Contains(sanitized, "test@password") {
		t.Fatalf("sanitized DSN contains escaped credential material")
	}
}

func TestSanitizeConnectionStringOmitsQueryParameters(t *testing.T) {
	dsn := "postgres://testuser:testpassword@db.example:5432/testdb?sslmode=disable&application_name=testapp&token=testtoken"

	sanitized := sanitizeConnectionString(dsn)
	if strings.Contains(sanitized, "sslmode") || strings.Contains(sanitized, "application_name") || strings.Contains(sanitized, "testtoken") {
		t.Fatalf("sanitized DSN contains query parameters")
	}
}

func TestSanitizeConnectionStringFailsClosed(t *testing.T) {
	for _, dsn := range []string{
		"",
		"not a DSN",
		"host=db.example port=5432 user=testuser password=testpassword dbname=testdb",
		"postgres://testuser:testpassword@:5432/testdb",
	} {
		if sanitized := sanitizeConnectionString(dsn); sanitized != redactedConnectionString {
			t.Fatalf("expected fail-closed redaction")
		}
	}
}
