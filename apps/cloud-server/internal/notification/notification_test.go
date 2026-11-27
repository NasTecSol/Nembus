package notification

import (
	"context"
	"strings"
	"testing"
)

func TestCleanPhoneNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"+966 50 123 4567", "966501234567"},
		{"966-50-123-4567", "966501234567"},
		{" +966501234567 ", "966501234567"},
	}

	for _, tt := range tests {
		got := cleanPhoneNumber(tt.input)
		if got != tt.expected {
			t.Errorf("cleanPhoneNumber(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestWhatsAppClientConstruction(t *testing.T) {
	client := NewWhatsAppClient(WhatsAppConfig{
		PhoneNumberID: "1234567890",
		AccessToken:   "test_token",
	})
	if client.cfg.ApiVersion != "v19.0" {
		t.Errorf("expected default api version v19.0, got %s", client.cfg.ApiVersion)
	}
}

func TestSESClientConstruction(t *testing.T) {
	client := NewSESClient(SESConfig{
		Host:      "email-smtp.me-central-1.amazonaws.com",
		Username:  "test_user",
		Password:  "test_pass",
		FromEmail: "no-reply@nembus.sa",
	})
	if client.cfg.Port != 587 {
		t.Errorf("expected default port 587, got %d", client.cfg.Port)
	}
	if client.cfg.FromName != "Nembus" {
		t.Errorf("expected default sender name Nembus, got %s", client.cfg.FromName)
	}
}

func TestWebSocketHubLifecycle(t *testing.T) {
	hub := NewWebSocketHub()
	stopCh := make(chan struct{})

	go hub.Run(stopCh)

	// Broadcast when empty should not panic
	hub.BroadcastToOrg(1, map[string]string{"test": "ok"})
	hub.SendToUser(1, 10, map[string]string{"test": "user_ok"})

	close(stopCh)
}

func TestWebSocketHubBroadcastEvictsSlowClientSafely(t *testing.T) {
	hub := NewWebSocketHub()
	stopCh := make(chan struct{})
	go hub.Run(stopCh)
	defer close(stopCh)

	slowClient := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 1),
		OrgID:  7,
		UserID: 42,
	}
	// Pre-fill buffer so broadcast triggers slow-client eviction.
	slowClient.Send <- []byte("prefill")
	hub.register <- slowClient

	// Trigger broadcast eviction and concurrent org/user sends to verify no map race
	// and no panic from sending on a closed client.Send channel.
	hub.broadcast <- []byte("evict")
	for i := 0; i < 20; i++ {
		hub.BroadcastToOrg(7, map[string]int{"i": i})
		hub.SendToUser(7, 42, map[string]int{"i": i})
	}
}

func TestSESClientMissingConfig(t *testing.T) {
	client := NewSESClient(SESConfig{})
	_, err := client.SendEmail(context.Background(), "user@test.com", "Subject", "<p>Hello</p>")
	if err == nil || !strings.Contains(err.Error(), "credentials not configured") {
		t.Errorf("expected credentials error, got %v", err)
	}
}

func TestCustomerCreatedWelcomeMessageFormat(t *testing.T) {
	custName := "Ahmad Al-Ghamdi"
	msg := customerWelcomeMessage(custName)
	if !strings.Contains(msg, custName) {
		t.Fatalf("expected message to contain customer name %q, got: %s", custName, msg)
	}
	emptyMsg := customerWelcomeMessage("")
	if !strings.Contains(emptyMsg, "Valued Customer") {
		t.Fatalf("expected default name fallback, got: %s", emptyMsg)
	}
}
