package handlers

import (
	"api/models"
	"api/utils"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	// "github.com/labstack/echo"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"gorm.io/gorm"
)

// --- Helper: Safe % Growth Calculation ---
func calculateGrowthRate(previous, current int64) float64 {
	if previous == 0 && current == 0 {
		return 0
	}
	if previous == 0 {
		return 100
	}
	change := float64(current-previous) / float64(previous) * 100
	return math.Round(change*100) / 100 // Round to 2 decimals
}

func GetDashboardOverview(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		now := time.Now()
		period := c.QueryParam("period")

		// Optional year override
		currentYear := now.Year()
		if yearParam := c.QueryParam("year"); yearParam != "" {
			if y, err := strconv.Atoi(yearParam); err == nil {
				currentYear = y
			}
		}

		var startDate, prevStart, prevEnd time.Time
		switch period {
		case "7d":
			startDate = now.AddDate(0, 0, -7)
			prevEnd = startDate
			prevStart = startDate.AddDate(0, 0, -7)
		case "1m":
			startDate = now.AddDate(0, -1, 0)
			prevEnd = startDate
			prevStart = startDate.AddDate(0, -1, 0)
		case "1yr":
			startDate = now.AddDate(-1, 0, 0)
			prevEnd = startDate
			prevStart = startDate.AddDate(-1, 0, 0)
		default:
			// Default to current year
			startDate = time.Date(currentYear, 1, 1, 0, 0, 0, 0, now.Location())
			prevEnd = startDate
			prevStart = time.Date(currentYear-1, 1, 1, 0, 0, 0, 0, now.Location())
		}

		// --- Initialize variables ---
		var (
			totalProducts, activeProducts, flaggedProducts, closedProducts, totalUsers                int64
			prevProducts, prevActiveProducts, prevFlaggedProducts, prevClosedProducts, prevTotalUsers int64
		)

		productModel := db.Model(&models.Products{})
		userModel := db.Model(&models.User{})

		// --- Current period queries ---
		if err := productModel.Where("created_at BETWEEN ? AND ?", startDate, now).Count(&totalProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count total products", err)
		}
		if err := productModel.Where("status = ? AND created_at BETWEEN ? AND ?", models.StatusOngoing, startDate, now).Count(&activeProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count active products", err)
		}
		if err := productModel.Where("status = ? AND created_at BETWEEN ? AND ?", models.StatusClosed, startDate, now).Count(&closedProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count closed products", err)
		}
		if err := productModel.Where("status = ? AND created_at BETWEEN ? AND ?", models.StatusRejected, startDate, now).Count(&flaggedProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count flagged products", err)
		}
		if err := userModel.Where("created_at BETWEEN ? AND ? AND role =?", startDate, now, models.RoleUser).Count(&totalUsers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count registered users", err)
		}

		stats := models.DashboardStats{
			TotalProductsListed:     totalProducts,
			ActiveProducts:          activeProducts,
			FlaggedReportedProducts: flaggedProducts,
			ClosedSoldProducts:      closedProducts,
			TotalRegisteredSellers:  totalUsers,
		}

		// --- Previous period for growth ---
		if err := productModel.Where("created_at BETWEEN ? AND ?", prevStart, prevEnd).Count(&prevProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count previous total products", err)
		}
		if err := productModel.Where("status = ? AND created_at BETWEEN ? AND ?", models.StatusOngoing, prevStart, prevEnd).Count(&prevActiveProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count previous active products", err)
		}
		if err := productModel.Where("status = ? AND created_at BETWEEN ? AND ?", models.StatusClosed, prevStart, prevEnd).Count(&prevClosedProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count previous closed products", err)
		}
		if err := productModel.Where("status = ? AND created_at BETWEEN ? AND ?", models.StatusRejected, prevStart, prevEnd).Count(&prevFlaggedProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count previous flagged products", err)
		}
		if err := userModel.Where("created_at BETWEEN ? AND ? AND role = ?", prevStart, prevEnd, models.RoleUser).Count(&prevTotalUsers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count previous registered users", err)
		}

		growth := models.DashboardGrowth{
			TotalProductsListed:     calculateGrowthRate(prevProducts, totalProducts),
			ActiveProducts:          calculateGrowthRate(prevActiveProducts, activeProducts),
			FlaggedReportedProducts: calculateGrowthRate(prevFlaggedProducts, flaggedProducts),
			ClosedSoldProducts:      calculateGrowthRate(prevClosedProducts, closedProducts),
			TotalRegisteredSellers:  calculateGrowthRate(prevTotalUsers, totalUsers),
		}

		// --- Monthly metrics (Jan–Dec) ---
		var signupMetrics, soldMetrics []models.MonthlyMetric
		for i := 1; i <= 12; i++ {
			month := time.Month(i)
			monthName := month.String()[:3]
			formattedMonth := fmt.Sprintf("%s %d", monthName, currentYear)

			var signups, sold int64
			if err := userModel.Where("EXTRACT(MONTH FROM created_at) = ? AND EXTRACT(YEAR FROM created_at) = ?", i, currentYear).Count(&signups).Error; err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch monthly signups", err)
			}
			if err := productModel.Where("EXTRACT(MONTH FROM created_at) = ? AND EXTRACT(YEAR FROM created_at) = ? AND status = ?", i, currentYear, models.StatusClosed).Count(&sold).Error; err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch monthly sales", err)
			}

			signupMetrics = append(signupMetrics, models.MonthlyMetric{Month: formattedMonth, Value: signups})
			soldMetrics = append(soldMetrics, models.MonthlyMetric{Month: formattedMonth, Value: sold})
		}

		metrics := models.DashboardMetrics{
			CustomerSignUpMetrics: signupMetrics,
			TotalSoldProducts:     soldMetrics,
		}

		response := models.DashboardResponse{
			Stats:   stats,
			Metrics: metrics,
			Growth:  growth,
		}

		return utils.ResponseSucess(c, http.StatusOK, "Dashboard Overview Retrieved", response)
	}
}

