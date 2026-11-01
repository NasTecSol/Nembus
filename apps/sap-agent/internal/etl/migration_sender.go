package etl

import (
	"context"

	"github.com/NasTecSol/nembus-sap/contracts"
)

// MigrationSender posts canonical migration batches to Nembus cloud (gRPC in production).
type MigrationSender interface {
	SendBatchWithRetry(ctx context.Context, payload *contracts.MigrationBatchPayload) (*contracts.MigrationBatchResponse, error)
}
