package models

import (
	"time"

	"github.com/google/uuid"

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

type ReferedBy struct {
	ID       uuid.UUID `json:"id"`
	UserName string    `json:"user_name"`
	Email    string    `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type RegisterRequest struct {
	UserName    string `json:"user_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Role        Role   `json:"role"`
	Password    string `json:"password"`
	ReferalCode string `json:"referral_code"`
}

type Suggestion struct {
	Type       string `json:"type"` // "keyword", "category", "brand"
	Text       string `json:"text"`
	Category   string `json:"category,omitempty"`
	Brand      string `json:"brand,omitempty"`
	University string `json:"university,omitempty"`
	Count      int    `json:"count,omitempty"`
	ProductID  string `json:"product_id,omitempty"`
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

type Contact struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	FirstName   string         `json:"first_name"`
	LastName    string         `json:"last_name"`
	Email       string         `json:"email"`
	PhoneNumber string         `json:"phone_number"`
	Message     string         `json:"message"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}
