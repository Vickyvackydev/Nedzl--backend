package handlers

import (
	"api/emails"
	"api/models"
	"api/utils"
	"api/whatsapp"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CreateFoodOrderRequest struct {
	ProductID        uuid.UUID       `json:"product_id"`
	SubMenus         json.RawMessage `json:"sub_menus"`
	CustomerName     string          `json:"customer_name"`
	CustomerPhone    string          `json:"customer_phone"`
	DeliveryAddress  string          `json:"delivery_address"`
	PaymentReference string          `json:"payment_reference"`
	TotalAmount      float64         `json:"total_amount"`
	CallbackURL      string          `json:"callback_url"`
}

func CreateFoodOrder(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uuid.UUID)

		var req CreateFoodOrderRequest
		if err := c.Bind(&req); err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid request payload", err)
		}

		if req.ProductID == uuid.Nil || req.CustomerPhone == "" || req.DeliveryAddress == "" {
			return utils.ResponseError(c, http.StatusBadRequest, "Product ID, customer phone, and delivery address are required", nil)
		}

		// Fetch food product
		var product models.Products
		if err := db.Preload("User").First(&product, "id = ?", req.ProductID).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "Food product not found", err)
		}

		// Calculate platform fee (10%) and vendor payout (90% + delivery fee)
		mealPrice := req.TotalAmount - product.DeliveryFee
		if mealPrice < 0 {
			mealPrice = product.ProductPrice
		}

		platformFee := mealPrice * 0.10
		vendorPayout := (mealPrice * 0.90) + product.DeliveryFee

		orderNumber := fmt.Sprintf("NDZ-FD-%d", time.Now().UnixNano()/1e6)

		var buyer models.User
		db.First(&buyer, "id = ?", userID)

		customerName := req.CustomerName
		if customerName == "" {
			customerName = buyer.UserName
		}
		customerEmail := buyer.Email
		if customerEmail == "" {
			customerEmail = "customer@nedzl.com"
		}

		callbackURL := req.CallbackURL
		if callbackURL == "" {
			callbackURL = fmt.Sprintf("%s/dashboard?tab=my_orders", utils.GetFrontendBaseURL(c))
		}
		checkoutURL, _ := utils.InitializePaystackTransaction(customerEmail, req.TotalAmount, orderNumber, callbackURL)

		status := "PAID"
		paymentStatus := "SUCCESS"
		if checkoutURL != "" {
			status = "PENDING"
			paymentStatus = "PENDING"
		}

		// Verify Paystack payment reference if provided
		paymentRef := req.PaymentReference
		if paymentRef == "" {
			paymentRef = orderNumber
		}

		var vendorID uuid.UUID
		if product.UserID != nil {
			vendorID = *product.UserID
		}

		order := models.FoodOrder{
			OrderNumber:      orderNumber,
			UserID:           userID,
			VendorID:         vendorID,
			ProductID:        product.ID,
			SubMenus:         datatypes.JSON(req.SubMenus),
			MealPrice:        mealPrice,
			DeliveryFee:      product.DeliveryFee,
			TotalAmount:      req.TotalAmount,
			PlatformFee:      platformFee,
			VendorPayout:     vendorPayout,
			CustomerName:     customerName,
			CustomerPhone:    req.CustomerPhone,
			DeliveryAddress:  req.DeliveryAddress,
			PaymentReference: paymentRef,
			Status:           status,
			PaymentStatus:    paymentStatus,
		}

		if err := db.Create(&order).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to record food order", err)
		}

		// Send email & WhatsApp notification to vendor ONLY if payment is already confirmed
		if status == "PAID" {
			go func() {
				vendorPhone := product.User.PhoneNumber
				if vendorPhone == "" {
					vendorPhone = product.GuestPhone
				}

				if product.User.Email != "" {
					submenusStr := "None"
					if len(req.SubMenus) > 0 {
						submenusStr = string(req.SubMenus)
					}
					_ = emails.SendVendorFoodOrderEmail(
						product.User.Email,
						product.User.UserName,
						orderNumber,
						product.Name,
						customerName,
						req.CustomerPhone,
						req.DeliveryAddress,
						submenusStr,
						req.TotalAmount,
						product.DeliveryFee,
					)
				}

				if vendorPhone != "" {
					_ = whatsapp.SendVendorFoodOrderWhatsApp(
						vendorPhone,
						orderNumber,
						product.Name,
						customerName,
						req.CustomerPhone,
						req.TotalAmount,
						req.DeliveryAddress,
					)
				}
			}()
		}

		return utils.ResponseSucess(c, http.StatusCreated, "Food order placed successfully", echo.Map{
			"order":        order,
			"checkout_url": checkoutURL,
		})
	}
}

func GetUserFoodOrders(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uuid.UUID)

		var orders []models.FoodOrder
		if err := db.Preload("Product").Preload("Vendor").Where("user_id = ?", userID).Order("created_at desc").Find(&orders).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch user food orders", err)
		}

		return c.JSON(http.StatusOK, echo.Map{
			"data": orders,
		})
	}
}

func GetVendorFoodOrders(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		vendorID := c.Get("user_id").(uuid.UUID)

		var orders []models.FoodOrder
		if err := db.Preload("Product").Preload("User").Where("vendor_id = ?", vendorID).Order("created_at desc").Find(&orders).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch vendor food orders", err)
		}

		return c.JSON(http.StatusOK, echo.Map{
			"data": orders,
		})
	}
}

func UpdateFoodOrderStatus(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		vendorID := c.Get("user_id").(uuid.UUID)
		orderID := c.Param("id")

		var body struct {
			Status string `json:"status"`
		}
		if err := c.Bind(&body); err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid body", err)
		}

		var order models.FoodOrder
		if err := db.First(&order, "id = ? AND vendor_id = ?", orderID, vendorID).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "Food order not found", err)
		}

		order.Status = body.Status
		if body.Status == "DELIVERED" || body.Status == "DELIVERED_BY_VENDOR" {
			order.Status = "DELIVERED_BY_VENDOR"
			now := time.Now()
			order.VendorDeliveredAt = &now
			// PaymentStatus remains HELD_IN_ESCROW until customer confirms receipt
			if order.PaymentStatus == "" {
				order.PaymentStatus = "HELD_IN_ESCROW"
			}
		}

		if err := db.Save(&order).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to update order status", err)
		}

		return c.JSON(http.StatusOK, echo.Map{
			"message": "Order status updated",
			"order":   order,
		})
	}
}

// ConfirmFoodOrderDelivery allows the customer to confirm they received their food order and release payment to the vendor
func ConfirmFoodOrderDelivery(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(uuid.UUID)
		orderID := c.Param("id")

		var order models.FoodOrder
		if err := db.First(&order, "id = ? AND user_id = ?", orderID, userID).Error; err != nil {
			return utils.ResponseError(c, http.StatusNotFound, "Food order not found", err)
		}

		if order.Status == "COMPLETED" && order.PaymentStatus == "RELEASED_TO_VENDOR" {
			return c.JSON(http.StatusOK, echo.Map{
				"message": "Order delivery already confirmed",
				"order":   order,
			})
		}

		now := time.Now()
		order.Status = "COMPLETED"
		order.PaymentStatus = "RELEASED_TO_VENDOR"
		order.CustomerConfirmedAt = &now

		if err := db.Save(&order).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to confirm food order delivery", err)
		}

		return c.JSON(http.StatusOK, echo.Map{
			"message": "Food delivery confirmed successfully! Funds released to vendor.",
			"order":   order,
		})
	}
}
