package notification

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestWhatsAppPolicyGates(t *testing.T) {
	rules := []repository.NotificationRule{{
		Channels:   []byte(`["whatsapp"]`),
		IsActive:   pgtype.Bool{Bool: true, Valid: true},
		Conditions: []byte(`{}`),
	}}
	if !whatsappPolicyAllows(rules, false) {
		t.Fatal("active whatsapp rule should allow a receipt without opt-in")
	}
	rules[0].Conditions = []byte(`{"require_whatsapp_opt_in":true}`)
	if whatsappPolicyAllows(rules, false) {
		t.Fatal("opt-in rule allowed a customer who has not opted in")
	}
	if !whatsappPolicyAllows(rules, true) {
		t.Fatal("opt-in rule rejected an opted-in customer")
	}
	if whatsappPolicyAllows(nil, true) {
		t.Fatal("missing rule was treated as permission to send")
	}
	emailOnly := []repository.NotificationRule{{
		Channels: []byte(`["email_ses"]`),
		IsActive: pgtype.Bool{Bool: true, Valid: true},
	}}
	if whatsappPolicyAllows(emailOnly, true) {
		t.Fatal("email rule allowed WhatsApp")
	}
}

func TestResolveCustomerPhone(t *testing.T) {
	if got := resolveCustomerPhone("966500000001", "966500000002", "966500000003"); got != "966500000001" {
		t.Fatalf("metadata phone: %s", got)
	}
	if got := resolveCustomerPhone("", "966500000002", "966500000003"); got != "966500000002" {
		t.Fatalf("order phone: %s", got)
	}
	if got := resolveCustomerPhone("  ", "", "966500000003"); got != "966500000003" {
		t.Fatalf("customer phone: %s", got)
	}
	if got := resolveCustomerPhone("", "", ""); got != "" {
		t.Fatalf("empty phone: %q", got)
	}
}

func TestReceiptParametersAndSummary(t *testing.T) {
	params := receiptParameters(saleReceipt{
		StoreName: "Main",
		Number:    "POS-9",
		Date:      "2026-11-03",
		Tax:       "1.5",
		Total:     "11.5",
		Currency:  "SAR",
		Lines:     []receiptLine{{Name: "Tea", Qty: "2", Total: "10"}},
	})
	if len(params) != 5 ||
		params[0] != "Main" || params[1] != "POS-9" || params[2] != "2026-11-03" ||
		params[3] != "Tea x2 10.00" || params[4] != "11.50 SAR" {
		t.Fatalf("parameters: %#v", params)
	}
	long := summarizeItems([]receiptLine{{Name: strings.Repeat("A", 1000), Qty: "1", Total: "1"}})
	if len(long) != 900 || !strings.HasSuffix(long, "...") {
		t.Fatalf("summary length %d", len(long))
	}
}

func TestReceiptTemplateIncludesQRHeader(t *testing.T) {
	var got map[string]interface{}
	var mediaHits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/media") {
			mediaHits++
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") || !bytesContains(body, "messaging_product") {
				t.Errorf("media upload was not a WhatsApp multipart request")
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "media-1"})
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode template: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"messages": []map[string]string{{"id": "wamid.1"}},
		})
	}))
	defer srv.Close()

	client := NewWhatsAppClient(WhatsAppConfig{PhoneNumberID: "123", AccessToken: "token"})
	client.baseURL = srv.URL
	png, err := RenderZATCAQR("TLV-QR-PAYLOAD")
	if err != nil || len(png) == 0 {
		t.Fatalf("render qr: %v len=%d", err, len(png))
	}
	mediaID, err := client.UploadMedia(context.Background(), "zatca-qr.png", "image/png", png)
	if err != nil || mediaID != "media-1" || mediaHits != 1 {
		t.Fatalf("upload media id=%s hits=%d err=%v", mediaID, mediaHits, err)
	}
	params := []string{"Main", "POS-9", "2026-11-03", "Tea x2 10.00", "11.50 SAR"}
	resp, err := client.SendTemplateMessageWithHeader(context.Background(), "+966 50 000 0001", "pos_receipt", "en", params, mediaID)
	if err != nil || resp == nil || len(resp.Messages) != 1 || resp.Messages[0].ID != "wamid.1" {
		t.Fatalf("send template: %+v err=%v", resp, err)
	}
	if got["to"] != "966500000001" {
		t.Fatalf("recipient: %#v", got["to"])
	}
	template, _ := got["template"].(map[string]interface{})
	components, _ := template["components"].([]interface{})
	if len(components) != 2 {
		t.Fatalf("components: %#v", template)
	}
	header, _ := components[0].(map[string]interface{})
	if header["type"] != "header" {
		t.Fatalf("header: %#v", header)
	}
	raw, _ := json.Marshal(header)
	if !strings.Contains(string(raw), "media-1") {
		t.Fatalf("header missing media id: %s", raw)
	}
	body, _ := components[1].(map[string]interface{})
	encoded, _ := json.Marshal(body)
	for _, param := range params {
		if !strings.Contains(string(encoded), param) {
			t.Fatalf("body missing %q in %s", param, encoded)
		}
	}
}

func bytesContains(haystack []byte, needle string) bool {
	return strings.Contains(string(haystack), needle)
}
