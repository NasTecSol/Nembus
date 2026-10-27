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
		bp.GET("/:id/addresses", h.ListPartnerAddresses)
		bp.GET("/:id/addresses/:addressId", h.GetPartnerAddress)
		bp.PUT("/:id/addresses/:addressId", h.UpdatePartnerAddress)
		bp.DELETE("/:id/addresses/:addressId", h.DeletePartnerAddress)
		bp.PATCH("/:id/addresses/:addressId/default", h.SetDefaultPartnerAddress)

		// Contacts
		bp.POST("/:id/contacts", h.AddPartnerContact)
		bp.GET("/:id/contacts", h.ListPartnerContacts)
		bp.GET("/:id/contacts/:contactId", h.GetPartnerContact)
		bp.PUT("/:id/contacts/:contactId", h.UpdatePartnerContact)
		bp.DELETE("/:id/contacts/:contactId", h.DeletePartnerContact)
		bp.PATCH("/:id/contacts/:contactId/primary", h.SetPrimaryPartnerContact)
	}

	// Standalone partner addresses routes
	addr := r.Group("/partner-addresses")
	{
		addr.POST("", h.CreatePartnerAddressStandalone)
		addr.GET("", h.ListPartnerAddressesStandalone)
		addr.GET("/:id", h.GetPartnerAddressByID)
		addr.PUT("/:id", h.UpdatePartnerAddressByID)
		addr.DELETE("/:id", h.DeletePartnerAddressByID)
		addr.PATCH("/:id/default", h.SetDefaultPartnerAddressByID)
	}

	// Standalone partner contacts routes
	ctc := r.Group("/partner-contacts")
	{
		ctc.POST("", h.CreatePartnerContactStandalone)
		ctc.GET("", h.ListPartnerContactsStandalone)
		ctc.GET("/:id", h.GetPartnerContactByID)
		ctc.PUT("/:id", h.UpdatePartnerContactByID)
		ctc.DELETE("/:id", h.DeletePartnerContactByID)
		ctc.PATCH("/:id/primary", h.SetPrimaryPartnerContactByID)
	}
}
