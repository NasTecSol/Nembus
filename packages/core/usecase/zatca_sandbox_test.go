package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Opt-in integration test against the public developer sandbox only. It creates
// an ephemeral test key/CSR, uses the documented test OTP, and submits synthetic
// sample data. No tenant database, real credentials or production API is used.
func TestZatcaDeveloperSandboxCompliance(t *testing.T) {
	if os.Getenv("NEMBUS_ZATCA_SANDBOX_TEST") != "1" { t.Skip("set NEMBUS_ZATCA_SANDBOX_TEST=1 to run the live developer sandbox test") }
	ctx,cancel:=context.WithTimeout(context.Background(),55*time.Second); defer cancel()
	client:=&http.Client{Timeout:25*time.Second}
	const base="https://gw-fatoora.zatca.gov.sa/e-invoicing/developer-portal"
	post:=func(path string,payload any,token,secret string)(int,[]byte){
		body,err:=json.Marshal(payload);if err!=nil{t.Fatal(err)}
		req,err:=http.NewRequestWithContext(ctx,"POST",base+path,bytes.NewReader(body));if err!=nil{t.Fatal(err)}
		req.Header.Set("Content-Type","application/json");req.Header.Set("Accept","application/json");req.Header.Set("Accept-Version","V2");req.Header.Set("Accept-Language","en")
		if token==""{req.Header.Set("OTP","123345")}else{req.SetBasicAuth(token,secret)}
		resp,err:=client.Do(req);if err!=nil{t.Fatal(err)};defer resp.Body.Close()
		data,err:=io.ReadAll(resp.Body);if err!=nil{t.Fatal(err)};return resp.StatusCode,data
	}
	key,err:=GenerateECDSAKeyPair();if err!=nil{t.Fatal(err)}
	csr,err:=GenerateCSR(key,ZatcaCSRParams{CommonName:"Nembus-Sandbox-Test",SerialNumber:"1-Nembus|2-POS|3-"+uuid.NewString(),OrganizationIdentifier:"399999999900003",OrgName:"Nembus Sandbox Test",OrgUnit:"Test Branch",CountryCode:"SA",InvoiceType:"1100",Location:"Riyadh",Industry:"Retail",IsSandbox:true})
	if err!=nil{t.Fatal(err)}
	status,body:=post("/compliance",map[string]string{"csr":base64.StdEncoding.EncodeToString([]byte(csr))},"","")
	if status!=200 && status!=201 {t.Fatalf("sandbox compliance certificate request HTTP %d: %s",status,body)}
	var credentials struct{Token string `json:"binarySecurityToken"`;Secret string `json:"secret"`}
	if err:=json.Unmarshal(body,&credentials);err!=nil || credentials.Token=="" || credentials.Secret=="" {t.Fatal("sandbox did not return credential pair")}
	input:=testInvoiceInput();input.SellerVATNumber="399999999900003";input.SellerName="Nembus Sandbox Test";input.ZatcaUUID=uuid.NewString();input.InvoiceNumber="TEST-"+uuid.NewString();input.IssueDate=time.Now().UTC()
	signed,err:=signZatcaDocument(input,key,&repository.ZatcaDeviceConfig{ComplianceCsid:pgtype.Text{String:credentials.Token,Valid:true}});if err!=nil{t.Fatal(err)}
	status,body=post("/compliance/invoices",map[string]string{"invoiceHash":signed.XMLHash,"uuid":signed.ZatcaUUID,"invoice":base64.StdEncoding.EncodeToString(signed.RawXML)},credentials.Token,credentials.Secret)
	var result struct{Validation struct{Status string `json:"status"`;Errors []json.RawMessage `json:"errorMessages"`;Warnings []json.RawMessage `json:"warningMessages"`} `json:"validationResults"`}
	if err:=json.Unmarshal(body,&result);err!=nil{t.Fatalf("sandbox check HTTP %d returned invalid JSON",status)}
	t.Logf("sandbox check HTTP %d: %s",status,body)
	if status!=200 || result.Validation.Status=="ERROR" || len(result.Validation.Errors)>0 || len(result.Validation.Warnings)>0 {t.Fatal("sandbox invoice did not pass without validation errors or warnings")}
}
