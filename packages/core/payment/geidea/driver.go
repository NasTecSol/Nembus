package geidea

import (
	"context"
	"encoding/json"
	"time"
)

// TransactionStatus constants
const (
	StatusInitiated = "initiated"
	StatusApproved  = "approved"
	StatusDeclined  = "declined"
	StatusUncertain = "uncertain"
	StatusVoided    = "voided"
	StatusReversed  = "reversed"
)

type PurchaseRequest struct {
	ECRReferenceID string                 `json:"ecr_reference_id"`
	Amount         float64                `json:"amount"`
	Currency       string                 `json:"currency"`
	Timeout        time.Duration          `json:"timeout"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type RefundRequest struct {
	ECRReferenceID string                 `json:"ecr_reference_id"`
	OriginalRRN    string                 `json:"original_rrn"`
	Amount         float64                `json:"amount"`
	Currency       string                 `json:"currency"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type VoidRequest struct {
	ECRReferenceID string                 `json:"ecr_reference_id"`
	OriginalRRN    string                 `json:"original_rrn"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type TransactionResponse struct {
	Status          string          `json:"status"`
	ResponseCode    string          `json:"response_code"`
	ResponseMessage string          `json:"response_message"`
	ECRReferenceID  string          `json:"ecr_reference_id"`
	RRN             string          `json:"rrn,omitempty"`
	AuthCode        string          `json:"auth_code,omitempty"`
	TID             string          `json:"tid,omitempty"`
	MID             string          `json:"mid,omitempty"`
	CardScheme      string          `json:"card_scheme,omitempty"`
	MaskedPAN       string          `json:"masked_pan,omitempty"`
	AID             string          `json:"aid,omitempty"`
	Amount          float64         `json:"amount"`
	RawResponse     json.RawMessage `json:"raw_response,omitempty"`
	Timestamp       time.Time       `json:"timestamp"`
}

type ReconciliationResponse struct {
	TotalCount   int       `json:"total_count"`
	TotalAmount  float64   `json:"total_amount"`
	BatchNumber  string    `json:"batch_number"`
	ResponseCode string    `json:"response_code"`
	Timestamp    time.Time `json:"timestamp"`
}

type TerminalDriver interface {
	Purchase(ctx context.Context, req *PurchaseRequest) (*TransactionResponse, error)
	Refund(ctx context.Context, req *RefundRequest) (*TransactionResponse, error)
	Void(ctx context.Context, req *VoidRequest) (*TransactionResponse, error)
	QueryStatus(ctx context.Context, ecrRef string) (*TransactionResponse, error)
	ReconcileBatch(ctx context.Context) (*ReconciliationResponse, error)
	CancelInFlight(ctx context.Context, ecrRef string) (*TransactionResponse, error)
	Ping(ctx context.Context) error
}
