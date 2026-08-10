package handlers

import (
	"api/emails"
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

	"path/filepath"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ProductResponse represents a safe product response without sensitive user data

// convertToProductResponse converts a Products model to a safe ProductResponse
func ConvertToProductResponse(product models.Products, isLiked bool) models.ProductResponse {
	var publicUserPtr *models.PublicUser
	if product.User.ID != uuid.Nil {
		publicUserPtr = &models.PublicUser{
			ID:            product.User.ID,
			UserName:      product.User.UserName,
			Email:         product.User.Email,
			Role:          string(product.User.Role),
			PhoneNumber:   product.User.PhoneNumber,
			ImageUrl:      product.User.ImageUrl,
			Location:      product.User.Location,
			IsVerified:    product.User.IsVerified,
			ReferralCode:  product.User.ReferralCode,
			ReferralBy:    product.User.ReferralBy,
			ReferralCount: product.User.ReferralCount,
			CreatedAt:     product.User.CreatedAt,
			UpdatedAt:     product.User.UpdatedAt,
			DeletedAt:     product.User.DeletedAt,
		}
	} else if product.IsGuestListing || product.GuestPhone != "" || product.GuestEmail != "" {
		publicUserPtr = &models.PublicUser{
			UserName:    "Guest Seller",
			Email:       product.GuestEmail,
			PhoneNumber: product.GuestPhone,
			IsVerified:  false,
			CreatedAt:   product.CreatedAt,
		}
	}

	return models.ProductResponse{
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
		User:              publicUserPtr,
		University:        product.University,
		CreatedAt:         product.CreatedAt,
		UpdatedAt:         product.UpdatedAt,
		DeletedAt:         product.DeletedAt,
		Views:             product.Views,
		Likes:             product.Likes,
		IsLikedByMe:       isLiked,
		IsDeletedByUser:   product.IsDeletedByUser,
		ProductType:       product.ProductType,
		SubMenus:          product.SubMenus,
		DeliveryFee:       product.DeliveryFee,
		OldPrice:          product.OldPrice,
		DiscountPercent:   product.DiscountPercent,
		GuestEmail:        product.GuestEmail,
		GuestPhone:        product.GuestPhone,
		IsGuestListing:    product.IsGuestListing,
		ServiceType:       product.ServiceType,
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
		university := c.FormValue("university")
		productType := c.FormValue("product_type")
		subMenusRaw := c.FormValue("sub_menus")
		deliveryFeeStr := c.FormValue("delivery_fee")
		serviceType := c.FormValue("service_type")

		if productType == "" {
			productType = "MARKET"
		}

		var deliveryFee float64
		if deliveryFeeStr != "" {
			deliveryFee, _ = strconv.ParseFloat(deliveryFeeStr, 64)
		}

		// if name == "" || productPriceStr == "" || marketPriceFromStr == "" || marketPriceToStr == "" || categoryName == "" || isNegotiableStr == "" || description == "" || state == "" || addressInState == "" || outstandingIssues == "" || condition == "" || brandName == "" {
		// 	return c.JSON(http.StatusBadRequest, echo.Map{"error": "All fields are required"})
		// }
		// Convert string values to float64
		productPrice, err := strconv.ParseFloat(productPriceStr, 64)
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid product price", err)
		}
		marketPriceFrom, err := strconv.ParseFloat(marketPriceFromStr, 64)
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid market price (from)", err)
		}
		marketPriceTo, err := strconv.ParseFloat(marketPriceToStr, 64)
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid market price (to)", err)
		}

		// Convert string ("true"/"false") to bool

		isNegotiable := strings.ToLower(isNegotiableStr) == "true"

		form, err := c.MultipartForm()
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid form data", err)
		}

		files := form.File["new_images"]
		if len(files) == 0 {
			return utils.ResponseError(c, http.StatusBadRequest, "You must upload at least one image", nil)
		}

		var imageUrls []string

		for _, file := range files {
			src, err := file.Open()
			if err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to open image", err)
			}

			tempFilePath := filepath.Join(
				os.TempDir(),
				uuid.New().String()+"_"+filepath.Base(file.Filename),
			)
			out, err := os.Create(tempFilePath)
			if err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to create temp file", err)
			}

			if _, err := io.Copy(out, src); err != nil {
				src.Close()
				out.Close()
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to copy image", err)
			}
			src.Close()
			out.Close()

			url, err := utils.UploadToCloudinary(tempFilePath, fmt.Sprintf("users/%s/products", userId.String()))
			if err != nil {
				log.Printf("Cloudinary upload failed: %v", err)
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to upload image", err)
			}
			if url == "" {
				return utils.ResponseError(c, http.StatusInternalServerError, "Received empty URL from Cloudinary", nil)
			}

			imageUrls = append(imageUrls, url)
			os.Remove(tempFilePath)
		}

		// Marshal the image URLs array into JSON for your datatypes.JSON field
		imageUrlsJSON, err := json.Marshal(imageUrls)
		if err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to process image URLs", err)
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
			University:        university,
			BrandName:         brandName,
			ImageUrls:         datatypes.JSON(imageUrlsJSON),
			Status:            models.Status(status),
			UserID:            &userId,
			ProductType:       productType,
			SubMenus:          datatypes.JSON([]byte(subMenusRaw)),
			DeliveryFee:       deliveryFee,
			ServiceType:       serviceType,
		}

		if err := db.Create(&products).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to save product", err)
		}

		// Preload the user data for the response
		if err := db.Preload("User").First(&products, products.ID).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to load product with user data", err)
		}

		// Convert to safe response without password
		response := ConvertToProductResponse(products, false)

		// Trigger Facebook Auto-Post and Search Alerts in background
		go func(p models.Products) {
			// 1. Facebook/Instagram Auto Post
			message := fmt.Sprintf("🛍️ New Product Alert: %s\n\nPrice: ₦%.2f\nCondition: %s\n\nCheck it out on Nedzl!", p.Name, p.ProductPrice, p.Condition)
			link := fmt.Sprintf("https://nedzl.com/product-details/%s", p.ID.String())
			igCaption := fmt.Sprintf("%s\n\nLink in bio or copy: %s", message, link)
			if len(imageUrls) > 0 {
				imageUrl := imageUrls[0]
				if err := utils.PostToFacebook(message, imageUrl, link); err != nil {
					log.Printf("Facebook auto-post failed for product %s: %v", p.ID, err)
				}
				if err := utils.PostToInstagram(igCaption, imageUrl); err != nil {
					log.Printf("Instagram auto-post failed for product %s: %v", p.ID, err)
				}
			}

			// 2. Search Alerts Matching
			var alerts []models.SearchAlert
			err := db.Where("(category = '' OR category = ?) AND (keyword = '' OR ? ILIKE '%' || keyword || '%')",
				p.CategoryName, p.Name).Find(&alerts).Error
			if err != nil {
				log.Printf("Failed to fetch search alerts: %v", err)
				return
			}
			for _, alert := range alerts {
				err := emails.SendSearchAlertNotificationMail(alert.Email, alert.Keyword, alert.Category, p.Name, p.ID.String())
				if err != nil {
					log.Printf("Failed to send search alert email to %s: %v", alert.Email, err)
				} else {
					db.Delete(&alert)
				}
			}
		}(products)

		return utils.ResponseSucess(c, http.StatusCreated, "Product created successfully", echo.Map{"products": response})
	}
}
func UpdateUserProduct(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userId := c.Get("user_id").(uuid.UUID)
		id := c.Param("id")

		// find matching product
		var existingProduct models.Products
		if err := db.Where("id = ? AND user_id = ?", id, userId).First(&existingProduct).Error; err != nil {
			return utils.ResponseError(c, http.StatusForbidden, "Unauthorized or not found", err)

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
		university := c.FormValue("university")
		productType := c.FormValue("product_type")
		subMenusRaw := c.FormValue("sub_menus")
		deliveryFeeStr := c.FormValue("delivery_fee")

		if productName == "" || productPrice == "" || categoryName == "" || description == "" || state == "" || addressInState == "" {
			return utils.ResponseError(c, http.StatusBadRequest, "Required fields are missing", nil)
		}

		productP, err := strconv.ParseFloat(productPrice, 64)
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid product price", err)
		}
		marketPFrom, _ := strconv.ParseFloat(marketPriceFrom, 64)
		if marketPFrom == 0 {
			marketPFrom = productP
		}
		marketPTo, _ := strconv.ParseFloat(marketPriceTo, 64)
		if marketPTo == 0 {
			marketPTo = productP
		}

		// Convert string ("true"/"false") to bool

		isNeg := strings.ToLower(isNegotiable) == "true"

		// Parse image URLs from form (frontend might send as JSON array string)
		imageUrlsStr := c.FormValue("image_urls") // e.g. ["https://res.cloudinary.com/...","https://..."]
		var imageUrls []string
		if imageUrlsStr != "" {
			_ = json.Unmarshal([]byte(imageUrlsStr), &imageUrls)
		}

		form, err := c.MultipartForm()
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid form input", err)
		}

		files := form.File["new_images"]
		if len(files) == 0 && len(imageUrlsStr) == 0 {
			return utils.ResponseError(c, http.StatusBadRequest, "You must upload at least one image", nil)
		}

		// var imageUrls []string

		for _, file := range files {
			src, err := file.Open()
			if err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to open image", err)
			}

			tempFilePath := filepath.Join(os.TempDir(), filepath.Base(file.Filename))
			out, err := os.Create(tempFilePath)
			if err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to create temp file", err)
			}

			if _, err := io.Copy(out, src); err != nil {
				src.Close()
				out.Close()
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to copy image", err)
			}
			src.Close()
			out.Close()

			url, err := utils.UploadToCloudinary(tempFilePath, fmt.Sprintf("users/%s/products", userId.String()))
			if err != nil {
				log.Printf("Cloudinary upload failed: %v", err)
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to upload image", err)
			}

			imageUrls = append(imageUrls, url)
			os.Remove(tempFilePath)
		}

		updatedImageUrlJson, err := json.Marshal(imageUrls)
		if err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to process image URLs", err)
		}

		// Detect Price Reduction / Price Slash
		if productP < existingProduct.ProductPrice && existingProduct.ProductPrice > 0 {
			oldP := existingProduct.ProductPrice
			disc := int(math.Round(((oldP - productP) / oldP) * 100))
			existingProduct.OldPrice = oldP
			existingProduct.DiscountPercent = disc

			// Notify users who searched for or viewed this product category in background
			go func(p models.Products, oldPrice float64, newPrice float64, discountPct int) {
				var alerts []models.SearchAlert
				_ = db.Where("(category = '' OR category = ?) AND (keyword = '' OR ? ILIKE '%' || keyword || '%')", p.CategoryName, p.Name).Find(&alerts).Error
				for _, alert := range alerts {
					_ = emails.SendPriceSlashNotificationMail(alert.Email, p.Name, p.ID.String(), oldPrice, newPrice, discountPct)
				}
			}(existingProduct, oldP, productP, disc)
		}

		existingProduct.Name = productName
		existingProduct.ProductPrice = productP
		existingProduct.MarketPriceFrom = marketPFrom
		existingProduct.MarketPriceTo = marketPTo
		existingProduct.CategoryName = categoryName
		if condition != "" {
			existingProduct.Condition = condition
		}
		if brandName != "" {
			existingProduct.BrandName = brandName
		}
		existingProduct.Description = description
		existingProduct.IsNegotiable = isNeg
		existingProduct.OutStandingIssues = outstandingIssues
		existingProduct.AddressInState = addressInState
		existingProduct.ImageUrls = updatedImageUrlJson
		if university != "" {
			existingProduct.University = university
		}

		if productType != "" {
			existingProduct.ProductType = productType
		}
		if subMenusRaw != "" {
			existingProduct.SubMenus = datatypes.JSON([]byte(subMenusRaw))
		}
		if deliveryFeeStr != "" {
			deliveryFee, _ := strconv.ParseFloat(deliveryFeeStr, 64)
			existingProduct.DeliveryFee = deliveryFee
		}

		if status != "" {
			existingProduct.Status = models.Status(status)
		}
		// Update only fields that were provided (prevent zero overwrite)
		if err := db.Save(&existingProduct).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to update product", err)
		}

		// Reload product with user info for response
		if err := db.Preload("User").First(&existingProduct, existingProduct.ID).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to load updated product", err)
		}

		response := ConvertToProductResponse(existingProduct, false)

		return utils.ResponseSucess(c, http.StatusOK, "Product updated successfully", echo.Map{"data": response})
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
		university := c.QueryParam("university")

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
		productType := c.QueryParam("product_type")
		if productType == "MARKET" {
			query = query.Where("product_type IS NULL OR product_type = 'MARKET' OR product_type = ''")
		} else if productType == "FOOD" {
			query = query.Where("product_type = ? OR category_name IN ('prepared-food', 'foodstuffs', 'fruits-vegetables')", "FOOD")
		} else if productType == "SERVICE" {
			query = query.Where("product_type = ? OR category_name = 'other-services'", "SERVICE")
		} else if productType != "" {
			query = query.Where("product_type = ?", productType)
		} else if category == "prepared-food" || category == "foodstuffs" || category == "fruits-vegetables" {
			query = query.Where("product_type = ?", "FOOD")
		} else if category == "other-services" {
			query = query.Where("product_type = ?", "SERVICE")
		} else {
			query = query.Where("product_type IS NULL OR product_type = 'MARKET' OR product_type = ''")
		}

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
		if university != "" {
			query = query.Where("university = ?", university)
		}

		if startDate != "" && endDate != "" {
			query = query.Where("created_at BETWEEN  ? AND ?", startDate, endDate)
		}
		if minPrice != "" && maxPrice != "" {
			query = query.Where("product_price BETWEEN  ? AND ?", minPrice, maxPrice)
		}
		serviceType := c.QueryParam("service_type")
		if serviceType != "" {
			query = query.Where("service_type ILIKE ?", "%"+serviceType+"%")
		}

		if keyword != "" {
			query = query.Where("product_name ILIKE ? OR description ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}

		// -- SECTION FILTER --
		section := c.QueryParam("section")
		if section != "" {
			now := time.Now().UTC()
			startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

			switch section {
			case "todays_deal", "todays-deal":
				query = query.Where("created_at >= ?", startOfDay)
			case "for_you", "for-you":
				query = query.Where("created_at < ?", startOfDay)
			case "discount_sales", "discount-sales", "discount":
				query = query.Where("discount_percent > 0 OR (old_price IS NOT NULL AND old_price > product_price)")
			}
		}

		query = query.Order("created_at DESC")
		// -- GET RESULTS --

		var total int64

		query.Model(&models.Products{}).Count(&total)
		if err := query.Limit(limit).Offset(offSet).Where("status = ? AND is_deleted_by_user = ?", models.StatusOngoing, false).Find(&products).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrieve products", err)
		}

		// -- CONVERT TO SAFE RESPONSE --
		var likedMap = make(map[uuid.UUID]bool)
		if userIdVal := c.Get("user_id"); userIdVal != nil {
			if uid, ok := userIdVal.(uuid.UUID); ok {
				var likedIDs []uuid.UUID
				db.Model(&models.ProductLike{}).Where("user_id = ?", uid).Pluck("product_id", &likedIDs)
				for _, id := range likedIDs {
					likedMap[id] = true
				}
			}
		}

		// Convert to safe responses without passwords
		var responses []models.ProductResponse
		for _, product := range products {
			isLiked := likedMap[product.ID]
			responses = append(responses, ConvertToProductResponse(product, isLiked))
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

		return utils.ResponseSucess(c, http.StatusOK, "Products retrieved successfully", echo.Map{
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
		productType := c.QueryParam("product_type")
		if productType == "MARKET" {
			query = query.Where("product_type IS NULL OR product_type = 'MARKET' OR product_type = ''")
		} else if productType == "FOOD" {
			query = query.Where("product_type = ? OR category_name IN ('prepared-food', 'foodstuffs', 'fruits-vegetables')", "FOOD")
		} else if productType == "SERVICE" {
			query = query.Where("product_type = ? OR category_name = 'other-services'", "SERVICE")
		} else if productType != "" && productType != "ALL" {
			query = query.Where("product_type = ?", productType)
		}

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

		query = query.Order("created_at DESC")
		var total int64
		query.Model(&models.Products{}).Count(&total)
		if err := query.Limit(limit).Offset(offset).Where("is_deleted_by_user = ?", false).Find(&products).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch user products", err)
		}

		var likedMap = make(map[uuid.UUID]bool)
		var likedIDs []uuid.UUID
		db.Model(&models.ProductLike{}).Where("user_id = ?", userId).Pluck("product_id", &likedIDs)
		for _, id := range likedIDs {
			likedMap[id] = true
		}

		var responses []models.ProductResponse
		for _, product := range products {
			isLiked := likedMap[product.ID]
			responses = append(responses, ConvertToProductResponse(product, isLiked))
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

		return utils.ResponseSucess(c, http.StatusOK, "User products retrieved successfully", echo.Map{
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
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid product id", err)
		}
		var product models.Products
		if err := db.Preload("User").First(&product, "id = ? AND is_deleted_by_user = ?", uid, false).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "Product not found", err)
		}

		// Increment view count
		db.Model(&product).Update("views", gorm.Expr("views + ?", 1))

		// Send email if first view and is_notified is false
		if !product.IsNotified && product.User.Email != "" {
			db.Model(&product).Update("is_notified", true)
			go func(email, uName, pName string, pID uuid.UUID) {
				_ = emails.SendProductViewedMail(email, uName, pName, pID.String())
			}(product.User.Email, product.User.UserName, product.Name, product.ID)
		}

		// Check if user has liked this product
		isLiked := false
		if userIdVal := c.Get("user_id"); userIdVal != nil {
			if uid, ok := userIdVal.(uuid.UUID); ok {
				var count int64
				db.Model(&models.ProductLike{}).Where("product_id = ? AND user_id = ?", product.ID, uid).Count(&count)
				isLiked = count > 0
			}
		}

		// Convert to safe response without password
		response := ConvertToProductResponse(product, isLiked)

		return utils.ResponseSucess(c, http.StatusOK, "Product fetched", echo.Map{"product": response})
	}
}

func GetUserProduct(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uuid.UUID)
		ProductID := c.Param("id")

		var product models.Products
		if err := db.Preload("User").Where("id = ? AND user_id = ? AND is_deleted_by_user = ?", ProductID, userID, false).First(&product).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "Product not found or unauthorized", err)
		}

		// Since this is GetUserProduct, it's the user's own product list view or detail
		// But they might still want to see if they liked it (unlikely but consistent)
		isLiked := false
		var count int64
		db.Model(&models.ProductLike{}).Where("product_id = ? AND user_id = ?", product.ID, userID).Count(&count)
		isLiked = count > 0

		// Convert to safe response without password
		response := ConvertToProductResponse(product, isLiked)

		return utils.ResponseSucess(c, http.StatusOK, "User product fetched successfully", echo.Map{"product": response})
	}

}

