package handler

import (
	"context"
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

type ProductEnrichmentReviewHandler struct {
	newService      productEnrichmentReviewServiceFactory
	repoFromContext func(context.Context) (productEnrichmentReviewRepository, bool)
	poolFromContext func(context.Context) (*pgxpool.Pool, bool)
}

type productEnrichmentReviewRepository interface {
	GetUser(context.Context, int32) (repository.User, error)
	CheckUserHasPermission(context.Context, repository.CheckUserHasPermissionParams) (bool, error)
	GetProductEnrichmentSuggestionByID(context.Context, repository.GetProductEnrichmentSuggestionByIDParams) (repository.ProductEnrichmentSuggestion, error)
}

type productEnrichmentReviewService interface {
	ListSuggestions(context.Context, int32, enrichment.ReviewListStatus, int32, int32) ([]enrichment.ReviewListItem, error)
	GetSuggestion(context.Context, int32, int32) (enrichment.ReviewDetail, error)
	ApproveSuggestion(context.Context, int32, int32, int32) (enrichment.ReviewDetail, error)
	RejectSuggestion(context.Context, int32, int32, int32) (enrichment.ReviewDetail, error)
}

type productEnrichmentReviewServiceFactory func(productEnrichmentReviewRepository, *pgxpool.Pool) (productEnrichmentReviewService, error)

type productEnrichmentReadAuthorization struct {
	repo             productEnrichmentReviewRepository
	pool             *pgxpool.Pool
	organizationID   int32
	reviewPermission bool
	applyPermission  bool
}

func NewProductEnrichmentReviewHandler() *ProductEnrichmentReviewHandler {
	return &ProductEnrichmentReviewHandler{
		repoFromContext: func(ctx context.Context) (productEnrichmentReviewRepository, bool) {
			repo, ok := middleware.RepositoryFromContext(ctx)
			return repo, ok
		},
		poolFromContext: middleware.TenantPoolFromContext,
		newService: func(repo productEnrichmentReviewRepository, pool *pgxpool.Pool) (productEnrichmentReviewService, error) {
			concreteRepo, ok := repo.(*repository.Queries)
			if !ok || concreteRepo == nil || pool == nil {
				return nil, errors.New("tenant review repository unavailable")
			}
			return enrichment.NewReviewService(repository.NewProductEnrichmentReviewStore(concreteRepo, pool)), nil
		},
	}
}

func (h *ProductEnrichmentReviewHandler) ListSuggestions(c *gin.Context) {
	auth, ok := h.authorizeRead(c)
	if !ok {
		return
	}
	statusValue, hasStatus := c.GetQuery("status")
	if !hasStatus {
		if auth.reviewPermission {
			statusValue = string(enrichment.ReviewStatusInReview)
		} else {
			statusValue = string(enrichment.ReviewStatusApproved)
		}
	}
	status := enrichment.ReviewListStatus(statusValue)
	if !auth.reviewPermission && (status != enrichment.ReviewStatusApproved && status != enrichment.ReviewStatusApplied) {
		c.JSON(http.StatusForbidden, utils.NewResponse(http.StatusForbidden, "product enrichment read scope does not allow requested status", gin.H{
			"code": "ENRICHMENT_READ_SCOPE_FORBIDDEN",
		}))
		return
	}
	if !status.Valid() {
		c.JSON(http.StatusBadRequest, utils.NewResponse(http.StatusBadRequest, "invalid review status", nil))
		return
	}
	limit, offset, ok := parseReviewPagination(c)
	if !ok {
		return
	}
	service, err := h.reviewService(auth.repo, auth.pool)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	items, err := service.ListSuggestions(c.Request.Context(), auth.organizationID, status, limit, offset)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "enrichment suggestions fetched successfully", gin.H{
		"items": items, "status": status, "limit": limit, "offset": offset,
	}))
}

