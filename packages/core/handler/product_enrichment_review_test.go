package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type reviewHandlerRepoFake struct {
	user        repository.User
	userErr     error
	permissions map[string]bool
	rows        map[int32]repository.ProductEnrichmentSuggestion
}

func (r *reviewHandlerRepoFake) GetUser(context.Context, int32) (repository.User, error) {
	return r.user, r.userErr
}

func (r *reviewHandlerRepoFake) CheckUserHasPermission(_ context.Context, params repository.CheckUserHasPermissionParams) (bool, error) {
	return r.permissions[params.Code], nil
}

func (r *reviewHandlerRepoFake) GetProductEnrichmentSuggestionByID(_ context.Context, params repository.GetProductEnrichmentSuggestionByIDParams) (repository.ProductEnrichmentSuggestion, error) {
	row, ok := r.rows[params.ID]
	if !ok || row.OrganizationID != params.OrganizationID {
		return repository.ProductEnrichmentSuggestion{}, pgx.ErrNoRows
	}
	return row, nil
}

type reviewHandlerServiceFake struct {
	listStatus    enrichment.ReviewListStatus
	listCalled    bool
	detailCalled  bool
	approveCalled bool
	rejectCalled  bool
	listItems     []enrichment.ReviewListItem
	detail        enrichment.ReviewDetail
	err           error
}

func (s *reviewHandlerServiceFake) ListSuggestions(_ context.Context, _ int32, status enrichment.ReviewListStatus, _, _ int32) ([]enrichment.ReviewListItem, error) {
	s.listCalled = true
	s.listStatus = status
	return s.listItems, s.err
}

func (s *reviewHandlerServiceFake) GetSuggestion(context.Context, int32, int32) (enrichment.ReviewDetail, error) {
	s.detailCalled = true
	return s.detail, s.err
}

func (s *reviewHandlerServiceFake) ApproveSuggestion(context.Context, int32, int32, int32) (enrichment.ReviewDetail, error) {
	s.approveCalled = true
	return s.detail, s.err
}

func (s *reviewHandlerServiceFake) RejectSuggestion(context.Context, int32, int32, int32) (enrichment.ReviewDetail, error) {
	s.rejectCalled = true
	return s.detail, s.err
}

func reviewHandlerTestUser() repository.User {
	return repository.User{ID: 5, OrganizationID: 41, IsActive: pgtype.Bool{Bool: true, Valid: true}}
}

func reviewHandlerSuggestion(id int32, organizationID int32, status enrichment.SuggestionStatus) repository.ProductEnrichmentSuggestion {
	return repository.ProductEnrichmentSuggestion{ID: id, OrganizationID: organizationID, Status: string(status)}
}

func newReviewHandlerForTest(repo *reviewHandlerRepoFake, service *reviewHandlerServiceFake) *ProductEnrichmentReviewHandler {
	return &ProductEnrichmentReviewHandler{
		repoFromContext: func(context.Context) (productEnrichmentReviewRepository, bool) { return repo, true },
		poolFromContext: func(context.Context) (*pgxpool.Pool, bool) { return nil, true },
		newService: func(productEnrichmentReviewRepository, *pgxpool.Pool) (productEnrichmentReviewService, error) {
			return service, nil
		},
	}
}

func newReviewHandlerTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Params = gin.Params{{Key: "id", Value: "17"}}
	return c, recorder
}

func withReviewHandlerAuth(c *gin.Context) {
	ctx := context.WithValue(c.Request.Context(), middleware.UserIDKey, "5")
	c.Request = c.Request.WithContext(ctx)
}

func reviewPermissions(review, apply bool) map[string]bool {
	return map[string]bool{
		enrichment.ReviewPermissionCode: review,
		enrichment.ApplyPermissionCode:  apply,
	}
}

func TestProductEnrichmentReviewHandlerListReadScopes(t *testing.T) {
	tests := []struct {
		name          string
		review        bool
		apply         bool
		query         string
		wantStatus    int
		wantReadState enrichment.ReviewListStatus
	}{
		{name: "review-only default remains in_review", review: true, query: "", wantStatus: http.StatusOK, wantReadState: enrichment.ReviewStatusInReview},
		{name: "review-only in_review", review: true, query: "?status=in_review", wantStatus: http.StatusOK, wantReadState: enrichment.ReviewStatusInReview},
		{name: "apply-only default is approved", apply: true, query: "", wantStatus: http.StatusOK, wantReadState: enrichment.ReviewStatusApproved},
		{name: "apply-only approved", apply: true, query: "?status=approved", wantStatus: http.StatusOK, wantReadState: enrichment.ReviewStatusApproved},
		{name: "apply-only applied", apply: true, query: "?status=applied", wantStatus: http.StatusOK, wantReadState: enrichment.ReviewStatusApplied},
		{name: "apply-only in_review denied", apply: true, query: "?status=in_review", wantStatus: http.StatusForbidden},
		{name: "neither denied", wantStatus: http.StatusForbidden},
		{name: "both preserve reviewer default", review: true, apply: true, query: "", wantStatus: http.StatusOK, wantReadState: enrichment.ReviewStatusInReview},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &reviewHandlerRepoFake{user: reviewHandlerTestUser(), permissions: reviewPermissions(test.review, test.apply)}
			service := &reviewHandlerServiceFake{}
			h := newReviewHandlerForTest(repo, service)
			c, recorder := newReviewHandlerTestContext(http.MethodGet, "/api/product-enrichment/suggestions"+test.query, "")
			withReviewHandlerAuth(c)

			h.ListSuggestions(c)

			if recorder.Code != test.wantStatus {
				t.Fatalf("unexpected status: got=%d want=%d body=%s", recorder.Code, test.wantStatus, recorder.Body.String())
			}
			if test.wantStatus == http.StatusOK && (!service.listCalled || service.listStatus != test.wantReadState) {
				t.Fatalf("unexpected list scope: called=%v status=%q want=%q", service.listCalled, service.listStatus, test.wantReadState)
			}
			if test.wantStatus != http.StatusOK && service.listCalled {
				t.Fatal("list service must not run for denied read scope")
			}
		})
	}
}

