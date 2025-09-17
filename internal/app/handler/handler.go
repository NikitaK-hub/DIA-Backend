package handler

import (
	"DIA_Backend/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetCosts(ctx *gin.Context) {
	var costs []repository.Cost
	var err error

	searchQuery := ctx.Query("query") // получаем значение из поля поиска
	if searchQuery == "" {            // если поле поиска пусто, то просто получаем из репозитория все записи
		costs, err = h.Repository.GetCosts()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		costs, err = h.Repository.GetCostsByTitle(searchQuery) // в ином случае ищем заказ по заголовку
		if err != nil {
			logrus.Error(err)
		}
	}

	request, err := h.Repository.GetRequest()
	countRequest := len(request)
	if err != nil {
		logrus.Error(err)
	}
	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"costs":         costs,
		"query":         searchQuery,
		"count_request": countRequest,
	})
}

func (h *Handler) GetCost(ctx *gin.Context) {
	idStr := ctx.Param("id") // получаем id заказа из урла
	// через двоеточие мы указываем параметры, которые потом сможем считать через функцию выше
	id, err := strconv.Atoi(idStr) // так как функция выше возвращает нам строку, нужно ее преобразовать в int
	if err != nil {
		logrus.Error(err)
	}
	cost, err := h.Repository.GetCost(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "cost.html", gin.H{
		"cost": cost,
	})
}

func (h *Handler) GetRequest(ctx *gin.Context) {
	request, err := h.Repository.GetRequest()
	if err != nil {
		logrus.Error(err)
	}

	costs, err := h.Repository.GetCosts()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "request.html", gin.H{
		"request": request,
		"costs":   costs,
	})
}
