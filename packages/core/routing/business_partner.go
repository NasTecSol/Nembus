package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterBusinessPartnerRoutes registers admin business partner routes.
func RegisterBusinessPartnerRoutes(r *gin.RouterGroup, h *handler.BusinessPartnerHandler) {
	bp := r.Group("/business-partners")
	{
		bp.POST("", h.CreateBusinessPartner)
		bp.GET("", h.ListBusinessPartners)
		bp.GET("/search", h.SearchBusinessPartners)
		bp.GET("/:id", h.GetBusinessPartner)
		bp.PUT("/:id", h.UpdateBusinessPartner)
		bp.DELETE("/:id", h.DeleteBusinessPartner)
		bp.PATCH("/:id/toggle", h.ToggleBusinessPartnerActive)

		// Addresses
		bp.POST("/:id/addresses", h.AddPartnerAddress)
		bp.PUT("/:id/addresses/:addressId", h.UpdatePartnerAddress)
		bp.DELETE("/:id/addresses/:addressId", h.DeletePartnerAddress)

		// Contacts
		bp.POST("/:id/contacts", h.AddPartnerContact)
		bp.PUT("/:id/contacts/:contactId", h.UpdatePartnerContact)
		bp.DELETE("/:id/contacts/:contactId", h.DeletePartnerContact)
	}
}
