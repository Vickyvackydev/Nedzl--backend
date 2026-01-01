package handlers

import (
	"api/models"
	"api/utils"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func Contact(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var contact models.Contact

		if err := c.Bind(&contact); err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid input", err)

		}

		if err := db.Create(&contact).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to create a contact mail", err)
		}

		return utils.ResponseSucess(c, http.StatusOK, "Contact mail created successfully", nil)
	}

}
