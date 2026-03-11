package auth

import "time"

// RegisterRequest — тело запроса регистрации.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest — тело запроса логина.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse — ответ с токеном.
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// RegisterResponse — ответ регистрации (для swagger).
type RegisterResponse struct {
	Data struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	} `json:"data"`
}

// UserResponse — ответ с данными пользователя (для swagger).
type UserResponse struct {
	Data struct {
		ID        string    `json:"id"`
		Email     string    `json:"email"`
		FirstName string    `json:"firstname"`
		LastName  string    `json:"lastname"`
		About     string    `json:"about"`
		Position  string    `json:"position"`
		Skills    []string  `json:"skills"`
		Role      string    `json:"role"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"created_at"`
	} `json:"data"`
}

// ErrorResponse — стандартный формат ошибки (для swagger).
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// OAuthCallbackResponse — ответ OAuth callback (редирект с токеном).
type OAuthCallbackResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}
