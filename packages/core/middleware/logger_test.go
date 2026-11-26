package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

type captureSlogHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (h *captureSlogHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *captureSlogHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, r.Clone())
	return nil
}
func (h *captureSlogHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureSlogHandler) WithGroup(_ string) slog.Handler      { return h }

func TestOTelGinMiddlewareQuietRoutesAndErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orig := slog.Default()
	capHandler := &captureSlogHandler{}
	slog.SetDefault(slog.New(capHandler))
	defer slog.SetDefault(orig)

	r := gin.New()
	r.Use(OTelGinMiddleware("test-service"))
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.GET("/api/v1/logs", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})
	r.POST("/api/print/receipt", func(c *gin.Context) {
		_ = c.Error(errors.New("print failed: printer offline"))
		c.Status(http.StatusInternalServerError)
	})
	r.GET("/api/orders", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 1. Quiet route returning 200 OK should NOT emit an INFO log
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if len(capHandler.records) != 0 {
		t.Fatalf("expected 0 log records for 200 /health, got %d", len(capHandler.records))
	}

	// 2. Quiet route returning 500 MUST still emit an ERROR log
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/logs", nil))
	if len(capHandler.records) != 1 || capHandler.records[0].Level != slog.LevelError {
		t.Fatalf("expected 1 ERROR log for 500 /api/v1/logs, got %+v", capHandler.records)
	}

	// 3. Print receipt 500 with c.Error must include the error attribute
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/print/receipt", nil))
	if len(capHandler.records) != 2 {
		t.Fatalf("expected 2 total records, got %d", len(capHandler.records))
	}
	var gotErrAttr string
	capHandler.records[1].Attrs(func(a slog.Attr) bool {
		if a.Key == "error" {
			gotErrAttr = a.Value.String()
		}
		return true
	})
	if gotErrAttr == "" {
		t.Fatalf("expected non-empty error attribute on 500 /api/print/receipt")
	}

	// 4. Normal business route returning 200 OK emits INFO log
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/orders", nil))
	if len(capHandler.records) != 3 || capHandler.records[2].Level != slog.LevelInfo {
		t.Fatalf("expected INFO log for 200 /api/orders, got %+v", capHandler.records)
	}
}
