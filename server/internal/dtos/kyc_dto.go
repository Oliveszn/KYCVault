package dtos

import (
	"time"

	"github.com/google/uuid"
)

type InitiateSessionRequest struct {
	// Country is the ISO 3166-1 alpha-2 country code, e.g. "NG", "GH", "KE".
	Country string `json:"country" binding:"required,len=2"`

	// IDType must match one of the models.IDType constants.
	IDType string `json:"id_type" binding:"required,oneof=national_id drivers_license passport residence_permit"`
}

type ReviewFaceRequest struct {
	Passed bool   `json:"passed"`
	Note   string `json:"note"`
}

type ReviewSessionRequest struct {
	// Note is an internal-only note visible only to admins in the audit log.
	Note string `json:"note" binding:"max=1000"`
}

type RejectSessionRequest struct {
	// Note is internal-only.
	Note string `json:"note" binding:"max=1000"`

	// Reason is user-facing it is sent in the rejection email and webhook payload.
	Reason string `json:"reason" binding:"required,min=10,max=500"`
}

// SessionResponse is the API representation of a KYCSession.
// It deliberately omits internal fields (vendor_raw_result, storage keys, etc.).
type SessionResponse struct {
	ID              uuid.UUID         `json:"id"`
	Status          string            `json:"status"`
	Country         string            `json:"country"`
	IDType          string            `json:"id_type"`
	AttemptNumber   int               `json:"attempt_number"`
	RejectionReason string            `json:"rejection_reason,omitempty"`
	ExpiresAt       *time.Time        `json:"expires_at,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	Documents       []DocumentSummary `json:"documents,omitempty"`
}

// DocumentSummary is the minimal document info embedded inside SessionResponse.
// The full DocumentResponse (with MIME type, size, etc.) lives in document_dto.go.
type DocumentSummary struct {
	ID     uuid.UUID `json:"id"`
	Side   string    `json:"side"`
	Status string    `json:"status"`
}