func DeleteUserProduct(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		UserId := c.Get("user_id").(uuid.UUID)
		id := c.Param("id")

		var product models.Products
		// Instead of deleting, we set IsDeletedByUser to true
		result := db.Model(&product).Where("id = ? AND user_id = ?", id, UserId).Update("is_deleted_by_user", true)
		if result.Error != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to mark product as deleted", result.Error)
		}

		if result.RowsAffected == 0 {
			return utils.ResponseError(c, http.StatusNotFound, "Product not found or unauthorized", nil)
		}

		return utils.ResponseSucess(c, http.StatusOK, "Product deleted successfully", nil)
	}
}

func SearchProducts(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		query := c.QueryParam("q")
		if query == "" {
			return utils.ResponseError(c, http.StatusBadRequest, "Missing search query", nil)
		}

		searchTerm := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"

		var products []models.Products

		if err := db.Where("(LOWER(name) LIKE ? OR LOWER(brand_name) LIKE ? OR LOWER(category_name) LIKE ? OR LOWER(university) LIKE ?) AND is_deleted_by_user = ?", searchTerm, searchTerm, searchTerm, searchTerm, false).Limit(20).Find(&products).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch search results", err)
		}

		// Extract unique categories with counts

		categoryMap := make(map[string]int)
		brandMap := make(map[string]int)
		universityMap := make(map[string]int)

		for _, p := range products {

			if p.CategoryName != "" {
				categoryMap[p.CategoryName]++
			}
			if p.BrandName != "" {
				categoryMap[p.BrandName]++
			}
			if p.University != "" {
				universityMap[p.University]++
			}

		}
		// response := map[string]interface{}{
		// 	"keywords":   []string{query},
		// 	"categories": categories,
		// 	"products":   products,
		// }

		var suggestions []models.Suggestion

		suggestions = append(suggestions, models.Suggestion{
			Type: "keyword",
			Text: query,
		})

		for category, count := range categoryMap {
			suggestions = append(suggestions, models.Suggestion{
				Type:     "category",
				Text:     query,
				Category: category,
				Count:    count,
			})
		}

		for brand, count := range brandMap {
			suggestions = append(suggestions, models.Suggestion{
				Type:  "brand",
				Text:  query,
				Brand: brand,
				Count: count,
			})
		}

		for university, count := range universityMap {
			suggestions = append(suggestions, models.Suggestion{
				Type:       "university",
				Text:       query,
				University: university,
				Count:      count,
			})
		}

		productLimit := 5
		if len(products) < 5 {

			productLimit = len(products)
		}

		for i := 0; i < productLimit; i++ {
			p := products[i]
			suggestions = append(suggestions, models.Suggestion{
				Type:      "product",
				Text:      p.Name,
				Category:  p.CategoryName,
				Brand:     p.BrandName,
				ProductID: p.ID.String(),
			})
		}

		response := map[string]interface{}{
			"query":       query,
			"suggestions": suggestions,
			"total":       len(products),
		}
		return utils.ResponseSucess(c, http.StatusOK, "Search Results Retrieved", response)

	}
}

