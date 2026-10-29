package zatca

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// Local mock CA fixture; never used against ZATCA.
func completionCertificate(t *testing.T, key *ecdsa.PrivateKey, serial int64) string {
	return completionCertificateWithSubject(t, &key.PublicKey, key, serial)
}
func completionCertificateWithSubject(t *testing.T, subject *ecdsa.PublicKey, signer *ecdsa.PrivateKey, serial int64) string {
	t.Helper()
	issuer, _ := asn1.Marshal(pkix.Name{CommonName: "Mock CA"}.ToRDNSequence())
	curve, _ := asn1.Marshal(asn1.ObjectIdentifier{1, 3, 132, 0, 10})
	algorithm := pkix.AlgorithmIdentifier{Algorithm: asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 2}}
	tbs := struct {
		Version   int `asn1:"optional,explicit,tag:0,default:0"`
		Serial    *big.Int
		Algorithm pkix.AlgorithmIdentifier
		Issuer    asn1.RawValue
		Validity  struct{ NotBefore, NotAfter time.Time }
		Subject   asn1.RawValue
		SPKI      struct {
			Algorithm pkix.AlgorithmIdentifier
			Key       asn1.BitString
		}
	}{Version: 2, Serial: big.NewInt(serial), Algorithm: algorithm, Issuer: asn1.RawValue{FullBytes: issuer}, Subject: asn1.RawValue{FullBytes: issuer}}
	tbs.Validity.NotBefore, tbs.Validity.NotAfter = time.Now().Add(-time.Hour), time.Now().Add(48*time.Hour)
	tbs.SPKI.Algorithm = pkix.AlgorithmIdentifier{Algorithm: asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}, Parameters: asn1.RawValue{FullBytes: curve}}
	pub := elliptic.Marshal(subject.Curve, subject.X, subject.Y)
	tbs.SPKI.Key = asn1.BitString{Bytes: pub, BitLength: len(pub) * 8}
	raw, err := asn1.Marshal(tbs)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(raw)
	sig, err := ecdsa.SignASN1(rand.Reader, signer, hash[:])
	if err != nil {
		t.Fatal(err)
	}
	cert, err := asn1.Marshal(struct {
		TBS       asn1.RawValue
		Algorithm pkix.AlgorithmIdentifier
		Signature asn1.BitString
	}{asn1.RawValue{FullBytes: raw}, algorithm, asn1.BitString{Bytes: sig, BitLength: len(sig) * 8}})
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(cert)
}

type completionDB struct {
	repository.DBTX
	device   repository.ZatcaDeviceConfig
	org      repository.Organization
	writes   int
	failSave bool
}

