package repository

import (
	"DIA_Backend/internal/app/ds"
	"fmt"

	"gorm.io/gorm"
)

type CostRepository struct {
	db *gorm.DB
}

func NewCostRepository(db *gorm.DB) *CostRepository {
	return &CostRepository{db: db}
}

func (r *CostRepository) GetCosts() ([]ds.Cost, error) {
	var costs []ds.Cost
	err := r.db.Find(&costs).Error
	if err != nil {
		return nil, err
	}
	if len(costs) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return costs, nil
}

func (r *CostRepository) GetCost(id int) (ds.Cost, error) {
	cost := ds.Cost{}
	err := r.db.Where("id = ?", id).First(&cost).Error
	if err != nil {
		return ds.Cost{}, err
	}
	return cost, nil
}

func (r *CostRepository) GetCostsByTitle(title string) ([]ds.Cost, error) {
	var costs []ds.Cost
	err := r.db.Where("title ILIKE ?", "%"+title+"%").Find(&costs).Error
	if err != nil {
		return nil, err
	}
	return costs, nil
}
