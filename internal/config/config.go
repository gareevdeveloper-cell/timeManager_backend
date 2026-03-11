package config

import (
	"fmt"
	"strconv"
	"time"

	"testik/pkg/env"
)

// Config — конфигурация приложения.
type Config struct {
	HTTPPort    int
	DatabaseURL string
	JWTSecret   string
	JWTExpires  time.Duration

	// MinIO — хранилище файлов (аватарки и т.п.)
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool
	MinIOPublicURL string

	// OAuth — авторизация через Google, Yandex, GitHub
	OAuthRedirectBase      string // базовый URL для callback (например https://api.example.com)
	OAuthFrontendRedirect   string // URL фронтенда для редиректа с токеном (?token=...)
	OAuthGoogleID          string
	OAuthGoogleSecret      string
	OAuthYandexID          string
	OAuthYandexSecret      string
	OAuthGitHubID          string
	OAuthGitHubSecret      string
}

// Load загружает конфигурацию из переменных окружения.
func Load() (*Config, error) {
	port, _ := strconv.Atoi(env.Get("HTTP_PORT", "8080"))
	expiresMin, _ := strconv.Atoi(env.Get("JWT_EXPIRES_MINUTES", "60"))

	dbURL := env.Get("DATABASE_URL", "")
	if dbURL == "" {
		dbURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			env.Get("DB_USER", "postgres"),
			env.Get("DB_PASSWORD", "postgres"),
			env.Get("DB_HOST", "localhost"),
			env.Get("DB_PORT", "5432"),
			env.Get("DB_NAME", "testik"),
		)
	}

	jwtSecret := env.Get("JWT_SECRET", "")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	minioEndpoint := env.Get("MINIO_ENDPOINT", "localhost:9000")
	minioAccessKey := env.Get("MINIO_ACCESS_KEY", "minioadmin")
	minioSecretKey := env.Get("MINIO_SECRET_KEY", "minioadmin")
	minioBucket := env.Get("MINIO_BUCKET", "testik")
	minioUseSSL := env.Get("MINIO_USE_SSL", "false") == "true"
	minioPublicURL := env.Get("MINIO_PUBLIC_URL", "")

	oauthRedirectBase := env.Get("OAUTH_REDIRECT_BASE", "http://localhost:8080")
	oauthFrontendRedirect := env.Get("OAUTH_FRONTEND_REDIRECT", "http://localhost:3000/auth/callback")
	oauthGoogleID := env.Get("OAUTH_GOOGLE_CLIENT_ID", "")
	oauthGoogleSecret := env.Get("OAUTH_GOOGLE_CLIENT_SECRET", "")
	oauthYandexID := env.Get("OAUTH_YANDEX_CLIENT_ID", "")
	oauthYandexSecret := env.Get("OAUTH_YANDEX_CLIENT_SECRET", "")
	oauthGitHubID := env.Get("OAUTH_GITHUB_CLIENT_ID", "")
	oauthGitHubSecret := env.Get("OAUTH_GITHUB_CLIENT_SECRET", "")

	return &Config{
		HTTPPort:    port,
		DatabaseURL: dbURL,
		JWTSecret:   jwtSecret,
		JWTExpires:  time.Duration(expiresMin) * time.Minute,

		MinIOEndpoint:  minioEndpoint,
		MinIOAccessKey: minioAccessKey,
		MinIOSecretKey: minioSecretKey,
		MinIOBucket:    minioBucket,
		MinIOUseSSL:    minioUseSSL,
		MinIOPublicURL: minioPublicURL,

		OAuthRedirectBase:     oauthRedirectBase,
		OAuthFrontendRedirect:  oauthFrontendRedirect,
		OAuthGoogleID:      oauthGoogleID,
		OAuthGoogleSecret:  oauthGoogleSecret,
		OAuthYandexID:      oauthYandexID,
		OAuthYandexSecret:  oauthYandexSecret,
		OAuthGitHubID:      oauthGitHubID,
		OAuthGitHubSecret:  oauthGitHubSecret,
	}, nil
}
