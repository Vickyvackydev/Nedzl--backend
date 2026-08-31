package models

import (
	"time"
)

type BulkEmailQueueItem struct {
	ID             uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	RecipientEmail string     `gorm:"type:varchar(255);not null;index" json:"recipient_email"`
	RecipientName  string     `gorm:"type:varchar(255)" json:"recipient_name"`
	Subject        string     `gorm:"type:varchar(255);not null" json:"subject"`
	HTMLContent    string     `gorm:"type:text;not null" json:"html_content"`
	Status         string     `gorm:"type:varchar(50);default:'PENDING';index" json:"status"` // PENDING, SENT, FAILED
	ErrorMessage   string     `gorm:"type:text" json:"error_message,omitempty"`
	SentAt         *time.Time `json:"sent_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
