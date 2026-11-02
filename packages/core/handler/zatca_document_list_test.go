package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type documentListDB struct {
	repository.DBTX
	total                     int64
	countOrg, listOrg         int32
	limit                     int32
	offset                    int64
	err                       error
	countCalls, listCalls     int
	countFilters, listFilters repository.ZatcaDocumentFilters
}

type documentCountRow struct {
	total int64
	err   error
}

func (r documentCountRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	*dest[0].(*int64) = r.total
	return nil
}

func (db *documentListDB) QueryRow(_ context.Context, sql string, args ...interface{}) pgx.Row {
	if !strings.Contains(sql, "CountZatcaDocuments") {
		panic("unexpected count query")
	}
	db.countCalls++
	db.countOrg = args[0].(int32)
	db.countFilters = repository.ZatcaDocumentFilters{StoreID: args[1].(int32), PosTerminalID: args[2].(int32)}
	return documentCountRow{db.total, db.err}
}

type documentListRows struct {
	pgx.Rows
	remaining int64
	id        int64
}

func (r *documentListRows) Close()     {}
func (r *documentListRows) Err() error { return nil }
func (r *documentListRows) Next() bool { return r.remaining > 0 }
func (r *documentListRows) Scan(dest ...interface{}) error {
	err := (documentStatusRow{entry: repository.ZatcaDocumentChain{
		ID: r.id, OrganizationID: 7, EntityType: "sales_order_v2",
		ZatcaStatus:   repository.NullZatcaDocStatus{ZatcaDocStatus: repository.ZatcaDocStatusReported, Valid: true},
		ZatcaResponse: []byte(`{"reportingStatus":"REPORTED"}`),
	}}).Scan(dest[:17]...)
	*dest[17].(*pgtype.Int4) = pgtype.Int4{Int32: 3, Valid: true}
	*dest[18].(*pgtype.Int4) = pgtype.Int4{Int32: 8, Valid: true}
	r.remaining--
	r.id--
	return err
}
func (db *documentListDB) Query(_ context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if !strings.Contains(sql, "ORDER BY id DESC LIMIT $4 OFFSET $5") {
		panic("pagination must be ordered and applied in SQL")
	}
	db.listCalls++
	db.listOrg, db.limit, db.offset = args[0].(int32), args[3].(int32), args[4].(int64)
	db.listFilters = repository.ZatcaDocumentFilters{StoreID: args[1].(int32), PosTerminalID: args[2].(int32)}
	remaining := db.total - db.offset
	if remaining > int64(db.limit) {
		remaining = int64(db.limit)
	}
	return &documentListRows{remaining: remaining, id: db.total - db.offset}, nil
}

func TestZatcaDocumentListPagination(t *testing.T) {
	for _, tc := range []struct {
		name, query                       string
		total, page, limit, offset, pages int64
		length                            int
		next                              bool
	}{
		{"defaults", "", 125, 1, 50, 0, 3, 50, true},
		{"more than fifty", "?page=2&limit=100", 125, 2, 100, 100, 2, 25, false},
		{"custom limit", "?limit=75", 125, 1, 75, 0, 2, 75, true},
		{"out of range", "?page=4", 125, 4, 50, 150, 3, 0, false},
		{"empty", "", 0, 1, 50, 0, 0, 0, false},
		{"large offset", "?page=2147483647&limit=2147483647", 125, 2147483647, 2147483647, 4611686011984936962, 1, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &documentListDB{total: tc.total}
			engine := gin.New()
			engine.Use(func(c *gin.Context) {
				c.Set(string(middleware.OrgIDKey), int32(7))
				c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), middleware.RepoKey, repository.New(db)))
			})
			engine.GET("/documents", NewZatcaHandler(nil).ListZatcaDocuments)
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, httptest.NewRequest("GET", "/documents"+tc.query, nil))
			if recorder.Code != 200 {
				t.Fatalf("body=%s", recorder.Body)
			}
			var body struct {
				Data struct {
					Documents          []repository.ZatcaDocumentChain `json:"documents"`
					Page, Limit, Total int64
					TotalPages         int64 `json:"total_pages"`
					HasNextPage        bool  `json:"has_next_page"`
				}
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			data := body.Data
			if data.Documents == nil || len(data.Documents) != tc.length || data.Page != tc.page || data.Limit != tc.limit || data.Total != tc.total || data.TotalPages != tc.pages || data.HasNextPage != tc.next {
				t.Fatalf("unexpected page: %s", recorder.Body)
			}
			if db.limit != int32(tc.limit) || db.offset != tc.offset || db.countOrg != 7 || db.listOrg != 7 || db.countCalls != 1 || db.listCalls != 1 {
				t.Fatalf("incorrect pagination or organization: %+v", db)
			}
			if len(data.Documents) > 0 && (data.Documents[0].ZatcaStatus.ZatcaDocStatus != repository.ZatcaDocStatusReported || string(data.Documents[0].ZatcaResponse) != `{"reportingStatus":"REPORTED"}`) {
				t.Fatal("saved response changed")
			}
		})
	}
}

