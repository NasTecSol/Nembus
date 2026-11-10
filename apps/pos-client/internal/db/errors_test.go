package db

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestWithStorageHint(t *testing.T) {
	for _, original := range []error{
		&pgconn.PgError{Code: "53100"},
		&os.PathError{Op: "write", Path: "backup.sql", Err: syscall.ENOSPC},
		errors.New("There is not enough space on the disk."),
	} {
		err := WithStorageHint(fmt.Errorf("clone failed: %w", original))
		if !strings.Contains(err.Error(), "free space") || !errors.Is(err, original) {
			t.Fatalf("expected actionable hint preserving the original error, got %v", err)
		}
	}
	other := &pgconn.PgError{Code: "42P01"}
	if WithStorageHint(other) != other || WithStorageHint(nil) != nil {
		t.Fatal("non-storage errors must remain unchanged")
	}
}
