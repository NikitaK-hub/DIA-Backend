package handler

import (
	"DIA_Backend/internal/app/repository"

	"github.com/gin-gonic/gin"
)

func RegisterHandlers(router *gin.Engine, repo *repository.Repository) {
	router.LoadHTMLGlob("./templates/*")
	router.Static("/static", "./resources")

	StageRequestHandler := NewCostRequestHandler(repo)
	StageHandler := NewCostHandler(repo)
	StagesHandler := NewCostsHandler(repo)

	StageRequestHandler.Register(router)
	StageHandler.Register(router)
	StagesHandler.Register(router)
}