func TestZatcaDocumentListInvalidPagination(t *testing.T) {
	engine := gin.New()
	engine.GET("/documents", NewZatcaHandler(nil).ListZatcaDocuments)
	for _, query := range []string{"page=0", "page=-1", "limit=0", "limit=-5", "limit=abc", "page=1.5", "limit=2147483648", "page=2147483648", "page="} {
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, httptest.NewRequest("GET", "/documents?"+query, nil))
		if recorder.Code != 400 {
			t.Fatalf("%s: %d %s", query, recorder.Code, recorder.Body)
		}
	}
}

func TestZatcaDocumentListMissingTenantAndDatabaseFailure(t *testing.T) {
	for _, missing := range []bool{true, false} {
		engine := gin.New()
		if !missing {
			engine.Use(func(c *gin.Context) {
				c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), middleware.RepoKey, repository.New(&documentListDB{err: errors.New("database offline")})))
			})
		}
		engine.Use(func(c *gin.Context) { c.Set(string(middleware.OrgIDKey), int32(7)) })
		engine.GET("/documents", NewZatcaHandler(nil).ListZatcaDocuments)
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, httptest.NewRequest("GET", "/documents", nil))
		if recorder.Code != 500 {
			t.Fatalf("body=%s", recorder.Body)
		}
	}
}

func TestZatcaDocumentListLocationFilters(t *testing.T) {
	for _, tc := range []struct {
		query   string
		filters repository.ZatcaDocumentFilters
	}{
		{"store_id=3", repository.ZatcaDocumentFilters{StoreID: 3}},
		{"pos_terminal_id=8", repository.ZatcaDocumentFilters{PosTerminalID: 8}},
		{"store_id=3&pos_terminal_id=8", repository.ZatcaDocumentFilters{StoreID: 3, PosTerminalID: 8}},
	} {
		t.Run(tc.query, func(t *testing.T) {
			db := &documentListDB{total: 3}
			engine := gin.New()
			engine.Use(func(c *gin.Context) {
				c.Set(string(middleware.OrgIDKey), int32(7))
				c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), middleware.RepoKey, repository.New(db)))
			})
			engine.GET("/documents", NewZatcaHandler(nil).ListZatcaDocuments)
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, httptest.NewRequest("GET", "/documents?"+tc.query+"&page=2&limit=2", nil))
			if recorder.Code != 200 || db.countFilters != tc.filters || db.listFilters != tc.filters || db.offset != 2 {
				t.Fatalf("filters not passed to count and list: %+v, body=%s", db, recorder.Body)
			}
			var body struct {
				Data struct {
					Documents  []repository.ZatcaDocumentListEntry `json:"documents"`
					Total      int64                               `json:"total"`
					TotalPages int64                               `json:"total_pages"`
				}
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if len(body.Data.Documents) != 1 || body.Data.Total != 3 || body.Data.TotalPages != 2 || body.Data.Documents[0].StoreID.Int32 != 3 || body.Data.Documents[0].PosTerminalID.Int32 != 8 {
				t.Fatalf("incorrect filtered response: %s", recorder.Body)
			}
		})
	}
}

func TestZatcaDocumentListInvalidLocationFilters(t *testing.T) {
	engine := gin.New()
	engine.GET("/documents", NewZatcaHandler(nil).ListZatcaDocuments)
	for _, field := range []string{"store_id", "pos_terminal_id"} {
		for _, value := range []string{"", "0", "-1", "abc", "1.5", "2147483648", "3&" + field + "=8"} {
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, httptest.NewRequest("GET", "/documents?"+field+"="+value, nil))
			if recorder.Code != 400 {
				t.Fatalf("%s=%s: %d %s", field, value, recorder.Code, recorder.Body)
			}
		}
	}
}

func TestZatcaDocumentListRequiresOrganization(t *testing.T) {
	for _, org := range []int32{0, -1} {
		db := &documentListDB{}
		engine := gin.New()
		engine.Use(func(c *gin.Context) {
			if org != 0 {
				c.Set(string(middleware.OrgIDKey), org)
			}
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), middleware.RepoKey, repository.New(db)))
		})
		engine.GET("/documents", NewZatcaHandler(nil).ListZatcaDocuments)
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, httptest.NewRequest("GET", "/documents?store_id=3", nil))
		if recorder.Code != 401 || db.countCalls != 0 || db.listCalls != 0 {
			t.Fatalf("missing organization allowed: %s", recorder.Body)
		}
	}
}
