package handlers

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"api/db"
	"api/emails"
	"api/models"
	"api/utils"

	"github.com/labstack/echo/v4"
)

type PaystackWebhookPayload struct {
	Event string `json:"event"`
	Data  struct {
		ID              int64   `json:"id"`
		Reference       string  `json:"reference"`
		Amount          float64 `json:"amount"` // in kobo
		Status          string  `json:"status"`
		GatewayResponse string  `json:"gateway_response"`
		Customer        struct {
			Email string `json:"email"`
		} `json:"customer"`
	} `json:"data"`
}

// HandlePaystackWebhook handles incoming webhook events from Paystack.
func HandlePaystackWebhook(c echo.Context) error {
	secretKey := os.Getenv("PAYSTACK_SECRET_KEY")

	// Read body bytes for HMAC signature verification
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Failed to read request body",
		})
	}

	// Verify Signature if Secret Key is set
	if secretKey != "" {
		signature := c.Request().Header.Get("x-paystack-signature")
		if signature == "" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"message": "Missing Paystack signature header",
			})
		}

		mac := hmac.New(sha512.New, []byte(secretKey))
		mac.Write(bodyBytes)
		expectedMAC := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"message": "Invalid Paystack webhook signature",
			})
		}
	}

	var payload PaystackWebhookPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Failed to parse webhook JSON",
		})
	}

	// Handle successful charge event
	if payload.Event == "charge.success" && payload.Data.Status == "success" {
		reference := payload.Data.Reference

		// 1. Check if reference belongs to a FoodOrder
		var foodOrder models.FoodOrder
		if err := db.DB.Preload("Product").Preload("Vendor").Where("payment_reference = ?", reference).First(&foodOrder).Error; err == nil {
			if foodOrder.Status != "PAID" {
				foodOrder.Status = "PAID"
				db.DB.Save(&foodOrder)

				// Send Email to Vendor
				go func(order models.FoodOrder) {
					subMenusStr := string(order.SubMenus)
					_ = emails.SendVendorFoodOrderEmail(
						order.Vendor.Email,
						order.Vendor.UserName,
						order.OrderNumber,
						order.Product.Name,
						order.CustomerName,
						order.CustomerPhone,
						order.DeliveryAddress,
						subMenusStr,
						order.TotalAmount,
						order.Product.DeliveryFee,
					)
				}(foodOrder)
			}
		}

		// 2. Check if reference belongs to a ServiceBooking
		var serviceBooking models.ServiceBooking
		if err := db.DB.Preload("Service").Preload("Artisan").Preload("User").Where("payment_reference = ?", reference).First(&serviceBooking).Error; err == nil {
			if serviceBooking.Status != "BOOKED" {
				serviceBooking.Status = "BOOKED"
				serviceBooking.PaymentStatus = "HELD_IN_ESCROW"
				db.DB.Save(&serviceBooking)

				// Send Email to Artisan
				go func(booking models.ServiceBooking) {
					customerName := booking.User.UserName
					if customerName == "" {
						customerName = "Nedzl Customer"
					}

					_ = emails.SendArtisanBookingNotificationEmail(
						booking.Artisan.Email,
						booking.Artisan.UserName,
						booking.BookingNumber,
						booking.Service.Name,
						customerName,
						booking.CustomerPhone,
						booking.ServiceAddress,
						booking.ScheduledDate,
						booking.BookingFee,
					)
				}(serviceBooking)
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "Webhook processed successfully",
	})
}

type ResolveBankRequest struct {
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
}

// ResolveBank verifies account number and bank code via Paystack API
func ResolveBank(c echo.Context) error {
	accountNumber := c.QueryParam("account_number")
	bankCode := c.QueryParam("bank_code")

	if accountNumber == "" || bankCode == "" {
		var req ResolveBankRequest
		if err := c.Bind(&req); err == nil {
			if req.AccountNumber != "" {
				accountNumber = req.AccountNumber
			}
			if req.BankCode != "" {
				bankCode = req.BankCode
			}
		}
	}

	if len(accountNumber) != 10 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  false,
			"message": "Account number must be 10 digits",
		})
	}

	accountName, err := utils.ResolveBankAccount(accountNumber, bankCode)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  false,
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "Account resolved successfully",
		"data": map[string]interface{}{
			"account_number": accountNumber,
			"account_name":   accountName,
		},
	})
}
