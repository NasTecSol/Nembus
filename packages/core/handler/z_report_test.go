package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/gin-gonic/gin"
)

func TestZReportRequestValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, query, user string
		org               int32
		status            int
	}{
		{"missing auth", "?cashier_session_id=1", "", 0, 401},
		{"invalid user", "?cashier_session_id=1", "abc", 1, 401},
		{"missing session", "", "1", 1, 400},
		{"empty session", "?cashier_session_id=", "1", 1, 400},
		{"negative session", "?cashier_session_id=-1", "1", 1, 400},
		{"overflow", "?cashier_session_id=2147483648", "1", 1, 400},
		{"duplicate", "?cashier_session_id=1&cashier_session_id=2", "1", 1, 400},
		{"bad optional", "?cashier_session_id=1&store_id=abc", "1", 1, 400},
		{"zero optional", "?cashier_session_id=1&pos_terminal_id=0", "1", 1, 400},
		{"missing repository", "?cashier_session_id=1&store_id=3&cashier_id=12&pos_terminal_id=7", "1", 1, 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/api/reports/z-report"+tc.query, nil)
			if tc.org > 0 {
				c.Set(string(middleware.OrgIDKey), tc.org)
			}
			if tc.user != "" {
				c.Set(string(middleware.UserIDKey), tc.user)
			}
			h := NewCashierSessionHandler(nil)
			h.GetZReport(c)
			if w.Code != tc.status {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("report must not be cached")
			}
		})
	}
}
