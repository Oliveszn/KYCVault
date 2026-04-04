package models

import (
	"time"

	"github.com/google/uuid"
)

type DocumentSide string

const (
	DocumentSideFront DocumentSide = "front"
	DocumentSideBack  DocumentSide = "back"
)

type DocumentStatus string

const (
	DocumentStatusPending  DocumentStatus = "pending"
	DocumentStatusAccepted DocumentStatus = "accepted"
	DocumentStatusRejected DocumentStatus = "rejected"
)

// /KYCdocument represents one captured image for front and back of an identity document
// /we dont store the raw image in the db, just the object storage key, actual file lives in s3
type KYCDocument struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SessionID uuid.UUID `gorm:"column:session_id;not null;index"               json:"session_id"`
	UserID    uuid.UUID `gorm:"column:user_id;not null;index"                  json:"user_id"` // denormalised for query convenience

	Side   DocumentSide   `gorm:"column:side;not null"                           json:"side"` // front | back
	Status DocumentStatus `gorm:"column:status;not null;default:pending"         json:"status"`

	//storagekey is the object key in s3, e.g "kyc/sessions/{session_id}/front.jpg", they shouldnt be exposed in api responses
	// StorageBucket is recorded so that if you ever migrate buckets, old records still resolve.
	StorageKey    string `gorm:"column:storage_key;not null;unique" json:"-"`
	StorageBucket string `gorm:"column:storage_bucket;not null"     json:"-"`

	// File metadata — captured at upload time.
	OriginalFilename string `gorm:"column:original_filename"          json:"original_filename"`
	MIMEType         string `gorm:"column:mime_type;not null"         json:"mime_type"` // e.g. "image/jpeg"
	FileSizeBytes    int64  `gorm:"column:file_size_bytes;not null"   json:"file_size_bytes"`
	Checksum         string `gorm:"column:checksum;not null"          json:"-"` // SHA-256 of raw file, for integrity checks

	// OCR / extraction populated after vendor processing.
	ExtractedData []byte `gorm:"column:extracted_data;type:jsonb" json:"-"` // e.g. { "id_number": "...", "dob": "..." }

	// Rejection reason if this specific document image failed validation.
	RejectionReason string `gorm:"column:rejection_reason" json:"rejection_reason,omitempty"`

	UploadedAt time.Time `gorm:"column:uploaded_at;autoCreateTime" json:"uploaded_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime"  json:"updated_at"`

	// Relations
	Session KYCSession `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE" json:"-"`
	User    User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"   json:"-"`
}

func (KYCDocument) TableName() string {
	return "kyc_documents"
}
