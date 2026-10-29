package handler

import (
	"net/http"
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
)

type BiometricHandler struct {
	useCase *usecase.BiometricUseCase
}

func NewBiometricHandler(uc *usecase.BiometricUseCase) *BiometricHandler {
	return &BiometricHandler{useCase: uc}
}

func (h *BiometricHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// EnrollFinger handles POST /api/users/:id/biometrics/enroll (and POST /api/biometrics/enroll)
// @Summary      Enroll user fingerprint template
// @Description  Stores a consolidated ISO/IEC 19794-2:2005 FMD after organization-wide 1:N duplicate check. Requires biometrics.enroll permission or X-Authorization-Token scoped to biometrics.enroll.
// @Tags         biometrics
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization header   string  true  "Bearer JWT"
// @Param        X-Authorization-Token header string false "Optional supervisor override token (biometrics.enroll)"
// @Param        id           path      int     false "User ID (when using /users/:id/biometrics/enroll)"
// @Param        request      body      EnrollFingerRequest  true  "Enrollment payload"
// @Success      200  {object}  EnrollFingerResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Router       /api/users/{id}/biometrics/enroll [post]
// @Router       /api/biometrics/enroll [post]
func (h *BiometricHandler) EnrollFinger(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	actorID, ok := middleware.AbortUnlessBiometricEnrollPermission(c, repo)
	if !ok {
		return
	}

	var req usecase.EnrollFingerInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body: "+err.Error(), nil))
		return
	}

	if idParam := c.Param("id"); idParam != "" && req.UserID <= 0 {
		if parsed, err := strconv.ParseInt(idParam, 10, 32); err == nil {
			req.UserID = int32(parsed)
		}
	}

	if req.EnrolledByUserID == nil && actorID > 0 {
		req.EnrolledByUserID = &actorID
	}
	if req.EnrolledByUserID == nil {
		if actorStr := c.GetString("user_id"); actorStr != "" {
			if parsed, err := strconv.ParseInt(actorStr, 10, 32); err == nil && parsed > 0 {
				actorID := int32(parsed)
				req.EnrolledByUserID = &actorID
			}
		}
	}

	resp := h.useCase.EnrollFinger(c.Request.Context(), req)
	c.JSON(resp.StatusCode, resp)
}

// ListUserBiometrics handles GET /api/users/:id/biometrics
// @Summary      List enrolled fingerprints for a user
// @Description  Returns active finger indices and metadata without raw FMD bytes.
// @Tags         biometrics
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        id           path      int     true  "User ID"
// @Success      200  {object}  ListUserBiometricsResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /api/users/{id}/biometrics [get]
func (h *BiometricHandler) ListUserBiometrics(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	resp := h.useCase.ListUserBiometrics(c.Request.Context(), int32(userID))
	c.JSON(resp.StatusCode, resp)
}

// DeleteUserBiometric handles DELETE /api/users/:id/biometrics/:fingerIndex
// @Summary      Deactivate a user fingerprint template
// @Description  Soft-deactivates the template (is_active=false) for sync propagation. Requires biometrics.enroll permission.
// @Tags         biometrics
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        id           path      int     true  "User ID"
// @Param        fingerIndex  path      int     true  "NIST finger index (1-10)"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/users/{id}/biometrics/{fingerIndex} [delete]
func (h *BiometricHandler) DeleteUserBiometric(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	if _, ok := middleware.AbortUnlessBiometricEnrollPermission(c, repo); !ok {
		return
	}
	userID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}
	fingerIdx, err := strconv.ParseInt(c.Param("fingerIndex"), 10, 16)
	if err != nil || fingerIdx < 1 || fingerIdx > 10 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "finger_index must be between 1 and 10", nil))
		return
	}

	resp := h.useCase.DeleteUserBiometric(c.Request.Context(), int32(userID), int16(fingerIdx))
	c.JSON(resp.StatusCode, resp)
}

