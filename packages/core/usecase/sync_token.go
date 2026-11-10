package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GenerateSyncToken returns a 32-byte random token (64 hex chars) and its SHA-256 hex digest.
func GenerateSyncToken() (rawToken string, tokenHash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate random sync token: %w", err)
	}
	rawToken = hex.EncodeToString(buf)
	tokenHash = HashSyncToken(rawToken)
	return rawToken, tokenHash, nil
}

// HashSyncToken returns the lowercase SHA-256 hex digest of rawToken.
func HashSyncToken(rawToken string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(rawToken)))
	return hex.EncodeToString(sum[:])
}

// IssueTerminalSyncToken rotates the per-terminal gRPC sync token on pos_terminals and
// returns the raw token (shown only once to the caller provisioning the POS).
func IssueTerminalSyncToken(ctx context.Context, pool *pgxpool.Pool, terminalID int32) (string, error) {
	if pool == nil {
		return "", fmt.Errorf("database pool is nil")
	}
	if terminalID <= 0 {
		return "", fmt.Errorf("terminal_id must be positive")
	}
	var exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pos_terminals WHERE id = $1)`, terminalID).Scan(&exists); err != nil {
		return "", fmt.Errorf("lookup terminal: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("pos_terminal %d not found", terminalID)
	}

	rawToken, tokenHash, err := GenerateSyncToken()
	if err != nil {
		return "", err
	}
	tag, err := pool.Exec(ctx, `
		UPDATE pos_terminals
		SET sync_token_hash = $1,
		    token_issued_at = NOW(),
		    updated_at      = NOW()
		WHERE id = $2
	`, tokenHash, terminalID)
	if err != nil {
		return "", fmt.Errorf("save terminal sync token hash: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return "", fmt.Errorf("pos_terminal %d not found", terminalID)
	}
	return rawToken, nil
}
