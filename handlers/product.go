package handlers

import (
	"api/models"
	"api/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ProductResponse represents a safe product response without sensitive user data
type ProductResponse struct {
	ID                uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	Name              string            `json:"product_name"`
	ProductPrice      float64           `json:"product_price"`
	MarketPriceFrom   float64           `json:"market_price_from"`
	MarketPriceTo     float64           `json:"market_price_to"`
	CategoryName      string            `json:"category_name"`
	IsNegotiable      bool              `json:"is_negotiable"`
	Description       string            `json:"description"`
	State             string            `json:"state"`
	AddressInState    string            `json:"address_in_state"`
	OutStandingIssues string            `json:"outstanding_issues"`
	ImageUrls         datatypes.JSON    `json:"image_urls"`
	Status            models.Status     `json:"status" gorm:"type:varchar(20);default:'UNDER_REVIEW'"`
	Condition         string            `json:"condition"`
	UserID            uuid.UUID         `json:"user_id"`
	BrandName         string            `json:"brand_name"`
	User              models.PublicUser `json:"user"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	DeletedAt         gorm.DeletedAt    `json:"deleted_at"`
}

// convertToProductResponse converts a Products model to a safe ProductResponse
func convertToProductResponse(product models.Products) ProductResponse {
	publicUser := models.PublicUser{
		ID:          product.User.ID,
		UserName:    product.User.UserName,
		Email:       product.User.Email,
		Role:        string(product.User.Role),
		PhoneNumber: product.User.PhoneNumber,
		ImageUrl:    product.User.ImageUrl,
		Location:    product.User.Location,
		CreatedAt:   product.User.CreatedAt,
		UpdatedAt:   product.User.UpdatedAt,
		DeletedAt:   product.User.DeletedAt,
	}

	return ProductResponse{
		ID:                product.ID,
		Name:              product.Name,
		ProductPrice:      product.ProductPrice,
		MarketPriceFrom:   product.MarketPriceFrom,
		MarketPriceTo:     product.MarketPriceTo,
		CategoryName:      product.CategoryName,
		IsNegotiable:      product.IsNegotiable,
		Description:       product.Description,
		State:             product.State,
		AddressInState:    product.AddressInState,
		OutStandingIssues: product.OutStandingIssues,
		ImageUrls:         product.ImageUrls,
		Status:            product.Status,
		Condition:         product.Condition,
		BrandName:         product.BrandName,
		UserID:            product.UserID,
		User:              publicUser,
		CreatedAt:         product.CreatedAt,
		UpdatedAt:         product.UpdatedAt,
		DeletedAt:         product.DeletedAt,
	}
}

func CreateProduct(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {

		userId := c.Get("user_id").(uuid.UUID)
		name := c.FormValue("product_name")
		productPriceStr := c.FormValue("product_price")
		marketPriceFromStr := c.FormValue("market_price_from")
		marketPriceToStr := c.FormValue("market_price_to")
		categoryName := c.FormValue("category_name")
		isNegotiableStr := c.FormValue("is_negotiable")
		description := c.FormValue("description")
		state := c.FormValue("state")
		addressInState := c.FormValue("address_in_state")
		outstandingIssues := c.FormValue("outstanding_issues")
		condition := c.FormValue("condition")
		brandName := c.FormValue("brand_name")
		status := c.FormValue("status")

		// if name == "" || productPriceStr == "" || marketPriceFromStr == "" || marketPriceToStr == "" || categoryName == "" || isNegotiableStr == "" || description == "" || state == "" || addressInState == "" || outstandingIssues == "" || condition == "" || brandName == "" {
		// 	return c.JSON(http.StatusBadRequest, echo.Map{"error": "All fields are required"})
		// }
		// Convert string values to float64
		productPrice, err := strconv.ParseFloat(productPriceStr, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid product price"})
		}
		marketPriceFrom, err := strconv.ParseFloat(marketPriceFromStr, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid market price (from)"})
		}
		marketPriceTo, err := strconv.ParseFloat(marketPriceToStr, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid market price (to)"})
		}

		// Convert string ("true"/"false") to bool

		isNegotiable := strings.ToLower(isNegotiableStr) == "true"

		form, err := c.MultipartForm()
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid form data"})
		}

		files := form.File["image_urls"]
		if len(files) == 0 {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "You must upload at least one image"})
		}

		var imageUrls []string

		for _, file := range files {
			src, err := file.Open()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to open image"})
			}

			tempFilePath := fmt.Sprintf("/tmp/%s", file.Filename)
			out, err := os.Create(tempFilePath)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create temp file"})
			}

			if _, err := io.Copy(out, src); err != nil {
				src.Close()
				out.Close()
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to copy image"})
			}
			src.Close()
			out.Close()

			url, err := utils.UploadToCloudinary(tempFilePath, fmt.Sprintf("users/%s/products", userId.String()))
			if err != nil {
				log.Printf("Cloudinary upload failed: %v", err)
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
			}

			imageUrls = append(imageUrls, url)
			os.Remove(tempFilePath)
		}

		// Marshal the image URLs array into JSON for your datatypes.JSON field
		imageUrlsJSON, err := json.Marshal(imageUrls)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to process image URLs"})
		}

		products := models.Products{
			Name:              name,
			ProductPrice:      productPrice,
			Description:       description,
			MarketPriceFrom:   marketPriceFrom,
			MarketPriceTo:     marketPriceTo,
			CategoryName:      categoryName,
			IsNegotiable:      isNegotiable,
			State:             state,
			AddressInState:    addressInState,
			OutStandingIssues: outstandingIssues,
			Condition:         condition,
			BrandName:         brandName,
			ImageUrls:         datatypes.JSON(imageUrlsJSON),
			Status:            models.Status(status),
			UserID:            userId,
		}

		if err := db.Create(&products).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to save product"})
		}

		// Preload the user data for the response
		if err := db.Preload("User").First(&products, products.ID).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to load product with user data"})
		}

		// Convert to safe response without password
		response := convertToProductResponse(products)

		return c.JSON(http.StatusCreated, echo.Map{"message": "Product created successfully", "products": response})
	}
}
func UpdateUserProduct(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userId := c.Get("user_id").(uuid.UUID)
		id := c.Param("id")

		// find matching product
		var existingProduct models.Products
		if err := db.Where("id = ? AND user_id = ?", id, userId).First(&existingProduct).Error; err != nil {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "Unauthorized or not found"})

		}

		productName := c.FormValue("product_name")
		productPrice := c.FormValue("product_price")
		marketPriceFrom := c.FormValue("market_price_from")
		marketPriceTo := c.FormValue("market_price_to")
		categoryName := c.FormValue("category_name")
		state := c.FormValue("state")
		addressInState := c.FormValue("address_in_state")
		outstandingIssues := c.FormValue("outstanding_issues")
		description := c.FormValue("description")
		isNegotiable := c.FormValue("is_negotiable")
		condition := c.FormValue("condition")
		brandName := c.FormValue("brand_name")
		status := c.FormValue("status")

		if productName == "" || productPrice == "" || marketPriceFrom == "" || marketPriceTo == "" ||
			categoryName == "" || isNegotiable == "" || description == "" || state == "" || brandName == "" ||
			addressInState == "" || condition == "" {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "All fields are required"})
		}

		productP, err := strconv.ParseFloat(productPrice, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid product price"})
		}
		marketPFrom, err := strconv.ParseFloat(marketPriceFrom, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid market price (from)"})
		}
		marketPTo, err := strconv.ParseFloat(marketPriceTo, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid market price (to)"})
		}

		// Convert string ("true"/"false") to bool

		isNeg := strings.ToLower(isNegotiable) == "true"

		form, err := c.MultipartForm()
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid form input"})
		}

		files := form.File["image_urls"]
		if len(files) == 0 {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "You must upload at least one image"})
		}

		var imageUrls []string

		for _, file := range files {
			src, err := file.Open()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to open image"})
			}

			tempFilePath := fmt.Sprintf("/tmp/%s", file.Filename)
			out, err := os.Create(tempFilePath)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create temp file"})
			}

			if _, err := io.Copy(out, src); err != nil {
				src.Close()
				out.Close()
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to copy image"})
			}
			src.Close()
			out.Close()

			url, err := utils.UploadToCloudinary(tempFilePath, fmt.Sprintf("users/%s/products", userId.String()))
			if err != nil {
				log.Printf("Cloudinary upload failed: %v", err)
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
			}

			imageUrls = append(imageUrls, url)
			os.Remove(tempFilePath)
		}

		updatedImageUrlJson, err := json.Marshal(imageUrls)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to process image URLs"})
		}

		existingProduct.Name = productName
		existingProduct.ProductPrice = productP
		existingProduct.MarketPriceFrom = marketPFrom
		existingProduct.MarketPriceTo = marketPTo
		existingProduct.CategoryName = categoryName
		existingProduct.Condition = condition
		existingProduct.BrandName = brandName
		existingProduct.Description = description
		existingProduct.IsNegotiable = isNeg
		existingProduct.OutStandingIssues = outstandingIssues
		existingProduct.AddressInState = addressInState
		existingProduct.ImageUrls = updatedImageUrlJson

		if status != "" {
			existingProduct.Status = models.Status(status)
		}
		// Update only fields that were provided (prevent zero overwrite)
		if err := db.Save(&existingProduct).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update product"})
		}

		// Reload product with user info for response
		if err := db.Preload("User").First(&existingProduct, existingProduct.ID).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to load updated product"})
		}

		response := convertToProductResponse(existingProduct)

		return c.JSON(http.StatusOK, echo.Map{"message": "Product updated successfully", "data": response})
	}
}
func GetAllProducts(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var products []models.Products
		// if err := db.Preload("User").Find(&products).Error; err != nil {
		// 	return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to retrieve products"})
		// }
		baseUrl := c.Scheme() + "://" + c.Request().Host + c.Path()
		query := db.Preload("User")

		search := c.QueryParam("search")
		category := c.QueryParam("category_name")
		status := c.QueryParam("status")
		startDate := c.QueryParam("start_date")
		state := c.QueryParam("state")
		endDate := c.QueryParam("end_date")
		minPrice := c.QueryParam("min_price")
		maxPrice := c.QueryParam("max_price")
		keyword := c.QueryParam("keyword")

		// -- PAGINATION -- //

		page, _ := strconv.Atoi(c.QueryParam("page"))
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		if page < 1 {
			page = 1
		}

		if limit < 1 {
			limit = 10
		}

		offSet := (page - 1) * limit

		// -- APPLY FILTERS --

		if search != "" {
			query = query.Where("name ILIKE ?", "%"+search+"%")
		}

		if category != "" {
			query = query.Where("category_name = ?", category)
		}

		if status != "" {
			query = query.Where("status = ?", status)
		}

		if state != "" {
			query = query.Where("state = ?", state)
		}

		if startDate != "" && endDate != "" {
			query = query.Where("created_at BETWEEN  ? AND ?", startDate, endDate)
		}
		if minPrice != "" && maxPrice != "" {
			query = query.Where("product_price BETWEEN  ? AND ?", minPrice, maxPrice)
		}
		if keyword != "" {
			query = query.Where("product_name ILIKE ? OR description ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}

		query = query.Order("created_at DESC")
		// -- GET RESULTS --

		var total int64

		query.Model(&models.Products{}).Count(&total)
		if err := query.Limit(limit).Offset(offSet).Find(&products).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to retrieve products"})
		}

		// -- CONVERT TO SAFE RESPONSE --

		// Convert to safe responses without passwords
		var responses []ProductResponse
		for _, product := range products {
			responses = append(responses, convertToProductResponse(product))
		}

		totalPages := int(math.Ceil(float64(total) / float64(limit)))
		var nextPageURL *string
		if page < totalPages {
			url := fmt.Sprintf("%s?page=%d&limit=%d", baseUrl, page+1, limit)
			nextPageURL = &url
		}

		var prevPageURL *string
		if page > 1 {
			url := fmt.Sprintf("%s?page=%d&limit=%d", baseUrl, page-1, limit)
			prevPageURL = &url
		}

		// --- Generate all pages ---

		pages := []map[string]interface{}{}
		for i := 1; i <= totalPages; i++ {

			pageURL := fmt.Sprintf("%s?page=%d&limit=%d", baseUrl, i, limit)
			pages = append(pages, map[string]interface{}{
				"page": i,
				"url":  pageURL,
			})

		}

		return c.JSON(http.StatusOK, echo.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"totalpages":  totalPages,
			"data":        responses,
			"nextPageUrl": nextPageURL,
			"prevPageUrl": prevPageURL,
			"pages":       pages,
		})
	}

}

func GetUserProducts(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userId := c.Get("user_id").(uuid.UUID)
		baseUrl := c.Scheme() + "://" + c.Request().Host + c.Path()
		var products []models.Products
		query := db.Preload("User").Where("user_id = ?", userId)

		// --- FILTER PARAMETERS ---
		search := c.QueryParam("search")
		category := c.QueryParam("category")
		status := c.QueryParam("status")
		startDate := c.QueryParam("start_date")
		state := c.QueryParam("state")
		endDate := c.QueryParam("end_date")
		minPrice := c.QueryParam("min_price")
		maxPrice := c.QueryParam("max_price")
		keyword := c.QueryParam("keyword")

		page, _ := strconv.Atoi(c.QueryParam("page"))
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 10
		}
		offset := (page - 1) * limit

		// --- APPLY FILTERS ---
		if search != "" {
			query = query.Where("name ILIKE ?", "%"+search+"%")
		}
		if category != "" {
			query = query.Where("category_name = ?", category)
		}
		if status != "" {
			query = query.Where("status = ?", status)
		}
		if state != "" {
			query = query.Where("state = ?", state)
		}
		if startDate != "" && endDate != "" {
			query = query.Where("created_at BETWEEN ? AND ?", startDate, endDate)
		}
		if minPrice != "" && maxPrice != "" {
			query = query.Where("product_price BETWEEN ? AND ?", minPrice, maxPrice)
		}
		if keyword != "" {
			query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}

		var total int64
		query.Model(&models.Products{}).Count(&total)
		if err := query.Limit(limit).Offset(offset).Find(&products).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch user products"})
		}

		var responses []ProductResponse
		for _, product := range products {
			responses = append(responses, convertToProductResponse(product))
		}

		totalPages := int(math.Ceil(float64(total) / float64(limit)))
		var nextPageURL *string
		if page < totalPages {
			url := fmt.Sprintf("%s?page=%d&limit=%d", baseUrl, page+1, limit)
			nextPageURL = &url
		}

		var prevPageURL *string
		if page > 1 {
			url := fmt.Sprintf("%s?page=%d&limit=%d", baseUrl, page-1, limit)
			prevPageURL = &url
		}

		// --- Generate all pages ---

		pages := []map[string]interface{}{}
		for i := 1; i <= totalPages; i++ {

			pageURL := fmt.Sprintf("%s?page=%d&limit=%d", baseUrl, i, limit)
			pages = append(pages, map[string]interface{}{
				"page": i,
				"url":  pageURL,
			})

		}

		return c.JSON(http.StatusOK, echo.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"totalPages":  int(math.Ceil(float64(total) / float64(limit))),
			"data":        responses,
			"pages":       pages,
			"nextPageUrl": nextPageURL,
			"prevPageUrl": prevPageURL,
		})
	}
}

func GetSingleProduct(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		uid, err := uuid.Parse(id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid product id"})
		}
		var product models.Products
		if err := db.Preload("User").First(&product, "id = ?", uid).Error; err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Product not found"})
		}

		// Convert to safe response without password
		response := convertToProductResponse(product)

		return c.JSON(http.StatusOK, echo.Map{"message": "Product fetched", "product": response})
	}
}

func GetUserProduct(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uuid.UUID)
		ProductID := c.Param("id")

		var product models.Products

		if err := db.Preload("User").Where("id = ? AND user_id =?", ProductID, userID).First(&product).Error; err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Product not found or unauthorized"})
		}

		// Convert to safe response without password
		response := convertToProductResponse(product)

		return c.JSON(http.StatusOK, echo.Map{"message": "User product fetched successfully", "product": response})
	}

}

func DeleteUserProduct(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		UserId := c.Get("user_id").(uuid.UUID)
		id := c.Param("id")

		var product models.Products

		if err := db.Where("id =? AND user_id =?", id, UserId).Delete(&product).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to delete product"})
		}

		return c.JSON(http.StatusOK, echo.Map{"message": "Product deleted successfully"})
	}
}

func GetTotalProductsByCatgory(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		category := c.QueryParam("category_name")
		var products models.Products

		if category == "" {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Category is required"})
		}

		var total int64
		if err := db.Model(&products).Where("category_name = ?", category).Count(&total).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to retrived product counts"})
		}

		return c.JSON(http.StatusOK, echo.Map{"message": "Products counts retrived", "category": category, "count": total})
	}
}
