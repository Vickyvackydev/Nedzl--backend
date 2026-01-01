package handlers

import (
	"api/models"
	"api/utils"
	"net/http"

	"github.com/google/uuid"
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

func GetContact(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var contact models.Contact

		if err := db.Find(&contact).Order("created_at DESC").Error; err != nil {
			return utils.ResponseError(c, 500, "Failed to retrieve contact lists", err)

		}

		return utils.ResponseSucess(c, http.StatusOK, "Contact lists retrieved successfully", echo.Map{"data": contact})
	}
}

func DeleteContact(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		ContactId := c.Param("id")

		Cuid, err := uuid.Parse(ContactId)
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid contact ID", err)
		}

		if err := db.Model(&models.Contact{}).Where("id = ?", Cuid).Delete(&models.Contact{}).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to delete contact mail", err)
		}

		return utils.ResponseSucess(c, http.StatusOK, "Contact mail deleted successfully", nil)
	}
}
