package handlers

import (
	"api/models"
	"api/utils"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func UpdateUserStatus(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		var body struct {
			Status string `json:"status"`
		}

		if err := c.Bind(&body); err != nil {
			return utils.ResponseError(c, 400, "Invalid input", err)
		}

		if body.Status == "" || !models.IsValidUserStatus(models.Status(body.Status)) {
			return utils.ResponseError(c, 400, "Invalid status", nil)
		}

		// Update status directly
		result := db.Model(&models.User{}).Where("id = ?", id).Update("status", body.Status)
		if result.Error != nil {
			return utils.ResponseError(c, 500, "Failed to update product status", result.Error)
		}

		if result.RowsAffected == 0 {
			return utils.ResponseError(c, 404, "User not found", nil)
		}

		return utils.ResponseSucess(c, 200, "User status updated", map[string]string{"status": body.Status})
	}
}
