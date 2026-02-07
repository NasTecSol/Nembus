package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRestaurantRoutes(r *gin.RouterGroup, h *handler.RestaurantHandler) {
	rest := r.Group("/restaurant")

	// Tables
	rest.POST("/tables", h.CreateTable)
	rest.GET("/tables/:id", h.GetTable)
	rest.PUT("/tables/:id", h.UpdateTable)
	rest.DELETE("/tables/:id", h.DeleteTable)

	// Menu Categories
	rest.POST("/menu-categories", h.CreateMenuCategory)
	rest.GET("/menu-categories/:id", h.GetMenuCategory)
	rest.PUT("/menu-categories/:id", h.UpdateMenuCategory)
	rest.DELETE("/menu-categories/:id", h.DeleteMenuCategory)
	rest.GET("/menu-categories/:category_id/items", h.ListMenuItems)

	// Menu Items
	rest.POST("/menu-items", h.CreateMenuItem)
	rest.GET("/menu-items/:id", h.GetMenuItem)
	rest.PUT("/menu-items/:id", h.UpdateMenuItem)
	rest.DELETE("/menu-items/:id", h.DeleteMenuItem)
	rest.GET("/menu-items/:item_id/modifiers", h.ListModifiers)

	// Modifiers
	rest.POST("/modifiers", h.CreateModifier)
	rest.GET("/modifiers/:id", h.GetModifier)
	rest.PUT("/modifiers/:id", h.UpdateModifier)
	rest.DELETE("/modifiers/:id", h.DeleteModifier)

	// Orders
	rest.POST("/orders", h.CreateOrder)
	rest.POST("/orders/online", h.CreateOnlineOrder)
	rest.GET("/orders/:id", h.GetOrder)
	rest.PUT("/orders/:id", h.UpdateOrder)
	rest.DELETE("/orders/:id", h.DeleteOrder)
	rest.PATCH("/orders/:id/status", h.UpdateOrderStatus)
	rest.POST("/orders/:id/settle", h.SettleOrder)

	// Order Items
	rest.GET("/order-items/:id", h.GetOrderItem)
	rest.PUT("/order-items/:id", h.UpdateOrderItem)
	rest.DELETE("/order-items/:id", h.DeleteOrderItem)

	// Recipes
	rest.POST("/recipes", h.CreateRecipe)
	rest.GET("/recipes/:id", h.GetRecipe)
	rest.PUT("/recipes/:id", h.UpdateRecipe)
	rest.DELETE("/recipes/:id", h.DeleteRecipe)
	rest.POST("/recipes/:id/ingredients", h.AddRecipeIngredient)

	// Recipe Ingredients
	rest.GET("/recipe-ingredients/:id", h.GetRecipeIngredient)
	rest.PUT("/recipe-ingredients/:id", h.UpdateRecipeIngredient)
	rest.DELETE("/recipe-ingredients/:id", h.DeleteRecipeIngredient)

	// Waste
	rest.POST("/waste-logs", h.CreateWasteLog)
	rest.GET("/waste-logs/:id", h.GetWasteLog)
	rest.PUT("/waste-logs/:id", h.UpdateWasteLog)
	rest.DELETE("/waste-logs/:id", h.DeleteWasteLog)

	// Kiosk
	rest.POST("/kiosk/sessions", h.CreateKioskSession)
	rest.GET("/kiosk/sessions/:token", h.GetKioskSession)
	rest.GET("/kiosk/sessions/id/:id", h.GetKioskSessionByID)
	rest.PUT("/kiosk/sessions/:id", h.UpdateKioskSession)
	rest.DELETE("/kiosk/sessions/:id", h.DeleteKioskSession)

	// Store-specific lookups
	stores := rest.Group("/stores/:store_id")
	{
		stores.GET("/tables", h.ListTables)
		stores.GET("/menu-categories", h.ListMenuCategories)
		stores.GET("/menu", h.GetFullMenu)
		stores.GET("/kds", h.GetKdsOrders)
		stores.GET("/waste-report", h.GetWasteReport)
		stores.GET("/recipes", h.ListRecipes)
	}
}
