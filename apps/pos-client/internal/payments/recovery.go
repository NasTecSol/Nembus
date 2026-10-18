package payments

import (
	"context"
	"fmt"
	"log"
)

// StartupRecovery scans for payment attempts left in 'processing' status (e.g. from app crash or power loss)
// and transitions them to 'unknown' so that they are never lost or silently assumed failed.
func StartupRecovery(ctx context.Context, store Store) (int64, error) {
	count, err := store.MarkOrphanProcessingAsUnknown(ctx)
	if err != nil {
		return 0, fmt.Errorf("startup payment recovery failed: %w", err)
	}

	if count > 0 {
		log.Printf("⚠️  [PAYMENT RECOVERY] Transitioned %d in-flight 'processing' payment attempts to 'unknown'.", count)
		log.Println("⚠️  [PAYMENT RECOVERY] Cashier action required: verify physical terminal roll or portal before retrying.")
	}

	return count, nil
}