func TestProductEnrichmentReviewHandlerApplyOnlyDetailScope(t *testing.T) {
	tests := []struct {
		name            string
		rowStatus       enrichment.SuggestionStatus
		rowOrganization int32
		wantStatus      int
		serviceCalled   bool
	}{
		{name: "approved", rowStatus: enrichment.SuggestionStatusApproved, rowOrganization: 41, wantStatus: http.StatusOK, serviceCalled: true},
		{name: "applied", rowStatus: enrichment.SuggestionStatusApplied, rowOrganization: 41, wantStatus: http.StatusOK, serviceCalled: true},
		{name: "in_review is scoped not found", rowStatus: enrichment.SuggestionStatusInReview, rowOrganization: 41, wantStatus: http.StatusNotFound},
		{name: "other organization is not visible", rowStatus: enrichment.SuggestionStatusApproved, rowOrganization: 99, wantStatus: http.StatusNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &reviewHandlerRepoFake{
				user:        reviewHandlerTestUser(),
				permissions: reviewPermissions(false, true),
				rows:        map[int32]repository.ProductEnrichmentSuggestion{17: reviewHandlerSuggestion(17, test.rowOrganization, test.rowStatus)},
			}
			service := &reviewHandlerServiceFake{}
			h := newReviewHandlerForTest(repo, service)
			c, recorder := newReviewHandlerTestContext(http.MethodGet, "/api/product-enrichment/suggestions/17", "")
			withReviewHandlerAuth(c)

			h.GetSuggestion(c)

			if recorder.Code != test.wantStatus || service.detailCalled != test.serviceCalled {
				t.Fatalf("unexpected detail authorization: status=%d called=%v body=%s", recorder.Code, service.detailCalled, recorder.Body.String())
			}
		})
	}
}

func TestProductEnrichmentReviewHandlerReviewerDetailRemainsAvailable(t *testing.T) {
	repo := &reviewHandlerRepoFake{user: reviewHandlerTestUser(), permissions: reviewPermissions(true, false)}
	service := &reviewHandlerServiceFake{}
	h := newReviewHandlerForTest(repo, service)
	c, recorder := newReviewHandlerTestContext(http.MethodGet, "/api/product-enrichment/suggestions/17", "")
	withReviewHandlerAuth(c)

	h.GetSuggestion(c)

	if recorder.Code != http.StatusOK || !service.detailCalled {
		t.Fatalf("review permission should preserve detail access: status=%d called=%v", recorder.Code, service.detailCalled)
	}
}

func TestProductEnrichmentReviewHandlerWritePermissionsRemainSeparate(t *testing.T) {
	tests := []struct {
		name        string
		review      bool
		apply       bool
		call        func(*ProductEnrichmentReviewHandler, *gin.Context)
		wantStatus  int
		wantService func(*reviewHandlerServiceFake) bool
	}{
		{name: "review-only approve allowed", review: true, call: (*ProductEnrichmentReviewHandler).ApproveSuggestion, wantStatus: http.StatusOK, wantService: func(s *reviewHandlerServiceFake) bool { return s.approveCalled }},
		{name: "review-only reject allowed", review: true, call: (*ProductEnrichmentReviewHandler).RejectSuggestion, wantStatus: http.StatusOK, wantService: func(s *reviewHandlerServiceFake) bool { return s.rejectCalled }},
		{name: "apply-only approve denied", apply: true, call: (*ProductEnrichmentReviewHandler).ApproveSuggestion, wantStatus: http.StatusForbidden, wantService: func(s *reviewHandlerServiceFake) bool { return !s.approveCalled }},
		{name: "apply-only reject denied", apply: true, call: (*ProductEnrichmentReviewHandler).RejectSuggestion, wantStatus: http.StatusForbidden, wantService: func(s *reviewHandlerServiceFake) bool { return !s.rejectCalled }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &reviewHandlerRepoFake{user: reviewHandlerTestUser(), permissions: reviewPermissions(test.review, test.apply)}
			service := &reviewHandlerServiceFake{}
			h := newReviewHandlerForTest(repo, service)
			c, recorder := newReviewHandlerTestContext(http.MethodPost, "/api/product-enrichment/suggestions/17", "")
			withReviewHandlerAuth(c)

			test.call(h, c)

			if recorder.Code != test.wantStatus || !test.wantService(service) {
				t.Fatalf("write permission separation failed: status=%d service=%+v body=%s", recorder.Code, service, recorder.Body.String())
			}
		})
	}
}

func TestProductEnrichmentReviewHandlerReadAuthorizationSanitizesLookupFailure(t *testing.T) {
	repo := &reviewHandlerRepoFake{user: reviewHandlerTestUser(), permissions: reviewPermissions(false, true), rows: map[int32]repository.ProductEnrichmentSuggestion{}}
	service := &reviewHandlerServiceFake{}
	h := newReviewHandlerForTest(repo, service)
	c, recorder := newReviewHandlerTestContext(http.MethodGet, "/api/product-enrichment/suggestions/17", "")
	withReviewHandlerAuth(c)
	repo.userErr = errors.New("database details must not leak")

	h.GetSuggestion(c)

	if recorder.Code != http.StatusInternalServerError || strings.Contains(recorder.Body.String(), "database details") {
		t.Fatalf("lookup failure was not sanitized: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