func GetUserDashboardOverview(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		now := time.Now()
		period := c.QueryParam("period")

		// Optional year override
		currentYear := now.Year()
		if yearParam := c.QueryParam("year"); yearParam != "" {
			if y, err := strconv.Atoi(yearParam); err == nil {
				currentYear = y
			}
		}

		var startDate, prevStart, prevEnd time.Time
		switch period {
		case "7d":
			startDate = now.AddDate(0, 0, -7)
			prevEnd = startDate
			prevStart = startDate.AddDate(0, 0, -7)
		case "1m":
			startDate = now.AddDate(0, -1, 0)
			prevEnd = startDate
			prevStart = startDate.AddDate(0, -1, 0)
		case "1yr":
			startDate = now.AddDate(-1, 0, 0)
			prevEnd = startDate
			prevStart = startDate.AddDate(-1, 0, 0)
		default:
			// Default to current year
			startDate = time.Date(currentYear, 1, 1, 0, 0, 0, 0, now.Location())
			prevEnd = startDate
			prevStart = time.Date(currentYear-1, 1, 1, 0, 0, 0, 0, now.Location())
		}

		// --- Initialize variables ---
		var (
			totalSellers, activeSellers, suspendedSellers, deactivatedUsers        int64
			prevSellers, prevActiveUsers, prevSuspendedUsers, prevDeactivatedUsers int64
		)

		// productModel := db.Model(&models.Products{})
		userModel := db.Model(&models.User{})

		// --- Current period queries ---

		if err := userModel.Where("status = ? AND created_at BETWEEN ? AND ? AND role = ?", models.UserActive, startDate, now, models.RoleUser).Count(&activeSellers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count active sellers", err)
		}
		if err := userModel.Where("status = ? AND created_at BETWEEN ? AND ?", models.UserSuspended, startDate, now).Count(&suspendedSellers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count closed suspended users", err)
		}
		if err := userModel.Where("status = ? AND created_at BETWEEN ? AND ?", models.UserDeactivated, startDate, now).Count(&deactivatedUsers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count deactivated users", err)
		}
		if err := userModel.Where("role = ?", startDate, now, models.RoleUser).Count(&totalSellers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count registered users", err)
		}

		stats := models.UserDashboardStats{
			TotalSellers:     totalSellers,
			ActiveSellers:    activeSellers,
			SuspendedUsers:   suspendedSellers,
			DeactivatedUsers: deactivatedUsers,
		}

		// --- Previous period for growth ---
		if err := userModel.Where("created_at BETWEEN ? AND ? AND role = ?", prevStart, prevEnd, models.RoleUser).Count(&prevSellers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count previous total products", err)
		}
		if err := userModel.Where("status = ? AND created_at BETWEEN ? AND ?", models.UserActive, prevStart, prevEnd).Count(&prevActiveUsers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count previous active products", err)
		}
		if err := userModel.Where("status = ? AND created_at BETWEEN ? AND ?", models.UserSuspended, prevStart, prevEnd).Count(&prevSuspendedUsers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count previous closed products", err)
		}
		if err := userModel.Where("status = ? AND created_at BETWEEN ? AND ?", models.StatusRejected, prevStart, prevEnd).Count(&prevDeactivatedUsers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count previous flagged products", err)
		}

		growth := models.UserDashboardGrowth{
			TotalSellers:     calculateGrowthRate(prevSellers, totalSellers),
			ActiveSellers:    calculateGrowthRate(prevActiveUsers, activeSellers),
			SuspendedUsers:   calculateGrowthRate(prevSuspendedUsers, suspendedSellers),
			DeactivatedUsers: calculateGrowthRate(prevDeactivatedUsers, deactivatedUsers),
			// TotalRegisteredSellers:  calculateGrowthRate(prevTotalUsers, totalUsers),
		}

		response := models.UserDashboardResponse{
			Stats: stats,

			Growth: growth,
		}

		return utils.ResponseSucess(c, http.StatusOK, "Dashboard Overview Retrieved", response)
	}
}

