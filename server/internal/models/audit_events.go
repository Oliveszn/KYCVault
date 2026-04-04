package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditEventType string

const (
	// Session lifecycle
	AuditEventSessionCreated AuditEventType = "session.created"
	AuditEventSessionExpired AuditEventType = "session.expired"

	// Status transitions
	AuditEventStatusChanged AuditEventType = "session.status_changed"

	// Document events
	AuditEventDocumentUploaded AuditEventType = "document.uploaded"
	AuditEventDocumentAccepted AuditEventType = "document.accepted"
	AuditEventDocumentRejected AuditEventType = "document.rejected"

	// Face verification events
	AuditEventFaceVerifyStarted AuditEventType = "face_verify.started"
	AuditEventFaceVerifyPassed  AuditEventType = "face_verify.passed"
	AuditEventFaceVerifyFailed  AuditEventType = "face_verify.failed"

	// Admin actions
	AuditEventManualApproval  AuditEventType = "admin.manual_approval"
	AuditEventManualRejection AuditEventType = "admin.manual_rejection"
	AuditEventReviewNoteAdded AuditEventType = "admin.review_note_added"

	// Webhook events
	AuditEventWebhookFired     AuditEventType = "webhook.fired"
	AuditEventWebhookDelivered AuditEventType = "webhook.delivered"
	AuditEventWebhookFailed    AuditEventType = "webhook.failed"
	AuditEventWebhookExhausted AuditEventType = "webhook.exhausted"

	// Auth events (supplement to refresh_tokens table)
	AuditEventLoginSuccess AuditEventType = "auth.login_success"
	AuditEventLoginFailure AuditEventType = "auth.login_failure"
	AuditEventTokenRevoked AuditEventType = "auth.token_revoked"
)

// AuditEvent is a strict append-only record. No row is ever updated or deleted.
// It forms a tamper-evident timeline of everything that happened in the system.
// Application code must NEVER issue UPDATE or DELETE against this table.
// Enforce this at the database level with a trigger if your threat model requires it.
type AuditEvent struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	// Actor who triggered this event. Nil for system-generated events (e.g. session expiry).
	ActorID   *uuid.UUID `gorm:"column:actor_id;index"   json:"actor_id,omitempty"`
	ActorRole string     `gorm:"column:actor_role"       json:"actor_role,omitempty"` // "user" | "admin" | "system"

	// Subject what entity this event is about.
	// Both can be set: e.g. an admin (actor) approving a session (subject).
	SessionID *uuid.UUID `gorm:"column:session_id;index"  json:"session_id,omitempty"`
	UserID    *uuid.UUID `gorm:"column:user_id;index"     json:"user_id,omitempty"`

	EventType AuditEventType `gorm:"column:event_type;not null;index" json:"event_type"`

	// Metadata is a free-form JSONB payload with event-specific detail.
	// Examples:
	//   session.status_changed → { "from": "doc_upload", "to": "face_verify" }
	//   face_verify.failed     → { "reason": "liveness_score_low", "score": 0.31 }
	//   webhook.fired          → { "delivery_id": "...", "endpoint_url": "..." }
	Metadata []byte `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`

	// Request context captured for every HTTP-triggered event.
	IPAddress string `gorm:"column:ip_address" json:"ip_address,omitempty"`
	UserAgent string `gorm:"column:user_agent" json:"user_agent,omitempty"`
	RequestID string `gorm:"column:request_id" json:"request_id,omitempty"` // trace/correlation ID

	// CreatedAt is the canonical event timestamp. Never autoUpdateTime.
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`

	// Relations (read-only, never used for writes)
	Actor   *User       `gorm:"foreignKey:ActorID"   json:"-"`
	Session *KYCSession `gorm:"foreignKey:SessionID" json:"-"`
	User    *User       `gorm:"foreignKey:UserID"    json:"-"`
}

func (AuditEvent) TableName() string {
	return "audit_events"
}
