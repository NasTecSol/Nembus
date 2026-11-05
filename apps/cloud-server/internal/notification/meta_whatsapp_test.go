package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWhatsAppLanguageCodesToTry(t *testing.T) {
	codes := WhatsAppLanguageCodesToTry("en")
	if len(codes) < 2 || codes[0] != "en" {
		t.Fatalf("en: %#v", codes)
	}
	codes = WhatsAppLanguageCodesToTry("en_US")
	if codes[0] != "en_US" {
		t.Fatalf("en_US primary: %#v", codes)
	}
}

func TestSendTemplateMessageWithHeaderLocalesRetries132001(t *testing.T) {
	var attempts []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		tmpl, _ := body["template"].(map[string]interface{})
		lang, _ := tmpl["language"].(map[string]interface{})
		code, _ := lang["code"].(string)
		attempts = append(attempts, code)
		if code == "en" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": "(#132001) Template name does not exist in the translation",
					"code":    132001,
				},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"messages": []map[string]string{{"id": "wamid.ok"}},
		})
	}))
	defer srv.Close()

	client := NewWhatsAppClient(WhatsAppConfig{PhoneNumberID: "123", AccessToken: "token"})
	client.baseURL = srv.URL
	resp, err := client.SendTemplateMessageWithHeaderLocales(
		context.Background(),
		"+966500000001",
		"nimbus_receipt_txt",
		WhatsAppLanguageCodesToTry("en"),
		[]string{"R-1", "2026-11-05", "10.00 SAR"},
		"",
	)
	if err != nil || resp == nil || len(resp.Messages) != 1 {
		t.Fatalf("resp=%+v err=%v attempts=%v", resp, err, attempts)
	}
	if len(attempts) < 2 || attempts[0] != "en" {
		t.Fatalf("expected en first, got %v", attempts)
	}
}