func (db *completionDB) QueryRow(_ context.Context, sql string, args ...interface{}) pgx.Row {
	if strings.Contains(sql, "GetZatcaDeviceConfig") {
		return submissionRow{db.device}
	}
	if strings.Contains(sql, "GetOrganization") {
		return submissionRow{db.org}
	}
	panic("unexpected query: " + sql)
}
func (db *completionDB) Exec(_ context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	if !strings.Contains(sql, "CompleteZatcaOnboarding") {
		panic("unexpected write")
	}
	db.writes++
	if db.failSave {
		return pgconn.CommandTag{}, fmt.Errorf("database unavailable")
	}
	db.device.ProductionCsid = pgtype.Text{String: args[2].(string), Valid: true}
	db.device.CsidExpiry = pgtype.Timestamptz{Time: args[3].(time.Time), Valid: true}
	var before, patch map[string]any
	_ = json.Unmarshal(db.device.Metadata, &before)
	_ = json.Unmarshal(args[4].(json.RawMessage), &patch)
	for k, v := range patch {
		before[k] = v
	}
	db.device.Metadata, _ = json.Marshal(before)
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func TestCompleteOnboardingChecksBeforeExchangeAndSavesPair(t *testing.T) {
	for _, scenario := range []string{"success", "already-completed", "validation-error", "http-error", "exchange-error", "missing-secret", "wrong-key", "save-error"} {
		t.Run(scenario, func(t *testing.T) {
			key, _ := usecase.GenerateECDSAKeyPair()
			pem, _ := usecase.EncodePrivateKeyPEM(key)
			compliance := completionCertificate(t, key, 1)
			production := completionCertificate(t, key, 2)
			if scenario == "wrong-key" {
				other, _ := usecase.GenerateECDSAKeyPair()
				production = completionCertificate(t, other, 2)
			}
			db := &completionDB{device: repository.ZatcaDeviceConfig{ID: 3, OrganizationID: 1, PrivateKeyPem: pem, ComplianceCsid: pgtype.Text{String: compliance, Valid: true}, ZatcaEnv: pgtype.Text{String: "simulation", Valid: true}, IsActive: pgtype.Bool{Bool: true, Valid: true}, Metadata: []byte(`{"compliance_secret":"cc-secret","compliance_request_id":"12345","onboarding_stage":"compliance","unrelated":"preserve"}`)}, org: repository.Organization{ID: 1, Name: "Seller", TaxID: pgtype.Text{String: "310123456700003", Valid: true}, Metadata: []byte(`{"zatca":{"enabled":true,"seller":{"commercial_registration_number":"1010123456","street":"Street","building_number":"1234","district":"District","city":"Riyadh","postal_code":"12345"}}}`)}, failSave: scenario == "save-error"}
			checks, exchanges := 0, 0
			gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user, secret, ok := r.BasicAuth()
				if !ok || user != compliance || secret != "cc-secret" || r.Header.Get("Accept-Version") != "V2" {
					t.Error("wrong saved credentials or version")
				}
				switch r.URL.Path {
				case "/compliance/invoices":
					checks++
					var payload InvoiceSubmissionRequest
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Error(err)
					}
					xml, err := base64.StdEncoding.DecodeString(payload.Invoice)
					if err != nil {
						t.Error(err)
					}
					canonical, err := usecase.CanonicalizeXML(xml)
					if err != nil || usecase.HashSHA256Base64(canonical) != payload.InvoiceHash {
						t.Error("invalid request hash")
					}
					if scenario == "http-error" {
						w.WriteHeader(401)
						return
					}
					if scenario == "validation-error" {
						fmt.Fprint(w, `{"validationResults":{"status":"ERROR","errorMessages":[{"code":"bad-sample"}]}}`)
						return
					}
					if scenario == "already-completed" {
						w.WriteHeader(406)
						kind := []string{"STANDARD", "STANDARD_CREDIT_NOTE", "STANDARD_DEBIT_NOTE", "SIMPLIFIED", "SIMPLIFIED_CREDIT_NOTE", "SIMPLIFIED_DEBIT_NOTE"}[checks-1]
						_ = json.NewEncoder(w).Encode(map[string]any{"validationResults": map[string]any{"status": "ERROR", "errorMessages": []map[string]string{{"code": "Submitted before", "category": "Compliance-Check", "message": "Compliance check already completed for " + kind + "."}}}})
						return
					}
					fmt.Fprint(w, `{"validationResults":{"status":"PASS","errorMessages":[]}}`)
				case "/production/csids":
					exchanges++
					if checks != 6 {
						t.Error("exchanged before all six checks")
					}
					var payload ProductionCSIDRequest
					_ = json.NewDecoder(r.Body).Decode(&payload)
					if payload.ComplianceRequestID != "12345" {
						t.Error("wrong saved request ID")
					}
					if scenario == "exchange-error" {
						w.WriteHeader(400)
						fmt.Fprint(w, `{"errors":["Missing-ComplianceSteps"]}`)
						return
					}
					secret := "pc-secret"
					if scenario == "missing-secret" {
						secret = ""
					}
					_ = json.NewEncoder(w).Encode(map[string]string{"binarySecurityToken": production, "secret": secret})
				default:
					t.Error("unexpected gateway path", r.URL.Path)
				}
			}))
			defer gateway.Close()
			service := testOnboardingService(nil, NewZatcaClient(gateway.URL), &usecase.ZatcaConfig{Enabled: true, Env: "simulation"})
			ctx := context.WithValue(context.Background(), middleware.RepoKey, repository.New(db))
			result, err := service.CompleteDeviceOnboarding(ctx, 3)
			if scenario != "success" && scenario != "already-completed" {
				if err == nil || result.ReportingReady {
					t.Fatal("failed onboarding reported ready")
				}
				if scenario == "validation-error" || scenario == "http-error" {
					if checks != 1 || exchanges != 0 {
						t.Fatal("continued after failed compliance")
					}
				}
				if scenario != "save-error" && db.writes != 0 {
					t.Fatal("saved failed credentials")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !result.ReportingReady || result.Stage != "production" || len(result.Checks) != 6 || db.writes != 1 {
				t.Fatal("onboarding incomplete")
			}
			if scenario == "already-completed" {
				for _, check := range result.Checks {
					if !check.AlreadyCompleted {
						t.Fatal("completed check not identified")
					}
				}
			}
			var metadata map[string]any
			_ = json.Unmarshal(db.device.Metadata, &metadata)
			if db.device.ProductionCsid.String != production || metadata["production_secret"] != "pc-secret" || metadata["secret"] != "pc-secret" || metadata["compliance_secret"] != "cc-secret" || metadata["unrelated"] != "preserve" {
				t.Fatal("credential pair or metadata lost")
			}
			data, _ := json.Marshal(result)
			if strings.Contains(string(data), "pc-secret") || strings.Contains(string(data), compliance) {
				t.Fatal("credentials exposed")
			}
			if _, err := service.CompleteDeviceOnboarding(ctx, 3); err != nil {
				t.Fatal(err)
			}
			if checks != 6 || exchanges != 1 || db.writes != 1 {
				t.Fatal("repeat completion issued new credentials")
			}
			if _, err := service.CompleteDeviceOnboarding(context.Background(), 3); err == nil {
				t.Fatal("accepted missing tenant")
			}
		})
	}
}

