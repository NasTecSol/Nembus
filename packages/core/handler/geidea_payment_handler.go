package handler

import (
	"net/http"

	"github.com/NasTecSol/nembus-core/payment/geidea"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
)

type GeideaPaymentHandler struct {
	uc *usecase.GeideaPaymentUseCase
}

func NewGeideaPaymentHandler(uc *usecase.GeideaPaymentUseCase) *GeideaPaymentHandler {
	return &GeideaPaymentHandler{uc: uc}
}

func (h *GeideaPaymentHandler) Purchase(c *gin.Context) {
	var in usecase.InitiateGeideaPurchaseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, resp, err := h.uc.Purchase(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "payment": payment, "response": resp})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment": payment, "terminal_response": resp})
}

func (h *GeideaPaymentHandler) ResolveUncertain(c *gin.Context) {
	var in usecase.ResolveUncertainInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.uc.ResolveUncertain(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment": payment})
}

func (h *GeideaPaymentHandler) Ping(c *gin.Context) {
	sim := geidea.NewSimulator(nil)
	if err := sim.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "offline", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "online"})
}
