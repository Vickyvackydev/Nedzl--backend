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
	StatusRejected Status = "REJCTED"
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
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at"`
}

type User struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	UserName    string         `json:"user_name"`
	Email       string         `json:"email"`
	PhoneNumber string         `json:"phone_number"`
	Role        Role           `json:"role"`
	Password    string         `json:"password"`
	ImageUrl    string         `json:"image_url"`
	Location    string         `json:"location"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
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
	Status            Status         `json:"status" gorm:"type:varchar(20);default:'UNDER_REVIEW'"`
	UserID            uuid.UUID      `json:"user_id" gorm:"type:uuid"`
	User              User           `json:"user" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	CreatedAt         time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" gorm:"index"`
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
