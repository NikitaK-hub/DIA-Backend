package handler

import (
	"DIA_Backend/internal/app/ds"
	"DIA_Backend/internal/app/repository"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CostRequestHandler struct {
	repo *repository.Repository
}

func NewCostRequestHandler(repository *repository.Repository) *CostRequestHandler {
	return &CostRequestHandler{repo: repository}
}

func (h *CostRequestHandler) Register(router *gin.Engine) {
	router.GET("/cost_request/:id", h.GetCostRequest)
	router.POST("/cost_request/:id", h.DeleteCostRequest)
}

type CostRequestTemplateEntry struct {
	Cost       ds.Cost
	Cost_price string
}

func NewCostRequestTemplateEntry(stageReqEntry *ds.Price_request_for_cost) *CostRequestTemplateEntry {
	return &CostRequestTemplateEntry{
		Cost:       stageReqEntry.Cost,
		Cost_price: strconv.FormatUint(uint64(stageReqEntry.Cost_price), 10),
	}
}

type CostRequestTemplate struct {
	ID                 uint64
	Entries            []CostRequestTemplateEntry
	Min_release_volume uint64
	Max_release_volume uint64
	Ratio              float64
}

func (h *CostRequestHandler) GetCostRequest(ctx *gin.Context) {
	costIDStr := ctx.Param("id")
	reqID, err := strconv.ParseUint(costIDStr, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusSeeOther, "/costs")
		return
	}

	costRequest, err := h.repo.CostRequest.GetCostRequestByID(reqID, 1)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusSeeOther, "/costs")
		return
	}

	costReqTemplate := CostRequestTemplate{
		ID:                 costRequest.ID,
		Min_release_volume: costRequest.Min_volume,
		Max_release_volume: costRequest.Max_volume,
		Ratio:              costRequest.Ratio,
	}

	for _, costReqToLamp := range costRequest.Price_request_for_cost {
		costReqTemplate.Entries = append(costReqTemplate.Entries, *NewCostRequestTemplateEntry(&costReqToLamp))
	}

	ctx.HTML(http.StatusOK, "cost_request.html", gin.H{
		"costRequest": &costReqTemplate,
	})
}

func (h *CostRequestHandler) DeleteCostRequest(ctx *gin.Context) {
	requestIDStr := ctx.PostForm("request-id")
	requestID, err := strconv.ParseUint(requestIDStr, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	err = h.repo.CostRequest.DeleteCostRequest(requestID, 1)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		logrus.Error(err)
		ctx.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/costs")
}
