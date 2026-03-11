package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"testik/pkg/response"
)

// Handler — HTTP-обработчики аутентификации.
type Handler struct {
	service               *AuthService
	oauthFrontendRedirect string
}

// HandlerConfig — конфигурация handler.
type HandlerConfig struct {
	OAuthFrontendRedirect string
}

// NewHandler создаёт handler.
func NewHandler(service *AuthService, cfg HandlerConfig) *Handler {
	frontendRedirect := cfg.OAuthFrontendRedirect
	if frontendRedirect == "" {
		frontendRedirect = "http://localhost:3000/auth/callback"
	}
	return &Handler{
		service:               service,
		oauthFrontendRedirect: frontendRedirect,
	}
}

// Register godoc
// @Summary Регистрация пользователя
// @Description Создаёт нового пользователя по email и паролю. Пароль должен быть не менее 8 символов.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "Email и пароль"
// @Success 201 {object} auth.RegisterResponse
// @Failure 400 {object} auth.ErrorResponse
// @Failure 409 {object} auth.ErrorResponse
// @Failure 500 {object} auth.ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	u, err := h.service.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == ErrUserAlreadyExists {
			response.Error(c, http.StatusConflict, "user_exists", "user with this email already exists")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "registration failed")
		return
	}

	response.Data(c, http.StatusCreated, gin.H{
		"user_id": u.ID.String(),
		"email":   u.Email,
	})
}

// Login godoc
// @Summary Вход в систему
// @Description Проверяет учётные данные и возвращает JWT access token.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Email и пароль"
// @Success 200 {object} auth.LoginResponse
// @Failure 400 {object} auth.ErrorResponse
// @Failure 401 {object} auth.ErrorResponse
// @Failure 500 {object} auth.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	token, expiresIn, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == ErrInvalidCredentials || err == ErrUserInactive {
			response.Error(c, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "login failed")
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken: token,
		ExpiresIn:   expiresIn,
		TokenType:   "Bearer",
	})
}

// OAuthRedirect godoc
// @Summary OAuth — редирект на провайдера
// @Description Редирект на страницу авторизации Google, Yandex или GitHub. Поддерживаемые provider: google, yandex, github.
// @Tags auth
// @Produce json
// @Param provider path string true "google | yandex | github"
// @Success 302 "Redirect to OAuth provider"
// @Failure 400 {object} auth.ErrorResponse
// @Failure 404 {object} auth.ErrorResponse
// @Router /api/v1/auth/{provider}/redirect [get]
func (h *Handler) OAuthRedirect(c *gin.Context) {
	provider := c.Param("provider")
	if provider != "google" && provider != "yandex" && provider != "github" {
		response.Error(c, http.StatusBadRequest, "invalid_provider", "provider must be google, yandex or github")
		return
	}

	state, err := randomState()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to generate state")
		return
	}

	redirectURL, ok := h.service.OAuthRedirectURL(provider, state)
	if !ok {
		response.Error(c, http.StatusNotFound, "oauth_disabled", "oauth provider not configured")
		return
	}

	c.SetCookie("oauth_state", state, 600, "/", "", false, true)
	c.SetCookie("oauth_provider", provider, 600, "/", "", false, true)

	c.Redirect(http.StatusFound, redirectURL)
}

// OAuthCallback godoc
// @Summary OAuth — callback от провайдера
// @Description Обрабатывает callback от OAuth-провайдера, создаёт/находит пользователя, редирект на фронтенд с токеном.
// @Tags auth
// @Produce json
// @Param provider path string true "google | yandex | github"
// @Param code query string true "Authorization code from provider"
// @Param state query string true "State (CSRF protection)"
// @Success 302 "Redirect to frontend with ?token=..."
// @Failure 400 {object} auth.ErrorResponse
// @Failure 401 {object} auth.ErrorResponse
// @Failure 500 {object} auth.ErrorResponse
// @Router /api/v1/auth/{provider}/callback [get]
func (h *Handler) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	if provider != "google" && provider != "yandex" && provider != "github" {
		redirectWithError(c, h.oauthFrontendRedirect, "invalid_provider")
		return
	}

	code := c.Query("code")
	if code == "" {
		redirectWithError(c, h.oauthFrontendRedirect, "missing_code")
		return
	}

	stateCookie, _ := c.Cookie("oauth_state")
	stateQuery := c.Query("state")
	if stateCookie != "" && stateQuery != "" && stateCookie != stateQuery {
		redirectWithError(c, h.oauthFrontendRedirect, "invalid_state")
		return
	}

	token, expiresIn, err := h.service.OAuthCallback(c.Request.Context(), provider, code)
	if err != nil {
		if err == ErrUserAlreadyExists {
			redirectWithError(c, h.oauthFrontendRedirect, "email_already_registered")
			return
		}
		if err == ErrUserInactive {
			redirectWithError(c, h.oauthFrontendRedirect, "user_inactive")
			return
		}
		redirectWithError(c, h.oauthFrontendRedirect, "oauth_failed")
		return
	}

	c.SetCookie("oauth_state", "", -1, "/", "", false, true)
	c.SetCookie("oauth_provider", "", -1, "/", "", false, true)

	redirectURL, _ := url.Parse(h.oauthFrontendRedirect)
	q := redirectURL.Query()
	q.Set("token", token)
	q.Set("expires_in", fmt.Sprintf("%d", expiresIn))
	redirectURL.RawQuery = q.Encode()

	c.Redirect(http.StatusFound, redirectURL.String())
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func redirectWithError(c *gin.Context, baseURL, errCode string) {
	u, _ := url.Parse(baseURL)
	q := u.Query()
	q.Set("error", errCode)
	u.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, u.String())
}
