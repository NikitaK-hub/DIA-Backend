package repository

import (
	"time"

	"gorm.io/gorm"

	"DIA_Backend/internal/app/ds"
	"errors"

	"github.com/sirupsen/logrus"
)

type CostRequestRepository struct {
	db *gorm.DB
}

func NewCostRequestRepository(db *gorm.DB) *CostRequestRepository {
	return &CostRequestRepository{db: db}
}

func (r *CostRequestRepository) GetDraftRequestInfo(userID uint64) (uint64, int, error) {
	var request ds.Cost_request
	err := r.db.
		Where("status = 1 AND id_user = ?", userID).
		First(&request).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, 0, nil
	} else if err != nil {
		return 0, 0, err
	}

	var count int64
	err = r.db.
		Model(&ds.Price_request_for_cost{}).
		Where("ID_request = ?", request.ID).
		Count(&count).Error

	if err != nil {
		return 0, 0, err
	}

	return request.ID, int(count), nil
}

func (r *CostRequestRepository) GetCostRequests(userID uint64, isModerator bool, statusFilter uint8, dateFrom, dateTo *time.Time) ([]ds.Cost_request, error) {
	var requests []ds.Cost_request

	query := r.db.
		Preload("Price_request_for_cost").
		Preload("Price_request_for_cost.Cost").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username")
		}).
		Preload("Moderator", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username")
		}).
		Where("status > 2").Order("status ASC")

	if !isModerator {
		query = query.Where("id_user = ?", userID)
	}

	if statusFilter != 0 {
		query = query.Where("status = ?", statusFilter)
	}

	if dateFrom != nil {
		query = query.Where("formed_at >= ?", dateFrom)
	}
	if dateTo != nil {
		query = query.Where("formed_at <= ?", dateTo)
	}

	err := query.Find(&requests).Error
	if err != nil {
		return nil, err
	}

	return requests, nil
}

func (r *CostRequestRepository) GetCostRequestByID(id uint64, ID_user uint64, isModerator bool) (*ds.Cost_request, error) {
	var request ds.Cost_request
	query := r.db.
		Preload("Price_request_for_cost").
		Preload("Price_request_for_cost.Cost").
		Where("status != 2")

	if !isModerator {
		query = query.Where("id_user = ?", ID_user)
	}

	err := query.First(&request, id).Error

	if err != nil {
		return nil, err
	}

	return &request, nil
}

func (r *CostRequestRepository) FormRequest(id uint64, userID uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var request ds.Cost_request
		err := tx.
			Preload("Price_request_for_cost").
			Where("id = ? AND id_user = ? AND status = 1", id, userID).
			First(&request).Error

		if err != nil {
			return err
		}

		if len(request.Price_request_for_cost) == 0 {
			return errors.New("at least one cost is required")
		}

		return tx.Model(&request).Updates(map[string]interface{}{
			"status":    3,
			"formed_at": time.Now(),
		}).Error
	})
}

func (r *CostRequestRepository) ResolveOrRejectRequest(id uint64, moderatorID uint64, status uint8) error {
	if status != 4 && status != 5 {
		return errors.New("invalid status for moderator action")
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var request ds.Cost_request
		err := tx.
			Preload("Price_request_for_cost").
			Preload("Price_request_for_cost.Cost").
			Where("id = ? AND status = 3", id).
			First(&request).Error

		if err != nil {
			return err
		}
		logrus.Info(request.ID)

		updates := map[string]any{
			"status":       status,
			"id_moderator": moderatorID,
			"closed_at":    time.Now(),
		}

		return tx.Model(&request).Updates(updates).Error
	})
	return err
}

func (r *CostRequestRepository) DeleteCostRequest(id uint64, userID uint64) error {
	return r.db.
		Model(&ds.Cost_request{}).
		Where("id = ? AND id_user = ? AND status = 1", id, userID).
		Update("status", 2).Error
}

