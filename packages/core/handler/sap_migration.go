package handler

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-sap/contracts"
)

type SAPMigrationHandler struct {
	uc *usecase.SAPMigrationUseCase
}

func NewSAPMigrationHandler(uc *usecase.SAPMigrationUseCase) *SAPMigrationHandler {
	return &SAPMigrationHandler{uc: uc}
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

	// Extract Org ID from header / context if set
	orgID := 1
	if headerOrg := c.GetHeader("x-tenant-id"); headerOrg != "" {
		if id, err := strconv.Atoi(headerOrg); err == nil && id > 0 {
			orgID = id
		}
	}

	resp, err := h.uc.IngestBatch(c.Request.Context(), orgID, &payload)
	if err != nil {
		log.Printf("[SAPMigrationHandler] IngestBatch ERROR: domain=%s batch_id=%s err=%v", payload.Domain, payload.BatchID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
