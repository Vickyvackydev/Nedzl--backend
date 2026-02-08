package utils

import (
	"api/models"
	"strings"

	"gorm.io/gorm"
)

func FindSimilarProducts(db *gorm.DB, product *models.Products, limit int) ([]models.Products, error) {

	var products []models.Products

	// we first extract keywords from product name

	keywords := strings.Fields(strings.ToLower(product.Name))

	query := db.Model(&products).Where("product_name = ?", product.Name).Where("id != ?", product.ID).Where("status = ?", models.StatusOngoing)

	// we apply the key word similarity

	keywordQuery := db

	for _, kw := range keywords {
		keywordQuery = keywordQuery.Or("product_name ILIKE ? OR description ILIKE ?", "%"+kw+"%", "%"+kw+"%")
	}

	query = query.Where(keywordQuery)

	// ranking the results by relevance (number of matching keywords)

	query = query.Order(`
	CASE
		WHEN product_name ILIKE '%` + keywords[0] + `%' THEN 1
		WHEN description ILIKE '%` + keywords[0] + `%' THEN 2
		ELSE 3
	 END,
	 created_at DESC`)

	err := query.Limit(limit).Preload("User").Find(&products).Error

	return products, err

}
