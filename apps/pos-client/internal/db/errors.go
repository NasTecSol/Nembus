package db

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
)

// WithStorageHint keeps the original error available for inspection while
// explaining how to recover from disk exhaustion during a clone or migration.
func WithStorageHint(err error) error {
	if err == nil {
		return nil
	}
	var state interface{ SQLState() string }
	if (errors.As(err, &state) && state.SQLState() == "53100") ||
		errors.Is(err, syscall.ENOSPC) ||
		strings.Contains(strings.ToLower(err.Error()), "no space left on device") ||
		strings.Contains(strings.ToLower(err.Error()), "not enough space on the disk") {
		return fmt.Errorf("insufficient disk space: free space on the drive containing .nembus/data and .nembus/backups, then retry the tenant clone: %w", err)
	}
	return err
}
