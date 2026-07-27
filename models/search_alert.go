package models

import (
	"time"

	"github.com/google/uuid"
)

type SearchAlert struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Email     string    `gorm:"not null" json:"email"`
	Keyword   string    `json:"keyword"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}
