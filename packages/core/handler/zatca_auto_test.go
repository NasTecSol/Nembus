package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeOrderReporter struct {
	entry *repository.ZatcaDocumentChain
	err   error
	calls int
}

func (f *fakeOrderReporter) SubmitSalesOrder(_ context.Context, _ uuid.UUID) (*repository.ZatcaDocumentChain, error) {
	f.calls++
	return f.entry, f.err
}

func TestAutoReportSalesOrderReturnsSavedGatewayResult(t *testing.T) {
	id := uuid.New()
	reporter := &fakeOrderReporter{entry: &repository.ZatcaDocumentChain{
		ZatcaStatus:   repository.NullZatcaDocStatus{ZatcaDocStatus: repository.ZatcaDocStatusReported, Valid: true},
		ZatcaResponse: []byte(`{"reportingStatus":"REPORTED"}`),
		QrCodeBase64:  pgtype.Text{String: "saved-tlv", Valid: true},
	}}
	result := autoReportSalesOrder(context.Background(), reporter, id)
	if reporter.calls != 1 || result["reported"] != true || result["qr_code"] != "saved-tlv" || result["reporting_status"] != "REPORTED" {
		t.Fatalf("unexpected automatic reporting result: %#v", result)
	}
}

func TestAutoReportSalesOrderExposesFailureWithoutClaimingSuccess(t *testing.T) {
	reporter := &fakeOrderReporter{err: errors.New("gateway unavailable")}
	result := autoReportSalesOrder(context.Background(), reporter, uuid.New())
	if result["reported"] != false || result["error"] != "gateway unavailable" {
		t.Fatalf("unexpected failure result: %#v", result)
	}
	reporter.err = usecase.ErrZatcaDisabled
	result = autoReportSalesOrder(context.Background(), reporter, uuid.New())
	if result["enabled"] != false || result["reported"] != false || result["error"] != nil {
		t.Fatalf("unexpected disabled result: %#v", result)
	}
}
