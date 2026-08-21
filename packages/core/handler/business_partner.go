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

type BusinessPartnerHandler struct {
	useCase *usecase.BusinessPartnerUseCase
}

func NewBusinessPartnerHandler(uc *usecase.BusinessPartnerUseCase) *BusinessPartnerHandler {
	return &BusinessPartnerHandler{
		useCase: uc,
	}
}

func (h *BusinessPartnerHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateBusinessPartner handles POST /business-partners
// @Summary      Create a new business partner
// @Description  Create a business partner record, optionally with addresses and contacts nested
// @Tags         business-partners
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        partner       body      CreateBusinessPartnerRequest  true  "Business partner data"
// @Success      201           {object}  BusinessPartnerResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/business-partners [post]
func (h *BusinessPartnerHandler) CreateBusinessPartner(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateBusinessPartnerRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body", nil))
		return
	}

	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata format", nil))
		return
	}

	var taxID string
	if req.TaxID != nil {
		taxID = *req.TaxID
	}
	var currencyCode string
	if req.CurrencyCode != nil {
		currencyCode = *req.CurrencyCode
	}

	var creditLimit float64
	if req.CreditLimit != nil && *req.CreditLimit != "" {
		cl, err := strconv.ParseFloat(*req.CreditLimit, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid credit_limit format", nil))
			return
		}
		creditLimit = cl
	}

	addresses := make([]usecase.PartnerAddressInput, len(req.Addresses))
	for i, a := range req.Addresses {
		var street, city, state, zip, country string
		if a.Street != nil {
			street = *a.Street
		}
		if a.City != nil {
			city = *a.City
		}
		if a.State != nil {
			state = *a.State
		}
		if a.ZipCode != nil {
			zip = *a.ZipCode
		}
		if a.CountryCode != nil {
			country = *a.CountryCode
		}
		addresses[i] = usecase.PartnerAddressInput{
			AddressName: a.AddressName,
			AddressType: a.AddressType,
			Street:      street,
			City:        city,
			State:       state,
			ZipCode:     zip,
			CountryCode: country,
			IsDefault:   a.IsDefault,
		}
	}

	contacts := make([]usecase.PartnerContactInput, len(req.Contacts))
	for i, ctc := range req.Contacts {
		var lastName, email, phone, pos string
		if ctc.LastName != nil {
			lastName = *ctc.LastName
		}
		if ctc.Email != nil {
			email = *ctc.Email
		}
		if ctc.Phone != nil {
			phone = *ctc.Phone
		}
		if ctc.Position != nil {
			pos = *ctc.Position
		}
		contacts[i] = usecase.PartnerContactInput{
			FirstName: ctc.FirstName,
			LastName:  lastName,
			Email:     email,
			Phone:     phone,
			Position:  pos,
			IsPrimary: ctc.IsPrimary,
		}
	}

	resp := h.useCase.CreateBusinessPartner(
		c.Request.Context(),
		usecase.CreateBusinessPartnerInput{
			OrganizationID: req.OrganizationID,
			Code:           req.Code,
			Name:           req.Name,
			PartnerRole:    req.PartnerRole,
			TaxID:          taxID,
			CurrencyCode:   currencyCode,
			CreditLimit:    creditLimit,
			PaymentTermsID: req.PaymentTermsID,
			SalesRepUserID: req.SalesRepUserID,
			IsActive:       req.IsActive,
			Metadata:       metaBytes,
			Addresses:      addresses,
			Contacts:       contacts,
		},
	)
	c.JSON(resp.StatusCode, resp)
}

// GetBusinessPartner handles GET /business-partners/:id
// @Summary      Get business partner by ID
// @Description  Retrieve a business partner with addresses and contacts nested by ID
// @Tags         business-partners
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Business partner ID"
// @Success      200           {object}  BusinessPartnerResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/business-partners/{id} [get]
func (h *BusinessPartnerHandler) GetBusinessPartner(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	resp := h.useCase.GetBusinessPartnerByID(c.Request.Context(), idStr)
	c.JSON(resp.StatusCode, resp)
}

// ListBusinessPartners handles GET /business-partners
// @Summary      List business partners
// @Description  List business partners filtered by role and organization
// @Tags         business-partners
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true  "Tenant identifier"
// @Param        Authorization    header    string  true  "Bearer token"
// @Param        organization_id  query     int     true  "Organization ID"
// @Param        partner_role     query     string  false "Filter by role (supplier, vendor, special_customer, corporate_group)"
// @Success      200              {array}   BusinessPartnerResponse
// @Failure      400              {object}  ErrorResponse
// @Failure      401              {object}  ErrorResponse
// @Failure      500              {object}  ErrorResponse
// @Router       /api/business-partners [get]
func (h *BusinessPartnerHandler) ListBusinessPartners(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	roleFilter := c.Query("partner_role")

	resp := h.useCase.ListBusinessPartners(c.Request.Context(), orgIDStr, roleFilter)
	c.JSON(resp.StatusCode, resp)
}

