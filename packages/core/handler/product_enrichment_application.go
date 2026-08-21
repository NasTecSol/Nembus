package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type productEnrichmentApplicationRepository interface {
	GetUser(context.Context, int32) (repository.User, error)
	CheckUserHasPermission(context.Context, repository.CheckUserHasPermissionParams) (bool, error)
}

type productEnrichmentApplicationService interface {
	ApplyApprovedSuggestion(context.Context, int32, int32, int32) (enrichment.ApplicationResult, error)
}

type productEnrichmentApplicationServiceFactory func(productEnrichmentApplicationRepository, *pgxpool.Pool) (productEnrichmentApplicationService, error)

type ProductEnrichmentApplicationHandler struct {
	newService      productEnrichmentApplicationServiceFactory
	repoFromContext func(context.Context) (productEnrichmentApplicationRepository, bool)
	poolFromContext func(context.Context) (*pgxpool.Pool, bool)
}

func NewProductEnrichmentApplicationHandler() *ProductEnrichmentApplicationHandler {
	return &ProductEnrichmentApplicationHandler{
		newService: func(repo productEnrichmentApplicationRepository, pool *pgxpool.Pool) (productEnrichmentApplicationService, error) {
			concreteRepo, ok := repo.(*repository.Queries)
			if !ok || concreteRepo == nil || pool == nil {
				return nil, errors.New("tenant application repository unavailable")
			}
			store := repository.NewProductEnrichmentApplicationStore(concreteRepo, pool)
			return enrichment.NewProductEnrichmentApplicationService(store), nil
		},
		repoFromContext: func(ctx context.Context) (productEnrichmentApplicationRepository, bool) {
			repo, ok := middleware.RepositoryFromContext(ctx)
			return repo, ok
		},
		poolFromContext: middleware.TenantPoolFromContext,
	}
}

type ProductEnrichmentApplicationResponse struct {
	SuggestionID   int32      `json:"suggestion_id"`
	ProductID      int32      `json:"product_id"`
	Status         string     `json:"status"`
	AppliedAt      *time.Time `json:"applied_at,omitempty"`
	ChangedFields  []string   `json:"changed_fields"`
	AlreadyApplied bool       `json:"already_applied"`
}

func (h *ProductEnrichmentApplicationHandler) ApplySuggestion(c *gin.Context) {
	if !requireEmptyApplicationBody(c) {
		return
	}

	userID, ok := parseAuthenticatedUserID(c)
	if !ok {
		return
	}
	if h == nil || h.repoFromContext == nil {
		writeApplicationInternalError(c)
		return
	}
	repo, ok := h.repoFromContext(c.Request.Context())
	if !ok || repo == nil {
		writeApplicationInternalError(c)
		return
	}

	user, err := repo.GetUser(c.Request.Context(), userID)
	if err != nil {
		// Do not disclose whether a user ID exists in another database.
		if errors.Is(err, pgx.ErrNoRows) {
			writeApplicationAuthLookupError(c)
		} else {
			writeApplicationInternalError(c)
		}
		return
	}
	if user.OrganizationID <= 0 {
		writeApplicationInternalError(c)
		return
	}
	if user.IsActive.Valid && !user.IsActive.Bool {
		c.JSON(http.StatusUnauthorized, utils.NewResponse(http.StatusUnauthorized, "authenticated user is inactive", nil))
		return
	}

	hasPermission, err := repo.CheckUserHasPermission(c.Request.Context(), repository.CheckUserHasPermissionParams{
		UserID: user.ID,
		Code:   enrichment.ApplyPermissionCode,
	})
	if err != nil {
		writeApplicationInternalError(c)
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, utils.NewResponse(http.StatusForbidden, "product enrichment apply permission required", gin.H{
			"code": "ENRICHMENT_APPLICATION_PERMISSION_REQUIRED",
		}))
		return
	}

	suggestionID, ok := parseReviewID(c)
	if !ok {
		return
	}
	if h.newService == nil || h.poolFromContext == nil {
		writeApplicationInternalError(c)
		return
	}
	pool, poolOK := h.poolFromContext(c.Request.Context())
	if !poolOK {
		writeApplicationInternalError(c)
		return
	}
	service, err := h.newService(repo, pool)
	if err != nil || service == nil {
		writeApplicationInternalError(c)
		return
	}

	result, err := service.ApplyApprovedSuggestion(c.Request.Context(), user.OrganizationID, suggestionID, user.ID)
	if err != nil {
		writeApplicationError(c, err)
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "enrichment suggestion applied", ProductEnrichmentApplicationResponse{
		SuggestionID:   result.SuggestionID,
		ProductID:      result.ProductID,
		Status:         string(result.Status),
		AppliedAt:      result.AppliedAt,
		ChangedFields:  safeApplicationChangedFields(result.ChangedFields),
		AlreadyApplied: result.AlreadyApplied,
	}))
}

func requireEmptyApplicationBody(c *gin.Context) bool {
	if c.Request.Body == nil {
		return true
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(http.StatusBadRequest, "invalid request body", nil))
		return false
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" || trimmed == "{}" {
		return true
	}
	c.JSON(http.StatusBadRequest, utils.NewResponse(http.StatusBadRequest, "apply accepts an empty body only", nil))
	return false
}

func safeApplicationChangedFields(fields []string) []string {
	allowed := map[string]struct{}{"brand_id": {}, "category_id": {}, "description": {}}
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		if _, ok := allowed[field]; ok {
			result = append(result, field)
		}
	}
	return result
}

func writeApplicationError(c *gin.Context, err error) {
	var applicationErr *enrichment.ApplicationError
	if !errors.As(err, &applicationErr) {
		writeApplicationInternalError(c)
		return
	}
	switch applicationErr.Code {
	case enrichment.ApplicationErrorNotFound:
		c.JSON(http.StatusNotFound, utils.NewResponse(http.StatusNotFound, "enrichment suggestion not found", gin.H{
			"code": "ENRICHMENT_APPLICATION_NOT_FOUND",
		}))
	case enrichment.ApplicationErrorNotApproved:
		writeApplicationConflict(c, "ENRICHMENT_APPLICATION_NOT_APPROVED", "enrichment suggestion is not approved")
	case enrichment.ApplicationErrorStale:
		writeApplicationConflict(c, "ENRICHMENT_APPLICATION_STALE", "enrichment suggestion source is stale")
	case enrichment.ApplicationErrorCanonicalConflict:
		writeApplicationConflict(c, "ENRICHMENT_APPLICATION_TARGET_INVALID", "enrichment application target is invalid")
	case enrichment.ApplicationErrorConditionalConflict:
		writeApplicationConflict(c, "ENRICHMENT_APPLICATION_CONFLICT", "enrichment application conflict")
	case enrichment.ApplicationErrorInvalidProposal:
		writeApplicationConflict(c, "ENRICHMENT_APPLICATION_TARGET_INVALID", "approved enrichment proposal is invalid")
	default:
		writeApplicationInternalError(c)
	}
}

func writeApplicationConflict(c *gin.Context, code, message string) {
	c.JSON(http.StatusConflict, utils.NewResponse(http.StatusConflict, message, gin.H{"code": code}))
}

func writeApplicationAuthLookupError(c *gin.Context) {
	// A tenant-local missing user is treated as an authentication failure; all
	// other lookup failures are intentionally sanitized as infrastructure errors.
	c.JSON(http.StatusUnauthorized, utils.NewResponse(http.StatusUnauthorized, "authenticated user not found", nil))
}

func writeApplicationInternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "enrichment application failed", nil))
}
