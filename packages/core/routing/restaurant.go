package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRestaurantRoutes(r *gin.RouterGroup, h *handler.RestaurantHandler) {
	rest := r.Group("/restaurant")

	// Tables
	rest.POST("/tables", h.CreateTable)
	rest.GET("/tables/:table_id", h.GetTable)
	rest.PUT("/tables/:table_id", h.UpdateTable)
	rest.DELETE("/tables/:table_id", h.DeleteTable)

	// Menu Categories
	rest.POST("/menu-categories", h.CreateMenuCategory)
	rest.GET("/menu-categories/:category_id", h.GetMenuCategory)
	rest.PUT("/menu-categories/:category_id", h.UpdateMenuCategory)
	rest.DELETE("/menu-categories/:category_id", h.DeleteMenuCategory)
	rest.GET("/menu-categories/:category_id/items", h.ListMenuItems)

	// Menu Items
	rest.POST("/menu-items", h.CreateMenuItem)
	rest.GET("/menu-items/:item_id", h.GetMenuItem)
	rest.PUT("/menu-items/:item_id", h.UpdateMenuItem)
	rest.DELETE("/menu-items/:item_id", h.DeleteMenuItem)
	rest.GET("/menu-items/:item_id/modifiers", h.ListModifiers)

	// Modifiers
	rest.POST("/modifiers", h.CreateModifier)
	rest.GET("/modifiers/:modifier_id", h.GetModifier)
	rest.PUT("/modifiers/:modifier_id", h.UpdateModifier)
	rest.DELETE("/modifiers/:modifier_id", h.DeleteModifier)

	// Menu Modifier Groups
	rest.POST("/menu-modifier-groups", h.CreateMenuModifierGroup)
	rest.GET("/menu-modifier-groups/:id", h.GetMenuModifierGroup)
	rest.PUT("/menu-modifier-groups/:id", h.UpdateMenuModifierGroup)
	rest.DELETE("/menu-modifier-groups/:id", h.DeleteMenuModifierGroup)

	// Orders
	rest.POST("/orders", h.CreateOrder)
	rest.POST("/orders/online", h.CreateOnlineOrder)
	rest.GET("/orders/:order_id", h.GetOrder)
	rest.PUT("/orders/:order_id", h.UpdateOrder)
	rest.DELETE("/orders/:order_id", h.DeleteOrder)
	rest.PATCH("/orders/:order_id/status", h.UpdateOrderStatus)
	rest.POST("/orders/:order_id/settle", h.SettleOrder)

	// Order Items
	rest.GET("/order-items/:order_item_id", h.GetOrderItem)
	rest.PUT("/order-items/:order_item_id", h.UpdateOrderItem)
	rest.DELETE("/order-items/:order_item_id", h.DeleteOrderItem)

	// Recipes
	rest.POST("/recipes", h.CreateRecipe)
	rest.GET("/recipes/:recipe_id", h.GetRecipe)
	rest.PUT("/recipes/:recipe_id", h.UpdateRecipe)
	rest.DELETE("/recipes/:recipe_id", h.DeleteRecipe)
	rest.POST("/recipes/:recipe_id/ingredients", h.AddRecipeIngredient)

	// Recipe Ingredients
	rest.GET("/recipe-ingredients/:recipe_ingredient_id", h.GetRecipeIngredient)
	rest.PUT("/recipe-ingredients/:recipe_ingredient_id", h.UpdateRecipeIngredient)
	rest.DELETE("/recipe-ingredients/:recipe_ingredient_id", h.DeleteRecipeIngredient)

	// Waste
	rest.POST("/waste-logs", h.CreateWasteLog)
	rest.GET("/waste-logs/:waste_log_id", h.GetWasteLog)
	rest.PUT("/waste-logs/:waste_log_id", h.UpdateWasteLog)
	rest.DELETE("/waste-logs/:waste_log_id", h.DeleteWasteLog)

	// Kiosk
	rest.POST("/kiosk/sessions", h.CreateKioskSession)
	rest.PUT("/kiosk/sessions/:session_id", h.UpdateKioskSession)
	rest.DELETE("/kiosk/sessions/:session_id", h.DeleteKioskSession)
	rest.GET("/kiosk/sessions/:token", h.GetKioskSession)
	rest.GET("/kiosk/sessions/id/:session_id", h.GetKioskSessionByID)

	// Store-specific lookups
	stores := rest.Group("/stores/:store_id")
	{
		stores.GET("/tables", h.ListTables)
		stores.GET("/menu-categories", h.ListMenuCategories)
		stores.GET("/menu", h.GetFullMenu)
		stores.GET("/kds", h.GetKdsOrders)
		stores.GET("/waste-report", h.GetWasteReport)
		stores.GET("/recipes", h.ListRecipes)
		stores.GET("/menu-modifier-groups", h.ListMenuModifierGroupsByStore)
	}
}
