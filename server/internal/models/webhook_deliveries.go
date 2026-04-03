package models

import (
	"time"

	"github.com/google/uuid"
)

type WebhookDeliveryStatus string

const (
	WebhookDeliveryStatusPending   WebhookDeliveryStatus = "pending"
	WebhookDeliveryStatusSuccess   WebhookDeliveryStatus = "success"
	WebhookDeliveryStatusFailed    WebhookDeliveryStatus = "failed"
	WebhookDeliveryStatusExhausted WebhookDeliveryStatus = "exhausted" // all retries consumed
)

// WebhookDelivery is an append-only log one row per delivery attempt.
// Never update an existing row. On retry, insert a new row with attempt_number + 1.
// This gives you a full audit trail of every HTTP call ever made.
type WebhookDelivery struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SessionID  uuid.UUID `gorm:"column:session_id;not null;index"   json:"session_id"`  // the KYC session that triggered this
	EndpointID uuid.UUID `gorm:"column:endpoint_id;not null;index"  json:"endpoint_id"` // which registered endpoint was called
	UserID     uuid.UUID `gorm:"column:user_id;not null;index"      json:"user_id"`     // denormalised for admin queries

	// EventType identifies the event, e.g. "kyc.approved", "kyc.rejected".
	EventType string `gorm:"column:event_type;not null" json:"event_type"`

	// Payload is the exact JSON body that was (or will be) sent.
	// Stored so that failed deliveries can be re-triggered without recomputing.
	Payload     []byte `gorm:"column:payload;type:jsonb;not null" json:"payload"`
	PayloadHash string `gorm:"column:payload_hash;not null"      json:"-"` // SHA-256 of payload bytes

	// Signature is the X-KYC-Signature header value we sent, for debugging verification failures.
	Signature string `gorm:"column:signature;not null" json:"-"`

	// IdempotencyKey is included in the payload headers so customers can deduplicate retries.
	// It is stable across all retry attempts for the same logical event.
	IdempotencyKey string `gorm:"column:idempotency_key;not null;index" json:"idempotency_key"`

	// Delivery outcome.
	Status         WebhookDeliveryStatus `gorm:"column:status;not null;default:pending" json:"status"`
	AttemptNumber  int                   `gorm:"column:attempt_number;not null;default:1" json:"attempt_number"`
	MaxAttempts    int                   `gorm:"column:max_attempts;not null;default:5"   json:"max_attempts"`
	HTTPStatusCode *int                  `gorm:"column:http_status_code"                  json:"http_status_code,omitempty"` // nil if network error
	ResponseBody   string                `gorm:"column:response_body"                     json:"response_body,omitempty"`    // first 2000 chars
	ErrorMessage   string                `gorm:"column:error_message"                     json:"error_message,omitempty"`    // transport-level error

	// Timing.
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`              // when first scheduled
	AttemptedAt *time.Time `gorm:"column:attempted_at"              json:"attempted_at,omitempty"`  // when this attempt fired
	DeliveredAt *time.Time `gorm:"column:delivered_at"              json:"delivered_at,omitempty"`  // set on success
	NextRetryAt *time.Time `gorm:"column:next_retry_at;index"       json:"next_retry_at,omitempty"` // nil if success or exhausted

	// Relations
	Session  KYCSession      `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE"  json:"-"`
	Endpoint WebhookEndpoint `gorm:"foreignKey:EndpointID;constraint:OnDelete:CASCADE" json:"-"`
	User     User            `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"     json:"-"`
}

func (WebhookDelivery) TableName() string {
	return "webhook_deliveries"
}

// Indexes:
//
//   -- Retry worker polls this to find jobs ready to fire.
//   CREATE INDEX idx_webhook_deliveries_retry_queue
//     ON webhook_deliveries (next_retry_at)
//     WHERE status = 'pending' AND next_retry_at IS NOT NULL;
//
//   -- Admin dashboard: all deliveries for a session.
//   CREATE INDEX idx_webhook_deliveries_session ON webhook_deliveries (session_id, created_at DESC);
//
//   -- Idempotency dedup check.
//   CREATE INDEX idx_webhook_deliveries_idempotency ON webhook_deliveries (idempotency_key);