func TestComplianceAcceptanceRequiresValidationNotOnlyHTTP200(t *testing.T) {
	for _, test := range []struct {
		body     string
		accepted bool
	}{
		{`{"validationResults":{"status":"PASS","errorMessages":[]}}`, true},
		{`{"validationResults":{"status":"WARNING","errorMessages":[],"warningMessages":[{"code":"test"}]}}`, true},
		{`{}`, false},
		{`{"validationResults":{"status":"PASS","errorMessages":[{"code":"bad"}]}}`, false},
		{`{"validationResults":{"status":"PASS"},"errors":[{"code":"bad"}]}`, false},
		{`{"validationResults":{"status":"ERROR"}}`, false},
	} {
		if got := complianceAccepted(&ReportingResponse{HTTPStatus: 200, RawResponse: []byte(test.body)}); got != test.accepted {
			t.Errorf("acceptance of %s = %v", test.body, got)
		}
	}
}

func TestAlreadyCompletedIsSpecificToGatewayResponseAndDocumentType(t *testing.T) {
	const gatewayBody = `{"validationResults":{"infoMessages":[],"warningMessages":[],"errorMessages":[{"type":"ERROR","code":"Submitted before","category":"Compliance-Check","message":"Compliance check already completed for SIMPLIFIED.","status":"ERROR"}],"status":"ERROR"},"reportingStatus":"NOT_REPORTED"}`
	for _, test := range []struct {
		name, body, sample string
		http               int
		want               bool
	}{
		{"earlier simplified check", gatewayBody, "simplified-invoice", 406, true},
		{"wrong document type", gatewayBody, "standard-invoice", 406, false},
		{"unrelated HTTP error", gatewayBody, "simplified-invoice", 401, false},
		{"other 406", strings.Replace(gatewayBody, "Submitted before", "Unsupported-Version", 1), "simplified-invoice", 406, false},
		{"wrong category", strings.Replace(gatewayBody, "Compliance-Check", "INVOICE_HASHING_ERRORS", 1), "simplified-invoice", 406, false},
		{"mixed errors", strings.Replace(gatewayBody, `"errorMessages":[`, `"errorMessages":[{"code":"invalid-invoice-hash"},`, 1), "simplified-invoice", 406, false},
		{"ambiguous message", strings.Replace(gatewayBody, "Compliance check already completed for SIMPLIFIED.", "Invoice submitted before", 1), "simplified-invoice", 406, false},
		{"malformed JSON", `<html>error</html>`, "simplified-invoice", 406, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := complianceAlreadyCompleted(&ReportingResponse{HTTPStatus: test.http, RawResponse: []byte(test.body)}, test.sample); got != test.want {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestCompleteOnboardingRejectsUnsafeDeviceBeforeGateway(t *testing.T) {
	for _, scenario := range []string{"disabled", "invalid-environment", "inactive", "revoked", "missing-request-id", "invalid-certificate"} {
		t.Run(scenario, func(t *testing.T) {
			db := &completionDB{device: repository.ZatcaDeviceConfig{ID: 3, OrganizationID: 1, IsActive: pgtype.Bool{Bool: true, Valid: true}, ZatcaEnv: pgtype.Text{String: "simulation", Valid: true}, ComplianceCsid: pgtype.Text{String: "not-a-real-certificate", Valid: true}, Metadata: []byte(`{"compliance_request_id":"123","compliance_secret":"secret"}`)}}
			db.org = repository.Organization{ID: 1, Metadata: []byte(`{"zatca":{"enabled":true}}`)}
			cfg := &usecase.ZatcaConfig{Enabled: true, Env: "simulation"}
			switch scenario {
			case "disabled":
				db.org.Metadata = []byte(`{"zatca":{"enabled":false}}`)
			case "invalid-environment":
				db.device.ZatcaEnv.String = "staging"
			case "inactive":
				db.device.IsActive.Bool = false
			case "revoked":
				db.device.IsRevoked = pgtype.Bool{Bool: true, Valid: true}
			case "missing-request-id":
				db.device.Metadata = []byte(`{"compliance_secret":"secret"}`)
			}
			service := testOnboardingService(nil, nil, cfg)
			ctx := context.WithValue(context.Background(), middleware.RepoKey, repository.New(db))
			if _, err := service.CompleteDeviceOnboarding(ctx, 3); err == nil {
				t.Fatal("accepted unsafe device")
			}
			if db.writes != 0 {
				t.Fatal("modified unsafe device")
			}
		})
	}
}
