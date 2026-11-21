package sap

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError represents a structured HTTP error response returned by SAP B1 Service Layer.
type APIError struct {
	StatusCode int
	Code       int
	Message    string
}

func (e *APIError) Error() string {
	if e.Code != 0 {
		return fmt.Sprintf("SAP API error (%d): %s [code %d]", e.StatusCode, e.Message, e.Code)
	}
	return fmt.Sprintf("SAP request returned HTTP %d: %s", e.StatusCode, e.Message)
}

// IsPermanentBusinessError returns true when err is an SAP Service Layer 4xx
// business/validation/configuration error (e.g. missing bin allocation or
// undefined numbering series) that will not succeed on immediate retries.
// Auth/session (401/403), rate-limit (429), 5xx, and transport errors return false.
func IsPermanentBusinessError(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	switch apiErr.StatusCode {
	case http.StatusBadRequest, http.StatusNotFound, http.StatusUnprocessableEntity:
		return true
	default:
		return false
	}
}
