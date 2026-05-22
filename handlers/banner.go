package handlers

import (
	"api/models"
	"api/utils"

	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func CreateBanner(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		targetUrl := c.FormValue("target_url")
		if targetUrl == "" {
			return utils.ResponseError(c, http.StatusBadRequest, "Target URL is required", nil)
		}

		file, err := c.FormFile("image")
		if err != nil || file == nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Banner image is required", err)
		}

		src, err := file.Open()
		if err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to open banner image", err)
		}
		defer src.Close()

		tempFilePath := filepath.Join(os.TempDir(), "banner_"+uuid.NewString()+filepath.Ext(file.Filename))
		out, err := os.Create(tempFilePath)
		if err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to create temp file", err)
		}
		defer out.Close()

		if _, err := io.Copy(out, src); err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to write temp file", err)
		}

		url, err := utils.UploadToCloudinary(tempFilePath, "banners")
		if err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to upload banner", err)
		}
		os.Remove(tempFilePath)

		banner := models.Banner{
			ImageUrl:  url,
			TargetUrl: targetUrl,
			IsActive:  true,
		}

		if err := db.Create(&banner).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to save banner to database", err)
		}

		return utils.ResponseSucess(c, http.StatusCreated, "Banner created successfully", banner)
	}
}

func GetBanners(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var banners []models.Banner
		if err := db.Where("is_active = ?", true).Order("created_at desc").Find(&banners).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrieve banners", err)
		}
		return utils.ResponseSucess(c, http.StatusOK, "Banners retrieved successfully", banners)
	}
}

func GetAdminBanners(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var banners []models.Banner
		if err := db.Order("created_at desc").Find(&banners).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrieve banners", err)
		}
		return utils.ResponseSucess(c, http.StatusOK, "Banners retrieved successfully", banners)
	}
}

func DeleteBanner(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return utils.ResponseError(c, http.StatusBadRequest, "Banner ID is required", nil)
		}

		var banner models.Banner
		if err := db.First(&banner, "id = ?", id).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "Banner not found", err)
		}

		if err := db.Delete(&banner).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to delete banner", err)
		}

		return utils.ResponseSucess(c, http.StatusOK, "Banner deleted successfully", nil)
	}
}

func ToggleBannerStatus(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return utils.ResponseError(c, http.StatusBadRequest, "Banner ID is required", nil)
		}

		var banner models.Banner
		if err := db.First(&banner, "id = ?", id).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "Banner not found", err)
		}

		banner.IsActive = !banner.IsActive
		if err := db.Save(&banner).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to update banner status", err)
		}

		return utils.ResponseSucess(c, http.StatusOK, "Banner status updated successfully", banner)
	}
}
