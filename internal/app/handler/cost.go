package handler

import (
	"net/http"
	"strconv"

	"DIA_Backend/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type CostHandler struct {
	repo *repository.Repository
}

func NewCostHandler(repository *repository.Repository) *CostHandler {
	return &CostHandler{repo: repository}
}

func (h *CostHandler) Register(router *gin.Engine) {
	router.GET("/cost/:id", h.GetCost)
}

func (h *CostHandler) GetCost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	costID, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}
	cost, err := h.repo.Cost.GetCost(costID)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "cost.html", gin.H{
		"cost": cost,
	})
}
