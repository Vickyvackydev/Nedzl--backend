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
		if err := userModel.Where("created_at BETWEEN ? AND ? AND role =?", startDate, now, "USER").Count(&totalUsers).Error; err != nil {
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
		if err := userModel.Where("created_at BETWEEN ? AND ? AND role = ?", prevStart, prevEnd, "USER").Count(&prevTotalUsers).Error; err != nil {
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
