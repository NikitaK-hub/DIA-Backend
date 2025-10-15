package repository

import (
	"DIA_Backend/internal/app/ds"
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type CostRepository struct {
	db          *gorm.DB
	minioClient *minio.Client
}

func NewCostRepository(db *gorm.DB, minioClient *minio.Client) *CostRepository {
	return &CostRepository{db: db}
}

func (r *CostRepository) GetCost(id uint64) (*ds.Cost, error) {
	var cost ds.Cost
	err := r.db.
		Where("is_deleted = false").
		First(&cost, id).Error

	if err != nil {
		return nil, err
	}
	return &cost, err
}

func (r *CostRepository) GetCosts(titleFilter string) ([]ds.Cost, error) {
	var costs []ds.Cost
	query := r.db.Where("is_deleted = false")
	if titleFilter != "" {
		query = query.Where("title ILIKE ?", "%"+titleFilter+"%")
	}
	err := query.Find(&costs).Error
	if err != nil {
		return nil, err
	}
	return costs, err
}

func (r *CostRepository) CreateCost(cost *ds.Cost) error {
	cost.IsDeleted = false
	return r.db.Create(cost).Error
}

func (r *CostRepository) UpdateCost(id uint64, data *ds.Cost) error {
	return r.db.Model(&ds.Cost{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(map[string]interface{}{
			"title":       data.Title,
			"info":        data.Info,
			"type_change": data.Type_change,
		}).Error
}

func (r *CostRepository) DeleteCost(id uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var cost ds.Cost
		if err := tx.First(&cost, id).Error; err != nil {
			return err
		}

		if cost.Img != "" {
			if err := r.deleteImageFile(cost.Img); err != nil {
				return err
			}
		}

		return tx.Model(&ds.Cost{}).Where("id = ?", id).Update("is_deleted", true).Error
	})
}

func (r *CostRepository) AddCostImage(id uint64, fileHeader *multipart.FileHeader) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var cost ds.Cost
		if err := tx.Where("is_deleted = false").First(&cost, id).Error; err != nil {
			return err
		}

		if cost.Img != "" {
			if err := r.deleteImageFile(cost.Img); err != nil {
				return err
			}
		}

		fileExt := filepath.Ext(fileHeader.Filename)
		newFileName := fmt.Sprintf("cost_%d_%d%s", id, time.Now().Unix(), fileExt)
		newFileName = strings.ToLower(newFileName)

		imageURL, err := r.SaveImageToMinio(newFileName, fileHeader)
		if err != nil {
			return err
		}

		return tx.Model(&cost).Update("image_url", imageURL).Error
	})
}

func (r *CostRepository) AddCostToDraftRequest(costID uint64, userID uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var cost ds.Cost
		if err := tx.Where("is_deleted = false").First(&cost, costID).Error; err != nil {
			return err
		}

		var request ds.Cost_request
		err := tx.Where("status = 1 AND user_id = ?", userID).First(&request).Error
		if err == gorm.ErrRecordNotFound {
			request = ds.Cost_request{
				Status:    1,
				ID_user:   userID,
				CreatedAt: time.Now(),
			}
			if err := tx.Create(&request).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		priceRequestForCost := ds.Price_request_for_cost{
			ID_request: request.ID,
			ID_cost:    costID,
			Cost_price: 0,
		}

		return tx.Create(&priceRequestForCost).Error
	})
}

const costImagesBucket = "costs"

func (r *CostRepository) SaveImageToMinio(fileName string, fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileSize := fileHeader.Size

	contentType := "application/octet-stream"
	if strings.HasSuffix(strings.ToLower(fileName), ".jpg") || strings.HasSuffix(strings.ToLower(fileName), ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(strings.ToLower(fileName), ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(strings.ToLower(fileName), ".gif") {
		contentType = "image/gif"
	}

	_, err = r.minioClient.PutObject(context.Background(), costImagesBucket, fileName, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s/%s/%s", os.Getenv("MINIO_HOST"), os.Getenv("MINIO_SERVER_PORT"), costImagesBucket, fileName), nil
}

func (r *CostRepository) deleteImageFile(imageURL string) error {
	if strings.Contains(imageURL, "localhost:9000") {
		fmt.Printf("Image deleted from MinIO: %s\n", imageURL)
	}
	return nil
}
