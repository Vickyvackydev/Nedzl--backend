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
	"api/whatsapp"

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
				foodOrder.PaymentStatus = "SUCCESS"
				db.DB.Save(&foodOrder)

				// Send Email & WhatsApp to Vendor upon successful payment
				go func(order models.FoodOrder) {
					subMenusStr := "None"
					if len(order.SubMenus) > 0 {
						subMenusStr = string(order.SubMenus)
					}
					if order.Vendor.Email != "" {
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
					}

					vendorPhone := order.Vendor.PhoneNumber
					if vendorPhone == "" {
						vendorPhone = order.Product.GuestPhone
					}
					if vendorPhone != "" {
						_ = whatsapp.SendVendorFoodOrderWhatsApp(
							vendorPhone,
							order.OrderNumber,
							order.Product.Name,
							order.CustomerName,
							order.CustomerPhone,
							order.TotalAmount,
							order.DeliveryAddress,
						)
					}
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

				// Send Email & WhatsApp to Artisan upon successful payment
				go func(booking models.ServiceBooking) {
					customerName := booking.User.UserName
					if customerName == "" {
						customerName = "Nedzl Customer"
					}

					if booking.Artisan.Email != "" {
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
					}

					artisanPhone := booking.Artisan.PhoneNumber
					if artisanPhone == "" {
						artisanPhone = booking.Service.GuestPhone
					}
					if artisanPhone != "" {
						_ = whatsapp.SendServiceBookingWhatsApp(
							artisanPhone,
							booking.BookingNumber,
							booking.Service.Name,
							customerName,
							booking.CustomerPhone,
							booking.ServiceAddress,
							booking.ScheduledDate.Format("2006-01-02 15:04"),
						)
					}
				}(serviceBooking)
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "Webhook processed successfully",
	})
}

// VerifyPaystackPayment verifies transaction reference with Paystack API and updates order/booking status
func VerifyPaystackPayment(c echo.Context) error {
	reference := c.Param("reference")
	if reference == "" {
		reference = c.QueryParam("reference")
	}
	if reference == "" {
		return utils.ResponseError(c, http.StatusBadRequest, "Transaction reference required", nil)
	}

	verified, msg, err := utils.VerifyPaystackTransaction(reference, 0)
	if !verified {
		return utils.ResponseError(c, http.StatusBadRequest, msg, err)
	}

	// 1. Check FoodOrder
	var foodOrder models.FoodOrder
	if err := db.DB.Preload("Product").Preload("Vendor").Where("payment_reference = ?", reference).First(&foodOrder).Error; err == nil {
		if foodOrder.Status != "PAID" {
			foodOrder.Status = "PAID"
			foodOrder.PaymentStatus = "SUCCESS"
			db.DB.Save(&foodOrder)

			go func(order models.FoodOrder) {
				subMenusStr := "None"
				if len(order.SubMenus) > 0 {
					subMenusStr = string(order.SubMenus)
				}
				if order.Vendor.Email != "" {
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
				}
				vendorPhone := order.Vendor.PhoneNumber
				if vendorPhone == "" {
					vendorPhone = order.Product.GuestPhone
				}
				if vendorPhone != "" {
					_ = whatsapp.SendVendorFoodOrderWhatsApp(
						vendorPhone,
						order.OrderNumber,
						order.Product.Name,
						order.CustomerName,
						order.CustomerPhone,
						order.TotalAmount,
						order.DeliveryAddress,
					)
				}
			}(foodOrder)
		}
		return utils.ResponseSucess(c, http.StatusOK, "Payment verified successfully", echo.Map{"order": foodOrder})
	}

	// 2. Check ServiceBooking
	var serviceBooking models.ServiceBooking
	if err := db.DB.Preload("Service").Preload("Artisan").Preload("User").Where("payment_reference = ?", reference).First(&serviceBooking).Error; err == nil {
		if serviceBooking.Status != "BOOKED" {
			serviceBooking.Status = "BOOKED"
			serviceBooking.PaymentStatus = "HELD_IN_ESCROW"
			db.DB.Save(&serviceBooking)

			go func(booking models.ServiceBooking) {
				customerName := booking.User.UserName
				if customerName == "" {
					customerName = "Nedzl Customer"
				}
				if booking.Artisan.Email != "" {
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
				}
				artisanPhone := booking.Artisan.PhoneNumber
				if artisanPhone == "" {
					artisanPhone = booking.Service.GuestPhone
				}
				if artisanPhone != "" {
					_ = whatsapp.SendServiceBookingWhatsApp(
						artisanPhone,
						booking.BookingNumber,
						booking.Service.Name,
						customerName,
						booking.CustomerPhone,
						booking.ServiceAddress,
						booking.ScheduledDate.Format("2006-01-02 15:04"),
					)
				}
			}(serviceBooking)
		}
		return utils.ResponseSucess(c, http.StatusOK, "Payment verified successfully", echo.Map{"booking": serviceBooking})
	}

	return utils.ResponseSucess(c, http.StatusOK, "Payment verified", nil)
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
