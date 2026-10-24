package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/NasTecSol/nembus-core/payment/geidea"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

type GeideaPaymentUseCase struct {
	repo         *repository.Queries
	driverLoader func(terminalConfig map[string]interface{}) (geidea.TerminalDriver, error)
}

func NewGeideaPaymentUseCase(repo *repository.Queries) *GeideaPaymentUseCase {
	return &GeideaPaymentUseCase{
		repo: repo,
		driverLoader: func(meta map[string]interface{}) (geidea.TerminalDriver, error) {
			isTest, _ := meta["is_test_mode"].(bool)
			if isTest {
				return geidea.NewSimulator(nil), nil
			}
			ip, _ := meta["terminal_ip"].(string)
			if ip == "" {
				ip = "127.0.0.1:8080"
			}
			return geidea.NewEcrTcpClient(ip, 90*time.Second, nil), nil
		},
	}
}

func (uc *GeideaPaymentUseCase) SetDriverLoader(loader func(meta map[string]interface{}) (geidea.TerminalDriver, error)) {
	uc.driverLoader = loader
}

type InitiateGeideaPurchaseInput struct {
	TransactionID  int32   `json:"transaction_id"`
	TerminalID     int32   `json:"terminal_id"`
	Amount         float64 `json:"amount"`
	ECRReferenceID string  `json:"ecr_reference_id"`
}

type ResolveUncertainInput struct {
	PaymentID      int32  `json:"payment_id"`
	ResolutionType string `json:"resolution_type"` // "query", "reversal", "manual_slip"
	RRN            string `json:"rrn,omitempty"`
	AuthCode       string `json:"auth_code,omitempty"`
	Notes          string `json:"notes,omitempty"`
}

