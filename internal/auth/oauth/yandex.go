package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// YandexProvider — OAuth через Yandex.
type YandexProvider struct{}

// Name возвращает имя провайдера.
func (YandexProvider) Name() string {
	return "yandex"
}

// Config возвращает конфигурацию OAuth2.
func (YandexProvider) Config(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"login:email", "login:info"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://oauth.yandex.com/authorize",
			TokenURL: "https://oauth.yandex.com/token",
		},
	}
}

// yandexUserInfo — ответ API info.
type yandexUserInfo struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DefaultEmail    string `json:"default_email"`
	Emails          []string `json:"emails"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	DisplayName     string `json:"display_name"`
	DefaultAvatarID string `json:"default_avatar_id"`
}

// FetchUserInfo получает данные пользователя из Yandex.
func (YandexProvider) FetchUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://login.yandex.ru/info?format=json", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "OAuth "+token.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yandex userinfo: status %d", resp.StatusCode)
	}

	var y yandexUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&y); err != nil {
		return nil, err
	}

	email := y.DefaultEmail
	if email == "" && len(y.Emails) > 0 {
		email = y.Emails[0]
	}

	avatarURL := ""
	if y.DefaultAvatarID != "" {
		avatarURL = "https://avatars.yandex.net/get-yapic/" + y.DefaultAvatarID + "/islands-200"
	}

	return &UserInfo{
		ProviderID: y.ID,
		Email:      email,
		FirstName:  y.FirstName,
		LastName:   y.LastName,
		AvatarURL:  avatarURL,
	}, nil
}
