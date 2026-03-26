package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"testik/pkg/env"
)

// postgresURLFromParts собирает DSN; пароль кодируется для спецсимволов в URL.
func postgresURLFromParts(user, password, host, port, dbName string) string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   host + ":" + port,
		Path:   "/" + dbName,
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	return u.String()
}

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
	OAuthRedirectBase     string // базовый URL для callback (например https://api.example.com)
	OAuthFrontendRedirect string // URL фронтенда для редиректа с токеном (?token=...)
	OAuthGoogleID         string
	OAuthGoogleSecret     string
	OAuthYandexID         string
	OAuthYandexSecret     string
	OAuthGitHubID         string
	OAuthGitHubSecret     string
}

// Load загружает конфигурацию из переменных окружения.
func Load() (*Config, error) {
	port, _ := strconv.Atoi(env.Get("HTTP_PORT", "8080"))
	expiresMin, _ := strconv.Atoi(env.Get("JWT_EXPIRES_MINUTES", "60"))

	dbName := env.Get("DB_NAME", "timemanager")
	// Если задан DB_HOST (типично Docker Compose / «соседний» контейнер), строку
	// подключения собираем только из DB_* — так устаревший или ошибочный DATABASE_URL
	// из .env не перекрывает пароль из окружения контейнера.
	var dbURL string
	if host := os.Getenv("DB_HOST"); host != "" {
		dbURL = postgresURLFromParts(
			env.Get("DB_USER", "postgres"),
			env.Get("DB_PASSWORD", "postgres"),
			host,
			env.Get("DB_PORT", "5432"),
			dbName,
		)
	} else {
		dbURL = env.Get("DATABASE_URL", "")
		if dbURL == "" {
			dbURL = postgresURLFromParts(
				env.Get("DB_USER", "postgres"),
				env.Get("DB_PASSWORD", "postgres"),
				"localhost",
				env.Get("DB_PORT", "5432"),
				dbName,
			)
		}
	}

	fmt.Println("dbURL: " + dbURL)

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
		OAuthFrontendRedirect: oauthFrontendRedirect,
		OAuthGoogleID:         oauthGoogleID,
		OAuthGoogleSecret:     oauthGoogleSecret,
		OAuthYandexID:         oauthYandexID,
		OAuthYandexSecret:     oauthYandexSecret,
		OAuthGitHubID:         oauthGitHubID,
		OAuthGitHubSecret:     oauthGitHubSecret,
	}, nil
}
