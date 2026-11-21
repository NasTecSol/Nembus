package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// DerefInt32 returns the value of an int32 pointer or 0 if it's nil.
func DerefInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

// DerefString returns the value of a string pointer or empty string if it's nil.
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// DerefBool returns the value of a bool pointer or false if it's nil.
func DerefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// FormatTimestamp formats a pgtype.Timestamp to RFC3339 string.
func FormatTimestamp(t interface{}) string {
	switch v := t.(type) {
	case pgtype.Timestamp:
		if !v.Valid {
			return ""
		}
		return v.Time.Format(time.RFC3339)
	case pgtype.Timestamptz:
		if !v.Valid {
			return ""
		}
		return v.Time.Format(time.RFC3339)
	default:
		return ""
	}
}

// FormatDate formats a pgtype.Date to YYYY-MM-DD string.
func FormatDate(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format("2006-01-02")
}

// Int32ToPgInt4 converts int32 pointer to pgtype.Int4.
func Int32ToPgInt4(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

// StringToPgText converts string pointer to pgtype.Text.
func StringToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// TimeToPgTimestamp converts time.Time pointer to pgtype.Timestamp.
func TimeToPgTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: true}
}

// TimePtrToPgTimestamp converts time.Time pointer to pgtype.Timestamp.
func TimePtrToPgTimestamp(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: *t, Valid: true}
}

// TimeToPgDate converts time.Time pointer to pgtype.Date.
func TimeToPgDate(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

// Float64ToPgNumeric converts float64 to pgtype.Numeric.
func Float64ToPgNumeric(v float64) pgtype.Numeric {
	var num pgtype.Numeric
	_ = num.Scan(fmt.Sprintf("%f", v))
	return num
}

// Float64PointerToPgNumeric converts *float64 to pgtype.Numeric.
func Float64PointerToPgNumeric(v *float64) pgtype.Numeric {
	if v == nil {
		return pgtype.Numeric{Valid: false}
	}
	var num pgtype.Numeric
	_ = num.Scan(fmt.Sprintf("%f", *v))
	return num
}

// NumericToString converts pgtype.Numeric to string.
func NumericToString(n pgtype.Numeric) string {
	if !n.Valid {
		return "0.00"
	}
	v, err := n.Value()
	if err != nil || v == nil {
		return "0.00"
	}
	return fmt.Sprintf("%v", v)
}

// Int4ToPtr converts pgtype.Int4 to *int32.
func Int4ToPtr(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	return &i.Int32
}

// TextToPtr converts pgtype.Text to *string.
func TextToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// BoolToPtr converts pgtype.Bool to *bool.
func BoolToPtr(b pgtype.Bool) *bool {
	if !b.Valid {
		return nil
	}
	return &b.Bool
}

// UUIDToPgUUID converts a uuid.UUID to pgtype.UUID (invalid if uuid.Nil).
func UUIDToPgUUID(u uuid.UUID) pgtype.UUID {
	if u == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: u, Valid: true}
}

// UUIDPtrToPgUUID converts a *uuid.UUID to pgtype.UUID.
func UUIDPtrToPgUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil || *u == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *u, Valid: true}
}

// PgUUIDToUUIDPtr converts a pgtype.UUID to *uuid.UUID.
func PgUUIDToUUIDPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	val := uuid.UUID(u.Bytes)
	return &val
}

// PgUUIDToUUID converts a pgtype.UUID to uuid.UUID (returns uuid.Nil if invalid).
func PgUUIDToUUID(u pgtype.UUID) uuid.UUID {
	if !u.Valid {
		return uuid.Nil
	}
	return uuid.UUID(u.Bytes)
}

// PgUUIDToStringPtr converts a pgtype.UUID to *string.
func PgUUIDToStringPtr(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := uuid.UUID(u.Bytes).String()
	return &s
}

// DerefUUID returns the value of a uuid.UUID pointer or uuid.Nil if nil.
func DerefUUID(u *uuid.UUID) uuid.UUID {
	if u == nil {
		return uuid.Nil
	}
	return *u
}

// ParseUUIDOrNil parses a string as uuid.UUID, returning uuid.Nil on error or empty string.
func ParseUUIDOrNil(s string) uuid.UUID {
	s = strings.TrimSpace(s)
	if s == "" {
		return uuid.Nil
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return u
}

// ParseUUIDPtr parses a string pointer as *uuid.UUID, returning nil if empty or invalid.
func ParseUUIDPtr(s *string) *uuid.UUID {
	if s == nil {
		return nil
	}
	u := ParseUUIDOrNil(*s)
	if u == uuid.Nil {
		return nil
	}
	return &u
}

// DetUUIDFromLegacyID matches pg_temp.det_uuid in the M1 UUID migration SQL.
func DetUUIDFromLegacyID(table string, oldID int64) uuid.UUID {
	sum := md5.Sum([]byte(fmt.Sprintf("nembus:v1:%s:%d", table, oldID)))
	h := hex.EncodeToString(sum[:])
	variant := "89ab"[int(sum[8]&3)]
	s := fmt.Sprintf("%s-%s-4%s-%c%s-%s", h[0:8], h[8:12], h[13:16], variant, h[17:20], h[20:32])
	return uuid.MustParse(s)
}

// TruncateRunes returns s truncated to at most max runes without splitting a multibyte rune.
func TruncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	n := 0
	for i := range s {
		if n == max {
			return s[:i]
		}
		n++
	}
	return s
}

// SanitizeTerminalCodeForDocNum normalizes a terminal code for document numbers (max 12 runes).
func SanitizeTerminalCodeForDocNum(terminalCode string) string {
	trimmed := strings.TrimSpace(terminalCode)
	if trimmed == "" {
		return "T00"
	}
	var b strings.Builder
	for _, r := range trimmed {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	res := b.String()
	if res == "" {
		return "T00"
	}
	return TruncateRunes(res, 12)
}



