package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ProductLike struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;index;not null" json:"product_id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Products struct {
	ID                uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	Name              string         `gorm:"column:name" json:"product_name"`
	ProductPrice      float64        `json:"product_price"`
	MarketPriceFrom   float64        `json:"market_price_from"`
	MarketPriceTo     float64        `json:"market_price_to"`
	CategoryName      string         `json:"category_name"`
	IsNegotiable      bool           `json:"is_negotiable"`
	Description       string         `json:"description"`
	State             string         `json:"state"`
	AddressInState    string         `json:"address_in_state"`
	OutStandingIssues string         `json:"outstanding_issues"`
	Condition         string         `json:"condition"`
	BrandName         string         `json:"brand_name"`
	ImageUrls         datatypes.JSON `json:"image_urls"`
	NewImages         datatypes.JSON `json:"new_images"`
	University        string         `json:"university"`
	Status            Status         `json:"status" gorm:"type:varchar(20);default:'ONGOING'"`
	UserID            *uuid.UUID     `json:"user_id" gorm:"type:uuid;default:null"`
	User              User           `json:"user" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt         time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
	ClosedAt          *time.Time     `gorm:"index"`
	Views             int64          `json:"views" gorm:"default:0"`
	Likes             int64          `json:"likes" gorm:"default:0"`
	IsDeletedByUser   bool           `json:"is_deleted_by_user" gorm:"default:false"`
	IsNotified        bool           `json:"is_notified" gorm:"default:false"`
	ProductType       string         `json:"product_type" gorm:"type:varchar(20);default:'MARKET'"` // MARKET, FOOD, SERVICE
	SubMenus          datatypes.JSON `json:"sub_menus"`
	DeliveryFee       float64        `json:"delivery_fee" gorm:"default:0"`
	OldPrice          float64        `json:"old_price" gorm:"default:0"`
	DiscountPercent   int            `json:"discount_percent" gorm:"default:0"`
	GuestEmail        string         `json:"guest_email" gorm:"size:255"`
	GuestPhone        string         `json:"guest_phone" gorm:"size:50"`
	IsGuestListing    bool           `json:"is_guest_listing" gorm:"default:false"`
	ServiceType       string         `json:"service_type" gorm:"type:varchar(100);default:''"`
}

type ProductResponse struct {
	ID                uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	Name              string         `json:"product_name"`
	ProductPrice      float64        `json:"product_price"`
	MarketPriceFrom   float64        `json:"market_price_from"`
	MarketPriceTo     float64        `json:"market_price_to"`
	CategoryName      string         `json:"category_name"`
	IsNegotiable      bool           `json:"is_negotiable"`
	Description       string         `json:"description"`
	State             string         `json:"state"`
	AddressInState    string         `json:"address_in_state"`
	OutStandingIssues string         `json:"outstanding_issues"`
	ImageUrls         datatypes.JSON `json:"image_urls"`
	Status            Status         `json:"status" gorm:"type:varchar(20);default:'UNDER_REVIEW'"`
	Condition         string         `json:"condition"`
	UserID            *uuid.UUID     `json:"user_id"`
	University        string         `json:"university"`
	BrandName         string         `json:"brand_name"`
	User              *PublicUser    `json:"user"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-"`
	Views             int64          `json:"views"`
	Likes             int64          `json:"likes"`
	IsLikedByMe       bool           `json:"is_liked_by_me"`
	IsDeletedByUser   bool           `json:"is_deleted_by_user"`
	ProductType       string         `json:"product_type"`
	SubMenus          datatypes.JSON `json:"sub_menus"`
	DeliveryFee       float64        `json:"delivery_fee"`
	OldPrice          float64        `json:"old_price"`
	DiscountPercent   int            `json:"discount_percent"`
	GuestEmail        string         `json:"guest_email"`
	GuestPhone        string         `json:"guest_phone"`
	IsGuestListing    bool           `json:"is_guest_listing"`
	ServiceType       string         `json:"service_type"`
}
type StoreSetting struct {
	ID                uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	BusinessName      string         `json:"business_name"`
	AboutCompany      string         `json:"about_company"`
	StoreName         string         `json:"store_name"`
	Address           string         `json:"address"`
	State             string         `json:"state"`
	Region            string         `json:"regoin"`
	HowDoWeLocateYou  string         `json:"how_do_we_locate_you"`
	BusinessHoursFrom string         `json:"business_hours_from"`
	BusinessHoursTo   string         `json:"business_hours_to"`
	UserID            uuid.UUID      `json:"user_id" gorm:"type:uuid"` // needed to link with the currently authenticated user
	User              User           `json:"user" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	CreatedAt         time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
