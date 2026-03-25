package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"testik/internal/auth"
	"testik/internal/auth/oauth"
	"testik/internal/files"
	httpserver "testik/internal/http"
	"testik/internal/config"
	"testik/internal/migrate"
	"testik/internal/organization"
	"testik/internal/project"
	"testik/internal/team"
	"testik/internal/user"
	"testik/internal/ws"
	"testik/pkg/storage"
)

// App — приложение с инициализированными зависимостями.
type App struct {
	cfg    *config.Config
	pool   *pgxpool.Pool
	router *httpserver.Router
}

// New создаёт и инициализирует приложение.
func New(cfg *config.Config) (*App, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}

	if err := migrate.Migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	minioStorage, err := storage.NewMinIOStorage(ctx, storage.MinIOConfig{
		Endpoint:        cfg.MinIOEndpoint,
		AccessKeyID:     cfg.MinIOAccessKey,
		SecretAccessKey: cfg.MinIOSecretKey,
		Bucket:          cfg.MinIOBucket,
		UseSSL:          cfg.MinIOUseSSL,
		PublicURL:       cfg.MinIOPublicURL,
	})
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("minio: %w", err)
	}

	authUserRepo := auth.NewPostgresUserRepository(pool)
	oauthProviders := []auth.OAuthProviderConfig{
		{
			ClientID:     cfg.OAuthGoogleID,
			ClientSecret: cfg.OAuthGoogleSecret,
			RedirectURL:  cfg.OAuthRedirectBase + "/api/v1/auth/google/callback",
			Provider:     oauth.GoogleProvider{},
		},
		{
			ClientID:     cfg.OAuthYandexID,
			ClientSecret: cfg.OAuthYandexSecret,
			RedirectURL:  cfg.OAuthRedirectBase + "/api/v1/auth/yandex/callback",
			Provider:     oauth.YandexProvider{},
		},
		{
			ClientID:     cfg.OAuthGitHubID,
			ClientSecret: cfg.OAuthGitHubSecret,
			RedirectURL:  cfg.OAuthRedirectBase + "/api/v1/auth/github/callback",
			Provider:     oauth.GitHubProvider{},
		},
	}
	authService := auth.NewAuthService(authUserRepo, auth.AuthServiceConfig{
		JWTSecret: cfg.JWTSecret,
		ExpiresIn: cfg.JWTExpires,
	}, oauthProviders)

	hub := ws.NewHub()
	go hub.Run()

	statusPublisher := ws.NewHubWorkStatusPublisher(hub)
	taskPublisher := ws.NewHubTaskPublisher(hub)
	wsHandler := ws.NewHandler(hub, authService)

	userRepo := user.NewPostgresUserRepository(pool)
	userSkillRepo := user.NewPostgresSkillRepository(pool)
	userStatusHistRepo := user.NewPostgresWorkStatusHistoryRepository(pool)
	taskRepo := project.NewPostgresTaskRepository(pool)
	userService := user.NewService(userRepo, taskRepo, userSkillRepo, userStatusHistRepo, minioStorage, statusPublisher)
	userHandler := user.NewHandler(userService)

	authHandler := auth.NewHandler(authService, auth.HandlerConfig{
		OAuthFrontendRedirect: cfg.OAuthFrontendRedirect,
	})

	orgRepo := organization.NewPostgresRepository(pool)
	orgService := organization.NewService(orgRepo, minioStorage)
	orgHandler := organization.NewHandler(orgService)

	teamRepo := team.NewPostgresRepository(pool)
	teamService := team.NewService(teamRepo, orgRepo, minioStorage)
	teamHandler := team.NewHandler(teamService)

	projectRepo := project.NewPostgresProjectRepository(pool)
	projectStatusRepo := project.NewPostgresProjectStatusRepository(pool)
	projectService := project.NewService(projectRepo, projectStatusRepo, taskRepo, teamRepo, taskPublisher)
	projectHandler := project.NewHandler(projectService)
	filesHandler := files.NewHandler(minioStorage)

	router := httpserver.NewRouter(httpserver.RouterDeps{
		AuthHandler:    authHandler,
		AuthService:    authService,
		UserHandler:    userHandler,
		OrgHandler:     orgHandler,
		TeamHandler:    teamHandler,
		ProjectHandler: projectHandler,
		FilesHandler:   filesHandler,
		WsHandler:      wsHandler,
	})

	return &App{
		cfg:    cfg,
		pool:   pool,
		router: router,
	}, nil
}

// Run запускает HTTP-сервер и блокируется до получения сигнала остановки.
func (a *App) Run() error {
	defer a.pool.Close()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.cfg.HTTPPort),
		Handler: a.router.Engine(),
	}

	go func() {
		log.Printf("server listening on :%d", a.cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	log.Println("server stopped")
	return nil
}
