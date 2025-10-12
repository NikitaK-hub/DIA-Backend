package main

import (
	"DIA_Backend/internal/api"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	err := godotenv.Load("deploy/.env")
	if err != nil {
		panic(err)
	}

	logrus.SetLevel(logrus.ErrorLevel)
	api.StartServer()
}
