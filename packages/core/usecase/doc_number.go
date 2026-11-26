package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
)

var weakLegacyDocNumRe = regexp.MustCompile(`^(?:SES-\d{8}-\d{1,4}|GUEST-\d+|POS-\d+|RET-\d+|ORD-\d+|WEB-\d+)$`)

// LookupTerminalCode resolves the terminal_code from pos_terminals when posTerminalID > 0,
// falling back to fallbackCode (or "T00" if empty).
func LookupTerminalCode(ctx context.Context, repo *repository.Queries, posTerminalID int32, fallbackCode string) string {
	if trimmed := strings.TrimSpace(fallbackCode); trimmed != "" {
		return SanitizeTerminalCode(trimmed)
	}
	if repo != nil && posTerminalID > 0 {
		if term, err := repo.GetPOSTerminal(ctx, posTerminalID); err == nil && strings.TrimSpace(term.TerminalCode) != "" {
			return SanitizeTerminalCode(term.TerminalCode)
		}
		return fmt.Sprintf("T%02d", posTerminalID)
	}
	return "T00"
}

// IsWeakOrEmptyDocNumber returns true when docNum is blank or matches a legacy
// non-terminal-scoped timestamp/3-digit-random format (e.g. "SES-20261121-123", "GUEST-1700000000").
func IsWeakOrEmptyDocNumber(docNum string) bool {
	trimmed := strings.TrimSpace(docNum)
	if trimmed == "" {
		return true
	}
	return weakLegacyDocNumRe.MatchString(strings.ToUpper(trimmed))
}


// SanitizeTerminalCode normalizes a terminal code for inclusion in offline-generated
// document numbers (e.g. "T01", "POS1"). If empty or containing no alphanumeric characters,
// it falls back to "T00".
func SanitizeTerminalCode(terminalCode string) string {
	return utils.SanitizeTerminalCodeForDocNum(terminalCode)
}

// GenerateTerminalDocNumber produces a collision-resistant, terminal-scoped business
// document number in the format: <PREFIX>-<TERMINAL>-<YYYYMMDD>-<6HEX>
// Example: POS-T01-20261121-A3F9B2
func GenerateTerminalDocNumber(prefix, terminalCode string) string {
	cleanPrefix := strings.ToUpper(strings.TrimSpace(prefix))
	if cleanPrefix == "" {
		cleanPrefix = "DOC"
	}
	cleanTerminal := SanitizeTerminalCode(terminalCode)
	datePart := time.Now().UTC().Format("20060102")

	var rnd [4]byte
	if _, err := rand.Read(rnd[:]); err != nil {
		ns := uint64(time.Now().UnixNano())
		rnd[0] = byte(ns)
		rnd[1] = byte(ns >> 8)
		rnd[2] = byte(ns >> 16)
		rnd[3] = byte(ns >> 24)
	}
	hexPart := strings.ToUpper(hex.EncodeToString(rnd[:]))

	return fmt.Sprintf("%s-%s-%s-%s", cleanPrefix, cleanTerminal, datePart, hexPart)
}
