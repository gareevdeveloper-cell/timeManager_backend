package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// GitHubProvider — OAuth через GitHub.
type GitHubProvider struct{}

// Name возвращает имя провайдера.
func (GitHubProvider) Name() string {
	return "github"
}

// Config возвращает конфигурацию OAuth2.
func (GitHubProvider) Config(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"user:email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}
}

// githubUserInfo — ответ API /user.
type githubUserInfo struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// FetchUserInfo получает данные пользователя из GitHub.
func (GitHubProvider) FetchUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github user: status %d", resp.StatusCode)
	}

	var g githubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&g); err != nil {
		return nil, err
	}

	email := g.Email
	if email == "" {
		email, err = fetchGitHubPrimaryEmail(ctx, token)
		if err != nil {
			return nil, err
		}
	}

	firstName, lastName := splitName(g.Name)

	return &UserInfo{
		ProviderID: fmt.Sprintf("%d", g.ID),
		Email:      email,
		FirstName:  firstName,
		LastName:   lastName,
		AvatarURL:  g.AvatarURL,
	}, nil
}

func splitName(name string) (first, last string) {
	for i, r := range name {
		if r == ' ' {
			return name[:i], name[i+1:]
		}
	}
	return name, ""
}

func fetchGitHubPrimaryEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github emails: status %d", resp.StatusCode)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}
	if len(emails) > 0 {
		return emails[0].Email, nil
	}
	return "", fmt.Errorf("no email from github")
}
