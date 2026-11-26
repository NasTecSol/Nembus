package sap

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestAPIErrorAndIsPermanentBusinessError(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		permanent bool
	}{
		{"sap 400 bin location", &APIError{StatusCode: http.StatusBadRequest, Code: -10, Message: "Fully allocate item"}, true},
		{"wrapped sap 400 numbering series", fmt.Errorf("post credit memo: %w", &APIError{StatusCode: http.StatusBadRequest, Code: -5002, Message: "define numbering series"}), true},
		{"sap 404", &APIError{StatusCode: http.StatusNotFound, Message: "not found"}, true},
		{"sap 422", &APIError{StatusCode: http.StatusUnprocessableEntity, Message: "invalid"}, true},
		{"sap 401 session expired", &APIError{StatusCode: http.StatusUnauthorized, Message: "session expired"}, false},
		{"sap 403 forbidden", &APIError{StatusCode: http.StatusForbidden, Message: "forbidden"}, false},
		{"sap 429 rate limit", &APIError{StatusCode: http.StatusTooManyRequests, Message: "too many requests"}, false},
		{"sap 500 server error", &APIError{StatusCode: http.StatusInternalServerError, Message: "internal"}, false},
		{"network error", errors.New("connection refused"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsPermanentBusinessError(tc.err); got != tc.permanent {
				t.Errorf("IsPermanentBusinessError(%v) = %v, want %v", tc.err, got, tc.permanent)
			}
		})
	}

	e := &APIError{StatusCode: 400, Code: -5002, Message: "define series"}
	want := "SAP API error (400): define series [code -5002]"
	if e.Error() != want {
		t.Errorf("APIError.Error() = %q, want %q", e.Error(), want)
	}
}