func GetTotalProductsByCatgory(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {

		var products models.Products

		type Result struct {
			Category string `json:"category"`
			Total    int64  `json:"total"`
		}

		var results []Result

		if err := db.Model(&products).Where("status = ?", models.StatusOngoing).Select("category_name as category, COUNT(*)  as total").Group("category_name").Scan(&results).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrieve product counts", err)
		}

		return utils.ResponseSucess(c, http.StatusOK, "Products counts retrieved", echo.Map{"results": results})
	}
}

func GetSimilarProducts(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		if limit <= 0 {
			limit = 10
		}

		var product models.Products
		if err := db.First(&product, "id = ?", id).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "Product not found", err)
		}

		similarProduct, err := utils.FindSimilarProducts(db, &product, limit)

		if err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to find similar products", err)
		}

		results := []models.ProductResponse{}

		for _, p := range similarProduct {
			results = append(results, ConvertToProductResponse(p, false))
		}
		return utils.ResponseSucess(c, http.StatusOK, "Similar products retrieved", results)
	}
}

func GetSearchResults(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		query := c.QueryParam("q")
		category := c.QueryParam("category")
		brand := c.QueryParam("brand")
		page := c.QueryParam("page")
		university := c.QueryParam("university")
		if page == "" {
			page = "1"
		}

		pageNum, _ := strconv.Atoi(page)
		limit := 20
		offset := (pageNum - 1) * limit

		dbQuery := db.Model(&models.Products{})

		if query != "" {
			searchTerm := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
			dbQuery = dbQuery.Where("LOWER(name) LIKE ? OR LOWER(brand_name) LIKE ? OR LOWER(category_name) LIKE ?", searchTerm, searchTerm, searchTerm)
		}
		if category != "" {
			dbQuery = dbQuery.Where("LOWER(category_name) = ?", strings.ToLower(category))
		}

		if brand != "" {
			dbQuery = dbQuery.Where("LOWER(brand_name) = ?", strings.ToLower(brand))
		}

		if university != "" {
			dbQuery = dbQuery.Where("LOWER(university) = ?", strings.ToLower(university))
		}

		var total int64
		dbQuery.Count(&total)

		var products []models.Products
		if err := dbQuery.Limit(limit).Offset(offset).Find(&products).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch products", err)

		}

		response := map[string]interface{}{
			"products":    products,
			"total":       total,
			"page":        pageNum,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		}

		return utils.ResponseSucess(c, http.StatusOK, "Products Retrieved", response)

	}
}

