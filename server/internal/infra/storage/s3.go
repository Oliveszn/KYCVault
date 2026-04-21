package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

type S3Client struct {
	client *s3.Client
	Bucket string
}

func NewS3Client(client *s3.Client, bucket string) *S3Client {
	return &S3Client{
		client: client,
		Bucket: bucket,
	}
}

func (s *S3Client) Put(ctx context.Context, bucket, key string, body io.Reader, size int64, contentType string) error {
	uploader := manager.NewUploader(s.client)
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:        &bucket,
		Key:           &key,
		Body:          body,
		ContentLength: &size,
		ContentType:   &contentType,
	})
	return err
}

func (s *S3Client) GetPresignedURL(ctx context.Context, bucket, key string, ttl time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(ttl))

	if err != nil {
		return "", fmt.Errorf("failed to presign url: %w", err)
	}

	return req.URL, nil
}

// Delete implements storage.Client
func (s *S3Client) Delete(ctx context.Context, bucket string, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.Bucket,
		Key:    &key,
	})
	return err
}

// KeyForSelfie builds the storage key for a face-capture image.
//
// Example: kyc/sessions/550e8400-e29b-41d4-a716-446655440000/selfie.jpg
func KeyForSelfie(sessionID, verificationID string) string {
	return fmt.Sprintf("kyc/sessions/%s/selfie_%s.jpg", sessionID, verificationID)
}

// KeyForDocument builds the storage key for a document image.
//
// Example: kyc/sessions/550e8400-e29b-41d4-a716-446655440000/front.jpg
func KeyForDocument(sessionID, docID, side, ext string) string {
	return fmt.Sprintf("kyc/sessions/%s/docs/%s_%s.%s", sessionID, docID, side, ext)
}
