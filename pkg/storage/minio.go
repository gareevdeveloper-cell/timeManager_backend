package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOConfig — конфигурация MinIO.
type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	UseSSL          bool
	// PublicURL — базовый URL для доступа к объектам (если отличается от Endpoint).
	// Например: https://cdn.example.com или https://minio.example.com
	PublicURL string
}

// MinIOStorage — реализация Storage для MinIO.
type MinIOStorage struct {
	client   *minio.Client
	bucket   string
	publicURL string
}

// NewMinIOStorage создаёт хранилище MinIO и создаёт bucket при необходимости.
func NewMinIOStorage(ctx context.Context, cfg MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
	}

	publicURL := cfg.PublicURL
	if publicURL == "" {
		scheme := "http"
		if cfg.UseSSL {
			scheme = "https"
		}
		publicURL = fmt.Sprintf("%s://%s/%s", scheme, cfg.Endpoint, cfg.Bucket)
	}
	publicURL = strings.TrimSuffix(publicURL, "/")

	return &MinIOStorage{
		client:    client,
		bucket:    cfg.Bucket,
		publicURL: publicURL,
	}, nil
}

// Upload загружает файл в MinIO.
func (s *MinIOStorage) Upload(ctx context.Context, path string, reader io.Reader, size int64, contentType string) (string, error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.client.PutObject(ctx, s.bucket, path, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	return s.publicURL + "/" + path, nil
}

// Delete удаляет файл из MinIO.
func (s *MinIOStorage) Delete(ctx context.Context, path string) error {
	return s.client.RemoveObject(ctx, s.bucket, path, minio.RemoveObjectOptions{})
}

// Get возвращает содержимое файла из MinIO. Caller должен закрыть reader.
func (s *MinIOStorage) Get(ctx context.Context, path string) (io.ReadCloser, *ObjectInfo, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("get object: %w", err)
	}
	info, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, nil, fmt.Errorf("stat object: %w", err)
	}
	oi := &ObjectInfo{
		ContentType: info.ContentType,
		Size:        info.Size,
	}
	if oi.ContentType == "" {
		oi.ContentType = "application/octet-stream"
	}
	return obj, oi, nil
}
