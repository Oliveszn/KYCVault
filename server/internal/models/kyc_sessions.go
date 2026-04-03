package models

import (
	"time"

	"github.com/google/uuid"
)

type KYCStatus string

const (
	KYCStatusInitiated  KYCStatus = "initiated"
	KYCStatusDocUpload  KYCStatus = "doc_upload"
	KYCStatusFaceVerify KYCStatus = "face_verify"
	KYCStatusInReview   KYCStatus = "in_review"
	KYCStatusApproved   KYCStatus = "approved"
	KYCStatusRejected   KYCStatus = "rejected"
)

type IDType string

const (
	IDTypeNationalID      IDType = "national_id"
	IDTypeDriversLicense  IDType = "drivers_license"
	IDTypePassport        IDType = "passport"
	IDTypeResidencePermit IDType = "residence_permit"
)

// KYCsessin is the root record for a single verification attmept
// a user can have multiple sessions that is after a rejection or re-submission
// but only one session may be in a non-terminal state a time
// we'll enforce this via the partail unique index below
type KYCSession struct {
	ID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID  uuid.UUID `gorm:"column:user_id;not null;index"                   json:"user_id"`
	Status  KYCStatus `gorm:"column:status;not null;default:initiated"        json:"status"`
	Country string    `gorm:"column:country;not null"                         json:"country"` // ISO 3166-1 alpha-2, e.g. "NG"
	IDType  IDType    `gorm:"column:id_type;not null"                         json:"id_type"`

	////Vendor fields populated once the external verification call completes.
	VendorName      string `gorm:"column:vendor_name"       json:"vendor_name"`       // e.g. "smile_identity", "onfido"
	VendorSessionID string `gorm:"column:vendor_session_id" json:"vendor_session_id"` // reference ID from vendor
	VendorRawResult []byte `gorm:"column:vendor_raw_result;type:jsonb" json:"-"`      // full vendor response, stored for audit

	// Review fields populated by an admin when status moves to approved/rejected.
	ReviewerID *uuid.UUID `gorm:"column:reviewer_id;index"  json:"reviewer_id,omitempty"`
	ReviewNote string     `gorm:"column:review_note"        json:"review_note,omitempty"`
	ReviewedAt *time.Time `gorm:"column:reviewed_at"        json:"reviewed_at,omitempty"`

	// Rejection reason by admin human-readable, will also be sent in the webhook payload.
	RejectionReason string `gorm:"column:rejection_reason" json:"rejection_reason,omitempty"`

	// Attempt tracking incremented on each re-submission.
	AttemptNumber int `gorm:"column:attempt_number;not null;default:1" json:"attempt_number"`

	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	ExpiresAt *time.Time `gorm:"column:expires_at"                json:"expires_at"` // session TTL, nil = no expiry

	User      User          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"     json:"-"`
	Reviewer  *User         `gorm:"foreignKey:ReviewerID"                             json:"-"`
	Documents []KYCDocument `gorm:"foreignKey:SessionID"                              json:"documents,omitempty"`
}

func (KYCSession) TableName() string {
	return "kyc_sessions"
}

// Indexes to add via AutoMigrate or raw SQL:
//
//   -- Only one active (non-terminal) session per user at a time.
//   CREATE UNIQUE INDEX idx_kyc_sessions_one_active_per_user
//     ON kyc_sessions (user_id)
//     WHERE status NOT IN ('approved', 'rejected');
//
//   -- Fast lookups by vendor reference (for webhook callbacks from vendor).
//   CREATE INDEX idx_kyc_sessions_vendor_session_id ON kyc_sessions (vendor_session_id)
//     WHERE vendor_session_id IS NOT NULL;
//
//   -- Status filtering for admin dashboard queue.
//   CREATE INDEX idx_kyc_sessions_status ON kyc_sessions (status);
