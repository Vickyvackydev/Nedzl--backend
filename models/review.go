package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

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
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations (optional but recommended)
	Product Products `gorm:"foreignKey:ProductID" json:"-"`
	User    *User    `gorm:"foreignKey:UserID" json:"-"`
}

type ReviewResponse struct {
	CustomerReview
	ProductDetails *ProductResponse `json:"product_details"`
}
