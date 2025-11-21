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

// parseStartEnd tries to parse ?start and ?end in YYYY-MM-DD format.
// If not provided, it derives start/end from period (7d, 1m, 1yr) or defaults to calendar year.
func parseStartEnd(c echo.Context) (start time.Time, end time.Time, err error) {
	now := time.Now().UTC()
	// Try explicit start/end first
	startParam := c.QueryParam("start")
	endParam := c.QueryParam("end")
	if startParam != "" && endParam != "" {
		start, err = time.ParseInLocation("2006-01-02", startParam, time.UTC)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		// set start to midnight UTC
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)

		end, err = time.ParseInLocation("2006-01-02", endParam, time.UTC)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		// set end to end of day UTC
		end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
		return start, end, nil
	}

	// fallback to period param
	period := c.QueryParam("period")
	yearParam := c.QueryParam("year")
	currentYear := now.Year()
	if yearParam != "" {
		if y, perr := strconv.Atoi(yearParam); perr == nil {
			currentYear = y
		}
	}

	switch period {
	case "7d":
		end = now.UTC()
		start = end.AddDate(0, 0, -7)
		return start, end, nil
	case "1m":
		end = now.UTC()
		start = end.AddDate(0, -1, 0)
		return start, end, nil
	case "1yr":
		end = now.UTC()
		start = end.AddDate(-1, 0, 0)
		return start, end, nil
	default:
		// default to calendar year of currentYear
		start = time.Date(currentYear, 1, 1, 0, 0, 0, 0, time.UTC)
		end = time.Date(currentYear, 12, 31, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
		return start, end, nil
	}
}

