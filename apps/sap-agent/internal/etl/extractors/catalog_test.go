package extractors

import (
	"testing"
	"time"
)

func TestSAPDeltaParams(t *testing.T) {
	since := time.Date(2026, 11, 11, 12, 54, 27, 750000000, time.Local)
	dateStr, timeInt, dateTimeStr := sapDeltaParams(since)

	if dateStr != "2026-11-11" {
		t.Fatalf("sinceDate: got %q want %q", dateStr, "2026-11-11")
	}
	if timeInt != 125427 {
		t.Fatalf("sinceTime: got %d want %d", timeInt, 125427)
	}
	if dateTimeStr != "2026-11-11T12:54:27" {
		t.Fatalf("sinceDateTime: got %q want %q", dateTimeStr, "2026-11-11T12:54:27")
	}

	_, zeroTime, zeroDateTime := sapDeltaParams(time.Time{})
	if zeroTime != 0 {
		t.Fatalf("zero time sinceTime: got %d want 0", zeroTime)
	}
	if len(zeroDateTime) != len("2006-01-02T00:00:00") || zeroDateTime[10:] != "T00:00:00" {
		t.Fatalf("zero time sinceDateTime: got %q want T00:00:00 suffix", zeroDateTime)
	}

	// Verify sapDeltaParams preserves SAP server's own timezone (does NOT shift to host time.Local).
	riyadh := time.FixedZone("AST", 3*3600)
	sapClock := time.Date(2026, 11, 27, 12, 15, 30, 0, riyadh)
	d, tm, dt := sapDeltaParams(sapClock)
	if d != "2026-11-27" || tm != 121530 || dt != "2026-11-27T12:15:30" {
		t.Fatalf("expected SAP +03:00 wall-clock preserved, got date=%q time=%d dt=%q", d, tm, dt)
	}

	naiveUTC := time.Date(2026, 11, 27, 12, 15, 30, 0, time.UTC)
	d, tm, dt = sapDeltaParams(naiveUTC)
	if d != "2026-11-27" || tm != 121530 || dt != "2026-11-27T12:15:30" {
		t.Fatalf("expected naive UTC wall-clock preserved without .Local() shift, got date=%q time=%d dt=%q", d, tm, dt)
	}
}
