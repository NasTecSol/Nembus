package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductEnrichmentReviewHandler struct{}

func NewProductEnrichmentReviewHandler() *ProductEnrichmentReviewHandler {
	return &ProductEnrichmentReviewHandler{}
}

func (h *ProductEnrichmentReviewHandler) ListSuggestions(c *gin.Context) {
	repo, pool, organizationID, _, ok := h.authorize(c)
	if !ok {
		return
	}
	status := enrichment.ReviewListStatus(c.DefaultQuery("status", string(enrichment.ReviewStatusInReview)))
	if !status.Valid() {
		c.JSON(http.StatusBadRequest, utils.NewResponse(http.StatusBadRequest, "invalid review status", nil))
		return
	}
	limit, offset, ok := parseReviewPagination(c)
	if !ok {
		return
	}
	service := enrichment.NewReviewService(repository.NewProductEnrichmentReviewStore(repo, pool))
	items, err := service.ListSuggestions(c.Request.Context(), organizationID, status, limit, offset)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "enrichment suggestions fetched successfully", gin.H{
		"items": items, "status": status, "limit": limit, "offset": offset,
	}))
}

func (h *ProductEnrichmentReviewHandler) GetSuggestion(c *gin.Context) {
	repo, pool, organizationID, _, ok := h.authorize(c)
	if !ok {
		return
	}
	suggestionID, ok := parseReviewID(c)
	if !ok {
		return
	}
	service := enrichment.NewReviewService(repository.NewProductEnrichmentReviewStore(repo, pool))
	detail, err := service.GetSuggestion(c.Request.Context(), organizationID, suggestionID)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "enrichment suggestion fetched successfully", detail))
}

func (h *ProductEnrichmentReviewHandler) ApproveSuggestion(c *gin.Context) {
	if !requireEmptyReviewBody(c) {
		return
	}
	repo, pool, organizationID, reviewerID, ok := h.authorize(c)
	if !ok {
		return
	}
	suggestionID, ok := parseReviewID(c)
	if !ok {
		return
	}
	service := enrichment.NewReviewService(repository.NewProductEnrichmentReviewStore(repo, pool))
	detail, err := service.ApproveSuggestion(c.Request.Context(), organizationID, suggestionID, reviewerID)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "enrichment suggestion approved", detail))
}

func (h *ProductEnrichmentReviewHandler) RejectSuggestion(c *gin.Context) {
	if !requireEmptyReviewBody(c) {
		return
	}
	repo, pool, organizationID, reviewerID, ok := h.authorize(c)
	if !ok {
		return
	}
	suggestionID, ok := parseReviewID(c)
	if !ok {
		return
	}
	service := enrichment.NewReviewService(repository.NewProductEnrichmentReviewStore(repo, pool))
	detail, err := service.RejectSuggestion(c.Request.Context(), organizationID, suggestionID, reviewerID)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "enrichment suggestion rejected", detail))
}

func (h *ProductEnrichmentReviewHandler) authorize(c *gin.Context) (*repository.Queries, *pgxpool.Pool, int32, int32, bool) {
	userID, ok := parseAuthenticatedUserID(c)
	if !ok {
		return nil, nil, 0, 0, false
	}
	repo, ok := middleware.RepositoryFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "tenant repository unavailable", nil))
		return nil, nil, 0, 0, false
	}
	user, err := repo.GetUser(c.Request.Context(), userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusUnauthorized, utils.NewResponse(http.StatusUnauthorized, "authenticated user not found", nil))
		} else {
			c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "authenticated user lookup failed", nil))
		}
		return nil, nil, 0, 0, false
	}
	if user.OrganizationID <= 0 {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "authenticated user organization is invalid", nil))
		return nil, nil, 0, 0, false
	}
	if user.IsActive.Valid && !user.IsActive.Bool {
		c.JSON(http.StatusUnauthorized, utils.NewResponse(http.StatusUnauthorized, "authenticated user is inactive", nil))
		return nil, nil, 0, 0, false
	}
	hasPermission, err := repo.CheckUserHasPermission(c.Request.Context(), repository.CheckUserHasPermissionParams{
		UserID: user.ID, Code: enrichment.ReviewPermissionCode,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "review permission check failed", nil))
		return nil, nil, 0, 0, false
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, utils.NewResponse(http.StatusForbidden, "product enrichment review permission required", nil))
		return nil, nil, 0, 0, false
	}
	pool, _ := middleware.TenantPoolFromContext(c.Request.Context())
	return repo, pool, user.OrganizationID, user.ID, true
}

func parseAuthenticatedUserID(c *gin.Context) (int32, bool) {
	raw, ok := middleware.AuthenticatedUserID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.NewResponse(http.StatusUnauthorized, "authentication required", nil))
		return 0, false
	}
	parsed, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || parsed <= 0 {
		c.JSON(http.StatusUnauthorized, utils.NewResponse(http.StatusUnauthorized, "invalid authenticated user", nil))
		return 0, false
	}
	return int32(parsed), true
}

func parseReviewID(c *gin.Context) (int32, bool) {
	parsed, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || parsed <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(http.StatusBadRequest, "invalid enrichment suggestion id", nil))
		return 0, false
	}
	return int32(parsed), true
}

func parseReviewPagination(c *gin.Context) (int32, int32, bool) {
	limit, err := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 32)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(http.StatusBadRequest, "limit must be between 1 and 100", nil))
		return 0, 0, false
	}
	offset, err := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)
	if err != nil || offset < 0 || offset > 100000 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(http.StatusBadRequest, "offset must be between 0 and 100000", nil))
		return 0, 0, false
	}
	return int32(limit), int32(offset), true
}

func requireEmptyReviewBody(c *gin.Context) bool {
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
	var fields map[string]json.RawMessage
	if json.Unmarshal(body, &fields) == nil && len(fields) == 0 {
		return true
	}
	c.JSON(http.StatusBadRequest, utils.NewResponse(http.StatusBadRequest, "approve and reject accept an empty body only", nil))
	return false
}

func writeReviewError(c *gin.Context, err error) {
	var reviewErr *enrichment.ReviewError
	if errors.As(err, &reviewErr) {
		switch reviewErr.Code {
		case enrichment.ReviewErrorNotFound:
			c.JSON(http.StatusNotFound, utils.NewResponse(http.StatusNotFound, "enrichment suggestion not found", nil))
		case enrichment.ReviewErrorConflict:
			c.JSON(http.StatusConflict, utils.NewResponse(http.StatusConflict, "enrichment suggestion review conflict", gin.H{"code": "ENRICHMENT_SUGGESTION_CONFLICT"}))
		case enrichment.ReviewErrorStale:
			c.JSON(http.StatusConflict, utils.NewResponse(http.StatusConflict, "enrichment suggestion source is stale", gin.H{"code": "ENRICHMENT_SUGGESTION_STALE"}))
		case enrichment.ReviewErrorNotReviewable:
			c.JSON(http.StatusConflict, utils.NewResponse(http.StatusConflict, "enrichment suggestion is not reviewable", gin.H{"code": "ENRICHMENT_SUGGESTION_NOT_REVIEWABLE"}))
		case enrichment.ReviewErrorBadRequest:
			c.JSON(http.StatusBadRequest, utils.NewResponse(http.StatusBadRequest, "invalid enrichment review request", nil))
		default:
			c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "enrichment review operation failed", nil))
		}
		return
	}
	c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "enrichment review operation failed", nil))
}
