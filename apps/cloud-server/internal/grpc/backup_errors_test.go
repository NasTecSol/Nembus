package grpc

import (
	"strings"
	"testing"
)

func TestBackupFailureHint(t *testing.T) {
	tests := []struct{ name, stderr, want string }{
		{"version", "pg_dump: error: aborting because of server version mismatch", "PostgreSQL version mismatch"},
		{"password", "connection to server failed: FATAL: password authentication failed for user secret_user", "authentication failed"},
		{"missing password", "fe_sendauth: no password supplied", "authentication failed"},
		{"access rules", "FATAL: no pg_hba.conf entry for host secret_host", "authentication failed"},
		{"permissions", "ERROR: permission denied for table secret_table", "permissions are insufficient"},
		{"row security", "ERROR: query would be affected by row-level security policy", "permissions are insufficient"},
		{"connection", "connection to server at secret_host failed: Connection refused", "connection failed"},
		{"dns", "could not translate host name secret_host to address", "connection failed"},
		{"missing database", "connection to server at secret_host failed: FATAL: database secret_database does not exist", "database was not found"},
		{"disk", "No space left on device", "insufficient disk space"},
		{"unknown", "unexpected failure postgres://secret_user:secret_password@secret_host/secret_database", "administrator must inspect"},
		{"empty", "", "administrator must inspect"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := backupFailureHint(tt.stderr)
			if !strings.Contains(got, tt.want) {
				t.Fatalf("got %q; want hint containing %q", got, tt.want)
			}
			if strings.Contains(got, "secret_") {
				t.Fatalf("diagnostic exposed connection details: %q", got)
			}
		})
	}
}
