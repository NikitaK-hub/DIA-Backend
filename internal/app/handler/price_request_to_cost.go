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

// @Summary      Create new cost request to cost handler
// @Description  Initialize handler for managing cost-request relationships
// @Tags         cost-request-costs
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

// @Summary      Remove cost from request
// @Description  Remove a cost from a cost request
// @Tags         cost-request-costs
// @Accept       json
// @Produce      json
// @Param        requestId path int true "Cost Request ID"
// @Param        costId path int true "Cost ID"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /cost-request-costs/{requestId}/costs/{costId} [delete]
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

	userUUID, _, ok := GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := h.repo.User.GetUserByUUID(userUUID)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	if err := h.repo.CostRequest.RemoveCostFromRequest(requestID, costID, user.ID); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove cost from request"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cost removed from request successfully"})
}

// @Summary      Update cost in request
// @Description  Update cost connection details in a cost request
// @Tags         cost-request-costs
// @Accept       json
// @Produce      json
// @Param        requestId path int true "Cost Request ID"
// @Param        costId path int true "Cost ID"
// @Param        request body UpdatePriceToRequestConnection true "Cost connection update data"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /cost-request-costs/{requestId}/costs/{costId} [put]
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

	userUUID, _, ok := GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := h.repo.User.GetUserByUUID(userUUID)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	if err := h.repo.CostRequest.UpdateRequestToCost(requestID, costID, user.ID, req.Cost_price); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update request cost"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Request cost updated successfully"})
}
