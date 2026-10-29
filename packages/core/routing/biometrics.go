package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

func RegisterBiometricRoutes(r *gin.RouterGroup, h *handler.BiometricHandler) {
	r.GET("/users/:id/biometrics", h.ListUserBiometrics)
	r.POST("/users/:id/biometrics/enroll", h.EnrollFinger)
	r.DELETE("/users/:id/biometrics/:fingerIndex", h.DeleteUserBiometric)
	r.POST("/biometrics/enroll", h.EnrollFinger)
}

