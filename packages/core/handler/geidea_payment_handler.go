package handler

import (
	"net/http"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/payment/geidea"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
)

type GeideaPaymentHandler struct {
	uc *usecase.GeideaPaymentUseCase
}

func NewGeideaPaymentHandler(uc *usecase.GeideaPaymentUseCase) *GeideaPaymentHandler {
	return &GeideaPaymentHandler{uc: uc}
}

func (h *GeideaPaymentHandler) useCaseForRequest(c *gin.Context) (*usecase.GeideaPaymentUseCase, bool) {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok || repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant repository not found in request context"})
		return nil, false
	}
	return h.uc.WithRepository(repo), true
}

func (h *GeideaPaymentHandler) Purchase(c *gin.Context) {
	var in usecase.InitiateGeideaPurchaseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uc, ok := h.useCaseForRequest(c)
	if !ok {
		return
	}
	payment, resp, err := uc.Purchase(c.Request.Context(), in)
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

	uc, ok := h.useCaseForRequest(c)
	if !ok {
		return
	}
	payment, err := uc.ResolveUncertain(c.Request.Context(), in)
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
