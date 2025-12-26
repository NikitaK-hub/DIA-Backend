package main

import (
	"DIA_Backend/internal/app/ds"
	"DIA_Backend/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load("deploy/.env")
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	err = db.AutoMigrate(
		&ds.Cost{},
		&ds.Cost_request{},
		&ds.Price_request_for_cost{},
		&ds.User{},
	)
	if err != nil {
		panic("cant migrate db")
	}
}
