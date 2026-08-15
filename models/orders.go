package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type FoodOrder struct {
	ID               uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OrderNumber      string         `gorm:"type:varchar(50);uniqueIndex" json:"order_number"`
	UserID           uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	User             User           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`
	VendorID         uuid.UUID      `gorm:"type:uuid;index;not null" json:"vendor_id"`
	Vendor           User           `gorm:"foreignKey:VendorID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"vendor"`
	ProductID        uuid.UUID      `gorm:"type:uuid;index;not null" json:"product_id"`
	Product          Products       `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product"`
	SubMenus         datatypes.JSON `json:"sub_menus"`
	MealPrice        float64        `json:"meal_price"`
	DeliveryFee      float64        `json:"delivery_fee"`
	TotalAmount      float64        `json:"total_amount"`
	PlatformFee      float64        `json:"platform_fee"` // 10%
	VendorPayout     float64        `json:"vendor_payout"`// 90% + delivery_fee
	CustomerName     string         `json:"customer_name"`
	CustomerPhone    string         `json:"customer_phone"`
	DeliveryAddress  string         `json:"delivery_address"`
	PaymentReference string         `gorm:"type:varchar(100)" json:"payment_reference"`
	Status              string         `gorm:"type:varchar(30);default:'PAID'" json:"status"` // PAID, PREPARING, OUT_FOR_DELIVERY, DELIVERED_BY_VENDOR, COMPLETED, CANCELLED
	VendorDeliveredAt   *time.Time     `json:"vendor_delivered_at"`
	CustomerConfirmedAt *time.Time     `json:"customer_confirmed_at"`
	DeliveryPIN         string         `gorm:"type:varchar(10)" json:"delivery_pin"`
	PaymentStatus       string         `gorm:"type:varchar(30);default:'HELD_IN_ESCROW'" json:"payment_status"` // HELD_IN_ESCROW, RELEASED_TO_VENDOR, REFUNDED
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
}

type ServiceBooking struct {
	ID                 uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	BookingNumber      string         `gorm:"type:varchar(50);uniqueIndex" json:"booking_number"`
	UserID             uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"` // Customer
	User               User           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`
	ArtisanID          uuid.UUID      `gorm:"type:uuid;index;not null" json:"artisan_id"` // Vendor/Artisan
	Artisan            User           `gorm:"foreignKey:ArtisanID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"artisan"`
	ServiceID          uuid.UUID      `gorm:"type:uuid;index;not null" json:"service_id"` // Product item
	Service            Products       `gorm:"foreignKey:ServiceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"service"`
	BookingFee         float64        `json:"booking_fee"`
	PlatformFee        float64        `json:"platform_fee"` // 10%
	ArtisanPayout      float64        `json:"artisan_payout"` // 90%
	ScheduledDate      time.Time      `json:"scheduled_date"`
	ServiceAddress     string         `json:"service_address"`
	CustomerPhone      string         `json:"customer_phone"`
	Notes              string         `json:"notes"`
	PaymentReference   string         `gorm:"type:varchar(100)" json:"payment_reference"`
	Status             string         `gorm:"type:varchar(30);default:'BOOKED'" json:"status"` // BOOKED, IN_PROGRESS, ARTISAN_COMPLETED, COMPLETED, CANCELLED
	ArtisanCompletedAt *time.Time     `json:"artisan_completed_at"`
	CompletedAt        *time.Time     `json:"completed_at"`
	PaymentStatus      string         `gorm:"type:varchar(30);default:'HELD_IN_ESCROW'" json:"payment_status"` // HELD_IN_ESCROW, RELEASED_TO_ARTISAN, REFUNDED
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"-" gorm:"index"`
}
