package models

import (
	"time"

	"github.com/google/uuid"
)

type FaceVerificationStatus string

const (
	FaceVerificationStatusPending FaceVerificationStatus = "pending"
	FaceVerificationStatusPassed  FaceVerificationStatus = "passed"
	FaceVerificationStatusFailed  FaceVerificationStatus = "failed"
)

// Faceverification recrds the result of the liveness check and face match against the submitted document. one record per seesion
// if the user retries face capture, the existing record is updated not duplicated, cos only the latest result matters and
// attempt history lives in auditevent
type FaceVerification struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SessionID uuid.UUID `gorm:"column:session_id;not null;uniqueIndex" json:"session_id"` // one-to-one with session
	UserID    uuid.UUID `gorm:"column:user_id;not null;index"           json:"user_id"`

	Status FaceVerificationStatus `gorm:"column:status;not null;default:pending" json:"status"`

	// Liveness, did the vendor confirm this is a real person, not a photo/video?
	LivenessScore  float64 `gorm:"column:liveness_score"    json:"liveness_score"` // 0.0–1.0
	LivenessPassed bool    `gorm:"column:liveness_passed"   json:"liveness_passed"`

	// Face match, how closely does the selfie match the document photo?
	MatchScore     float64 `gorm:"column:match_score"       json:"match_score"`     // 0.0–1.0
	MatchThreshold float64 `gorm:"column:match_threshold"   json:"match_threshold"` // threshold used at time of check
	MatchPassed    bool    `gorm:"column:match_passed"      json:"match_passed"`

	// Storage key for the selfie image, same pattern as KYCDocument.
	SelfieStorageKey    string `gorm:"column:selfie_storage_key;not null" json:"-"`
	SelfieStorageBucket string `gorm:"column:selfie_storage_bucket;not null" json:"-"`
	SelfieChecksum      string `gorm:"column:selfie_checksum;not null"    json:"-"`

	// Vendor fields.
	VendorName      string `gorm:"column:vendor_name"       json:"vendor_name"`
	VendorRequestID string `gorm:"column:vendor_request_id" json:"vendor_request_id"`
	VendorRawResult []byte `gorm:"column:vendor_raw_result;type:jsonb" json:"-"`

	// Attempt tracking, how many times the user has tried face capture in this session.
	AttemptCount int `gorm:"column:attempt_count;not null;default:1" json:"attempt_count"`

	FailureReason string `gorm:"column:failure_reason" json:"failure_reason,omitempty"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relations
	Session KYCSession `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE" json:"-"`
	User    User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"   json:"-"`
}

func (FaceVerification) TableName() string {
	return "face_verifications"
}
