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

func TestAutoReportSalesOrderReceiptQRRequiresReportedGatewayQR(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   repository.ZatcaDocStatus
		response string
		err      error
		wantQR   string
	}{
		{"reported with returned QR", repository.ZatcaDocStatusReported, `{"reportingStatus":"REPORTED","qrCode":"gateway-qr"}`, nil, "gateway-qr"},
		{"reported without returned QR", repository.ZatcaDocStatusReported, `{"reportingStatus":"REPORTED"}`, nil, ""},
		{"pending with QR", repository.ZatcaDocStatusPending, `{"reportingStatus":"REPORTED","qrCode":"gateway-qr"}`, nil, ""},
		{"rejected", repository.ZatcaDocStatusRejected, `{"reportingStatus":"NOT_REPORTED","qrCode":"gateway-qr"}`, errors.New("rejected"), ""},
		{"gateway status not reported", repository.ZatcaDocStatusReported, `{"reportingStatus":"NOT_REPORTED","qrCode":"gateway-qr"}`, nil, ""},
		{"invalid response", repository.ZatcaDocStatusReported, `{`, nil, ""},
		{"submission error", repository.ZatcaDocStatusReported, `{"reportingStatus":"REPORTED","qrCode":"gateway-qr"}`, errors.New("save failed"), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reporter := &fakeOrderReporter{err: tc.err, entry: &repository.ZatcaDocumentChain{
				ZatcaStatus:   repository.NullZatcaDocStatus{ZatcaDocStatus: tc.status, Valid: true},
				ZatcaResponse: []byte(tc.response),
				QrCodeBase64:  pgtype.Text{String: "local-qr-must-not-print", Valid: true},
			}}
			result := autoReportSalesOrder(context.Background(), reporter, uuid.New())
			qr, _ := result["receipt_qr_code_base64"].(string)
			if qr != tc.wantQR || reporter.calls != 1 {
				t.Fatalf("receipt QR = %q, want %q; gateway calls = %d", qr, tc.wantQR, reporter.calls)
			}
		})
	}
}
