package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/handler"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
)

func TestInventoryTotalsRoute(t *testing.T) {
	r := gin.New()
	RegisterInventoryStockRoutes(r.Group("/api"), handler.NewInventoryStockHandler(usecase.NewInventoryStockUseCase()))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/inventory-stock/totals", nil))
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "organization_id") {
		t.Fatalf("totals route did not reach filter validation: %d %s", w.Code, w.Body.String())
	}
}
