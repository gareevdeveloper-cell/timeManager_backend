package storage

import (
	"context"
	"io"
)

// ObjectInfo — метаданные объекта при чтении.
type ObjectInfo struct {
	ContentType string
	Size        int64
}

// Storage — интерфейс хранилища файлов (MinIO, S3 и т.п.).
type Storage interface {
	// Upload загружает файл и возвращает путь/URL для доступа.
	Upload(ctx context.Context, path string, reader io.Reader, size int64, contentType string) (string, error)
	// Delete удаляет файл по пути.
	Delete(ctx context.Context, path string) error
	// Get возвращает содержимое файла по пути. Caller должен закрыть reader.
	Get(ctx context.Context, path string) (io.ReadCloser, *ObjectInfo, error)
}