func (uc *GeideaPaymentUseCase) Purchase(ctx context.Context, in InitiateGeideaPurchaseInput) (*repository.PosPayment, *geidea.TransactionResponse, error) {
	if in.Amount <= 0 {
		return nil, nil, errors.New("payment amount must be greater than zero")
	}

	// 1. Fetch Terminal & Metadata
	terminal, err := uc.repo.GetPOSTerminal(ctx, in.TerminalID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load terminal %d: %w", in.TerminalID, err)
	}

	var meta map[string]interface{}
	if len(terminal.Metadata) > 0 {
		_ = json.Unmarshal(terminal.Metadata, &meta)
	}
	if meta == nil {
		meta = make(map[string]interface{})
	}

	driver, err := uc.driverLoader(meta)
	if err != nil {
		return nil, nil, err
	}

	// 2. Insert initial payment in 'initiated' status
	var amountNum pgtype.Numeric
	_ = amountNum.Scan(fmt.Sprintf("%.2f", in.Amount))

	payment, err := uc.repo.CreateGeideaPayment(ctx, repository.CreateGeideaPaymentParams{
		TransactionID:  in.TransactionID,
		PaymentMethod:  "card",
		PaymentGateway: pgtype.Text{String: "geidea", Valid: true},
		Amount:         amountNum,
		Status:         pgtype.Text{String: geidea.StatusInitiated, Valid: true},
		EcrReferenceID: pgtype.Text{String: in.ECRReferenceID, Valid: true},
		Metadata:       terminal.Metadata,
		RawResponse:    nil,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create initiated payment: %w", err)
	}

	// 3. Dispatch purchase over driver
	resp, driverErr := driver.Purchase(ctx, &geidea.PurchaseRequest{
		ECRReferenceID: in.ECRReferenceID,
		Amount:         in.Amount,
		Currency:       "SAR",
		Timeout:        90 * time.Second,
	})

	if driverErr != nil {
		// Network drop / timeout -> immediately trigger QueryStatus before marking uncertain
		queryResp, qErr := driver.QueryStatus(ctx, in.ECRReferenceID)
		if qErr == nil && queryResp.Status == geidea.StatusApproved {
			upd, _ := uc.repo.UpdateGeideaPaymentApproved(ctx, repository.UpdateGeideaPaymentApprovedParams{
				ID:          payment.ID,
				Rrn:         pgtype.Text{String: queryResp.RRN, Valid: true},
				AuthCode:    pgtype.Text{String: queryResp.AuthCode, Valid: true},
				Tid:         pgtype.Text{String: queryResp.TID, Valid: true},
				Mid:         pgtype.Text{String: queryResp.MID, Valid: true},
				CardScheme:  pgtype.Text{String: queryResp.CardScheme, Valid: true},
				MaskedPan:   pgtype.Text{String: queryResp.MaskedPAN, Valid: true},
				Aid:         pgtype.Text{String: queryResp.AID, Valid: true},
				RawResponse: queryResp.RawResponse,
			})
			return &upd, queryResp, nil
		}

		// Mark uncertain
		upd, _ := uc.repo.UpdateGeideaPaymentStatus(ctx, repository.UpdateGeideaPaymentStatusParams{
			ID:          payment.ID,
			Status:      pgtype.Text{String: geidea.StatusUncertain, Valid: true},
			RawResponse: nil,
		})
		return &upd, &geidea.TransactionResponse{
			Status:          geidea.StatusUncertain,
			ECRReferenceID:  in.ECRReferenceID,
			ResponseMessage: driverErr.Error(),
		}, nil
	}

	// 4. Handle clean driver response
	if resp.Status == geidea.StatusApproved {
		upd, err := uc.repo.UpdateGeideaPaymentApproved(ctx, repository.UpdateGeideaPaymentApprovedParams{
			ID:          payment.ID,
			Rrn:         pgtype.Text{String: resp.RRN, Valid: true},
			AuthCode:    pgtype.Text{String: resp.AuthCode, Valid: true},
			Tid:         pgtype.Text{String: resp.TID, Valid: true},
			Mid:         pgtype.Text{String: resp.MID, Valid: true},
			CardScheme:  pgtype.Text{String: resp.CardScheme, Valid: true},
			MaskedPan:   pgtype.Text{String: resp.MaskedPAN, Valid: true},
			Aid:         pgtype.Text{String: resp.AID, Valid: true},
			RawResponse: resp.RawResponse,
		})
		return &upd, resp, err
	}

	// Declined
	upd, err := uc.repo.UpdateGeideaPaymentStatus(ctx, repository.UpdateGeideaPaymentStatusParams{
		ID:          payment.ID,
		Status:      pgtype.Text{String: geidea.StatusDeclined, Valid: true},
		RawResponse: resp.RawResponse,
	})
	return &upd, resp, err
}

func (uc *GeideaPaymentUseCase) ResolveUncertain(ctx context.Context, in ResolveUncertainInput) (*repository.PosPayment, error) {
	payment, err := uc.repo.GetPosPayment(ctx, in.PaymentID)
	if err != nil {
		return nil, err
	}

	switch in.ResolutionType {
	case "manual_slip":
		if in.RRN == "" {
			return nil, errors.New("RRN required for manual slip verification")
		}
		raw, _ := json.Marshal(map[string]string{"resolution": "manual_slip", "notes": in.Notes})
		upd, err := uc.repo.UpdateGeideaPaymentApproved(ctx, repository.UpdateGeideaPaymentApprovedParams{
			ID:          payment.ID,
			Rrn:         pgtype.Text{String: in.RRN, Valid: true},
			AuthCode:    pgtype.Text{String: in.AuthCode, Valid: true},
			Tid:         payment.Tid,
			Mid:         payment.Mid,
			CardScheme:  pgtype.Text{String: "mada", Valid: true},
			MaskedPan:   payment.MaskedPan,
			Aid:         payment.Aid,
			RawResponse: raw,
		})
		return &upd, err

	case "reversal":
		raw, _ := json.Marshal(map[string]string{"resolution": "reversal", "notes": in.Notes})
		upd, err := uc.repo.UpdateGeideaPaymentStatus(ctx, repository.UpdateGeideaPaymentStatusParams{
			ID:          payment.ID,
			Status:      pgtype.Text{String: geidea.StatusReversed, Valid: true},
			RawResponse: raw,
		})
		return &upd, err

	case "query":
		if !payment.EcrReferenceID.Valid || payment.EcrReferenceID.String == "" {
			return nil, errors.New("no ecr reference found for query")
		}
		driver, err := uc.driverLoader(nil)
		if err != nil {
			return nil, err
		}
		resp, err := driver.QueryStatus(ctx, payment.EcrReferenceID.String)
		if err != nil {
			return nil, err
		}
		if resp.Status == geidea.StatusApproved {
			upd, err := uc.repo.UpdateGeideaPaymentApproved(ctx, repository.UpdateGeideaPaymentApprovedParams{
				ID:          payment.ID,
				Rrn:         pgtype.Text{String: resp.RRN, Valid: true},
				AuthCode:    pgtype.Text{String: resp.AuthCode, Valid: true},
				Tid:         pgtype.Text{String: resp.TID, Valid: true},
				Mid:         pgtype.Text{String: resp.MID, Valid: true},
				CardScheme:  pgtype.Text{String: resp.CardScheme, Valid: true},
				MaskedPan:   pgtype.Text{String: resp.MaskedPAN, Valid: true},
				Aid:         pgtype.Text{String: resp.AID, Valid: true},
				RawResponse: resp.RawResponse,
			})
			return &upd, err
		}
		upd, err := uc.repo.UpdateGeideaPaymentStatus(ctx, repository.UpdateGeideaPaymentStatusParams{
			ID:          payment.ID,
			Status:      pgtype.Text{String: resp.Status, Valid: true},
			RawResponse: resp.RawResponse,
		})
		return &upd, err

	default:
		return nil, fmt.Errorf("unknown resolution type: %s", in.ResolutionType)
	}
}
