package handlers

import (
	"api/emails"
	"api/models"
	"api/utils"
	"api/whatsapp"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type CreateServiceBookingRequest struct {
	ServiceID        uuid.UUID `json:"service_id"`
	ScheduledDate    time.Time `json:"scheduled_date"`
	ServiceAddress   string    `json:"service_address"`
	CustomerPhone    string    `json:"customer_phone"`
	Notes            string    `json:"notes"`
	BookingFee       float64   `json:"booking_fee"`
	PaymentReference string    `json:"payment_reference"`
	CallbackURL      string    `json:"callback_url"`
}

func CreateServiceBooking(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uuid.UUID)

		var req CreateServiceBookingRequest
		if err := c.Bind(&req); err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid request body", err)
		}

		if req.ServiceID == uuid.Nil || req.ServiceAddress == "" || req.CustomerPhone == "" {
			return utils.ResponseError(c, http.StatusBadRequest, "Service ID, address, and customer phone are required", nil)
		}

		// Fetch service product
		var service models.Products
		if err := db.Preload("User").First(&service, "id = ?", req.ServiceID).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "Service not found", err)
		}

		fee := req.BookingFee
		if fee <= 0 {
			fee = service.ProductPrice
		}

		platformFee := fee * 0.10
		artisanPayout := fee * 0.90

		bookingNumber := fmt.Sprintf("NDZ-BK-%d", time.Now().UnixNano()/1e6)

		var customer models.User
		db.First(&customer, "id = ?", userID)
		customerEmail := customer.Email
		if customerEmail == "" {
			customerEmail = "customer@nedzl.com"
		}

		callbackURL := req.CallbackURL
		if callbackURL == "" {
			callbackURL = fmt.Sprintf("%s/dashboard?tab=service_bookings", utils.GetFrontendBaseURL(c))
		}
		checkoutURL, _ := utils.InitializePaystackTransaction(customerEmail, fee, bookingNumber, callbackURL)

		status := "BOOKED"
		paymentStatus := "HELD_IN_ESCROW"
		if checkoutURL != "" {
			status = "PENDING"
			paymentStatus = "PENDING"
		}

		paymentRef := req.PaymentReference
		if paymentRef == "" {
			paymentRef = bookingNumber
		}

		if req.ScheduledDate.IsZero() {
			req.ScheduledDate = time.Now().Add(24 * time.Hour)
		}

		var artisanID uuid.UUID
		if service.UserID != nil {
			artisanID = *service.UserID
		}

		booking := models.ServiceBooking{
			BookingNumber:    bookingNumber,
			UserID:           userID,
			ArtisanID:        artisanID,
			ServiceID:        service.ID,
			BookingFee:       fee,
			PlatformFee:      platformFee,
			ArtisanPayout:    artisanPayout,
			ScheduledDate:    req.ScheduledDate,
			ServiceAddress:   req.ServiceAddress,
			CustomerPhone:    req.CustomerPhone,
			Notes:            req.Notes,
			PaymentReference: paymentRef,
			Status:           status,
			PaymentStatus:    paymentStatus,
		}

		if err := db.Create(&booking).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to create service booking", err)
		}

		// Notify artisan via email & WhatsApp ONLY if booking is already confirmed
		if status == "BOOKED" {
			go func() {
				if service.User.Email != "" {
					customerName := customer.UserName
					if customerName == "" {
						customerName = "Valued Customer"
					}

					_ = emails.SendArtisanBookingNotificationEmail(
						service.User.Email,
						service.User.UserName,
						bookingNumber,
						service.Name,
						customerName,
						req.CustomerPhone,
						req.ServiceAddress,
						req.ScheduledDate,
						fee,
					)
					
					_ = whatsapp.SendServiceBookingWhatsApp(
						service.User.PhoneNumber,
						bookingNumber,
						service.Name,
						customerName,
						req.CustomerPhone,
						req.ServiceAddress,
						req.ScheduledDate.Format("2006-01-02 15:04"),
					)
				}
			}()
		}

		return utils.ResponseSucess(c, http.StatusCreated, "Service booking created successfully", echo.Map{
			"booking":      booking,
			"checkout_url": checkoutURL,
		})
	}
}

func GetUserServiceBookings(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uuid.UUID)

		var bookings []models.ServiceBooking
		if err := db.Preload("Service").Preload("Artisan").Where("user_id = ?", userID).Order("created_at desc").Find(&bookings).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch user service bookings", err)
		}

		return c.JSON(http.StatusOK, echo.Map{
			"data": bookings,
		})
	}
}

func GetArtisanServiceBookings(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		artisanID := c.Get("user_id").(uuid.UUID)

		var bookings []models.ServiceBooking
		if err := db.Preload("Service").Preload("User").Where("artisan_id = ?", artisanID).Order("created_at desc").Find(&bookings).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch artisan service bookings", err)
		}

		return c.JSON(http.StatusOK, echo.Map{
			"data": bookings,
		})
	}
}

func ArtisanCompleteBooking(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		artisanID := c.Get("user_id").(uuid.UUID)
		bookingID := c.Param("id")

		var booking models.ServiceBooking
		if err := db.First(&booking, "id = ? AND artisan_id = ?", bookingID, artisanID).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "Booking not found", err)
		}

		now := time.Now()
		booking.Status = "ARTISAN_COMPLETED"
		booking.ArtisanCompletedAt = &now

		if err := db.Save(&booking).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to update booking", err)
		}

		return c.JSON(http.StatusOK, echo.Map{
			"message": "Service marked as completed by artisan. Waiting for customer confirmation or 24-hour auto-payout.",
			"booking": booking,
		})
	}
}

func CustomerCompleteBooking(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uuid.UUID)
		bookingID := c.Param("id")

		var booking models.ServiceBooking
		if err := db.First(&booking, "id = ? AND user_id = ?", bookingID, userID).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "Booking not found", err)
		}

		now := time.Now()
		booking.Status = "COMPLETED"
		booking.CompletedAt = &now
		booking.PaymentStatus = "RELEASED_TO_ARTISAN"

		if err := db.Save(&booking).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to release escrow payout", err)
		}

		return c.JSON(http.StatusOK, echo.Map{
			"message": "Service completed successfully! Payout released to artisan.",
			"booking": booking,
		})
	}
}
