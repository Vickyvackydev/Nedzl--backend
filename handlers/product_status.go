package handlers

import (
	"api/models"
	"api/utils"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func UpdateProductStatus(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		var body struct {
			Status string `json:"status"`
		}

		if err := c.Bind(&body); err != nil {
			return utils.ResponseError(c, 400, "Invalid input", err)
		}

		if body.Status == "" || !models.IsValidStatus(models.Status(body.Status)) {
			return utils.ResponseError(c, 400, "Invalid status", nil)
		}

		// Prepare update data
		updateData := map[string]interface{}{
			"status": body.Status,
		}

		// If status is CLOSED, set closed_at timestamp
		if models.Status(body.Status) == models.StatusClosed {
			now := time.Now()
			updateData["closed_at"] = &now
		}

		// Update with all fields
		result := db.Model(&models.Products{}).Where("id = ?", id).Updates(updateData)
		if result.Error != nil {
			return utils.ResponseError(c, 500, "Failed to update product status", result.Error)
		}

		if result.RowsAffected == 0 {
			return utils.ResponseError(c, 404, "Product not found", nil)
		}

		return utils.ResponseSucess(c, 200, "Product status updated", map[string]string{"status": body.Status})
	}
}