// --- Admin Dashboard (Products + Users) ---
func GetDashboardOverview(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		now := time.Now().UTC()

		startDate, endDate, err := parseStartEnd(c)
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid start/end date format. Use YYYY-MM-DD", err)
		}

		// determine previous period (same length right before current start)
		periodLen := endDate.Sub(startDate)
		prevEnd := startDate.Add(-time.Nanosecond)
		prevStart := prevEnd.Add(-periodLen).Add(time.Nanosecond)

		// prepare counters
		var (
			totalProducts, productsNewInPeriod                          int64
			activeProducts, closedProducts, flaggedProducts             int64
			prevTotalProducts, prevProductsNewInPrevPeriod              int64
			prevActiveProducts, prevClosedProducts, prevFlaggedProducts int64
		)

		productModel := db.Model(&models.Products{})
		userModel := db.Model(&models.User{})

		// --- Products: totals & new in period ---
		// total products ever created in period (new products)
		if err := productModel.Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&productsNewInPeriod).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count new products in period", err)
		}

		// total products overall by status (lifetime counts)
		if err := productModel.Where("status = ?", models.StatusOngoing).Count(&activeProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count active products", err)
		}
		// closed (lifetime)
		if err := productModel.Where("status = ?", models.StatusClosed).Count(&closedProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count closed products", err)
		}
		// flagged/rejected (lifetime)
		if err := productModel.Where("status = ?", models.StatusRejected).Count(&flaggedProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count flagged products", err)
		}

		// total products in current period (for "TotalProductsListed" metric we interpret as new products in period)
		totalProducts = productsNewInPeriod

		// --- Users: new users in period ---
		var totalUsersInPeriod int64
		if err := userModel.Where("created_at BETWEEN ? AND ? AND role = ?", startDate, endDate, models.RoleUser).Count(&totalUsersInPeriod).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count registered users in period", err)
		}

		stats := models.DashboardStats{
			TotalProductsListed:     totalProducts,
			ActiveProducts:          activeProducts,
			FlaggedReportedProducts: flaggedProducts,
			ClosedSoldProducts:      closedProducts,
			TotalRegisteredSellers:  totalUsersInPeriod,
		}

		// --- Previous period counts for growth ---
		if err := productModel.Where("created_at BETWEEN ? AND ?", prevStart, prevEnd).Count(&prevProductsNewInPrevPeriod).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count previous new products", err)
		}

		// For status-based growth — we attempt to count how many products entered that status during the period.
		// closed products -> use closed_at (preferred) otherwise fallback to created_at if closed_at is null
		if err := productModel.Where("(status = ? AND closed_at BETWEEN ? AND ?) OR (status = ? AND closed_at IS NULL AND created_at BETWEEN ? AND ?)",
			models.StatusClosed, startDate, endDate, models.StatusClosed, startDate, endDate).Count(&closedProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count closed products in period", err)
		}
		if err := productModel.Where("(status = ? AND closed_at BETWEEN ? AND ?) OR (status = ? AND closed_at IS NULL AND created_at BETWEEN ? AND ?)",
			models.StatusClosed, prevStart, prevEnd, models.StatusClosed, prevStart, prevEnd).Count(&prevClosedProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count closed products in previous period", err)
		}

		// For ongoing and rejected statuses we can't reliably know "entered status at" time without an audit log.
		// We'll approximate "entered in period" by checking current status + created_at in period (reasonable fallback).
		if err := productModel.Where("status = ? AND created_at BETWEEN ? AND ?", models.StatusOngoing, startDate, endDate).Count(&activeProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count ongoing products in period", err)
		}
		if err := productModel.Where("status = ? AND created_at BETWEEN ? AND ?", models.StatusOngoing, prevStart, prevEnd).Count(&prevActiveProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count ongoing products in previous period", err)
		}

		if err := productModel.Where("status = ? AND created_at BETWEEN ? AND ?", models.StatusRejected, startDate, endDate).Count(&flaggedProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count flagged products in period", err)
		}
		if err := productModel.Where("status = ? AND created_at BETWEEN ? AND ?", models.StatusRejected, prevStart, prevEnd).Count(&prevFlaggedProducts).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count flagged products in previous period", err)
		}

		// User previous period counts
		if err := userModel.Where("created_at BETWEEN ? AND ? AND role = ?", prevStart, prevEnd, models.RoleUser).Count(&prevTotalUsers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count previous registered users", err)
		}

		// --- Build growth object ---
		growth := models.DashboardGrowth{
			TotalProductsListed:     calculateGrowthRate(prevProductsNewInPrevPeriod, productsNewInPeriod),
			ActiveProducts:          calculateGrowthRate(prevActiveProducts, activeProducts),
			ClosedSoldProducts:      calculateGrowthRate(prevClosedProducts, closedProducts),
			FlaggedReportedProducts: calculateGrowthRate(prevFlaggedProducts, flaggedProducts),
			TotalRegisteredSellers:  calculateGrowthRate(prevTotalUsers, totalUsersInPeriod),
		}

		// --- Monthly metrics: use currentYear (or year param if provided) ---
		yearParam := c.QueryParam("year")
		currentYear := now.Year()
		if yearParam != "" {
			if y, perr := strconv.Atoi(yearParam); perr == nil {
				currentYear = y
			}
		}

		var signupMetrics, soldMetrics []models.MonthlyMetric
		for m := 1; m <= 12; m++ {
			month := time.Month(m)
			monthName := month.String()[:3]
			formatted := fmt.Sprintf("%s %d", monthName, currentYear)

			var signups int64
			var sold int64

			if err := userModel.Where("EXTRACT(MONTH FROM created_at) = ? AND EXTRACT(YEAR FROM created_at) = ? AND role = ?", m, currentYear, models.RoleUser).Count(&signups).Error; err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch monthly signups", err)
			}

			// count closed products by closed_at month, fallback to created_at if closed_at is null
			if err := productModel.Where("(EXTRACT(MONTH FROM closed_at) = ? AND EXTRACT(YEAR FROM closed_at) = ? AND status = ?) OR (closed_at IS NULL AND EXTRACT(MONTH FROM created_at) = ? AND EXTRACT(YEAR FROM created_at) = ? AND status = ?)",
				m, currentYear, models.StatusClosed, m, currentYear, models.StatusClosed).Count(&sold).Error; err != nil {
				return utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch monthly sales", err)
			}

			signupMetrics = append(signupMetrics, models.MonthlyMetric{Month: formatted, Value: signups})
			soldMetrics = append(soldMetrics, models.MonthlyMetric{Month: formatted, Value: sold})
		}

		metrics := models.DashboardMetrics{
			CustomerSignUpMetrics: signupMetrics,
			TotalSoldProducts:     soldMetrics,
		}

		response := models.DashboardResponse{
			Stats:   stats,
			Growth:  growth,
			Metrics: metrics,
		}

		return utils.ResponseSucess(c, http.StatusOK, "Dashboard Overview Retrieved", response)
	}
}

