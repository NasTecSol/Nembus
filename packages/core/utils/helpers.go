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
	_ = num.Scan(time.Duration(0).String()) // initialize
	_ = num.Scan(v)
	return num
}

// Float64PointerToPgNumeric converts *float64 to pgtype.Numeric.
func Float64PointerToPgNumeric(v *float64) pgtype.Numeric {
	if v == nil {
		return pgtype.Numeric{Valid: false}
	}
	var num pgtype.Numeric
	_ = num.Scan(*v)
	return num
}

