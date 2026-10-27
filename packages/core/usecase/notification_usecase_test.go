package usecase

import (
	"testing"
)

func TestNotificationTemplateInterpolation(t *testing.T) {
	uc := NewNotificationUseCase(nil)

	subjectTmpl := "Order #{{order_number}} Approved"
	bodyTmpl := "Hello {{customer_name}}, your order #{{order_number}} of SAR {{total_amount}} has been approved."

	data := map[string]interface{}{
		"order_number":  "SO-2026-0099",
		"customer_name": "Nasser Al-Otaibi",
		"total_amount":  "1550.00",
	}

	subj, body := uc.RenderTemplate(subjectTmpl, bodyTmpl, data)

	expectedSubj := "Order #SO-2026-0099 Approved"
	if subj != expectedSubj {
		t.Errorf("RenderTemplate subject = %q, want %q", subj, expectedSubj)
	}

	expectedBody := "Hello Nasser Al-Otaibi, your order #SO-2026-0099 of SAR 1550.00 has been approved."
	if body != expectedBody {
		t.Errorf("RenderTemplate body = %q, want %q", body, expectedBody)
	}
}

func TestNotificationTemplateMissingKey(t *testing.T) {
	uc := NewNotificationUseCase(nil)

	subjectTmpl := "Receipt #{{receipt_number}} for {{missing_key}}"
	bodyTmpl := "Total: {{total}}"

	data := map[string]interface{}{
		"receipt_number": "TX-1001",
	}

	subj, body := uc.RenderTemplate(subjectTmpl, bodyTmpl, data)

	if subj != "Receipt #TX-1001 for {{missing_key}}" {
		t.Errorf("expected missing key unreplaced, got %q", subj)
	}
	if body != "Total: {{total}}" {
		t.Errorf("expected missing total unreplaced, got %q", body)
	}
}
