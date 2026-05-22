package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Banner struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ImageUrl  string         `json:"image_url"`
	TargetUrl string         `json:"target_url"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
