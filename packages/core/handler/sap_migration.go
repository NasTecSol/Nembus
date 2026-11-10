package handler

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NasTecSol/nembus-core/middleware/manager"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-sap/contracts"
)

type SAPMigrationHandler struct {
	defaultUC     *usecase.SAPMigrationUseCase
	tenantManager *manager.Manager
	masterPool    *pgxpool.Pool
}

func NewSAPMigrationHandler(uc *usecase.SAPMigrationUseCase, tm *manager.Manager, defaultPool *pgxpool.Pool) *SAPMigrationHandler {
	return &SAPMigrationHandler{
		defaultUC:     uc,
		tenantManager: tm,
		masterPool:    defaultPool,
	}
}

func (h *SAPMigrationHandler) IngestBatch(c *gin.Context) {
	var reader io.Reader = c.Request.Body

	// Handle Gzip Content-Encoding if sent by agent
	if c.GetHeader("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid gzip payload: " + err.Error()})
			return
		}
		defer gz.Close()
		reader = gz
	}

	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read payload: " + err.Error()})
		return
	}

	var payload contracts.MigrationBatchPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse batch JSON: " + err.Error()})
		return
	}

	// 1. Resolve Tenant Slug from headers or query
	slug := strings.TrimSpace(c.GetHeader("x-tenant-slug"))
	if slug == "" {
		slug = strings.TrimSpace(c.GetHeader("X-Tenant-Slug"))
	}
	if slug == "" {
		slug = strings.TrimSpace(c.Query("tenant"))
	}
	if slug == "" {
		headerTenant := strings.TrimSpace(c.GetHeader("x-tenant-id"))
		if headerTenant != "" {
			if _, err := strconv.Atoi(headerTenant); err != nil {
				slug = headerTenant
			}
		}
	}

	// 2. Select appropriate UseCase (tenant pool if slug resolved, otherwise default)
	targetUC := h.defaultUC
	if slug != "" && h.tenantManager != nil {
		pool, err := h.tenantManager.GetPool(c.Request.Context(), slug)
		if err == nil && pool != nil {
			targetUC = usecase.NewSAPMigrationUseCase(pool)
		} else {
			log.Printf("[SAPMigrationHandler] Warning: failed to resolve pool for tenant '%s': %v. Falling back to default pool.", slug, err)
		}
	}

	// 3. Extract Organization ID
	orgID := 1
	if headerOrg := c.GetHeader("x-organization-id"); headerOrg != "" {
		if id, err := strconv.Atoi(headerOrg); err == nil && id > 0 {
			orgID = id
		}
	} else if headerTenant := c.GetHeader("x-tenant-id"); headerTenant != "" {
		if id, err := strconv.Atoi(headerTenant); err == nil && id > 0 {
			orgID = id
		}
	}
	if payload.OrganizationID > 0 {
		orgID = payload.OrganizationID
	}

	resp, err := targetUC.IngestBatch(c.Request.Context(), orgID, &payload)
	if err != nil {
		log.Printf("[SAPMigrationHandler] IngestBatch ERROR: domain=%s batch_id=%s err=%v", payload.Domain, payload.BatchID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
