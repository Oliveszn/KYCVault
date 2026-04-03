package models

import (
	"time"

	"github.com/google/uuid"
)

// Each token is single use and rotated on every refresh cycle, revoked tokens are reatined for audit purposes and blacklist enforcemnet
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"column:user_id;not null;index"   json:"user_id"`
	TokenHash string    `gorm:"column:token_hash;not null;unique" json:"-"` // we store this in SHA-256 hash, we never store raw token
	ExpiresAt time.Time `gorm:"column:expires_at;not null"      json:"expires_at"`
	Revoked   bool      `gorm:"column:revoked;not null;default:false" json:"revoked"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	// The IP and user-agent are captured for anomaly detection / audit trails.
	IPAddress string `gorm:"column:ip_address" json:"ip_address"`
	UserAgent string `gorm:"column:user_agent" json:"user_agent"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
