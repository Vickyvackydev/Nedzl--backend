package handlers

import (
	"api/models"
	"api/utils"
	"encoding/json"

	"io"

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

		userIDStr, _ := c.Get("userID").(string)

		productID := c.FormValue("product_id")
		experience := c.FormValue("experience")
		reviewTitle := c.FormValue("review_title")
		customerName := c.FormValue("customer_name")
		review := c.FormValue("review")
		isPublicStr := c.FormValue("is_public")

		if productID == "" || experience == "" || reviewTitle == "" ||
			customerName == "" || review == "" || isPublicStr == "" {

			return utils.ResponseError(c, 400, "All fields are required", nil)
		}

		isPublic := strings.ToLower(isPublicStr) == "true"

		//  Ensure product exists
		var product models.Products
		if err := db.Where("id = ?", productID).First(&product).Error; err != nil {
			return utils.ResponseError(c, 404, "Product not found", err)
		}

		//  Prevent duplicate reviews
		if userIDStr != "" {
			var existing models.CustomerReview
			db.Where("user_id = ? AND product_id = ? AND is_public = ?", userIDStr, productID, isPublic).
				First(&existing)
			if existing.ID != uuid.Nil {
				return utils.ResponseError(c, 400, "You already submitted this type of review", nil)
			}
		}

		form, err := c.MultipartForm()
		if err != nil {
			return utils.ResponseError(c, 400, "Invalid form data", err)
		}

		var imageUrls []string
		files := form.File["images"]

		for _, file := range files {
			src, _ := file.Open()
			temp := "/tmp/" + file.Filename
			dst, _ := os.Create(temp)
			io.Copy(dst, src)
			src.Close()
			dst.Close()

			url, err := utils.UploadToCloudinary(temp, "reviews/"+file.Filename)
			if err != nil {
				return utils.ResponseError(c, 500, "Failed upload", err)
			}

			imageUrls = append(imageUrls, url)
			os.Remove(temp)
		}

		imgJson, _ := json.Marshal(imageUrls)

		reviewRecord := models.CustomerReview{
			Experience:   experience,
			ReviewTitle:  reviewTitle,
			CustomerName: customerName,
			Review:       review,
			Images:       datatypes.JSON(imgJson),
			IsPublic:     isPublic,
			ProductID:    uuid.MustParse(productID),
		}

		if userIDStr != "" {
			uid := uuid.MustParse(userIDStr)
			reviewRecord.UserID = &uid
		}

		if err := db.Create(&reviewRecord).Error; err != nil {
			return utils.ResponseError(c, 500, "Failed to create review", err)
		}

		return utils.ResponseSucess(c, 201, "Review created", reviewRecord)
	}
}

func GetPublicReviews(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {

		productId := c.Param("product_id")
		var reviews []models.CustomerReview
		// Convert string → UUID
		productUUID, err := uuid.Parse(productId)
		if err != nil {
			return utils.ResponseError(c, 400, "Invalid product ID format", err)
		}

		if err := db.Where("product_id = ? AND is_public = true", productUUID).Order("created_at DESC").Find(&reviews).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch reviews", err)
		}

		var product models.Products
		if err := db.Where("id = ?", productId).Where(&product).Error; err != nil {
			return utils.ResponseError(c, 404, "Product not found", err)
		}

		productDetails := ConvertToProductResponse(product)
		var response []models.ReviewResponse

		for _, r := range reviews {

			response = append(response, models.ReviewResponse{
				CustomerReview: r,
				ProductDetails: &productDetails,
			})

		}

		return utils.ResponseSucess(c, http.StatusOK, "Public reviews fetched successfully", response)
	}
}
func GetPrivateReviews(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {

		userID := c.Get("userID").(string)

		var reviews []models.CustomerReview
		if err := db.Where("user_id = ? AND is_public = false", userID).
			Order("created_at DESC").
			Find(&reviews).Error; err != nil {
			return utils.ResponseError(c, 500, "Failed to fetch reviews", err)
		}

		return utils.ResponseSucess(c, 200, "Private reviews fetched", reviews)
	}
}
