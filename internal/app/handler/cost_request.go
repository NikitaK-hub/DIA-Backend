package handler

import (
	"DIA_Backend/internal/app/ds"
	"DIA_Backend/internal/app/repository"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
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
	Ratio        float64   `json:"Ratio"`
}

type CostRequestResponse struct {
	ID                 uint64                       `json:"id"`
	PriceRequestToCost []PriceRequestToCostResponse `json:"PriceRequestToCost"`
	Min_volume         uint64                       `json:"Min_volume"`
	Max_volume         uint64                       `json:"Max_volume"`
	Cost_price         uint64                       `json:"Cost_price"`
}

type CostRequestDetailResponse struct {
	ID                  uint64                             `json:"id"`
	CreatedAt           time.Time                          `json:"created_at"`
	Min_volume          uint64                             `json:"Min_volume"`
	Max_volume          uint64                             `json:"Max_volume"`
	Ratio               float64                            `json:"Ratio"`
	PriceRequestToCosts []PriceRequestToCostDetailResponse `json:"price_request_to_costs"`
	Status              uint8                              `json:"status"`
}

type PriceRequestToCostDetailResponse struct {
	Cost_price float64 `json:"cost_price"`
	CostTitle  string  `json:"cost_title"`
	CostImg    string  `json:"img"`
	CostID     uint64  `json:"cost_id"`
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

// @Summary      Get draft request info
// @Description  Get information about current user's draft request
// @Tags         cost-requests
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  CostRequestInfoResponse
// @Failure      500  {object}  map[string]interface{}
// @Router       /cost-requests/costRequestInfo [get]
func (h *CostRequestHandler) GetCostRequestInfo(ctx *gin.Context) {
	jwtStr := ctx.GetHeader("Authorization")
	const jwtPrefix = "Bearer "

	var userUUID = "Bearer "

	if strings.HasPrefix(jwtStr, jwtPrefix) {
		jwtStr = jwtStr[len(jwtPrefix):]

		claims := &ds.JWTClaims{}
		token, err := jwt.ParseWithClaims(jwtStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.repo.GetJWTSecret()), nil
		})

		if err == nil && token.Valid && !claims.IsRefresh {
			userUUID = claims.UserUUID.String()
		}
	}

	if userUUID == "" {
		ctx.JSON(http.StatusOK, CostRequestInfoResponse{
			RequestID: 0,
			ItemCount: -1,
		})
	}

	user, err := h.repo.User.GetUserByUUID(userUUID)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	requestID, itemCount, err := h.repo.CostRequest.GetDraftRequestInfo(user.ID)
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

// @Summary      Get cost requests
// @Description  Get a list of cost requests with optional filtering
// @Tags         cost-requests
// @Accept       json
// @Produce      json
// @Param        status query int false "Filter by status"
// @Param        date_from query string false "Filter by date from (YYYY-MM-DD)"
// @Param        date_to query string false "Filter by date to (YYYY-MM-DD)"
// @Security     BearerAuth
// @Success      200  {array}   CostsRequestsFilterResponse
// @Failure      500  {object}  map[string]interface{}
// @Router       /cost-requests [get]
func (h *CostRequestHandler) GetCostRequests(ctx *gin.Context) {
	userUUID, scopes, ok := GetUserFromContext(ctx)
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

	isModerator := false
	for _, scope := range scopes {
		if scope == "resolve:requests" || scope == "reject:requests" || scope == "manage:users" {
			isModerator = true
			break
		}
	}

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

	requests, err := h.repo.CostRequest.GetCostRequests(user.ID, isModerator, statusFilter, dateFrom, dateTo)
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
			Ratio:        costRequest.Ratio,
		})
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary      Get cost request by ID
// @Description  Get detailed information about a specific cost request
// @Tags         cost-requests
// @Accept       json
// @Produce      json
// @Param        id path int true "Cost Request ID"
// @Security     BearerAuth
// @Success      200  {object}  CostRequestDetailResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /cost-requests/{id} [get]
func (h *CostRequestHandler) GetCostRequestByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	userUUID, scopes, ok := GetUserFromContext(ctx)
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

	isModerator := false
	for _, scope := range scopes {
		if scope == "resolve:requests" || scope == "reject:requests" || scope == "manage:users" {
			isModerator = true
			break
		}
	}

	request, err := h.repo.CostRequest.GetCostRequestByID(id, user.ID, isModerator)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Cost request not found"})
		return
	}

	response := CostRequestDetailResponse{
		ID:         request.ID,
		CreatedAt:  request.CreatedAt,
		Min_volume: request.Min_volume,
		Max_volume: request.Max_volume,
		Ratio:      request.Ratio,
		Status:     request.Status,
	}

	for _, priceToRequest := range request.Price_request_for_cost {
		costDetail := PriceRequestToCostDetailResponse{
			CostTitle:  priceToRequest.Cost.Title,
			Cost_price: priceToRequest.Cost_price,
			CostImg:    priceToRequest.Cost.Img,
			CostID:     priceToRequest.Cost.ID,
		}
		response.PriceRequestToCosts = append(response.PriceRequestToCosts, costDetail)
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary      Update cost request
// @Description  Update an existing cost request
// @Tags         cost-requests
// @Accept       json
// @Produce      json
// @Param        id path int true "Cost Request ID"
// @Param        request body UpdateCostRequestResponse true "Cost request update data"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /cost-requests/{id} [put]
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

// @Summary      Form cost request
// @Description  Form a draft cost request into a submitted request
// @Tags         cost-requests
// @Accept       json
// @Produce      json
// @Param        id path int true "Cost Request ID"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /cost-requests/{id}/form [put]
func (h *CostRequestHandler) FormCostRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost request ID"})
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

	if err := h.repo.CostRequest.FormRequest(id, user.ID); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cost request formed successfully"})
}

// @Summary      Resolve cost request
// @Description  Resolve a cost request (moderator action)
// @Tags         cost-requests
// @Accept       json
// @Produce      json
// @Param        id path int true "Cost Request ID"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /cost-requests/{id}/resolve [put]
func (h *CostRequestHandler) ResolveCostRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost request ID"})
		return
	}

	userUUID, scopes, ok := GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	hasScope := slices.Contains(scopes, "resolve:requests")
	if !hasScope {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions. Moderator role required"})
		return
	}

	user, err := h.repo.User.GetUserByUUID(userUUID)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	deliveryDate := time.Now().AddDate(0, 1, 0)

	calculatedRatio, err := h.repo.CostRequest.ResolveOrRejectRequest(id, user.ID, 4)
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

// @Summary      Reject cost request
// @Description  Reject a cost request (moderator action)
// @Tags         cost-requests
// @Accept       json
// @Produce      json
// @Param        id path int true "Cost Request ID"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /cost-requests/{id}/reject [put]
func (h *CostRequestHandler) RejectCostRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost request ID"})
		return
	}

	userUUID, scopes, ok := GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	hasScope := slices.Contains(scopes, "reject:requests")
	if !hasScope {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions. Moderator role required"})
		return
	}

	user, err := h.repo.User.GetUserByUUID(userUUID)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	_, err = h.repo.CostRequest.ResolveOrRejectRequest(id, user.ID, 5)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cost request rejected successfully"})
}

// @Summary      Delete cost request
// @Description  Delete a cost request
// @Tags         cost-requests
// @Accept       json
// @Produce      json
// @Param        id path int true "Cost Request ID"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /cost-requests/{id} [delete]
func (h *CostRequestHandler) DeleteCostRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost request ID"})
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

	if err := h.repo.CostRequest.DeleteRequest(id, user.ID); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cost request"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cost request deleted successfully"})
}