// SearchBusinessPartners handles GET /business-partners/search
// @Summary      Search business partners
// @Description  Search business partners by name or code
// @Tags         business-partners
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true  "Tenant identifier"
// @Param        Authorization    header    string  true  "Bearer token"
// @Param        organization_id  query     int     true  "Organization ID"
// @Param        q                query     string  true  "Search query"
// @Param        limit            query     int     false "Limit (default 10)"
// @Success      200              {array}   BusinessPartnerResponse
// @Failure      400              {object}  ErrorResponse
// @Router       /api/business-partners/search [get]
func (h *BusinessPartnerHandler) SearchBusinessPartners(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	query := c.Query("q")
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		limit = 10
	}

	resp := h.useCase.SearchBusinessPartners(c.Request.Context(), orgIDStr, query, int32(limit))
	c.JSON(resp.StatusCode, resp)
}

// UpdateBusinessPartner handles PUT /business-partners/:id
// @Summary      Update a business partner
// @Description  Update core business partner fields
// @Tags         business-partners
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Business partner ID"
// @Param        partner       body      UpdateBusinessPartnerRequest  true  "Updated partner data"
// @Success      200           {object}  BusinessPartnerResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/business-partners/{id} [put]
func (h *BusinessPartnerHandler) UpdateBusinessPartner(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")

	var req UpdateBusinessPartnerRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body", nil))
		return
	}

	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata format", nil))
		return
	}

	var code, name, role, taxID, currency string
	if req.Code != nil {
		code = *req.Code
	}
	if req.Name != nil {
		name = *req.Name
	}
	if req.PartnerRole != nil {
		role = *req.PartnerRole
	}
	if req.TaxID != nil {
		taxID = *req.TaxID
	}
	if req.CurrencyCode != nil {
		currency = *req.CurrencyCode
	}

	var creditLimit float64
	if req.CreditLimit != nil && *req.CreditLimit != "" {
		cl, err := strconv.ParseFloat(*req.CreditLimit, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid credit_limit format", nil))
			return
		}
		creditLimit = cl
	}

	var isActive bool
	if req.IsActive != nil {
		isActive = *req.IsActive
	} else {
		isActive = true
	}

	resp := h.useCase.UpdateBusinessPartner(
		c.Request.Context(),
		idStr,
		usecase.UpdateBusinessPartnerInput{
			Code:           code,
			Name:           name,
			PartnerRole:    role,
			TaxID:          taxID,
			CurrencyCode:   currency,
			CreditLimit:    creditLimit,
			PaymentTermsID: req.PaymentTermsID,
			SalesRepUserID: req.SalesRepUserID,
			IsActive:       isActive,
			Metadata:       metaBytes,
		},
	)
	c.JSON(resp.StatusCode, resp)
}

// DeleteBusinessPartner handles DELETE /business-partners/:id
// @Summary      Delete a business partner
// @Description  Delete a business partner by ID (cascades addresses/contacts)
// @Tags         business-partners
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Business partner ID"
// @Success      200           {object}  SuccessResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/business-partners/{id} [delete]
func (h *BusinessPartnerHandler) DeleteBusinessPartner(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	resp := h.useCase.DeleteBusinessPartner(c.Request.Context(), idStr)
	c.JSON(resp.StatusCode, resp)
}

// ToggleBusinessPartnerActive handles PATCH /business-partners/:id/toggle
// @Summary      Toggle business partner active state
// @Description  Toggle the is_active status of a business partner
// @Tags         business-partners
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Business partner ID"
// @Param        body          body      ToggleBusinessPartnerActiveRequest true "State data"
// @Success      200           {object}  BusinessPartnerResponse
// @Router       /api/business-partners/{id}/toggle [patch]
func (h *BusinessPartnerHandler) ToggleBusinessPartnerActive(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	var req ToggleBusinessPartnerActiveRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body", nil))
		return
	}

	resp := h.useCase.ToggleBusinessPartnerActive(c.Request.Context(), idStr, req.IsActive)
	c.JSON(resp.StatusCode, resp)
}

// Addresses management handlers

