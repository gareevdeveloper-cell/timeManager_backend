package oauth

import (
	"context"

	"golang.org/x/oauth2"
)

// UserInfo — данные пользователя от OAuth-провайдера.
type UserInfo struct {
	ProviderID string
	Email      string
	FirstName  string
	LastName   string
	AvatarURL  string
}

// Provider — интерфейс OAuth-провайдера.
type Provider interface {
	Name() string
	Config(clientID, clientSecret, redirectURL string) *oauth2.Config
	FetchUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error)
}
