package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
