package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PublicUser struct {
	ID            uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	UserName      string         `json:"user_name"`
	Email         string         `json:"email"`
	Role          string         `json:"role"`
	PhoneNumber   string         `json:"phone_number"`
	ImageUrl      string         `json:"image_url"`
	Location      string         `json:"location"`
	ReferralCode  string         `gorm:"uniqueIndex" json:"referral_code"`
	ReferralBy    *ReferedBy     `gorm:"jsonb" json:"referral_by"`
	ReferralCount int64          `json:"referral_count"`
	Status        Status         `json:"status" gorm:"type:varchar(20);default:'ACTIVE'"`
	IsVerified    bool           `gorm:"default:false" json:"is_verified"`
	CreatedAt     time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

type User struct {
	ID                       uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	UserName                 string     `json:"user_name"`
	Email                    string     `json:"email"`
	PhoneNumber              string     `json:"phone_number"`
	Role                     Role       `json:"role"`
	Password                 string     `json:"password"`
	ImageUrl                 string     `json:"image_url"`
	Location                 string     `json:"location"`
	ReferralCode             string     `gorm:"uniqueIndex" json:"referral_code"`
	ReferralBy               *ReferedBy `gorm:"type:jsonb" json:"referral_by"`
	ReferralCount            int64      `json:"referral_count"`
	EmailVerified            bool       `gorm:"default:false" json:"email_verified"`
	IsVerified               bool       `gorm:"default:false" json:"is_verified"`
	EmailToken               string     `gorm:"size:255" json:"email_token"`
	EmailTokenExpiry         *time.Time `json:"email_token_expiry"`
	Status                   Status     `json:"status" gorm:"type:varchar(20);default:'ACTIVE'"`
	PasswordResetToken       string     `gorm:"size:255"`
	PasswordResetTokenExpiry *time.Time
	CreatedAt                time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt                time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt                gorm.DeletedAt `json:"-" gorm:"index"`
}

type UserDetailsResponse struct {
	UserDetail   PublicUser        `json:"user_details"`
	Metrics      UserProductStats  `json:"metrics"`
	StoreDetails *UserStoreDetails `json:"store_details"`
}
