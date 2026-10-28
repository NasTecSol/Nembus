package zatca

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestPOSReportingRequestUsesSavedInvoiceAndCredentials(t *testing.T) {
	id := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost || req.URL.Path != "/invoices/reporting/single" || req.Header.Get("Accept-Version") != "V2" || req.Header.Get("Clearance-Status") != "0" {
			t.Error("incorrect ZATCA reporting route or headers")
		}
		user, secret, ok := req.BasicAuth()
		if !ok || user != "production-csid" || secret != "production-secret" {
			t.Error("production credentials were not used")
		}
		var body struct {
			InvoiceHash string `json:"invoiceHash"`
			UUID        string `json:"uuid"`
			Invoice     string `json:"invoice"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body.InvoiceHash != "saved-hash" || body.UUID != id.String() || body.Invoice != base64.StdEncoding.EncodeToString([]byte("<Invoice/>")) {
			t.Error("saved invoice identity or XML changed")
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"reportingStatus":"REPORTED"}`))
	}))
	defer server.Close()
	reporter := NewReporter(nil)
	entry := &repository.ZatcaDocumentChain{XmlHash: "saved-hash", ZatcaUuid: id, SignedXml: pgtype.Text{String: "<Invoice/>", Valid: true}}
	raw, status, reportingStatus, err := reporter.report(context.Background(), server.URL, "production-csid", "production-secret", entry)
	if err != nil || status != http.StatusAccepted || reportingStatus != "REPORTED" || !json.Valid(raw) {
		t.Fatalf("report result: HTTP %d, status %q, error %v", status, reportingStatus, err)
	}
}
