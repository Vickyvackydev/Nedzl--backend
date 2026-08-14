package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BankAccountItem struct {
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	AccountName   string `json:"account_name"`
	IsDefault     bool   `json:"is_default"`
}

type PublicUser struct {
	ID            uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	UserName      string         `json:"user_name"`
	Email         string         `json:"email"`
	Role          string         `json:"role"`
	PhoneNumber   string         `json:"phone_number"`
	ImageUrl      string         `json:"image_url"`
	Location      string         `json:"location"`
	BankName      string         `json:"bank_name"`
	AccountNumber string         `json:"account_number"`
	AccountName   string         `json:"account_name"`
	BankAccounts  datatypes.JSON `gorm:"type:jsonb" json:"bank_accounts"`
	ReferralCode  string         `gorm:"uniqueIndex" json:"referral_code"`
	ReferralBy    *ReferedBy     `gorm:"jsonb" json:"referral_by"`
	ReferralCount int64          `json:"referral_count"`
	Status        Status         `json:"status" gorm:"type:varchar(20);default:'ACTIVE'"`
	IsVerified    bool           `gorm:"default:false" json:"is_verified"`
	StudentIDCard string         `json:"student_id_card"`
	CreatedAt     time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

type User struct {
	ID                       uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	UserName                 string         `json:"user_name"`
	Email                    string         `json:"email"`
	PhoneNumber              string         `json:"phone_number"`
	Role                     Role           `json:"role"`
	Password                 string         `json:"-"`
	ImageUrl                 string         `json:"image_url"`
	Location                 string         `json:"location"`
	BankName                 string         `json:"bank_name"`
	AccountNumber            string         `json:"account_number"`
	AccountName              string         `json:"account_name"`
	BankAccounts             datatypes.JSON `gorm:"type:jsonb" json:"bank_accounts"`
	ReferralCode             string     `gorm:"uniqueIndex" json:"referral_code"`
	ReferralBy               *ReferedBy `gorm:"type:jsonb" json:"referral_by"`
	ReferralCount            int64      `json:"referral_count"`
	EmailVerified            bool       `gorm:"default:false" json:"email_verified"`
	IsVerified               bool       `gorm:"default:false" json:"is_verified"`
	EmailToken               string     `gorm:"size:255" json:"-"`
	EmailTokenExpiry         *time.Time `json:"-"`
	Status                   Status     `json:"status" gorm:"type:varchar(20);default:'ACTIVE'"`
	FailedLoginAttempts      int        `gorm:"default:0" json:"-"`
	PasswordResetToken       string     `gorm:"size:255" json:"-"`
	PasswordResetTokenExpiry *time.Time `json:"-"`
	StudentIDCard            string         `json:"student_id_card"`
	CreatedAt                time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt                time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt                gorm.DeletedAt `json:"-" gorm:"index"`
}

type UserDetailsResponse struct {
	UserDetail   PublicUser        `json:"user_details"`
	Metrics      UserProductStats  `json:"metrics"`
	StoreDetails *UserStoreDetails `json:"store_details"`
}