func ToggleLike(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		productID := c.Param("id")
		puid, err := uuid.Parse(productID)
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid product id", err)
		}

		uID, ok := c.Get("user_id").(uuid.UUID)
		if !ok {
			return utils.ResponseError(c, http.StatusUnauthorized, "Unauthorized", nil)
		}

		var like models.ProductLike
		err = db.Where("product_id = ? AND user_id = ?", puid, uID).First(&like).Error

		if err == gorm.ErrRecordNotFound {
			// Like the product
			newLike := models.ProductLike{
				ProductID: puid,
				UserID:    uID,
			}
			if err := db.Create(&newLike).Error; err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Could not like product", err)
			}
			// Update likes count on product
			db.Model(&models.Products{}).Where("id = ?", puid).Update("likes", gorm.Expr("likes + ?", 1))
			return utils.ResponseSucess(c, http.StatusOK, "Product liked", nil)
		} else if err == nil {
			// Unlike the product
			if err := db.Delete(&like).Error; err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Could not unlike product", err)
			}
			// Update likes count on product
			db.Model(&models.Products{}).Where("id = ?", puid).Update("likes", gorm.Expr("likes - ?", 1))
			return utils.ResponseSucess(c, http.StatusOK, "Product unliked", nil)
		}

		return utils.ResponseError(c, http.StatusInternalServerError, "Database error", err)
	}
}

