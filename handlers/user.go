package handlers

import (
	"api/db"
	"api/emails"
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
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrieve users", err)
		}

		// convert to public format
		publicUsers := make([]models.PublicUser, len(users))

		for i, u := range users {
			publicUsers[i] = models.PublicUser{
				ID:            u.ID,
				UserName:      u.UserName,
				Email:         u.Email,
				Role:          string(u.Role),
				ImageUrl:      u.ImageUrl,
				Location:      u.Location,
				Status:        u.Status,
				IsVerified:    u.IsVerified,
				ReferralCode:  u.ReferralCode,
				ReferralBy:    u.ReferralBy,
				ReferralCount: u.ReferralCount,
				CreatedAt:     u.CreatedAt,
				UpdatedAt:     u.UpdatedAt,
			}

		}
		return utils.ResponseSucess(c, http.StatusOK, "Users retrieved successfully", publicUsers)
	}

}

func GetUserById(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		uid, err := uuid.Parse(id)
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid user id", err)
		}

		var user models.User
		if err := db.First(&user, "id = ?", uid).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "User not found", err)
		}

		response := models.PublicUser{
			ID:            user.ID,
			UserName:      user.UserName,
			Email:         user.Email,
			Role:          string(user.Role),
			ImageUrl:      user.ImageUrl,
			PhoneNumber:   user.PhoneNumber,
			Location:      user.Location,
			Status:        user.Status,
			IsVerified:    user.IsVerified,
			ReferralCode:  user.ReferralCode,
			ReferralBy:    user.ReferralBy,
			ReferralCount: user.ReferralCount,
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
		}

		return utils.ResponseSucess(c, http.StatusOK, "User retrieved successfully", response)
	}

}

func UpdateUser(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Get("user_id").(uuid.UUID)
		var user models.User

		// find existing user
		if err := db.First(&user, id).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "User not found", err)
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
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to open file", err)
			}
			defer src.Close()

			tempFilePath := filepath.Join(os.TempDir(), filepath.Base(file.Filename))
			out, err := os.Create(tempFilePath)
			if err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to create temp file", err)
			}
			defer out.Close()

			if _, err := io.Copy(out, src); err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to write temp file", err)
			}

			fmt.Println("Uploading to Cloudinary:", tempFilePath)

			url, err := utils.UploadToCloudinary(tempFilePath, fmt.Sprintf("users/%s/image", id.String()))
			if err != nil {
				fmt.Println("Cloudinary upload failed:", err)
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to upload image", err)
			}

			user.ImageUrl = url
			os.Remove(tempFilePath)
		}

		// save changes

		if err := db.Save(&user).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to update user", err)
		}

		response := models.PublicUser{
			ID:            user.ID,
			UserName:      user.UserName,
			Email:         user.Email,
			PhoneNumber:   user.PhoneNumber,
			Role:          string(user.Role),
			ImageUrl:      user.ImageUrl,
			Location:      user.Location,
			ReferralCount: user.ReferralCount,
			IsVerified:    user.IsVerified,
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
		}

		return utils.ResponseSucess(c, http.StatusOK, "User updated successfully", echo.Map{"user": response})
	}
}

func Me(c echo.Context) error {
	// Get user ID from context (set by middleware)
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return utils.ResponseError(c, http.StatusInternalServerError, "Invalid user context", nil)
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return utils.ResponseError(c, http.StatusNotFound, "User not found", err)
	}

	// Convert to PublicUser for response
	publicUser := models.PublicUser{
		ID:            user.ID,
		UserName:      user.UserName,
		Email:         user.Email,
		Role:          string(user.Role),
		PhoneNumber:   user.PhoneNumber,
		ImageUrl:      user.ImageUrl,
		Location:      user.Location,
		IsVerified:    user.IsVerified,
		ReferralCode:  user.ReferralCode,
		ReferralBy:    user.ReferralBy,
		ReferralCount: user.ReferralCount,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		DeletedAt:     user.DeletedAt,
	}

	return utils.ResponseSucess(c, http.StatusOK, "User Details Retrieved", echo.Map{"user": publicUser})
}

func DeleteUser(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		if id == "" {
			return utils.ResponseError(c, 400, "ID can not be null", nil)

		}

		var user models.User
		result := db.Delete(&user, "id = ?", id)
		if result.Error != nil {
			return utils.ResponseError(c, 500, "Failed to delete User", result.Error)
		}

		if result.RowsAffected == 0 {
			return utils.ResponseError(c, 404, "User not found", nil)
		}

		return utils.ResponseSucess(c, 200, "User deleted successfully", nil)
	}
}

func VerifyUser(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		uid, err := uuid.Parse(id)
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid user id", err)
		}

		var user models.User

		if err := db.First(&user, "id = ?", uid).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "User not found", err)
		}

		user.IsVerified = true

		if err := db.Save(&user).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to verify user", err)
		}
		emails.SendAccountVerifiedMail(user.Email, user.UserName)

		return utils.ResponseSucess(c, http.StatusOK, "User verified successfully", nil)
	}
}
