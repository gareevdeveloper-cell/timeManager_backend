package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"

	"testik/internal/auth/oauth"
	"testik/internal/domain"
)

// JWTClaims — claims для access token.
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// AuthService — сервис аутентификации.
type AuthService struct {
	repo       UserRepository
	jwtSecret  []byte
	expiresIn  time.Duration
	bcryptCost int
	oauth      map[string]oauthConfig
}

type oauthConfig struct {
	config   *oauth2.Config
	provider oauth.Provider
}

// AuthServiceConfig — конфигурация сервиса.
type AuthServiceConfig struct {
	JWTSecret  string
	ExpiresIn  time.Duration
	BcryptCost int
}

// OAuthProviderConfig — конфиг одного OAuth-провайдера.
type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Provider     oauth.Provider
}

// NewAuthService создаёт AuthService.
func NewAuthService(repo UserRepository, cfg AuthServiceConfig, oauthProviders []OAuthProviderConfig) *AuthService {
	if cfg.BcryptCost == 0 {
		cfg.BcryptCost = bcrypt.DefaultCost
	}
	oauthMap := make(map[string]oauthConfig)
	for _, p := range oauthProviders {
		if p.ClientID != "" && p.ClientSecret != "" {
			oauthMap[p.Provider.Name()] = oauthConfig{
				config:   p.Provider.Config(p.ClientID, p.ClientSecret, p.RedirectURL),
				provider: p.Provider,
			}
		}
	}
	return &AuthService{
		repo:       repo,
		jwtSecret:  []byte(cfg.JWTSecret),
		expiresIn:  cfg.ExpiresIn,
		bcryptCost: cfg.BcryptCost,
		oauth:      oauthMap,
	}
}

// Register создаёт нового пользователя.
func (s *AuthService) Register(ctx context.Context, email, password string) (*domain.User, error) {
	_, err := s.repo.GetByEmail(ctx, email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	u := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         domain.UserRoleUser,
		Status:       domain.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Login проверяет учётные данные и возвращает access token.
func (s *AuthService) Login(ctx context.Context, email, password string) (token string, expiresIn int64, err error) {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", 0, ErrInvalidCredentials
		}
		return "", 0, err
	}

	if u.Status != domain.UserStatusActive {
		return "", 0, ErrUserInactive
	}

	if u.OAuthProvider != "" {
		return "", 0, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", 0, ErrInvalidCredentials
	}

	exp := time.Now().Add(s.expiresIn)
	claims := JWTClaims{
		UserID: u.ID.String(),
		Email:  u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = t.SignedString(s.jwtSecret)
	if err != nil {
		return "", 0, err
	}
	return token, int64(s.expiresIn.Seconds()), nil
}

// GetUserByID возвращает пользователя по ID.
func (s *AuthService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

// OAuthRedirectURL возвращает URL для редиректа на OAuth-провайдера (если провайдер настроен).
func (s *AuthService) OAuthRedirectURL(provider, state string) (string, bool) {
	oc, ok := s.oauth[provider]
	if !ok {
		return "", false
	}
	opts := []oauth2.AuthCodeOption{oauth2.AccessTypeOffline}
	if provider == "google" {
		opts = append(opts, oauth2.SetAuthURLParam("prompt", "consent"))
	}
	return oc.config.AuthCodeURL(state, opts...), true
}

// OAuthCallback обменивает code на токен, получает userinfo, создаёт/находит пользователя и возвращает JWT.
func (s *AuthService) OAuthCallback(ctx context.Context, provider, code string) (token string, expiresIn int64, err error) {
	oc, ok := s.oauth[provider]
	if !ok {
		return "", 0, ErrInvalidCredentials
	}

	tok, err := oc.config.Exchange(ctx, code)
	if err != nil {
		return "", 0, err
	}

	info, err := oc.provider.FetchUserInfo(ctx, tok)
	if err != nil {
		return "", 0, err
	}

	if info.Email == "" {
		return "", 0, errors.New("oauth: email required")
	}

	u, err := s.repo.GetByOAuthID(ctx, provider, info.ProviderID)
	if err == nil {
		if u.Status != domain.UserStatusActive {
			return "", 0, ErrUserInactive
		}
		return s.issueToken(u)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", 0, err
	}

	existing, _ := s.repo.GetByEmail(ctx, info.Email)
	if existing != nil {
		return "", 0, ErrUserAlreadyExists
	}

	now := time.Now()
	u = &domain.User{
		ID:              uuid.New(),
		Email:           info.Email,
		OAuthProvider:   provider,
		OAuthProviderID: info.ProviderID,
		FirstName:       info.FirstName,
		LastName:        info.LastName,
		AvatarURL:       info.AvatarURL,
		Role:            domain.UserRoleUser,
		Status:          domain.UserStatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return "", 0, err
	}
	return s.issueToken(u)
}

func (s *AuthService) issueToken(u *domain.User) (string, int64, error) {
	exp := time.Now().Add(s.expiresIn)
	claims := JWTClaims{
		UserID: u.ID.String(),
		Email:  u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := t.SignedString(s.jwtSecret)
	if err != nil {
		return "", 0, err
	}
	return token, int64(s.expiresIn.Seconds()), nil
}

// ValidateToken проверяет JWT и возвращает user_id.
func (s *AuthService) ValidateToken(tokenString string) (userID string, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return "", jwt.ErrTokenInvalidClaims
	}
	return claims.UserID, nil
}
