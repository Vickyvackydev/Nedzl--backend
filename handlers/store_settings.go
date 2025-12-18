package handlers

import (
	"api/models"
	"api/utils"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type StoreResponse struct {
	ID                uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	BusinessName      string            `json:"business_name"`
	AboutCompany      string            `json:"about_company"`
	StoreName         string            `json:"store_name"`
	Address           string            `json:"address"`
	State             string            `json:"state"`
	HowDoWeLocateYou  string            `json:"how_do_we_locate_you"`
	BusinessHoursFrom string            `json:"business_hours_from"`
	BusinessHoursTo   string            `json:"business_hours_to"`
	Region            string            `json:"region"`
	UserID            uuid.UUID         `json:"user_id"` // needed to link with the currently authenticated user
	User              models.PublicUser `json:"user"`

	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func convertToStoreResponse(settings models.StoreSetting) StoreResponse {
	publicUser := models.PublicUser{
		ID:          settings.User.ID,
		UserName:    settings.User.UserName,
		Email:       settings.User.Email,
		Role:        string(settings.User.Role),
		PhoneNumber: settings.User.PhoneNumber,
		ImageUrl:    settings.User.ImageUrl,
		Location:    settings.User.Location,
		Status:      settings.User.Status,
		IsVerified:  settings.User.IsVerified,
		CreatedAt:   settings.User.CreatedAt,
		UpdatedAt:   settings.User.UpdatedAt,
		DeletedAt:   settings.User.DeletedAt,
	}

	return StoreResponse{
		ID:                settings.ID,
		BusinessName:      settings.BusinessName,
		AboutCompany:      settings.AboutCompany,
		StoreName:         settings.StoreName,
		State:             settings.State,
		Address:           settings.Address,
		HowDoWeLocateYou:  settings.HowDoWeLocateYou,
		BusinessHoursFrom: settings.BusinessHoursFrom,
		BusinessHoursTo:   settings.BusinessHoursTo,
		Region:            settings.Region,
		UserID:            settings.UserID,
		User:              publicUser,
		CreatedAt:         settings.CreatedAt,
		UpdatedAt:         settings.UpdatedAt,
		DeletedAt:         settings.DeletedAt,
	}
}
func CreateStoreSettings(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var body models.StoreSetting
		if err := c.Bind(&body); err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid input", err)
		}

		// user must be authenticated; set foreign key
		uid, ok := c.Get("user_id").(uuid.UUID)
		if !ok {
			return utils.ResponseError(c, http.StatusUnauthorized, "Unauthorized", nil)
		}

		if body.BusinessName == "" || body.AboutCompany == "" || body.StoreName == "" || body.Address == "" || body.State == "" || body.HowDoWeLocateYou == "" || body.BusinessHoursFrom == "" || body.BusinessHoursTo == "" {
			return utils.ResponseError(c, http.StatusBadRequest, "All fields are required", nil)
		}

		body.UserID = uid

		if err := db.Create(&body).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to create store settings", err)
		}

		// Load related user for response
		if err := db.Preload("User").First(&body, "id = ?", body.ID).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to load created settings", err)
		}
		response := convertToStoreResponse(body)

		return utils.ResponseSucess(c, http.StatusCreated, "Store settings created successfully", echo.Map{"settings": response})

	}

}

func UpdateStoreSettings(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uuid.UUID)
		ID := c.Param("id")
		uid, err := uuid.Parse(ID)
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid id", err)
		}

		var storeSettings models.StoreSetting
		if err := c.Bind(&storeSettings); err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid input", err)
		}

		var existing models.StoreSetting
		if err := db.Where("user_id = ? AND id = ?", userID, uid).First(&existing).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "Store settings not found", err)
		}

		// update allowed fields
		existing.BusinessName = storeSettings.BusinessName
		existing.AboutCompany = storeSettings.AboutCompany
		existing.StoreName = storeSettings.StoreName
		existing.Address = storeSettings.Address
		existing.State = storeSettings.State
		existing.Region = storeSettings.Region
		existing.HowDoWeLocateYou = storeSettings.HowDoWeLocateYou
		existing.BusinessHoursFrom = storeSettings.BusinessHoursFrom
		existing.BusinessHoursTo = storeSettings.BusinessHoursTo

		if err := db.Save(&existing).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to update store settings", err)
		}

		if err := db.Preload("User").First(&existing, "id = ?", existing.ID).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to load updated settings", err)
		}

		return utils.ResponseSucess(c, http.StatusOK, "Store settings updated successfully", echo.Map{"data": existing})

	}

}

func GetStoreSettings(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		ID := c.Param("id")
		uid, err := uuid.Parse(ID)
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid product id", err)
		}

		var storeSettings models.StoreSetting

		if err := db.Preload("User").Where("user_id = ?", uid).First(&storeSettings).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "Store settings not found", err)
		}

		response := convertToStoreResponse(storeSettings)

		return utils.ResponseSucess(c, http.StatusOK, "User store settings retrieved", response)
	}

}
