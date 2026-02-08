package utils

import (
	"api/models"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func FindSimilarProducts(db *gorm.DB, product *models.Products, limit int) ([]models.Products, error) {
	var products []models.Products

	// Extract keywords from product name
	keywords := strings.Fields(strings.ToLower(product.Name))
	if len(keywords) == 0 && product.CategoryName == "" {
		return products, nil
	}

	// Build the base query: exclude current product and only active ones
	query := db.Model(&models.Products{}).
		Where("id != ?", product.ID).
		Where("status = ?", models.StatusOngoing)

	// Broadened Filter: (Match any significant keyword OR Same category)
	filter := db.Where("category_name = ?", product.CategoryName)
	for _, kw := range keywords {
		// Only use keywords longer than 2 characters to avoid noise (a, the, in, etc.)
		if len(kw) > 2 {
			filter = filter.Or("name ILIKE ?", "%"+kw+"%").Or("description ILIKE ?", "%"+kw+"%")
		}
	}
	query = query.Where(filter)

	// Multi-tiered Ranking (Order by relevance)
	// We escape single quotes for literal SQL injection safety in the ORDER BY clause
	escapedName := strings.ReplaceAll(product.Name, "'", "''")
	escapedCat := strings.ReplaceAll(product.CategoryName, "'", "''")

	orderBy := fmt.Sprintf(`
		CASE 
			WHEN name = '%s' THEN 0           -- Highest: Exact name match
			WHEN name ILIKE '%%%s%%' THEN 1   -- High: Contains full original name
			`, escapedName, escapedName)

	if len(keywords) > 0 {
		firstKw := strings.ReplaceAll(keywords[0], "'", "''")
		orderBy += fmt.Sprintf("WHEN name ILIKE '%%%s%%' THEN 2 -- Medium: Matches first keyword\n", firstKw)
	}

	orderBy += fmt.Sprintf(`
			WHEN category_name = '%s' THEN 3  -- Low: Same category match
			ELSE 4 
		END, 
		created_at DESC`, escapedCat)

	err := query.Order(orderBy).
		Limit(limit).
		Preload("User").
		Find(&products).Error

	return products, err
}
