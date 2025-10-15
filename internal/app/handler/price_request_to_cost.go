package handler

import (
	"DIA_Backend/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RequestCostHandler struct {
	repo *repository.Repository
}

func NewPriceRequestToCostHandler(repo *repository.Repository) *RequestCostHandler {
	return &RequestCostHandler{
		repo: repo,
	}
}

type PriceRequestToCostResponse struct {
	Cost       []CostResponse `json:"cost"`
	Cost_price float64        `json:"cost_price"`
}

type UpdatePriceToRequestConnection struct {
	Cost_price *float64 `json:"cost_price"`
}

func (h *RequestCostHandler) RemovePriceToRequestConnection(ctx *gin.Context) {
	requestIDStr := ctx.Param("requestId")
	costIDStr := ctx.Param("costId")

	requestID, err := strconv.ParseUint(requestIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	costID, err := strconv.ParseUint(costIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost ID"})
		return
	}

	userID := GetFixedUserID()
	if err := h.repo.CostRequest.RemoveCostFromRequest(requestID, costID, userID); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove cost from request"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cost removed from request successfully"})
}

func (h *RequestCostHandler) UpdatePriceToRequestConnection(ctx *gin.Context) {
	requestIDStr := ctx.Param("requestId")
	costIDStr := ctx.Param("costId")

	requestID, err := strconv.ParseUint(requestIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	costID, err := strconv.ParseUint(costIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost ID"})
		return
	}

	var req UpdatePriceToRequestConnection
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost request data"})
		return
	}

	userID := GetFixedUserID()
	if err := h.repo.CostRequest.UpdateRequestToCost(requestID, costID, userID, req.Cost_price); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update request cost"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Request cost updated successfully"})
}
