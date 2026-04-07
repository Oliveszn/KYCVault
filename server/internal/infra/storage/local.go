package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type LocalClient struct {
	basePath string
}

func NewLocalClient(basePath string) *LocalClient {
	return &LocalClient{basePath: basePath}
}

func (l *LocalClient) Put(ctx context.Context, bucket, key string, body io.Reader, size int64, contentType string) error {
	fullPath := filepath.Join(l.basePath, bucket, key)

	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	return err
}

func (l *LocalClient) GetPresignedURL(ctx context.Context, bucket, key string, ttl time.Duration) (string, error) {
	// No real signing locally — just return a static URL
	url := fmt.Sprintf("http://localhost:8000/files/%s/%s", bucket, key)
	return url, nil
}

func (l *LocalClient) Delete(ctx context.Context, bucket, key string) error {
	fullPath := filepath.Join(l.basePath, bucket, key)
	return os.Remove(fullPath)
}
