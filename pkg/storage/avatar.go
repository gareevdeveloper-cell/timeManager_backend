package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
)

const (
	MaxAvatarSize   = 5 * 1024 * 1024 // 5 MB
	allowedAvatarCT = "image/jpeg,image/png,image/webp,image/gif"
)

// AvatarInput — данные аватарки для загрузки.
type AvatarInput struct {
	Reader      io.Reader
	Size        int64
	ContentType string
}

// FilesAPIPath — префикс для URL аватарок через API-прокси.
const FilesAPIPath = ""

// UploadAvatar загружает аватарку с валидацией в MinIO.
// Возвращает путь для хранения: users/{id}/avatar.ext, organizations/{id}/avatar.ext, teams/{id}/avatar.ext
func UploadAvatar(ctx context.Context, st Storage, pathPrefix string, entityID uuid.UUID, avatar *AvatarInput) (string, error) {
	if avatar.Size <= 0 || avatar.Size > MaxAvatarSize {
		return "", ErrInvalidAvatar
	}
	if !strings.Contains(allowedAvatarCT, avatar.ContentType) {
		return "", ErrInvalidAvatar
	}
	ext := extFromContentType(avatar.ContentType)
	path := fmt.Sprintf("%s/%s/avatar%s", pathPrefix, entityID.String(), ext)
	if _, err := st.Upload(ctx, path, avatar.Reader, avatar.Size, avatar.ContentType); err != nil {
		return "", err
	}
	return path, nil
}

// AvatarURLForResponse возвращает URL для ответа API. Если path — наш путь (users/, organizations/, teams/),
// добавляет префикс /api/v1/files. OAuth-аватары (http/https) возвращаются без изменений.
func AvatarURLForResponse(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return FilesAPIPath + "/" + path
}

func extFromContentType(ct string) string {
	switch strings.ToLower(strings.TrimSpace(ct)) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".jpg"
	}
}
