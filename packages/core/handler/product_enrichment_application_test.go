package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type applicationHandlerRepoFake struct {
	user           repository.User
	userErr        error
	permission     bool
	permissionErr  error
	permissionCode string
	permissionUser int32
}

func (r *applicationHandlerRepoFake) GetUser(context.Context, int32) (repository.User, error) {
	return r.user, r.userErr
}

func (r *applicationHandlerRepoFake) CheckUserHasPermission(_ context.Context, params repository.CheckUserHasPermissionParams) (bool, error) {
	r.permissionCode = params.Code
	r.permissionUser = params.UserID
	return r.permission, r.permissionErr
}

type applicationHandlerServiceFake struct {
	result       enrichment.ApplicationResult
	err          error
	called       bool
	organization int32
	suggestion   int32
	applier      int32
}

func (s *applicationHandlerServiceFake) ApplyApprovedSuggestion(_ context.Context, organizationID, suggestionID, applierUserID int32) (enrichment.ApplicationResult, error) {
	s.called = true
	s.organization = organizationID
	s.suggestion = suggestionID
	s.applier = applierUserID
	return s.result, s.err
}

func newApplicationHandlerTestContext(body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/product-enrichment/suggestions/17/apply", strings.NewReader(body))
	c.Params = gin.Params{{Key: "id", Value: "17"}}
	return c, recorder
}

func withApplicationAuth(c *gin.Context, userID string) {
	ctx := context.WithValue(c.Request.Context(), middleware.UserIDKey, userID)
	c.Request = c.Request.WithContext(ctx)
}

func newApplicationHandlerForTest(repo *applicationHandlerRepoFake, service *applicationHandlerServiceFake) *ProductEnrichmentApplicationHandler {
	return &ProductEnrichmentApplicationHandler{
		repoFromContext: func(context.Context) (productEnrichmentApplicationRepository, bool) { return repo, true },
		poolFromContext: func(context.Context) (*pgxpool.Pool, bool) { return nil, true },
		newService: func(productEnrichmentApplicationRepository, *pgxpool.Pool) (productEnrichmentApplicationService, error) {
			return service, nil
		},
	}
}

func applicationHandlerTestUser() repository.User {
	return repository.User{ID: 5, OrganizationID: 41, IsActive: pgtype.Bool{Bool: true, Valid: true}}
}

func TestProductEnrichmentApplicationHandlerPassesTrustedIDsToE1(t *testing.T) {
	repo := &applicationHandlerRepoFake{user: applicationHandlerTestUser(), permission: true}
	appliedAt := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	service := &applicationHandlerServiceFake{result: enrichment.ApplicationResult{
		SuggestionID: 17, ProductID: 95, Status: enrichment.SuggestionStatusApplied,
		AppliedAt: &appliedAt, ChangedFields: []string{"brand_id", "description"},
	}}
	h := newApplicationHandlerForTest(repo, service)
	c, recorder := newApplicationHandlerTestContext("")
	withApplicationAuth(c, "5")

	h.ApplySuggestion(c)

	if recorder.Code != http.StatusOK || !service.called {
		t.Fatalf("expected successful E1 application, status=%d called=%v body=%s", recorder.Code, service.called, recorder.Body.String())
	}
	if service.organization != 41 || service.suggestion != 17 || service.applier != 5 {
		t.Fatalf("E1 received unexpected trusted IDs: org=%d suggestion=%d applier=%d", service.organization, service.suggestion, service.applier)
	}
	if repo.permissionCode != enrichment.ApplyPermissionCode || repo.permissionUser != 5 {
		t.Fatalf("permission check was not exact: code=%q user=%d", repo.permissionCode, repo.permissionUser)
	}
	var envelope struct {
		Data ProductEnrichmentApplicationResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Status != string(enrichment.SuggestionStatusApplied) || envelope.Data.ProductID != 95 || envelope.Data.AlreadyApplied || len(envelope.Data.ChangedFields) != 2 {
		t.Fatalf("unexpected safe response: %+v", envelope.Data)
	}
}

func TestProductEnrichmentApplicationHandlerRequiresOnlyApplyPermission(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "review-only user", body: ""},
		{name: "admin role without permission", body: ""},
		{name: "products manage without permission", body: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &applicationHandlerRepoFake{user: applicationHandlerTestUser(), permission: false}
			service := &applicationHandlerServiceFake{}
			h := newApplicationHandlerForTest(repo, service)
			c, recorder := newApplicationHandlerTestContext(test.body)
			withApplicationAuth(c, "5")

			h.ApplySuggestion(c)

			if recorder.Code != http.StatusForbidden || service.called {
				t.Fatalf("expected 403 without apply permission, status=%d called=%v", recorder.Code, service.called)
			}
			if repo.permissionCode != enrichment.ApplyPermissionCode {
				t.Fatalf("handler checked wrong permission: %q", repo.permissionCode)
			}
		})
	}
}

