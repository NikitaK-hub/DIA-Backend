package api

import (
	"DIA_Backend/internal/app/handler"
	"DIA_Backend/internal/app/repository"
	"log"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()

	// добавляем наш html/шаблон
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	templatesDir := filepath.Join(rootDir, "templates/*")
	r.LoadHTMLGlob(templatesDir)

	// Настраиваем статические файлы с абсолютным путем
	resourcesPath := filepath.Join(rootDir, "resources")
	r.Static("/static", resourcesPath)

	r.GET("/index", handler.GetCosts)
	r.GET("/cost/:id", handler.GetCost)
	r.GET("/request", handler.GetRequest)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	log.Println("Server down")
}
