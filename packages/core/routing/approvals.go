package router

import (
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/gin-gonic/gin"
)

func RegisterApprovalsRoutes(r *gin.RouterGroup, h *handler.ApprovalsHandler) {
	g := r.Group("/approvals")
	// Register the literal /policies routes before the :id parameter routes.
	g.GET("/policies", h.Policies)
	g.GET("/policies/:domain", h.GetPolicy)
	g.PUT("/policies/:domain", h.SavePolicy)
	g.DELETE("/policies/:domain", h.DeletePolicy)
	g.POST("/submit", h.Submit)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.DELETE("/:id", h.Delete)
	g.POST("/:id/approve", h.Approve)
	g.POST("/:id/reject", h.Reject)
	g.POST("/:id/cancel", h.Cancel)
}
