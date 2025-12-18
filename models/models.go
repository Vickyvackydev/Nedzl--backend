package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Role represents the allowed user roles in the system.
// Only the declared constants are considered valid values.
//

type Role string
type Status string

const (
	StatusOngoing  Status = "ONGOING"
	StatusReview   Status = "UNDER_REVIEW"
	StatusClosed   Status = "CLOSED"
	StatusRejected Status = "REJECTED"
)
const (
	UserActive      Status = "ACTIVE"
	UserSuspended   Status = "SUSPENDED"
	UserDeactivated Status = "DEACTIVATED"
)

const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
)

func IsValidRole(r Role) bool {
	switch r {
	case RoleAdmin, RoleUser:
		return true

	default:
		return false

	}

}
func IsValidStatus(s Status) bool {
	switch s {
	case StatusOngoing, StatusRejected, StatusReview, StatusClosed:
		return true

	default:
		return false

	}

}
func IsValidUserStatus(s Status) bool {
	switch s {
	case UserActive, UserDeactivated, UserSuspended:
		return true

	default:
		return false

	}

}

type RegisterRequest struct {
	UserName    string `json:"user_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Role        Role   `json:"role"`
	Password    string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PublicUser struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	UserName    string         `json:"user_name"`
	Email       string         `json:"email"`
	Role        string         `json:"role"`
	PhoneNumber string         `json:"phone_number"`
	ImageUrl    string         `json:"image_url"`
	Location    string         `json:"location"`
	Status      Status         `json:"status" gorm:"type:varchar(20);default:'ACTIVE'"`
	CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type User struct {
	ID            uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	UserName      string         `json:"user_name"`
	Email         string         `json:"email"`
	PhoneNumber   string         `json:"phone_number"`
	Role          Role           `json:"role"`
	Password      string         `json:"password"`
	ImageUrl      string         `json:"image_url"`
	Location      string         `json:"location"`
	EmailVerified bool           `gorm:"default:false" json:"email_verified"`
	IsVerified    bool           `gorm:"default:false" json:"is_verified"`
	EmailToken    string         `gorm:"size:255" json:"email_token"`
	Status        Status         `json:"status" gorm:"type:varchar(20);default:'ACTIVE'"`
	CreatedAt     time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type Products struct {
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
	Condition         string         `json:"condition"`
	BrandName         string         `json:"brand_name"`
	ImageUrls         datatypes.JSON `json:"image_urls"`
	NewImages         datatypes.JSON `json:"new_images"`
	Status            Status         `json:"status" gorm:"type:varchar(20);default:'ONGOING'"`
	UserID            uuid.UUID      `json:"user_id" gorm:"type:uuid"`
	User              User           `json:"user" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	CreatedAt         time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	ClosedAt          *time.Time     `gorm:"index"`
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
	UserID            uuid.UUID      `json:"user_id"`
	BrandName         string         `json:"brand_name"`
	User              PublicUser     `json:"user"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at"`
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

type Suggestion struct {
	Type      string `json:"type"` // "keyword", "category", "brand"
	Text      string `json:"text"`
	Category  string `json:"category,omitempty"`
	Brand     string `json:"brand,omitempty"`
	Count     int    `json:"count,omitempty"`
	ProductID string `json:"product_id,omitempty"`
}

type DashboardResponse struct {
	Stats   DashboardStats   `json:"stats"`
	Growth  DashboardGrowth  `json:"growth"`
	Metrics DashboardMetrics `json:"metrics"`
}

type DashboardStats struct {
	TotalProductsListed     int64 `json:"total_product_listed"`
	ActiveProducts          int64 `json:"active_products"`
	ClosedSoldProducts      int64 `json:"closed_sold_products"`
	FlaggedReportedProducts int64 `json:"flagged_reported_products"`
	TotalRegisteredSellers  int64 `json:"total_registered_sellers"`
}
type DashboardGrowth struct {
	TotalProductsListed     float64 `json:"total_product_listed"`
	ActiveProducts          float64 `json:"active_products"`
	ClosedSoldProducts      float64 `json:"closed_sold_products"`
	FlaggedReportedProducts float64 `json:"flagged_reported_products"`
	TotalRegisteredSellers  float64 `json:"total_registered_sellers"`
}

type MonthlyMetric struct {
	Month string `json:"month"`
	Value int64  `json:"value"`
}

type DashboardMetrics struct {
	CustomerSignUpMetrics []MonthlyMetric `json:"customer_signup_metrics"`
	TotalSoldProducts     []MonthlyMetric `json:"total_sold_products"`
}

type UserDashboardResponse struct {
	Stats  UserDashboardStats  `json:"user_stats"`
	Growth UserDashboardGrowth `json:"growth"`
}

type UserDashboardStats struct {
	TotalSellers     int64 `json:"total_sellers"`
	ActiveSellers    int64 `json:"active_sellers"`
	SuspendedUsers   int64 `json:"suspended_users"`
	DeactivatedUsers int64 `json:"deactivated_users"`
}

type UserDashboardGrowth struct {
	TotalSellers     float64 `json:"total_sellers"`
	ActiveSellers    float64 `json:"active_sellers"`
	SuspendedUsers   float64 `json:"suspended_users"`
	DeactivatedUsers float64 `json:"deactivated_users"`
}

type UserDashboardUsers struct {
	User           PublicUser `json:"user"`
	ListedProducts int64      `json:"listed_products"`
	SoldProducts   int64      `json:"sold_products"`
}

type UserProductStats struct {
	TotalProductsListed int64 `json:"total_products_listed"`
	ActiveProducts      int64 `json:"active_products"`
	SoldProducts        int64 `json:"sold_products"`
	FlaggedProducts     int64 `json:"flagged_products"`
}

type UserStoreDetails struct {
	ID                uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	BusinessName      string    `json:"business_name"`
	AboutCompany      string    `json:"about_company"`
	StoreName         string    `json:"store_name"`
	Address           string    `json:"address"`
	State             string    `json:"state"`
	Region            string    `json:"regoin"`
	HowDoWeLocateYou  string    `json:"how_do_we_locate_you"`
	BusinessHoursFrom string    `json:"business_hours_from"`
	BusinessHoursTo   string    `json:"business_hours_to"`
	UserID            uuid.UUID `json:"user_id" gorm:"type:uuid"` // needed to link with the currently authenticated user

	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
type UserDetailsResponse struct {
	UserDetail   PublicUser        `json:"user_details"`
	Metrics      UserProductStats  `json:"metrics"`
	StoreDetails *UserStoreDetails `json:"store_details"`
}

type FeaturedSection struct {
	ID           uuid.UUID                `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BoxNumber    int                      `gorm:"unique" json:"box_number"`
	CategoryName string                   `json:"category_name"`
	Description  string                   `json:"description"`
	Products     []FeaturedSectionProduct `gorm:"foreignKey:FeaturedSectionID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type FeaturedSectionProduct struct {
	ID                uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	FeaturedSectionID uuid.UUID `gorm:"type:uuid;index;not null"`
	ProductID         uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt         time.Time

	FeaturedSection FeaturedSection `gorm:"constraint:OnDelete:CASCADE;"`
}

type CustomerReview struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Experience   string         `json:"experience"`
	UserID       *uuid.UUID     `gorm:"type:uuid;index" json:"user_id"`
	ProductID    uuid.UUID      `gorm:"type:uuid;index;not null" json:"product_id"`
	ReviewTitle  string         `json:"review_title"`
	CustomerName string         `json:"customer_name"`
	Review       string         `json:"review"`
	Images       datatypes.JSON `json:"images"`
	IsPublic     bool           `json:"is_public" gorm:"default:true"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relations (optional but recommended)
	Product Products `gorm:"foreignKey:ProductID" json:"-"`
	User    *User    `gorm:"foreignKey:UserID" json:"-"`
}

type ReviewResponse struct {
	CustomerReview
	ProductDetails *ProductResponse `json:"product_details"`
}
