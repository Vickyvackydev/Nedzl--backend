package utils

import (
	"api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CountUserProducts(db *gorm.DB, UserID uuid.UUID, status ...string) (int64, error) {
	var count int64

	query := db.Model(&models.Products{}).Where("user_id = ?", UserID)

	if len(status) > 0 {
		query = query.Where("status = ?", status[0])
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