func CreateSearchAlert(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			Email    string `json:"email"`
			Keyword  string `json:"keyword"`
			Category string `json:"category"`
		}

		if err := c.Bind(&req); err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid input", err)
		}

		if req.Email == "" {
			return utils.ResponseError(c, http.StatusBadRequest, "Email is required", nil)
		}

		alert := models.SearchAlert{
			Email:    req.Email,
			Keyword:  req.Keyword,
			Category: req.Category,
		}

		if err := db.Create(&alert).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to create search alert", err)
		}

		return utils.ResponseSucess(c, http.StatusCreated, "Search alert created successfully", nil)
	}
}

func GetPublicStats(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var productCount int64
		var sellerCount int64

		if err := db.Model(&models.Products{}).Where("status = ? AND is_deleted_by_user = ?", models.StatusOngoing, false).Count(&productCount).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count products", err)
		}

		if err := db.Model(&models.User{}).Where("role = ? AND status = ?", models.RoleUser, models.UserActive).Count(&sellerCount).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count vendors", err)
		}

		return utils.ResponseSucess(c, http.StatusOK, "Public stats fetched successfully", echo.Map{
			"total_products": productCount,
			"total_sellers":  sellerCount,
		})
	}
}

// CreateGuestProduct allows unregistered users to quickly list 1 physical selling item
func CreateGuestProduct(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		guestEmail := strings.TrimSpace(c.FormValue("guest_email"))
		guestPhone := strings.TrimSpace(c.FormValue("guest_phone"))

		if guestEmail == "" || guestPhone == "" {
			return utils.ResponseError(c, http.StatusBadRequest, "Guest email and phone number are required", nil)
		}

		// Enforce limit: Unregistered guests can only list 1 product
		var count int64
		db.Model(&models.Products{}).Where("is_guest_listing = ? AND (guest_email = ? OR guest_phone = ?)", true, guestEmail, guestPhone).Count(&count)
		if count >= 1 {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status":  false,
				"message": "Unregistered users can only list 1 product. Please register or login to manage and list more products.",
			})
		}

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
		university := c.FormValue("university")

		// Guest listings are restricted strictly to selling items (MARKET)
		productType := "MARKET"

		if name == "" || productPriceStr == "" || categoryName == "" || description == "" {
			return utils.ResponseError(c, http.StatusBadRequest, "Required fields are missing", nil)
		}

		productPrice, err := strconv.ParseFloat(productPriceStr, 64)
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid product price", err)
		}
		marketPriceFrom, _ := strconv.ParseFloat(marketPriceFromStr, 64)
		if marketPriceFrom == 0 {
			marketPriceFrom = productPrice
		}
		marketPriceTo, _ := strconv.ParseFloat(marketPriceToStr, 64)
		if marketPriceTo == 0 {
			marketPriceTo = productPrice
		}

		isNegotiable := strings.ToLower(isNegotiableStr) == "true"

		form, err := c.MultipartForm()
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid form input", err)
		}

		files := form.File["new_images"]
		if len(files) == 0 {
			return utils.ResponseError(c, http.StatusBadRequest, "You must upload at least one image", nil)
		}

		var imageUrls []string
		for _, file := range files {
			src, err := file.Open()
			if err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to open image", err)
			}

			tempFilePath := filepath.Join(os.TempDir(), uuid.New().String()+"_"+filepath.Base(file.Filename))
			out, err := os.Create(tempFilePath)
			if err != nil {
				src.Close()
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to create temp file", err)
			}

			if _, err := io.Copy(out, src); err != nil {
				src.Close()
				out.Close()
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to copy image", err)
			}
			src.Close()
			out.Close()

			url, err := utils.UploadToCloudinary(tempFilePath, "guest_products")
			if err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to upload image", err)
			}
			imageUrls = append(imageUrls, url)
			os.Remove(tempFilePath)
		}

		imageUrlsJSON, _ := json.Marshal(imageUrls)

		product := models.Products{
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
			University:        university,
			BrandName:         brandName,
			ImageUrls:         datatypes.JSON(imageUrlsJSON),
			Status:            models.StatusOngoing,
			ProductType:       productType,
			GuestEmail:        guestEmail,
			GuestPhone:        guestPhone,
			IsGuestListing:    true,
		}

		if err := db.Create(&product).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to save guest product", err)
		}

		// Send Email to Guest User informing them of listing and encouraging account registration
		go func(email, pName, pID string) {
			_ = emails.SendGuestProductListedEmail(email, pName, pID)
		}(guestEmail, name, product.ID.String())

		response := ConvertToProductResponse(product, false)

		func(p models.Products) {
			// 1. Facebook/Instagram Auto Post
			message := fmt.Sprintf("🛍️ New Product Alert: %s\n\nPrice: ₦%.2f\nCondition: %s\n\nCheck it out on Nedzl!", p.Name, p.ProductPrice, p.Condition)
			link := fmt.Sprintf("https://nedzl.com/product-details/%s", p.ID.String())
			igCaption := fmt.Sprintf("%s\n\nLink in bio or copy: %s", message, link)
			if len(imageUrls) > 0 {
				imageUrl := imageUrls[0]
				if err := utils.PostToFacebook(message, imageUrl, link); err != nil {
					log.Printf("Facebook auto-post failed for product %s: %v", p.ID, err)
				}
				if err := utils.PostToInstagram(igCaption, imageUrl); err != nil {
					log.Printf("Instagram auto-post failed for product %s: %v", p.ID, err)
				}
			}
		}(product)
		return utils.ResponseSucess(c, http.StatusCreated, "Product listed successfully as guest", echo.Map{"products": response})
	}
}
