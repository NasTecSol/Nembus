package handler

import (
	"net/http"

	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// parseUUIDParam reads a Gin path parameter as a UUID and writes a 400 response on failure.
func parseUUIDParam(c *gin.Context, param string) (uuid.UUID, bool) {
	raw := c.Param(param)
	if raw == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, param+": missing", nil))
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid "+param+": expected a valid UUID", nil))
		return uuid.Nil, false
	}
	return id, true
}