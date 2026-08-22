package handler

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-sap/contracts"
)

type SAPMigrationHandler struct {
	enrichmentEnabled        bool
	newEnrichmentCoordinator func(*repository.Queries) enrichment.EnrichmentEnqueuer
}

func NewSAPMigrationHandler(enrichmentEnabled bool) *SAPMigrationHandler {
	return &SAPMigrationHandler{
		enrichmentEnabled: enrichmentEnabled,
		newEnrichmentCoordinator: func(tenantRepo *repository.Queries) enrichment.EnrichmentEnqueuer {
			return enrichment.NewProductEnrichmentCoordinator(repository.NewProductEnrichmentStore(tenantRepo))
		},
	}
}

// configureEnrichment wires the optional post-commit Stage 2A seam for this
// request's tenant. The application-level flag is intentionally checked here,
// before the coordinator can load a snapshot or create a suggestion. A
// disabled enrichment feature leaves SAP migration fully independent.
func (h *SAPMigrationHandler) configureEnrichment(uc *usecase.SAPMigrationUseCase, tenantRepo *repository.Queries) {
	if h == nil || !h.enrichmentEnabled || uc == nil || h.newEnrichmentCoordinator == nil {
		return
	}
	uc.SetProductEnrichmentCoordinator(h.newEnrichmentCoordinator(tenantRepo))
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

	identity, ok := middleware.TrustedMachineIdentityFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Trusted machine identity required"})
		return
	}
	if payload.OrganizationID < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OrganizationID must be positive when supplied"})
		return
	}
	if payload.OrganizationID > 0 && int32(payload.OrganizationID) != identity.OrganizationID {
		c.JSON(http.StatusConflict, gin.H{"error": "payload OrganizationID does not match the authenticated organization"})
		return
	}
	payload.OrganizationID = int(identity.OrganizationID)

	pool, ok := middleware.TenantPoolFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tenant database unavailable"})
		return
	}
	tenantRepo, ok := middleware.RepositoryFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tenant repository unavailable"})
		return
	}
	// These objects are immutable after construction and are tied to the exact
	// tenant pool selected by the authenticated request.
	uc := usecase.NewSAPMigrationUseCase(pool)
	h.configureEnrichment(uc, tenantRepo)

	resp, err := uc.IngestBatch(c.Request.Context(), int(identity.OrganizationID), &payload)
	if err != nil {
		log.Printf("[SAPMigrationHandler] IngestBatch ERROR: domain=%s batch_id=%s err=%v", payload.Domain, payload.BatchID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SAP migration failed"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
