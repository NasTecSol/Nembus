package usecase

import (
	"regexp"
	"strings"
	"testing"
)

func TestGenerateTerminalDocNumber(t *testing.T) {
	re := regexp.MustCompile(`^[A-Z]+-[A-Z0-9]+-\d{8}-[0-9A-F]{6}$`)

	tests := []struct {
		name         string
		prefix       string
		terminalCode string
		wantPrefix   string
		wantTerm     string
	}{
		{
			name:         "standard POS terminal",
			prefix:       "POS",
			terminalCode: "T01",
			wantPrefix:   "POS",
			wantTerm:     "T01",
		},
		{
			name:         "empty terminal fallback to T00",
			prefix:       "RET",
			terminalCode: "",
			wantPrefix:   "RET",
			wantTerm:     "T00",
		},
		{
			name:         "sanitizes punctuation and lowercase in terminal code",
			prefix:       "ses",
			terminalCode: "pos-term_02",
			wantPrefix:   "SES",
			wantTerm:     "POSTERM02",
		},
		{
			name:         "customer code prefix",
			prefix:       "CUST",
			terminalCode: "K1",
			wantPrefix:   "CUST",
			wantTerm:     "K1",
		},
	}

	seen := make(map[string]struct{})
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GenerateTerminalDocNumber(tc.prefix, tc.terminalCode)
			if !re.MatchString(got) {
				t.Fatalf("GenerateTerminalDocNumber(%q, %q) = %q, does not match pattern %s", tc.prefix, tc.terminalCode, got, re.String())
			}
			parts := strings.Split(got, "-")
			if len(parts) != 4 {
				t.Fatalf("expected 4 hyphen-separated parts, got %d in %q", len(parts), got)
			}
			if parts[0] != tc.wantPrefix {
				t.Errorf("prefix = %q, want %q", parts[0], tc.wantPrefix)
			}
			if parts[1] != tc.wantTerm {
				t.Errorf("terminal = %q, want %q", parts[1], tc.wantTerm)
			}
			if _, exists := seen[got]; exists {
				t.Errorf("duplicate doc number generated: %q", got)
			}
			seen[got] = struct{}{}
		})
	}
}