func (h *ProductEnrichmentReviewHandler) GetSuggestion(c *gin.Context) {
	auth, ok := h.authorizeRead(c)
	if !ok {
		return
	}
	suggestionID, ok := parseReviewID(c)
	if !ok {
		return
	}
	if !auth.reviewPermission {
		row, err := auth.repo.GetProductEnrichmentSuggestionByID(c.Request.Context(), repository.GetProductEnrichmentSuggestionByIDParams{
			OrganizationID: auth.organizationID,
			ID:             suggestionID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, utils.NewResponse(http.StatusNotFound, "enrichment suggestion not found", nil))
			} else {
				c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "enrichment suggestion lookup failed", nil))
			}
			return
		}
		if enrichment.SuggestionStatus(row.Status) != enrichment.SuggestionStatusApproved && enrichment.SuggestionStatus(row.Status) != enrichment.SuggestionStatusApplied {
			// Keep forbidden statuses indistinguishable from a missing suggestion.
			c.JSON(http.StatusNotFound, utils.NewResponse(http.StatusNotFound, "enrichment suggestion not found", nil))
			return
		}
	}
	service, err := h.reviewService(auth.repo, auth.pool)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	detail, err := service.GetSuggestion(c.Request.Context(), auth.organizationID, suggestionID)
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
	service, err := h.reviewService(repo, pool)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	detail, err := service.ApproveSuggestion(c.Request.Context(), organizationID, suggestionID, reviewerID)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "enrichment suggestion approved and applied", detail))
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
	service, err := h.reviewService(repo, pool)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	detail, err := service.RejectSuggestion(c.Request.Context(), organizationID, suggestionID, reviewerID)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "enrichment suggestion rejected", detail))
}

// ListMachineSuggestions is the SAP Agent-facing review read path. It is
// mounted only behind tenant-bound SAP machine authentication; the browser
// never receives the machine token.
func (h *ProductEnrichmentReviewHandler) ListMachineSuggestions(c *gin.Context) {
	repo, pool, organizationID, ok := h.authorizeMachine(c, enrichment.ReviewPermissionCode)
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
	service, err := h.reviewService(repo, pool)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	items, err := service.ListSuggestions(c.Request.Context(), organizationID, status, limit, offset)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "enrichment suggestions fetched successfully", gin.H{
		"items": items, "status": status, "limit": limit, "offset": offset,
	}))
}

func (h *ProductEnrichmentReviewHandler) GetMachineSuggestion(c *gin.Context) {
	repo, pool, organizationID, ok := h.authorizeMachine(c, enrichment.ReviewPermissionCode)
	if !ok {
		return
	}
	suggestionID, ok := parseReviewID(c)
	if !ok {
		return
	}
	service, err := h.reviewService(repo, pool)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	detail, err := service.GetSuggestion(c.Request.Context(), organizationID, suggestionID)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "enrichment suggestion fetched successfully", detail))
}

func (h *ProductEnrichmentReviewHandler) ApproveMachineSuggestion(c *gin.Context) {
	if !requireEmptyReviewBody(c) {
		return
	}
	repo, pool, organizationID, ok := h.authorizeMachine(c, enrichment.ReviewPermissionCode)
	if !ok {
		return
	}
	suggestionID, ok := parseReviewID(c)
	if !ok {
		return
	}
	service, err := h.reviewService(repo, pool)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	detail, err := service.ApproveSuggestion(c.Request.Context(), organizationID, suggestionID, 0)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "enrichment suggestion approved and applied", detail))
}

func (h *ProductEnrichmentReviewHandler) RejectMachineSuggestion(c *gin.Context) {
	if !requireEmptyReviewBody(c) {
		return
	}
	repo, pool, organizationID, ok := h.authorizeMachine(c, enrichment.ReviewPermissionCode)
	if !ok {
		return
	}
	suggestionID, ok := parseReviewID(c)
	if !ok {
		return
	}
	service, err := h.reviewService(repo, pool)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	detail, err := service.RejectSuggestion(c.Request.Context(), organizationID, suggestionID, 0)
	if err != nil {
		writeReviewError(c, err)
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "enrichment suggestion rejected", detail))
}

func (h *ProductEnrichmentReviewHandler) reviewService(repo productEnrichmentReviewRepository, pool *pgxpool.Pool) (productEnrichmentReviewService, error) {
	if h == nil || h.newService == nil {
		return nil, errors.New("review service unavailable")
	}
	return h.newService(repo, pool)
}

func (h *ProductEnrichmentReviewHandler) authorize(c *gin.Context) (productEnrichmentReviewRepository, *pgxpool.Pool, int32, int32, bool) {
	userID, ok := parseAuthenticatedUserID(c)
	if !ok {
		return nil, nil, 0, 0, false
	}
	if h == nil || h.repoFromContext == nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "tenant repository unavailable", nil))
		return nil, nil, 0, 0, false
	}
	repo, ok := h.repoFromContext(c.Request.Context())
	if !ok || repo == nil {
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
	var pool *pgxpool.Pool
	if h.poolFromContext != nil {
		pool, _ = h.poolFromContext(c.Request.Context())
	}
	return repo, pool, user.OrganizationID, user.ID, true
}

