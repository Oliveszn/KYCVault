package models

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID  `gorm:"column:user_id;not null;index"`
	SessionID *uuid.UUID `gorm:"column:session_id;index"`

	Type    string `gorm:"column:type;not null"`    // "kyc_approved", "kyc_rejected"
	Message string `gorm:"column:message;not null"` // "Your identity has been verified"
	Read    bool   `gorm:"column:read;default:false"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (Notification) TableName() string {
	return "notifications"
}