func GetDashboardUsers(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		query := db.Model(&models.User{})
		page, _ := strconv.Atoi(c.QueryParam("page"))
		limit, _ := strconv.Atoi(c.QueryParam("limit"))

		name := c.QueryParam("name")
		phone := c.QueryParam("phone_number")
		status := c.QueryParam("status")
		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 10
		}
		offset := (page - 1) * limit

		var users []models.User

		if name != "" {
			query = query.Where("user_name ILIKE = ?", "%"+name+"%")
		}

		if phone != "" {
			query = query.Where("phone_number ILIKE = ?", "%"+phone+"%")

		}

		if status != "" {
			query = query.Where("status = ?", status)
		}

		query = query.Where("role = ?", models.RoleUser)
		// count total

		var total int64

		if err := query.Count(&total).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrieve total users count", err)
		}

		// arrange newer data above older ones
		query = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users)

		// calculate total pages

		totalPages := int(math.Ceil(float64(total) / float64(limit)))

		if err := query.Where("role = ?", models.RoleUser).Find(&users).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrieve users", err)

		}

		var result []models.UserDashboardUsers

		for _, user := range users {

			var listedCount, soldCount int64
			if err := db.Model(&models.Products{}).Where("user_id = ?", user.ID).Count(&listedCount).Error; err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrive user product counts", err)
			}
			if err := db.Model(&models.Products{}).Where("user_id = ? AND status = ?", user.ID, models.StatusClosed).Count(&soldCount).Error; err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrive user sold product counts", err)
			}

			result = append(result, models.UserDashboardUsers{
				User: models.PublicUser{
					ID:          user.ID,
					UserName:    user.UserName,
					Email:       user.Email,
					Role:        string(user.Role),
					PhoneNumber: user.PhoneNumber,
					Location:    user.Location,
					Status:      user.Status,
					ImageUrl:    user.ImageUrl,
					CreatedAt:   user.CreatedAt,
					UpdatedAt:   user.UpdatedAt,
					DeletedAt:   user.DeletedAt,
				},
				ListedProducts: listedCount,
				SoldProducts:   soldCount,
			})

		}

		return utils.ResponseSucess(c, http.StatusOK, "Users fetched successfully", echo.Map{
			"data":  result,
			"total": total,
			"meta": map[string]interface{}{
				"page":       page,
				"limit":      limit,
				"totalPages": totalPages,
			},
		})

	}
}

func GetActiveProductsUsers(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		query := db.Model(&models.Products{})
		page, _ := strconv.Atoi(c.QueryParam("page"))
		limit, _ := strconv.Atoi(c.QueryParam("limit"))

		name := c.QueryParam("name")
		phone := c.QueryParam("phone_number")
		status := c.QueryParam("status")
		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 10
		}
		offset := (page - 1) * limit

		var users []models.User

		if name != "" {
			query = query.Where("name ILIKE = ?", "%"+name+"%")
		}

		if phone != "" {
			query = query.Where("phone_number ILIKE = ?", "%"+phone+"%")

		}

		if status != "" {
			query = query.Where("status = ?", status)
		}
		// count total
		var total int64

		if err := query.Count(&total).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrieve total users count", err)
		}

		query = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users)

		// calculate total pages

		totalPages := int(math.Ceil(float64(total) / float64(limit)))

		if err := query.Where("role = ?", models.RoleUser).Find(&users).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrieve users", err)

		}

		var result []models.UserDashboardUsers

		for _, user := range users {

			var listedCount, soldCount int64
			if err := db.Model(&models.Products{}).Where("user_id = ?", user.ID).Count(&listedCount).Error; err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrive user product counts", err)
			}
			if err := db.Model(&models.Products{}).Where("user_id = ? AND status = ?", user.ID, models.StatusClosed).Count(&soldCount).Error; err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrive user sold product counts", err)
			}

			result = append(result, models.UserDashboardUsers{
				User: models.PublicUser{
					ID:          user.ID,
					UserName:    user.UserName,
					Email:       user.Email,
					Role:        string(user.Role),
					PhoneNumber: user.PhoneNumber,
					Location:    user.Location,
					Status:      user.Status,
					ImageUrl:    user.ImageUrl,
					CreatedAt:   user.CreatedAt,
					UpdatedAt:   user.UpdatedAt,
					DeletedAt:   user.DeletedAt,
				},
				ListedProducts: listedCount,
				SoldProducts:   soldCount,
			})

		}

		return utils.ResponseSucess(c, http.StatusOK, "Users fetched successfully", echo.Map{
			"data":        result,
			"total_count": total,
			"meta": map[string]interface{}{
				"page":       page,
				"limit":      limit,
				"totalPages": totalPages,
			},
		})

	}
}

