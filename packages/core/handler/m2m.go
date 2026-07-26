package handler

import (
	"net/http"

	"github.com/NasTecSol/nembus-core/middleware"

	"github.com/gin-gonic/gin"
)

type M2MHandler struct{}

func NewM2MHandler() *M2MHandler {
	return &M2MHandler{}
}

type CreateM2MRequest struct {
	ClientID   string   `json:"client_id" binding:"required"`
	ClientName string   `json:"client_name" binding:"required"`
	Scopes     []string `json:"scopes"`
	Years      int      `json:"years"`
}

// CreateToken handles POST /api/m2m/tokens
// @Summary      Create M2M token
// @Description  Register a new client application and generate a long-lived JWT token
// @Tags         m2m
// @Accept       json
// @Produce      json
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        request       body      CreateM2MRequest  true  "M2M client registration data"
// @Success      200  {object}  middleware.M2MClient
// @Failure      400  {object}  gin.H
// @Failure      401  {object}  gin.H
// @Failure      500  {object}  gin.H
// @Router       /api/m2m/tokens [post]
func (h *M2MHandler) CreateToken(c *gin.Context) {
	tenantID := c.GetHeader("x-tenant-id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-tenant-id header required"})
		return
	}

	var req CreateM2MRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	years := req.Years
	if years <= 0 {
		years = 5
	}

	tokenString, err := middleware.GenerateM2MToken(req.ClientID, req.ClientName, tenantID, req.Scopes, years)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token", "details": err.Error()})
		return
	}

	client := middleware.M2MClient{
		ClientID:   req.ClientID,
		ClientName: req.ClientName,
		TenantID:   tenantID,
		Scopes:     req.Scopes,
		IsActive:   true,
		Token:      tokenString,
	}

	if err := middleware.SaveM2MClient(client); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save client registry", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client)
}

// ListTokens handles GET /api/m2m/tokens
// @Summary      List M2M tokens
// @Description  List all registered M2M tokens for the current tenant
// @Tags         m2m
// @Produce      json
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Success      200  {array}   middleware.M2MClient
// @Failure      400  {object}  gin.H
// @Failure      500  {object}  gin.H
// @Router       /api/m2m/tokens [get]
func (h *M2MHandler) ListTokens(c *gin.Context) {
	tenantID := c.GetHeader("x-tenant-id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-tenant-id header required"})
		return
	}

	clients, err := middleware.LoadM2MClients()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load clients", "details": err.Error()})
		return
	}

	// Filter by tenant
	var filtered []middleware.M2MClient
	for _, client := range clients {
		if client.TenantID == tenantID {
			filtered = append(filtered, client)
		}
	}

	if filtered == nil {
		filtered = []middleware.M2MClient{}
	}

	c.JSON(http.StatusOK, filtered)
}