// --- Seller/User Dashboard (Users metrics) ---
func GetUserDashboardOverview(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		now := time.Now().UTC()
		startDate, endDate, err := parseStartEnd(c)
		if err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid start/end date format. Use YYYY-MM-DD", err)
		}

		periodLen := endDate.Sub(startDate)
		prevEnd := startDate.Add(-time.Nanosecond)
		prevStart := prevEnd.Add(-periodLen).Add(time.Nanosecond)

		userModel := db.Model(&models.User{})

		var (
			totalSellersInPeriod, activeSellers, suspendedSellers, deactivatedUsers    int64
			prevTotalSellers, prevActiveSellers, prevSuspendedSellers, prevDeactivated int64
		)

		// Total sellers created in the period
		if err := userModel.Where("created_at BETWEEN ? AND ? AND role = ?", startDate, endDate, models.RoleUser).Count(&totalSellersInPeriod).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count total sellers in period", err)
		}

		// Current counts by status (lifetime current state)
		if err := userModel.Where("status = ? AND role = ?", models.UserActive, models.RoleUser).Count(&activeSellers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count active sellers", err)
		}
		if err := userModel.Where("status = ? AND role = ?", models.UserSuspended, models.RoleUser).Count(&suspendedSellers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count suspended sellers", err)
		}
		if err := userModel.Where("status = ? AND role = ?", models.UserDeactivated, models.RoleUser).Count(&deactivatedUsers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count deactivated sellers", err)
		}

		// Previous period counts (for growth). Because we may not have status-change timestamps,
		// we approximate "entered status during prev period" by users with that current status AND created in prev period.
		if err := userModel.Where("created_at BETWEEN ? AND ? AND role = ?", prevStart, prevEnd, models.RoleUser).Count(&prevTotalSellers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count previous total sellers", err)
		}
		if err := userModel.Where("status = ? AND role = ? AND created_at BETWEEN ? AND ?", models.UserActive, models.RoleUser, prevStart, prevEnd).Count(&prevActiveSellers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count previous active sellers", err)
		}
		if err := userModel.Where("status = ? AND role = ? AND created_at BETWEEN ? AND ?", models.UserSuspended, models.RoleUser, prevStart, prevEnd).Count(&prevSuspendedSellers).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count previous suspended sellers", err)
		}
		if err := userModel.Where("status = ? AND role = ? AND created_at BETWEEN ? AND ?", models.UserDeactivated, models.RoleUser, prevStart, prevEnd).Count(&prevDeactivated).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to count previous deactivated sellers", err)
		}

		stats := models.UserDashboardStats{
			TotalSellers:     totalSellersInPeriod,
			ActiveSellers:    activeSellers,
			SuspendedUsers:   suspendedSellers,
			DeactivatedUsers: deactivatedUsers,
		}

		growth := models.UserDashboardGrowth{
			TotalSellers:     calculateGrowthRate(prevTotalSellers, totalSellersInPeriod),
			ActiveSellers:    calculateGrowthRate(prevActiveSellers, activeSellers),
			SuspendedUsers:   calculateGrowthRate(prevSuspendedSellers, suspendedSellers),
			DeactivatedUsers: calculateGrowthRate(prevDeactivated, deactivatedUsers),
		}

		response := models.UserDashboardResponse{
			Stats:  stats,
			Growth: growth,
		}

		return utils.ResponseSucess(c, http.StatusOK, "User Dashboard Overview Retrieved", response)
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

func DeleteAdminProduct(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {

		id := c.Param("id")

		var product models.Products

		if err := db.Where("id =?", id).Delete(&product).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to delete product", err)
		}

		return utils.ResponseSucess(c, http.StatusOK, "Product deleted successfully", nil)
	}
}
