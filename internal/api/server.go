package api

import (
	"DIA_Backend/internal/app/handler"
	"DIA_Backend/internal/app/repository"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func StartServer() {
	repo, err := repository.NewRepository()
	if err != nil {
		panic(err)
	}
	defer repository.CloseDBConn(repo)
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	handler.RegisterHandlers(router, repo)
	logrus.Debug("Starting server")
	router.Run(":" + os.Getenv("LISTEN_PORT"))
	logrus.Debug("Server stopped")
}
