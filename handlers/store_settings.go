package handlers

import (
	"api/models"
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
	CreatedAt         time.Time         `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time         `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt         gorm.DeletedAt    `json:"deleted_at" gorm:"index"`
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
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
		}

		// user must be authenticated; set foreign key
		uid, ok := c.Get("user_id").(uuid.UUID)
		if !ok {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
		}

		if body.BusinessName == "" || body.AboutCompany == "" || body.StoreName == "" || body.Address == "" || body.State == "" || body.HowDoWeLocateYou == "" || body.BusinessHoursFrom == "" || body.BusinessHoursTo == "" {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "All fields are required"})
		}

		body.UserID = uid

		if err := db.Create(&body).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create store settings"})
		}

		// Load related user for response
		if err := db.Preload("User").First(&body, "id = ?", body.ID).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to load created settings"})
		}
		response := convertToStoreResponse(body)

		return c.JSON(http.StatusCreated, echo.Map{"message": "Store settings created successfully", "settings": response})

	}

}

func UpdateStoreSettings(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uuid.UUID)
		ID := c.Param("id")
		uid, err := uuid.Parse(ID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid id"})
		}

		var storeSettings models.StoreSetting
		if err := c.Bind(&storeSettings); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
		}

		var existing models.StoreSetting
		if err := db.Where("user_id = ? AND id = ?", userID, uid).First(&existing).Error; err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Store settings not found"})
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
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update store settings"})
		}

		if err := db.Preload("User").First(&existing, "id = ?", existing.ID).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to load updated settings"})
		}

		return c.JSON(http.StatusOK, echo.Map{"message": "Store settings updated successfully", "data": existing})

	}

}

func GetStoreSettings(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		ID := c.Param("id")
		uid, err := uuid.Parse(ID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid product id"})
		}

		var storeSettings models.StoreSetting

		if err := db.Preload("User").Where("user_id =?", uid).First(&storeSettings).Error; err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Store settings not found"})
		}

		return c.JSON(http.StatusOK, echo.Map{"message": "User store settings retrieved", "store_settings": storeSettings})
	}

}
