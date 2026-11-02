package router

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/handler"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type referenceTypeRouteDB struct{ repository.DBTX }

func (referenceTypeRouteDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return referenceTypeRouteRow{}
}

type referenceTypeRouteRow struct{}

func (referenceTypeRouteRow) Scan(...any) error {
	return errors.New("reference type count reached")
}

func TestStockMovementReferenceTypeRoutes(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), middleware.RepoKey, repository.New(referenceTypeRouteDB{}))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	RegisterStockMovementRoutes(r.Group("/api"), handler.NewStockMovementHandler(usecase.NewStockMovementsUseCase()))
	for _, path := range []string{"reference-type", "by-reference-type"} {
		for _, tc := range []struct {
			query   string
			status  int
			message string
		}{
			{"", http.StatusBadRequest, "reference_type is required"},
			{"?reference_type=invoice&page=1&limit=100", http.StatusInternalServerError, "reference type count reached"},
		} {
			t.Run(path+tc.query, func(t *testing.T) {
				w := httptest.NewRecorder()
				r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/stock-movements/"+path+tc.query, nil))
				if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.message) {
					t.Fatalf("unexpected route response %d: %s", w.Code, w.Body.String())
				}
			})
		}
	}
}
