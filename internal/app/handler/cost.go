package handler

import (
	"DIA_Backend/internal/app/ds"
	"DIA_Backend/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type CostHandler struct {
	repo *repository.Repository
}

func NewCostHandler(repository *repository.Repository) *CostHandler {
	return &CostHandler{repo: repository}
}

type CreateCostRequest struct {
	Title       string `json:"title" binding:"required"`
	Info        string `json:"info"`
	Type_change bool   `json:"type_change" binding:"required"`
}

type UpdateCostRequest struct {
	Title       *string `json:"title"`
	Info        *string `json:"info"`
	Type_change *bool   `json:"type_change"`
}

type CostsFilterResponse struct {
	ID       uint64 `json:"id"`
	Title    string `json:"title"`
	Info     string `json:"info"`
	ImageURL string `json:"image_url"`
}

type CostResponse struct {
	ID          uint64 `json:"id"`
	Title       string `json:"title"`
	Type_change bool   `json:"type_change"`
	Info        string `json:"info"`
	Img         string `json:"image_url"`
}

// @Description  Get a list of all cost with optional title search
// @Tags         costs
// @Accept       json
// @Produce      json
// @Param        title query string false "Search cost by title"
// @Security     BearerAuth
// @Success      200  {array}   CostsFilterResponse
// @Failure      500  {object}  map[string]interface{}
// @Router       /costs [get]
func (h *CostHandler) GetCosts(ctx *gin.Context) {
	searchQuery := ctx.Query("title")

	costs, err := h.repo.Cost.GetCosts(searchQuery)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get costs"})
		return
	}

	var response []CostsFilterResponse
	for _, cost := range costs {
		response = append(response, CostsFilterResponse{
			ID:       cost.ID,
			Title:    cost.Title,
			Info:     cost.Info,
			ImageURL: cost.Img,
		})
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary      Get cost by ID
// @Description  Get detailed information about a specific cost
// @Tags         costs
// @Accept       json
// @Produce      json
// @Param        id path int true "Cost ID"
// @Security     BearerAuth
// @Success      200  {object}  CostResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /costs/{id} [get]
func (h *CostHandler) GetCostByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost ID"})
		return
	}

	cost, err := h.repo.Cost.GetCost(id)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Cost not found"})
		return
	}

	responce := CostResponse{
		ID:          cost.ID,
		Title:       cost.Title,
		Type_change: cost.Type_change,
		Info:        cost.Info,
		Img:         cost.Img,
	}
	ctx.JSON(http.StatusOK, responce)
}

// @Summary      Create a new cost
// @Description  Create a new cost with the provided data
// @Tags         costs
// @Accept       json
// @Produce      json
// @Param        request body CreateCostRequest true "Cost creation data"
// @Security     BearerAuth
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /costs [post]
func (h *CostHandler) CreateCost(ctx *gin.Context) {
	var req CreateCostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("CreateCost validation failed: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	cost := &ds.Cost{
		Title:       req.Title,
		Info:        req.Info,
		Type_change: req.Type_change,
	}

	if err := h.repo.Cost.CreateCost(cost); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cost"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Cost created successfully",
		"id_cost": cost.ID,
	})
}

// @Summary      Update cost
// @Description  Update an existing cost with new data
// @Tags         costs
// @Accept       json
// @Produce      json
// @Param        id path int true "Cost ID"
// @Param        request body UpdateCostRequest true "Cost update data"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /costs/{id} [put]
func (h *CostHandler) UpdateCost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost ID"})
		return
	}

	var req UpdateCostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	costData := &ds.Cost{}
	if req.Title != nil {
		costData.Title = *req.Title
	}
	if req.Info != nil {
		costData.Info = *req.Info
	}
	if req.Type_change != nil {
		costData.Type_change = *req.Type_change
	}
	if err := h.repo.Cost.UpdateCost(id, costData); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cost"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cost updated successfully"})
}

// @Summary      Delete cost
// @Description  Delete a cost by ID
// @Tags         costs
// @Accept       json
// @Produce      json
// @Param        id path int true "Cost ID"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /costs/{id} [delete]
func (h *CostHandler) DeleteCost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost ID"})
		return
	}

	if err := h.repo.Cost.DeleteCost(id); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cost"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cost deleted successfully"})
}

// @Summary      Add image to cost
// @Description  Upload and attach an image to a cost
// @Tags         costs
// @Accept       multipart/form-data
// @Produce      json
// @Param        id path int true "Cost ID"
// @Param        image formData file true "Cost image file"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /costs/{id}/image [post]
func (h *CostHandler) AddCostImage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost ID"})
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}

	if err := h.repo.Cost.AddCostImage(id, file); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add image"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Image added successfully"})
}

// @Summary      Add cost to draft request
// @Description  Add a cost to the current user's draft request
// @Tags         costs
// @Accept       json
// @Produce      json
// @Param        id path int true "Cost ID"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /costs/{id}/add-to-request [post]
func (h *CostHandler) AddCostToDraftRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logrus.Errorf("AddCostToDraftRequest validation failed: %v", err)
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

	if err := h.repo.Cost.AddCostToDraftRequest(id, user.ID); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add cost to draft request"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cost added to draft request successfully"})
}
