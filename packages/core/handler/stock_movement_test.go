package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
)

func TestStockMovementsByReferenceTypeRequiresType(t *testing.T) {
	for _, query := range []string{"", "?reference_type=", "?reference_type=%20%20"} {
		t.Run(query, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/stock-movements/by-reference-type"+query, nil)
			h := NewStockMovementHandler(usecase.NewStockMovementsUseCase())
			h.ListStockMovementsByReferenceType(ctx)
			if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "reference_type is required") {
				t.Fatalf("expected missing reference type error, got %d: %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}