// AddPartnerAddress handles POST /business-partners/:id/addresses
func (h *BusinessPartnerHandler) AddPartnerAddress(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	partnerIDStr := c.Param("id")

	var req CreatePartnerAddressDTO
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid address data", nil))
		return
	}

	var street, city, state, zip, country string
	if req.Street != nil {
		street = *req.Street
	}
	if req.City != nil {
		city = *req.City
	}
	if req.State != nil {
		state = *req.State
	}
	if req.ZipCode != nil {
		zip = *req.ZipCode
	}
	if req.CountryCode != nil {
		country = *req.CountryCode
	}

	resp := h.useCase.AddPartnerAddress(
		c.Request.Context(),
		partnerIDStr,
		usecase.PartnerAddressInput{
			AddressName: req.AddressName,
			AddressType: req.AddressType,
			Street:      street,
			City:        city,
			State:       state,
			ZipCode:     zip,
			CountryCode: country,
			IsDefault:   req.IsDefault,
		},
	)
	c.JSON(resp.StatusCode, resp)
}

// UpdatePartnerAddress handles PUT /business-partners/:id/addresses/:addressId
func (h *BusinessPartnerHandler) UpdatePartnerAddress(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	addressIDStr := c.Param("addressId")

	var req CreatePartnerAddressDTO
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid address data", nil))
		return
	}

	var street, city, state, zip, country string
	if req.Street != nil {
		street = *req.Street
	}
	if req.City != nil {
		city = *req.City
	}
	if req.State != nil {
		state = *req.State
	}
	if req.ZipCode != nil {
		zip = *req.ZipCode
	}
	if req.CountryCode != nil {
		country = *req.CountryCode
	}

	resp := h.useCase.UpdatePartnerAddress(
		c.Request.Context(),
		addressIDStr,
		usecase.PartnerAddressInput{
			AddressName: req.AddressName,
			AddressType: req.AddressType,
			Street:      street,
			City:        city,
			State:       state,
			ZipCode:     zip,
			CountryCode: country,
			IsDefault:   req.IsDefault,
		},
	)
	c.JSON(resp.StatusCode, resp)
}

// DeletePartnerAddress handles DELETE /business-partners/:id/addresses/:addressId
func (h *BusinessPartnerHandler) DeletePartnerAddress(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	addressIDStr := c.Param("addressId")
	resp := h.useCase.DeletePartnerAddress(c.Request.Context(), addressIDStr)
	c.JSON(resp.StatusCode, resp)
}

// Contacts management handlers

// AddPartnerContact handles POST /business-partners/:id/contacts
func (h *BusinessPartnerHandler) AddPartnerContact(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	partnerIDStr := c.Param("id")

	var req CreatePartnerContactDTO
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid contact data", nil))
		return
	}

	var lastName, email, phone, pos string
	if req.LastName != nil {
		lastName = *req.LastName
	}
	if req.Email != nil {
		email = *req.Email
	}
	if req.Phone != nil {
		phone = *req.Phone
	}
	if req.Position != nil {
		pos = *req.Position
	}

	resp := h.useCase.AddPartnerContact(
		c.Request.Context(),
		partnerIDStr,
		usecase.PartnerContactInput{
			FirstName: req.FirstName,
			LastName:  lastName,
			Email:     email,
			Phone:     phone,
			Position:  pos,
			IsPrimary: req.IsPrimary,
		},
	)
	c.JSON(resp.StatusCode, resp)
}

// UpdatePartnerContact handles PUT /business-partners/:id/contacts/:contactId
func (h *BusinessPartnerHandler) UpdatePartnerContact(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	contactIDStr := c.Param("contactId")

	var req CreatePartnerContactDTO
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid contact data", nil))
		return
	}

	var lastName, email, phone, pos string
	if req.LastName != nil {
		lastName = *req.LastName
	}
	if req.Email != nil {
		email = *req.Email
	}
	if req.Phone != nil {
		phone = *req.Phone
	}
	if req.Position != nil {
		pos = *req.Position
	}

	resp := h.useCase.UpdatePartnerContact(
		c.Request.Context(),
		contactIDStr,
		usecase.PartnerContactInput{
			FirstName: req.FirstName,
			LastName:  lastName,
			Email:     email,
			Phone:     phone,
			Position:  pos,
			IsPrimary: req.IsPrimary,
		},
	)
	c.JSON(resp.StatusCode, resp)
}

// DeletePartnerContact handles DELETE /business-partners/:id/contacts/:contactId
func (h *BusinessPartnerHandler) DeletePartnerContact(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	contactIDStr := c.Param("contactId")
	resp := h.useCase.DeletePartnerContact(c.Request.Context(), contactIDStr)
	c.JSON(resp.StatusCode, resp)
}
