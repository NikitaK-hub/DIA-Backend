package repository

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"DIA_Backend/internal/app/ds"
	"errors"
)

type CostRequestRepository struct {
	db *gorm.DB
}

func NewCostRequestRepository(db *gorm.DB) *CostRequestRepository {
	return &CostRequestRepository{db: db}
}

func (r *CostRequestRepository) GetCostRequestIDEntryCountByUserID(ID_user int) (int, int, error) {
	var cost_request ds.Cost_request
	err := r.db.
		Model(&ds.Cost_request{}).
		Where("status = 1 and ID_user = ?", ID_user).
		Take(&cost_request).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, 0, nil
	} else if err != nil {
		return 0, 0, err
	}

	var count int64
	err = r.db.
		Model(&ds.Cost_request{}).
		Where("id = ?", cost_request.ID).
		Joins("Price_request_for_cost").
		Count(&count).Error

	if err != nil {
		return 0, 0, err
	}

	return int(cost_request.ID), int(count), nil
}

func (r *CostRequestRepository) GetCostRequestByID(id uint64, ID_user uint64) (*ds.Cost_request, error) {
	var costRequest ds.Cost_request
	err := r.db.
		Preload("Price_request_for_cost").
		Preload("Price_request_for_cost.Cost").
		Where("status = 1 and ID_user = ?", ID_user).
		First(&costRequest, id).Error

	if err != nil {
		return nil, err
	}

	return &costRequest, nil
}

func (r *CostRequestRepository) AddCostToCostRequest(ID_cost uint64, ID_user uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var cost ds.Cost
		err := tx.First(&cost, ID_cost).Error
		if err != nil {
			return err
		}

		var costRequest ds.Cost_request
		err = tx.
			Where("status = 1 and ID_user = ?", ID_user).
			Take(&costRequest).Error
		notFound := errors.Is(err, gorm.ErrRecordNotFound)
		if err != nil && !notFound {
			return err
		}

		if notFound {
			costRequest = ds.Cost_request{User: ds.User{ID: ID_user}}

			err = tx.Create(&costRequest).Error
			if err != nil {
				return err
			}

		}

		costRequestToCost := ds.Price_request_for_cost{ID_request: costRequest.ID, ID_cost: ID_cost}
		err = tx.Create(&costRequestToCost).Error
		if err != nil {
			return err
		}

		return nil
	})
}

func (r *CostRequestRepository) DeleteCostRequest(ID_request uint64, ID_user uint64) error {
	logrus.Infof("Executing SQL UPDATE for cost request ID: %d, user ID: %d", ID_request, ID_user)

	query := "UPDATE cost_requests SET status = 2 WHERE status = 1 AND id = ? AND ID_user = ?"
	result := r.db.Exec(query, ID_request, ID_user)

	if result.Error != nil {
		logrus.Errorf("SQL UPDATE failed: %v", result.Error)
		return result.Error
	}

	rowsAffected := result.RowsAffected
	logrus.Infof("SQL UPDATE completed, rows affected: %d", rowsAffected)

	if rowsAffected == 0 {
		logrus.Warnf("No rows affected - cost request not found or already deleted (ID: %d, User: %d)", ID_request, ID_user)
		return gorm.ErrRecordNotFound
	}

	logrus.Infof("Successfully updated cost request status to 2 for ID: %d", ID_request)
	return nil
}
