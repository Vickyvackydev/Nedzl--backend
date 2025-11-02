package handlers

import (
	"api/db"
	"api/models"
	"api/utils"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func GetUsers(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// rows, err := db.Query("SELECT id, name, email FROM users")

		var users []models.User

		if err := db.Find(&users).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to retrieve users"})
		}

		// convert to public format
		publicUsers := make([]models.PublicUser, len(users))

		for i, u := range users {
			publicUsers[i] = models.PublicUser{
				ID:       u.ID,
				UserName: u.UserName,
				Email:    u.Email,
				Role:     string(u.Role),
				ImageUrl: u.ImageUrl,
				Location: u.Location,
			}

		}
		return c.JSON(http.StatusOK, publicUsers)
	}

}

func GetUserById(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		uid, err := uuid.Parse(id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid user id"})
		}

		var user models.User
		if err := db.First(&user, "id = ?", uid).Error; err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
		}

		return c.JSON(http.StatusOK, echo.Map{"id": user.ID, "user_name": user.UserName, "email": user.Email})
	}

}

func UpdateUser(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Get("user_id").(uuid.UUID)
		var user models.User

		// find existing user
		if err := db.First(&user, id).Error; err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
		}

		name := c.FormValue("user_name")
		email := c.FormValue("email")
		phone := c.FormValue("phone_number")
		location := c.FormValue("location")

		// update only if new values provided
		if name != "" {
			user.UserName = name
		}
		if email != "" {
			user.Email = email
		}
		if phone != "" {
			user.PhoneNumber = phone
		}
		if location != "" {
			user.Location = location
		}

		// handle optional image upload

		file, err := c.FormFile("image_url")
		if err == nil && file != nil {
			src, err := file.Open()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to open file"})
			}
			defer src.Close()

			tempFilePath := filepath.Join(os.TempDir(), file.Filename)
			out, err := os.Create(tempFilePath)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create temp file"})
			}
			defer out.Close()

			if _, err := io.Copy(out, src); err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to write temp file"})
			}

			fmt.Println("Uploading to Cloudinary:", tempFilePath)

			url, err := utils.UploadToCloudinary(tempFilePath, fmt.Sprintf("users/%s/image", id.String()))
			if err != nil {
				fmt.Println("Cloudinary upload failed:", err)
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
			}

			user.ImageUrl = url
			os.Remove(tempFilePath)
		}

		// save changes

		if err := db.Save(&user).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update user"})
		}

		response := models.PublicUser{
			ID:          user.ID,
			UserName:    user.UserName,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			Role:        string(user.Role),
			ImageUrl:    user.ImageUrl,
			Location:    user.Location,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}

		return c.JSON(http.StatusOK, echo.Map{
			"message": "User updated successfully",
			"user":    response,
		})
	}
}

func Me(c echo.Context) error {
	// Get user ID from context (set by middleware)
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Invalid user context"})
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
	}

	// Convert to PublicUser for response
	publicUser := models.PublicUser{
		ID:          user.ID,
		UserName:    user.UserName,
		Email:       user.Email,
		Role:        string(user.Role),
		PhoneNumber: user.PhoneNumber,
		ImageUrl:    user.ImageUrl,
		Location:    user.Location,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		DeletedAt:   user.DeletedAt,
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "User Details Retrieved", "user": publicUser})
}
