package utils

import (
	"time"

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
