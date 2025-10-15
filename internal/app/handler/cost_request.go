package handler

import (
	"DIA_Backend/internal/app/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type CostRequestHandler struct {
	repo *repository.Repository
}

type CostsRequestsFilterResponse struct {
	ID           uint64    `json:"id"`
	Status       uint8     `json:"Status"`
	ID_user      uint64    `json:"UserID"`
	ID_moderator uint64    `json:"ModeratorID"`
	Min_volume   uint64    `json:"Min_volume"`
	Max_volume   uint64    `json:"Max_volume"`
	CreatedAt    time.Time `json:"CreatedAt"`
	FormedAt     time.Time `json:"FormedAt"`
	ClosedAt     time.Time `json:"ClosedAt"`
}

type CostRequestResponse struct {
	ID                 uint64                       `json:"id"`
	PriceRequestToCost []PriceRequestToCostResponse `json:"PriceRequestToCost"`
	Min_volume         uint64                       `json:"Min_volume"`
	Max_volume         uint64                       `json:"Max_volume"`
}

type CostRequestDetailResponse struct {
	ID                  uint64                             `json:"id"`
	CreatedAt           time.Time                          `json:"created_at"`
	Min_volume          uint64                             `json:"Min_volume"`
	Max_volume          uint64                             `json:"Max_volume"`
	PriceRequestToCosts []PriceRequestToCostDetailResponse `json:"price_request_to_costs"`
}

type PriceRequestToCostDetailResponse struct {
	Cost_price float64 `json:"cost_price"`
	CostTitle  string  `json:"cost_title"`
}

func NewCostRequestHandler(repository *repository.Repository) *CostRequestHandler {
	return &CostRequestHandler{repo: repository}
}

type CostRequestInfoResponse struct {
	RequestID uint64 `json:"request_id"`
	ItemCount int    `json:"item_count"`
}

type UpdateCostRequestResponse struct {
	Min_volume *uint64 `json:"Min_volume"`
	Max_volume *uint64 `json:"Max_volume"`
}

func (h *CostRequestHandler) GetCostRequestInfo(ctx *gin.Context) {
	userID := GetFixedUserID()
	requestID, itemCount, err := h.repo.CostRequest.GetDraftRequestInfo(userID)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cost request info"})
		return
	}

	ctx.JSON(http.StatusOK, CostRequestInfoResponse{
		RequestID: requestID,
		ItemCount: itemCount,
	})
}

func (h *CostRequestHandler) GetCostRequests(ctx *gin.Context) {
	var statusFilter uint8
	if statusStr := ctx.Query("status"); statusStr != "" {
		if status, err := strconv.ParseUint(statusStr, 10, 8); err == nil {
			statusFilter = uint8(status)
		}
	}

	var dateFrom, dateTo *time.Time
	if dateFromStr := ctx.Query("date_from"); dateFromStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			dateFrom = &parsed
		}
	}
	if dateToStr := ctx.Query("date_to"); dateToStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateToStr); err == nil {
			dateTo = &parsed
		}
	}

	requests, err := h.repo.CostRequest.GetCostRequests(statusFilter, dateFrom, dateTo)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cost requests"})
		return
	}

	var response []CostsRequestsFilterResponse
	for _, costRequest := range requests {
		response = append(response, CostsRequestsFilterResponse{
			ID:           costRequest.ID,
			Status:       costRequest.Status,
			ID_user:      costRequest.ID_user,
			ID_moderator: costRequest.ID_moderator,
			CreatedAt:    costRequest.CreatedAt,
			FormedAt:     costRequest.FormedAt,
			ClosedAt:     costRequest.ClosedAt,
			Min_volume:   costRequest.Min_volume,
			Max_volume:   costRequest.Max_volume,
		})
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *CostRequestHandler) GetCostRequestByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	userID := GetFixedUserID()
	request, err := h.repo.CostRequest.GetCostRequestByID(id, userID)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Cost request not found"})
		return
	}

	// Transform to response with only required fields
	response := CostRequestDetailResponse{
		ID:         request.ID,
		CreatedAt:  request.CreatedAt,
		Min_volume: request.Min_volume,
		Max_volume: request.Max_volume,
	}

	// Transform CostRequestToCost items
	for _, priceToRequest := range request.Price_request_for_cost {
		costDetail := PriceRequestToCostDetailResponse{
			CostTitle:  priceToRequest.Cost.Title,
			Cost_price: priceToRequest.Cost_price,
		}
		response.PriceRequestToCosts = append(response.PriceRequestToCosts, costDetail)
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *CostRequestHandler) UpdateCostRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost request ID"})
		return
	}

	var req UpdateCostRequestResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost request data"})
		return
	}

	if err := h.repo.CostRequest.UpdateCostRequest(id, req.Min_volume, req.Max_volume); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cost request"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cost request updated successfully"})
}

func (h *CostRequestHandler) FormCostRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost request ID"})
		return
	}

	userID := GetFixedUserID()
	if err := h.repo.CostRequest.FormRequest(id, userID); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cost request formed successfully"})
}

func (h *CostRequestHandler) ResolveCostRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost request ID"})
		return
	}

	// ratioCalculationResult := h.repo.CostRequest.CalculateRatio(id)

	deliveryDate := time.Now().AddDate(0, 1, 0)

	moderatorID := uint64(2)
	calculatedRatio, err := h.repo.CostRequest.ResolveOrRejectRequest(id, moderatorID, 4)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"message": "Cost request resolved successfully",
		"calculated_data": gin.H{
			"ratio calculation result": calculatedRatio,
			"delivery_date":            deliveryDate.Format("2006-01-02"),
		},
	}
	ctx.JSON(http.StatusOK, response)
}

func (h *CostRequestHandler) RejectCostRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost request ID"})
		return
	}
	moderatorID := uint64(2)
	_, err = h.repo.CostRequest.ResolveOrRejectRequest(id, moderatorID, 5)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cost request rejected successfully"})
}

func (h *CostRequestHandler) DeleteCostRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost request ID"})
		return
	}

	userID := GetFixedUserID()
	if err := h.repo.CostRequest.DeleteRequest(id, userID); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cost request"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cost request deleted successfully"})
}
