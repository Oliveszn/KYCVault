package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	// "time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

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

//	func (s *S3Client) Put(ctx context.Context, bucket, key string, body io.Reader, size int64, contentType string) error {
//		_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
//			Bucket:        &bucket,
//			Key:           &key,
//			Body:          body,
//			ContentType:   &contentType,
//			ContentLength: &size,
//		})
//		return err
//	}
//
// Upload implements storage.Client
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
	// For now, simple public URL (later we can add real presigning)
	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key)
	return url, nil
}

// func (s *S3Client) GetPresignedURL(ctx context.Context, bucket, key string, ttl time.Duration) string {
// 	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key)
// 	return url, nil
// }

//	func (s *S3Client) Delete(ctx context.Context, bucket, key string) error {
//		_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
//			Bucket: &bucket,
//			Key:    &key,
//		})
//		return err
//	}
//
// Delete implements storage.Client
func (s *S3Client) Delete(ctx context.Context, bucket string, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.Bucket,
		Key:    &key,
	})
	return err
}
