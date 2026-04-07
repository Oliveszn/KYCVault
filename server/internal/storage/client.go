package storage

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Client is the interface the document and face services depend on.
// The concrete implementation (S3, GCS, local disk for dev) lives in
// internal/infrastructure/storage/ and is injected at startup in main.go.
// Nothing in the service layer imports an AWS or GCS SDK directly.
type Client interface {
	// Put uploads raw bytes under the given key. It is the caller's
	// responsibility to choose a key that is unique and path-safe.
	Put(ctx context.Context, bucket, key string, body io.Reader, size int64, contentType string) error

	// GetPresignedURL returns a time-limited URL for direct browser download.
	// Used by the admin dashboard to view document images without routing
	// the file through your API server.
	GetPresignedURL(ctx context.Context, bucket, key string, ttl time.Duration) (string, error)

	// Delete removes an object. Used if an upload is later found to be
	// fraudulent and must be purged from storage.
	Delete(ctx context.Context, bucket, key string) error
}

// KeyForDocument builds the canonical storage key for a document image.
// Centralising this here means the service and the admin viewer always
// agree on where a file lives.
//
// Example: kyc/sessions/550e8400-e29b-41d4-a716-446655440000/front.jpg
func KeyForDocument(sessionID, docID, side, ext string) string {
	return fmt.Sprintf("kyc/sessions/%s/docs/%s_%s.%s", sessionID, docID, side, ext)
}

// KeyForSelfie builds the canonical storage key for a face-capture image.
//
// Example: kyc/sessions/550e8400-e29b-41d4-a716-446655440000/selfie.jpg
func KeyForSelfie(sessionID, verificationID string) string {
	return fmt.Sprintf("kyc/sessions/%s/selfie_%s.jpg", sessionID, verificationID)
}
