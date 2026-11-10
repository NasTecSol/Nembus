package geidea

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

type Simulator struct {
	mu            sync.RWMutex
	transactions  map[string]*TransactionResponse
	shouldDrop    bool
	simulateDelay time.Duration
	statusHook    func(status string)
}

func NewSimulator(statusHook func(status string)) *Simulator {
	return &Simulator{
		transactions:  make(map[string]*TransactionResponse),
		statusHook:    statusHook,
		simulateDelay: 10 * time.Millisecond,
	}
}

func (s *Simulator) SetDropConnection(drop bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shouldDrop = drop
}

func (s *Simulator) notify(status string) {
	if s.statusHook != nil {
		s.statusHook(status)
	}
}

func (s *Simulator) Purchase(ctx context.Context, req *PurchaseRequest) (*TransactionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.notify("WAITING_FOR_CARD")
	time.Sleep(s.simulateDelay)
	s.notify("PIN_ENTRY")
	time.Sleep(s.simulateDelay)
	s.notify("PROCESSING")

	if s.shouldDrop {
		return nil, errors.New("simulated network connection drop")
	}

	// Trigger decline if amount has cents == 0.99
	cents := int(math.Round(req.Amount*100)) % 100
	if cents == 99 {
		resp := &TransactionResponse{
			Status:          StatusDeclined,
			ResponseCode:    "51",
			ResponseMessage: "DECLINED_INSUFFICIENT_FUNDS",
			ECRReferenceID:  req.ECRReferenceID,
			Amount:          req.Amount,
			Timestamp:       time.Now(),
		}
		raw, _ := json.Marshal(resp)
		resp.RawResponse = raw
		s.transactions[req.ECRReferenceID] = resp
		return resp, nil
	}

	resp := &TransactionResponse{
		Status:          StatusApproved,
		ResponseCode:    "00",
		ResponseMessage: "APPROVED",
		ECRReferenceID:  req.ECRReferenceID,
		RRN:             fmt.Sprintf("999%09d", time.Now().UnixNano()%1000000000),
		AuthCode:        "OK7890",
		TID:             "TID88801",
		MID:             "MID99901",
		CardScheme:      "mada",
		MaskedPAN:       "588845******1234",
		AID:             "A0000002281010",
		Amount:          req.Amount,
		Timestamp:       time.Now(),
	}
	raw, _ := json.Marshal(resp)
	resp.RawResponse = raw
	s.transactions[req.ECRReferenceID] = resp
	return resp, nil
}

func (s *Simulator) Refund(ctx context.Context, req *RefundRequest) (*TransactionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	resp := &TransactionResponse{
		Status:          StatusApproved,
		ResponseCode:    "00",
		ResponseMessage: "REFUND_APPROVED",
		ECRReferenceID:  req.ECRReferenceID,
		RRN:             req.OriginalRRN,
		AuthCode:        "RF1234",
		TID:             "TID88801",
		MID:             "MID99901",
		CardScheme:      "mada",
		Amount:          req.Amount,
		Timestamp:       time.Now(),
	}
	raw, _ := json.Marshal(resp)
	resp.RawResponse = raw
	s.transactions[req.ECRReferenceID] = resp
	return resp, nil
}

func (s *Simulator) Void(ctx context.Context, req *VoidRequest) (*TransactionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	resp := &TransactionResponse{
		Status:          StatusVoided,
		ResponseCode:    "00",
		ResponseMessage: "VOID_APPROVED",
		ECRReferenceID:  req.ECRReferenceID,
		RRN:             req.OriginalRRN,
		Timestamp:       time.Now(),
	}
	raw, _ := json.Marshal(resp)
	resp.RawResponse = raw
	s.transactions[req.ECRReferenceID] = resp
	return resp, nil
}

func (s *Simulator) QueryStatus(ctx context.Context, ecrRef string) (*TransactionResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if resp, ok := s.transactions[ecrRef]; ok {
		return resp, nil
	}
	return &TransactionResponse{
		Status:          StatusUncertain,
		ResponseCode:    "99",
		ResponseMessage: "TRANSACTION_NOT_FOUND",
		ECRReferenceID:  ecrRef,
		Timestamp:       time.Now(),
	}, nil
}

func (s *Simulator) ReconcileBatch(ctx context.Context) (*ReconciliationResponse, error) {
	return &ReconciliationResponse{
		TotalCount:   len(s.transactions),
		TotalAmount:  1000.00,
		BatchNumber:  "B0001",
		ResponseCode: "00",
		Timestamp:    time.Now(),
	}, nil
}

func (s *Simulator) CancelInFlight(ctx context.Context, ecrRef string) (*TransactionResponse, error) {
	return &TransactionResponse{
		Status:          StatusDeclined,
		ResponseCode:    "98",
		ResponseMessage: "CANCELLED_BY_CASHIER",
		ECRReferenceID:  ecrRef,
		Timestamp:       time.Now(),
	}, nil
}

func (s *Simulator) Ping(ctx context.Context) error {
	return nil
}
