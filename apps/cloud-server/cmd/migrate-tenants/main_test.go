package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestDSNToURL(t *testing.T) {
	tests := []struct {
		name         string
		connStr      string
		hostOverride string
		want         string
		wantErr      bool
	}{
		{
			name:         "keyword DSN with docker host override",
			connStr:      "host=postgres user=nembus_admin_user password=secret dbname=qitaf",
			hostOverride: "postgres=localhost",
			want:         "postgres://nembus_admin_user:secret@localhost:5432/qitaf?sslmode=disable",
		},
		{
			name:         "keyword DSN carrying explicit sslmode",
			connStr:      "host=postgres user=u dbname=d sslmode=require",
			hostOverride: "postgres=localhost",
			want:         "postgres://u:@localhost:5432/d?sslmode=require",
		},
		{
			name:         "URL input unchanged when host does not match override",
			connStr:      "postgres://u:p@localhost:5432/db?sslmode=disable",
			hostOverride: "postgres=localhost",
			want:         "postgres://u:p@localhost:5432/db?sslmode=disable",
		},
		{
			name:         "URL input without sslmode stays unmodified",
			connStr:      "postgres://u:p@10.0.0.5:5432/db",
			hostOverride: "postgres=localhost",
			want:         "postgres://u:p@10.0.0.5:5432/db",
		},
		{
			name:         "override disabled",
			connStr:      "host=postgres user=u dbname=d",
			hostOverride: "",
			want:         "postgres://u:@postgres:5432/d",
		},
		{
			name:         "unix socket rejected",
			connStr:      "host=/var/run/postgresql user=u dbname=d",
			hostOverride: "",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dsnToURL(tt.connStr, tt.hostOverride)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("dsnToURL() expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("dsnToURL() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("dsnToURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFirstMigrationVersion(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{
		"20260813124500.sql",
		"20260818093727_remove_customer_bp.sql",
		"atlas.sum",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("-- x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := firstMigrationVersion(dir)
	if err != nil {
		t.Fatalf("firstMigrationVersion() error: %v", err)
	}
	if got != "20260813124500" {
		t.Errorf("firstMigrationVersion() = %q, want %q", got, "20260813124500")
	}

	empty := t.TempDir()
	if _, err := firstMigrationVersion(empty); err == nil {
		t.Error("firstMigrationVersion() expected error on empty dir")
	}
}

func TestIsMissingDatabase(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "3D000", Message: `database "qitaf3" does not exist`}
	if !isMissingDatabase(pgErr) {
		t.Error("isMissingDatabase() should detect raw PgError 3D000")
	}
	if !isMissingDatabase(fmt.Errorf("check atlas_schema_revisions: %w", pgErr)) {
		t.Error("isMissingDatabase() should detect wrapped PgError")
	}
	if !isMissingDatabase(errors.New(`failed to connect: FATAL: database "qitaf3" does not exist (SQLSTATE 3D000)`)) {
		t.Error("isMissingDatabase() should detect plain string form")
	}
	if isMissingDatabase(errors.New("connection refused")) {
		t.Error("isMissingDatabase() must not match connection refused")
	}
	if isMissingDatabase(errors.New(`relation "foo" does not exist`)) {
		t.Error("isMissingDatabase() must not match relation-not-exists")
	}
}
