package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// GoogleProvider — OAuth через Google.
type GoogleProvider struct{}

// Name возвращает имя провайдера.
func (GoogleProvider) Name() string {
	return "google"
}

// Config возвращает конфигурацию OAuth2.
func (GoogleProvider) Config(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}
}

// googleUserInfo — ответ API userinfo.
type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// FetchUserInfo получает данные пользователя из Google.
func (GoogleProvider) FetchUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo: status %d", resp.StatusCode)
	}

	var g googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&g); err != nil {
		return nil, err
	}

	return &UserInfo{
		ProviderID: g.ID,
		Email:      g.Email,
		FirstName:  g.GivenName,
		LastName:   g.FamilyName,
		AvatarURL:  g.Picture,
	}, nil
}
