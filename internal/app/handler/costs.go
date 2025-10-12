package handler

import (
	"DIA_Backend/internal/app/ds"
	"DIA_Backend/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type CostsHandler struct {
	repo *repository.Repository
}

func NewCostsHandler(repository *repository.Repository) *CostsHandler {
	return &CostsHandler{repo: repository}
}

func (h *CostsHandler) Register(router *gin.Engine) {
	router.GET("/costs", h.GetCosts)
	router.POST("/costs", h.AddCostToRequest)
}

func (h *CostsHandler) GetCosts(ctx *gin.Context) {
	var costs []ds.Cost
	var err error

	searchCostTitle := ctx.Query("cost_title")
	if searchCostTitle == "" {
		costs, err = h.repo.Cost.GetCosts()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		costs, err = h.repo.Cost.GetCostsByTitle(searchCostTitle)
		if err != nil {
			logrus.Error(err)
		}
	}

	costRequestID, costRequestEntryCount, err := h.repo.CostRequest.GetCostRequestIDEntryCountByUserID(1)
	if err != nil {
		logrus.Error(err)
	}
	ctx.HTML(http.StatusOK, "costs.html", gin.H{
		"costs":                 costs,
		"cost_title":            searchCostTitle,
		"costRequestID":         costRequestID,
		"costRequestEntryCount": costRequestEntryCount,
	})
}

func (h *CostsHandler) AddCostToRequest(ctx *gin.Context) {
	costIDStr := ctx.PostForm("cost-id")
	costID, err := strconv.ParseUint(costIDStr, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusBadRequest)
		ctx.Redirect(http.StatusSeeOther, "/costs")
		return
	}
	err = h.repo.CostRequest.AddCostToCostRequest(costID, 1)
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusBadRequest)
		ctx.Redirect(http.StatusSeeOther, "/costs")
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/costs")
}
