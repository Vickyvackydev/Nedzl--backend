package handlers

import (
	"api/models"
	"api/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func CreateReview(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {

		userID, _ := c.Get("userID").(string)
		productID := c.FormValue("product_id")
		experience := c.FormValue("experience")
		reviewTitle := c.FormValue("review_title")
		customerName := c.FormValue("customer_name")
		review := c.FormValue("review")
		isPublicStr := c.FormValue("is_public")

		if productID == "" || experience == "" || reviewTitle == "" || customerName == "" || review == "" || isPublicStr == "" {
			return c.JSON(400, map[string]string{"error": "All fields are required"})
		}

		isPublic := strings.ToLower(isPublicStr) == "true"

		form, err := c.MultipartForm()
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid form data", err)
		}

		files := form.File["images"]
		var imageUrls []string

		for _, file := range files {
			src, err := file.Open()

			if err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to open image", err)
			}

			tempFilePath := fmt.Sprintf("/tmp/%s", file.Filename)
			out, err := os.Create(tempFilePath)
			if err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to create temp file", err)
			}

			if _, err := io.Copy(out, src); err != nil {
				src.Close()
				out.Close()
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to write temp file", err)
			}
			src.Close()
			out.Close()

			url, err := utils.UploadToCloudinary(tempFilePath, fmt.Sprintf("reviews/%s/%s", userID, file.Filename))
			if err != nil {
				log.Printf("Cloudinary upload failed: %v", err)
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to upload image", err)
			}

			imageUrls = append(imageUrls, url)

			os.Remove(tempFilePath)

		}

		imageUrlJSON, err := json.Marshal(imageUrls)
		if err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to process image URLs", err)
		}

		var product models.Products
		if err := db.First(&product).Where("id = ?", productID).Error; err != nil {
			return utils.ResponseError(c, 404, "Product not found", err)
		}

		if isPublicStr != "true" {
			var existingCustomerReview models.CustomerReview

			err := db.Where("user_id = ? AND product_id AND is_public = false", userID, productID).First(&existingCustomerReview).Error

			if err == nil {
				return utils.ResponseError(c, 400, "You have already submitted a private review for this product", nil)
			}
		}
		if isPublicStr == "true" {
			var existingCustomerReview models.CustomerReview

			err := db.Where("user_id = ? AND product_id AND is_public = true", userID, productID).First(&existingCustomerReview).Error

			if err == nil {
				return utils.ResponseError(c, 400, "You have already submitted a private review for this product", nil)
			}
		}
		reviewRecord := models.CustomerReview{
			Experience:   experience,
			ReviewTitle:  reviewTitle,
			CustomerName: customerName,
			Review:       review,
			Images:       datatypes.JSON(imageUrlJSON),
			IsPublic:     isPublic,
		}

		if userID != "" {
			reviewRecord.UserID = uuid.MustParse(userID)

		}

		if err := db.Create(&reviewRecord).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to create review", err)
		}
		return utils.ResponseSucess(c, http.StatusCreated, "Review created successfully", reviewRecord)

	}
}

func GetPublicReviews(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {

		productId := c.Param("product_id")
		var reviews []models.CustomerReview

		if err := db.Where("product_id = ? AND is_public = false", productId).Order("created_at DESC").Find(&reviews).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch reviews", err)
		}

		var product models.Products
		if err := db.First(&product).Where("id = ?", productId).Error; err != nil {
			return utils.ResponseError(c, 404, "Product not found", err)
		}

		productDetails := ConvertToProductResponse(product)

		for review, _ := range reviews {

			reviews[review].ProductDetails = &productDetails

		}

		return utils.ResponseSucess(c, http.StatusOK, "Public reviews fetched successfully", reviews)
	}
}
func GetPrivateReviews(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {

		userID := c.Get("user_id").(string)

		var reviews []models.CustomerReview

		if err := db.Where("user_id = ? AND is_public = false", userID).Order("created_at DESC").Find(&reviews).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch reviews", err)
		}

		return utils.ResponseSucess(c, http.StatusOK, "Public reviews fetched successfully", reviews)
	}
}
