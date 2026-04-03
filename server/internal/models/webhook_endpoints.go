package models

import (
	"time"

	"github.com/google/uuid"
)

// WebhookEndpoint is a customer-registered URL that receives KYC event notifications.
// One customer (user with role="customer" or an org entity) can register multiple endpoints,
// e.g. one for production and one for staging.

// webhookendpoint is a customer registered url that receives kyc event notifications
// one customer user with role=customer/user can register multiple ennpoints, one for staging and one for production
type WebhookEndpoint struct {
	ID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID uuid.UUID `gorm:"column:user_id;not null;index" json:"user_id"` // the customer who owns this endpoint

	EndpointURL string `gorm:"column:endpoint_url;not null" json:"endpoint_url"` // must be HTTPS in production

	// SecretHash is a SHA-256 hash of the HMAC signing secret.
	// We never store the raw secret — it is shown to the customer only once at creation.
	// On each delivery, the worker fetches this hash and uses the raw secret from an
	// in-memory secrets manager (e.g. AWS Secrets Manager / Vault).
	SecretHash string `gorm:"column:secret_hash;not null" json:"-"`

	// EventTypes is a JSONB array of event strings this endpoint subscribes to.
	// e.g. ["kyc.approved", "kyc.rejected"]
	// An empty array or null means "subscribe to all events".
	EventTypes []byte `gorm:"column:event_types;type:jsonb" json:"event_types"`

	IsActive bool   `gorm:"column:is_active;not null;default:true" json:"is_active"`
	Label    string `gorm:"column:label"                           json:"label"` // human-readable name, e.g. "Production"

	// Timeout for delivery attempts, in seconds. Defaults to 10.
	TimeoutSeconds int `gorm:"column:timeout_seconds;not null;default:10" json:"timeout_seconds"`

	// MaxRetries caps how many times we attempt delivery. Defaults to 5.
	MaxRetries int `gorm:"column:max_retries;not null;default:5" json:"max_retries"`

	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index"          json:"deleted_at,omitempty"` // soft delete

	// Relations
	User       User              `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Deliveries []WebhookDelivery `gorm:"foreignKey:EndpointID"                        json:"deliveries,omitempty"`
}

func (WebhookEndpoint) TableName() string {
	return "webhook_endpoints"
}

// Indexes:
//
//   -- Active endpoints only, for dispatcher lookups.
//   CREATE INDEX idx_webhook_endpoints_active ON webhook_endpoints (user_id)
//     WHERE is_active = true AND deleted_at IS NULL;
