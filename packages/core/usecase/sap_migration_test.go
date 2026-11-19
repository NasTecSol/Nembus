package usecase

import "testing"

func TestIsZeroOrExemptSAPVatGroup(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{code: "O0", want: true},
		{code: "s0", want: true},
		{code: " X0 ", want: true},
		{code: "OUT0", want: true},
		{code: "zero", want: true},
		{code: "EXEMPT", want: true},
		{code: "e", want: true},
		{code: " ex ", want: true},
		{code: "O1", want: false},
		{code: "S15", want: false},
		{code: "", want: false},
	}

	for _, tc := range tests {
		if got := isZeroOrExemptSAPVatGroup(tc.code); got != tc.want {
			t.Errorf("isZeroOrExemptSAPVatGroup(%q) = %v, want %v", tc.code, got, tc.want)
		}
	}
}