func TestProductEnrichmentApplicationHandlerRejectsClientControlledFields(t *testing.T) {
	repo := &applicationHandlerRepoFake{user: applicationHandlerTestUser(), permission: true}
	service := &applicationHandlerServiceFake{}
	h := newApplicationHandlerForTest(repo, service)
	for _, body := range []string{
		`{"organization_id":999,"applier_user_id":999,"brand_id":1,"description":"client value"}`,
		`{"force":true,"ignore_stale":true,"product_type":"fixed_asset"}`,
	} {
		c, recorder := newApplicationHandlerTestContext(body)
		withApplicationAuth(c, "5")
		h.ApplySuggestion(c)
		if recorder.Code != http.StatusBadRequest || service.called {
			t.Fatalf("expected empty-body rejection, status=%d called=%v body=%s", recorder.Code, service.called, recorder.Body.String())
		}
	}
}

func TestProductEnrichmentApplicationHandlerAlreadyAppliedIsSuccess(t *testing.T) {
	repo := &applicationHandlerRepoFake{user: applicationHandlerTestUser(), permission: true}
	service := &applicationHandlerServiceFake{result: enrichment.ApplicationResult{
		SuggestionID: 17, ProductID: 95, Status: enrichment.SuggestionStatusApplied, AlreadyApplied: true,
	}}
	h := newApplicationHandlerForTest(repo, service)
	c, recorder := newApplicationHandlerTestContext("{}")
	withApplicationAuth(c, "5")

	h.ApplySuggestion(c)

	if recorder.Code != http.StatusOK || !service.called || !strings.Contains(recorder.Body.String(), `"already_applied":true`) {
		t.Fatalf("expected idempotent 200 response, status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestProductEnrichmentApplicationHandlerMissingTenantUserFailsSafely(t *testing.T) {
	repo := &applicationHandlerRepoFake{userErr: pgx.ErrNoRows}
	service := &applicationHandlerServiceFake{}
	h := newApplicationHandlerForTest(repo, service)
	c, recorder := newApplicationHandlerTestContext("")
	withApplicationAuth(c, "5")

	h.ApplySuggestion(c)

	if recorder.Code != http.StatusUnauthorized || service.called || strings.Contains(recorder.Body.String(), "SQL") {
		t.Fatalf("expected safe tenant-local user failure, status=%d called=%v body=%s", recorder.Code, service.called, recorder.Body.String())
	}
}

func TestProductEnrichmentApplicationHandlerRejectsMalformedSuggestionID(t *testing.T) {
	repo := &applicationHandlerRepoFake{user: applicationHandlerTestUser(), permission: true}
	service := &applicationHandlerServiceFake{}
	h := newApplicationHandlerForTest(repo, service)
	c, recorder := newApplicationHandlerTestContext("")
	c.Params = gin.Params{{Key: "id", Value: "not-an-id"}}
	withApplicationAuth(c, "5")

	h.ApplySuggestion(c)

	if recorder.Code != http.StatusBadRequest || service.called {
		t.Fatalf("expected malformed ID rejection, status=%d called=%v", recorder.Code, service.called)
	}
}

func TestProductEnrichmentApplicationHandlerMapsE1Errors(t *testing.T) {
	tests := []struct {
		name string
		code enrichment.ApplicationErrorCode
		want int
	}{
		{name: "not found", code: enrichment.ApplicationErrorNotFound, want: http.StatusNotFound},
		{name: "not approved", code: enrichment.ApplicationErrorNotApproved, want: http.StatusConflict},
		{name: "stale", code: enrichment.ApplicationErrorStale, want: http.StatusConflict},
		{name: "target", code: enrichment.ApplicationErrorCanonicalConflict, want: http.StatusConflict},
		{name: "conditional", code: enrichment.ApplicationErrorConditionalConflict, want: http.StatusConflict},
		{name: "invalid proposal", code: enrichment.ApplicationErrorInvalidProposal, want: http.StatusConflict},
		{name: "persistence", code: enrichment.ApplicationErrorPersistence, want: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &applicationHandlerRepoFake{user: applicationHandlerTestUser(), permission: true}
			service := &applicationHandlerServiceFake{err: &enrichment.ApplicationError{Code: test.code, Err: errors.New("must not leak")}}
			h := newApplicationHandlerForTest(repo, service)
			c, recorder := newApplicationHandlerTestContext("")
			withApplicationAuth(c, "5")

			h.ApplySuggestion(c)

			if recorder.Code != test.want || strings.Contains(recorder.Body.String(), "must not leak") {
				t.Fatalf("unexpected sanitized mapping: status=%d body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestProductEnrichmentApplicationHandlerRejectsUnauthenticatedRequest(t *testing.T) {
	repo := &applicationHandlerRepoFake{user: applicationHandlerTestUser(), permission: true}
	service := &applicationHandlerServiceFake{}
	h := newApplicationHandlerForTest(repo, service)
	c, recorder := newApplicationHandlerTestContext("")

	h.ApplySuggestion(c)

	if recorder.Code != http.StatusUnauthorized || service.called {
		t.Fatalf("expected unauthenticated request to stop before E1, status=%d called=%v", recorder.Code, service.called)
	}
}
