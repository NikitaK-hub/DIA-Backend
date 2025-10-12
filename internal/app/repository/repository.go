package repository

import (
	"DIA_Backend/internal/app/dsn"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	Cost        *CostRepository
	CostRequest *CostRequestRepository
}

func NewRepository() (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Repository{
		Cost:        NewCostRepository(db),
		CostRequest: NewCostRequestRepository(db),
	}, nil
}
