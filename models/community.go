package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CommunityMessage struct {
	ID          uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	SenderName  string            `gorm:"size:100;not null" json:"sender_name"`
	SenderEmail string            `gorm:"size:255" json:"sender_email"`
	UserID      *uuid.UUID        `gorm:"type:uuid" json:"user_id"`
	Message     string            `gorm:"type:text;not null" json:"message"`
	ReplyToID   *uuid.UUID        `gorm:"type:uuid" json:"reply_to_id"`
	ReplyTo     *CommunityMessage `gorm:"foreignKey:ReplyToID" json:"reply_to"`
	Reactions   datatypes.JSON    `gorm:"type:jsonb" json:"reactions"` // e.g. [{"emoji":"❤️","count":2,"users":["UserA"]}]
	CreatedAt   time.Time         `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time         `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt    `json:"-" gorm:"index"`
}
