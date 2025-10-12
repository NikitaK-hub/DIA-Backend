package api

import (
	"DIA_Backend/internal/app/handler"
	"DIA_Backend/internal/app/repository"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	repo, err := repository.NewRepository()
	if err != nil {
		panic(err)
	}
	router := gin.Default()
	handler.RegisterHandlers(router, repo)
	logrus.Debug("Starting server")
	router.Run(":" + os.Getenv("LISTEN_PORT"))
	logrus.Debug("Server stopped")
}