func (h *ProductEnrichmentReviewHandler) authorizeRead(c *gin.Context) (*productEnrichmentReadAuthorization, bool) {
	userID, ok := parseAuthenticatedUserID(c)
	if !ok {
		return nil, false
	}
	if h == nil || h.repoFromContext == nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "tenant repository unavailable", nil))
		return nil, false
	}
	repo, ok := h.repoFromContext(c.Request.Context())
	if !ok || repo == nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "tenant repository unavailable", nil))
		return nil, false
	}
	user, err := repo.GetUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, utils.NewResponse(http.StatusUnauthorized, "authenticated user not found", nil))
		} else {
			c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "authenticated user lookup failed", nil))
		}
		return nil, false
	}
	if user.OrganizationID <= 0 {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "authenticated user organization is invalid", nil))
		return nil, false
	}
	if user.IsActive.Valid && !user.IsActive.Bool {
		c.JSON(http.StatusUnauthorized, utils.NewResponse(http.StatusUnauthorized, "authenticated user is inactive", nil))
		return nil, false
	}
	reviewPermission, err := repo.CheckUserHasPermission(c.Request.Context(), repository.CheckUserHasPermissionParams{
		UserID: user.ID,
		Code:   enrichment.ReviewPermissionCode,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "review permission check failed", nil))
		return nil, false
	}
	applyPermission, err := repo.CheckUserHasPermission(c.Request.Context(), repository.CheckUserHasPermissionParams{
		UserID: user.ID,
		Code:   enrichment.ApplyPermissionCode,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "apply permission check failed", nil))
		return nil, false
	}
	if !reviewPermission && !applyPermission {
		c.JSON(http.StatusForbidden, utils.NewResponse(http.StatusForbidden, "product enrichment read permission required", gin.H{
			"code": "ENRICHMENT_READ_PERMISSION_REQUIRED",
		}))
		return nil, false
	}
	var pool *pgxpool.Pool
	if h.poolFromContext != nil {
		pool, _ = h.poolFromContext(c.Request.Context())
	}
	return &productEnrichmentReadAuthorization{
		repo:             repo,
		pool:             pool,
		organizationID:   user.OrganizationID,
		reviewPermission: reviewPermission,
		applyPermission:  applyPermission,
	}, true
}

func (h *ProductEnrichmentReviewHandler) authorizeMachine(c *gin.Context, requiredScope string) (productEnrichmentReviewRepository, *pgxpool.Pool, int32, bool) {
	identity, ok := middleware.TrustedMachineIdentityFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.NewResponse(http.StatusUnauthorized, "machine authentication required", nil))
		return nil, nil, 0, false
	}
	if !machineHasScope(c, requiredScope) {
		c.JSON(http.StatusForbidden, utils.NewResponse(http.StatusForbidden, "machine scope required", gin.H{"code": "ENRICHMENT_MACHINE_SCOPE_REQUIRED"}))
		return nil, nil, 0, false
	}
	if h == nil || h.repoFromContext == nil || h.poolFromContext == nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "tenant review repository unavailable", nil))
		return nil, nil, 0, false
	}
	repo, ok := h.repoFromContext(c.Request.Context())
	if !ok || repo == nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "tenant review repository unavailable", nil))
		return nil, nil, 0, false
	}
	pool, ok := h.poolFromContext(c.Request.Context())
	if !ok || pool == nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(http.StatusInternalServerError, "tenant review database unavailable", nil))
		return nil, nil, 0, false
	}
	return repo, pool, identity.OrganizationID, true
}

func machineHasScope(c *gin.Context, required string) bool {
	raw, ok := c.Get("scopes")
	if !ok {
		return false
	}
	scopes, ok := raw.([]string)
	if !ok {
		return false
	}
	for _, scope := range scopes {
		// The SAP Agent's existing tenant-bound machine credential is the
		// trusted transport for both migration and its local review UI. A
		// dedicated enrichment scope remains supported for least-privilege
		// deployments; sap:migration is accepted for backward-compatible agent
		// credentials without changing tenant/org binding.
		if scope == required || (required == enrichment.ReviewPermissionCode && scope == "sap:migration") || scope == "*" {
			return true
		}
	}
	return false
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