func GetAdminProducts(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var products []models.Products

		// if err := db.Preload("User").Find(&products).Error; err != nil {
		// 	return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to retrieve products"})
		// }
		baseUrl := c.Scheme() + "://" + c.Request().Host + c.Path()
		query := db.Preload("User")
		user_id := c.QueryParam("user_id")

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

		if user_id != "" {
			query = query.Where("user_id = ?", user_id)
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
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrieve products", err)
		}

		// -- CONVERT TO SAFE RESPONSE --

		// Convert to safe responses without passwords
		var responses []ProductResponse
		for _, product := range products {
			responses = append(responses, ConvertToProductResponse(product))
		}

		totalPages := int(math.Ceil(float64(total) / float64(limit)))
		// var nextPageURL *string
		// if page < totalPages {
		// 	url := fmt.Sprintf("%s?page=%d&limit=%d", baseUrl, page+1, limit)
		// 	nextPageURL = &url
		// }

		// var prevPageURL *string
		// if page > 1 {
		// 	url := fmt.Sprintf("%s?page=%d&limit=%d", baseUrl, page-1, limit)
		// 	prevPageURL = &url
		// }

		// --- Generate all pages ---

		pages := []map[string]interface{}{}
		for i := 1; i <= totalPages; i++ {

			pageURL := fmt.Sprintf("%s?page=%d&limit=%d", baseUrl, i, limit)
			pages = append(pages, map[string]interface{}{
				"page": i,
				"url":  pageURL,
			})

		}

		return utils.ResponseSucess(c, http.StatusOK, "Products fetched successfully", echo.Map{
			"data":        responses,
			"total_count": total,
			"meta": map[string]interface{}{
				"page":       page,
				"limit":      limit,
				"totalPages": totalPages,
			},
		})
	}

}

func GetUserDetails(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		uid, err := uuid.Parse(id)
		if err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Invalid User id", err)

		}

		var user models.User
		var storeSettings models.StoreSetting
		// productModel := db.Model(&models.Products{})

		if err := db.First(&user, "id = ?", uid).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "User not found", err)
		}
		storeSettingsErr := db.First(&storeSettings, "user_id = ?", uid).Error

		// var totalProductListed, activeProduct, soldProducts, flaggedProducts int64
		totalProductListed, err := utils.CountUserProducts(db, uid)
		if err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrieve user products count", err)
		}
		activeProduct, err := utils.CountUserProducts(db, uid, string(models.StatusOngoing))
		if err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrieve user active products count", err)
		}
		soldProducts, err := utils.CountUserProducts(db, uid, string(models.StatusClosed))
		if err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrieve user closed products count", err)
		}
		flaggedProducts, err := utils.CountUserProducts(db, uid, string(models.StatusRejected))
		if err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to retrieve user flagged products count", err)
		}

		var storeResponse *models.UserStoreDetails = nil

		userResponse := models.PublicUser{
			ID:        user.ID,
			UserName:  user.UserName,
			Email:     user.Email,
			Role:      string(user.Role),
			ImageUrl:  user.ImageUrl,
			Location:  user.Location,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		if storeSettingsErr == nil {
			storeResponse = &models.UserStoreDetails{
				ID:                storeSettings.ID,
				BusinessName:      storeSettings.BusinessName,
				AboutCompany:      storeSettings.AboutCompany,
				StoreName:         storeSettings.StoreName,
				Address:           storeSettings.Address,
				State:             storeSettings.State,
				HowDoWeLocateYou:  storeSettings.HowDoWeLocateYou,
				BusinessHoursFrom: storeSettings.BusinessHoursFrom,
				BusinessHoursTo:   storeSettings.BusinessHoursTo,
				Region:            storeSettings.Region,
				UserID:            user.ID,
				CreatedAt:         storeSettings.CreatedAt,
				UpdatedAt:         storeSettings.UpdatedAt,
			}
		}

		userMetrics := models.UserProductStats{
			TotalProductsListed: totalProductListed,
			ActiveProducts:      activeProduct,
			SoldProducts:        soldProducts,
			FlaggedProducts:     flaggedProducts,
		}

		response := models.UserDetailsResponse{
			UserDetail:   userResponse,
			Metrics:      userMetrics,
			StoreDetails: storeResponse,
		}

		return utils.ResponseSucess(c, http.StatusOK, "User Details Retrieved", response)

	}
}