func (r *CostRequestRepository) RemoveCostFromRequest(requestID uint64, costID uint64, userID uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var request ds.Cost_request
		err := tx.
			Where("id = ? AND id_user = ? AND status = 1", requestID, userID).
			First(&request).Error

		if err != nil {
			return err
		}

		return tx.
			Where("id_request = ? AND id_cost = ?", requestID, costID).
			Delete(&ds.Price_request_for_cost{}).Error
	})
}

func (r *CostRequestRepository) UpdateRequestToCost(requestID uint64, costID uint64, userID uint64, cost_price *float64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var request ds.Cost_request
		err := tx.
			Where("id = ? AND id_user = ? AND status = 1", requestID, userID).
			First(&request).Error

		if err != nil {
			return err
		}

		updates := make(map[string]any)
		if cost_price != nil {
			updates["cost_price"] = *cost_price
		}

		if len(updates) == 0 {
			return nil
		}

		return tx.
			Model(&ds.Price_request_for_cost{}).
			Where("id_request = ? AND id_cost = ?", requestID, costID).
			Updates(updates).Error
	})
}

func (r *CostRequestRepository) UpdateCostRequest(id uint64, Min_volume *uint64, Max_volume *uint64) error {
	updates := make(map[string]any)

	if Min_volume != nil {
		updates["Min_volume"] = *Min_volume
	}

	if Max_volume != nil {
		updates["Max_volume"] = *Max_volume
	}

	if len(updates) == 0 {
		return nil
	}

	return r.db.
		Model(&ds.Cost_request{}).
		Where("id = ? AND status != 2", id).
		Updates(updates).Error
}

func (r *CostRequestRepository) UpdateCostRequestEmission(id uint64, Ratio float64) error {
	updates := make(map[string]any)
	updates["ratio"] = Ratio

	if len(updates) == 0 {
		return nil
	}

	return r.db.
		Model(&ds.Cost_request{}).
		Where("id = ? AND status = 4", id).
		Updates(updates).Error
}

func (r *CostRequestRepository) CalculateRatio(requestID uint64) float64 {
	var priceRequestToCosts []ds.Price_request_for_cost

	err := r.db.
		Preload("Cost").
		Where("id_request = ?", requestID).
		Find(&priceRequestToCosts).Error

	if err != nil {
		return 0
	}

	var CalculationResultFixCosts float64 = 0
	var CalculationResultMinChangeCosts float64 = 0
	var CalculationResultMaxChangeCosts float64 = 0
	for _, priceRequestToCost := range priceRequestToCosts {
		cost := priceRequestToCost.Cost
		request := priceRequestToCost.Cost_request
		if cost.Type_change {
			CalculationResultMinChangeCosts += priceRequestToCost.Cost_price * float64(request.Min_volume)
			CalculationResultMaxChangeCosts += priceRequestToCost.Cost_price * float64(request.Max_volume)
		}
		if !cost.Type_change {
			CalculationResultFixCosts += priceRequestToCost.Cost_price
		}
	}
	var totalEmission float64 = (CalculationResultFixCosts + CalculationResultMinChangeCosts) / (CalculationResultFixCosts + CalculationResultMaxChangeCosts)
	return totalEmission
}

func (r *CostRequestRepository) AddCostToCostRequest(costID uint64, userID uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var cost ds.Cost
		err := tx.First(&cost, costID).Error
		if err != nil {
			return err
		}

		var costRequest ds.Cost_request
		err = tx.
			Where("status = 1 AND id_user = ?", userID).
			Take(&costRequest).Error

		notFound := errors.Is(err, gorm.ErrRecordNotFound)
		if err != nil && !notFound {
			return err
		}

		if notFound {
			costRequest = ds.Cost_request{
				ID_user: userID,
			}
			err := tx.Create(&costRequest).Error
			if err != nil {
				return err
			}
		}

		costRequestToCost := ds.Price_request_for_cost{
			ID_request: costRequest.ID,
			ID_cost:    costID,
		}
		return tx.Create(&costRequestToCost).Error
	})
}

func (r *CostRequestRepository) DeleteRequest(id uint64, userID uint64) error {
	return r.db.
		Model(&ds.Cost_request{}).
		Where("id = ? AND id_user = ? AND status = 1", id, userID).
		Update("status", 2).Error
}
