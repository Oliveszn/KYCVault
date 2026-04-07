package storage

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Client is the interface the document and face services depend on.
type Client interface {
	// Put uploads raw bytes under the given key
	Put(ctx context.Context, bucket, key string, body io.Reader, size int64, contentType string) error

	// GetPresignedURL returns a time-limited URL for direct browser download Used by the admin dashboard to view document images without routing the file
	GetPresignedURL(ctx context.Context, bucket, key string, ttl time.Duration) (string, error)

	// Delete removes an object
	Delete(ctx context.Context, bucket, key string) error
}

// KeyForDocument builds the storage key for a document image.
//
// Example: kyc/sessions/550e8400-e29b-41d4-a716-446655440000/front.jpg
func KeyForDocument(sessionID, docID, side, ext string) string {
	return fmt.Sprintf("kyc/sessions/%s/docs/%s_%s.%s", sessionID, docID, side, ext)
}

// KeyForSelfie builds the storage key for a face-capture image.
//
// Example: kyc/sessions/550e8400-e29b-41d4-a716-446655440000/selfie.jpg
func KeyForSelfie(sessionID, verificationID string) string {
	return fmt.Sprintf("kyc/sessions/%s/selfie_%s.jpg", sessionID, verificationID)
}
