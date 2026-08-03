package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	githubProvider    = "github"
	githubAuthorize   = "https://github.com/login/oauth/authorize"
	githubToken       = "https://github.com/login/oauth/access_token"
	githubProfileURL  = "https://api.github.com/user"
	githubEmailList   = "https://api.github.com/user/emails"
	maxOAuthBodyBytes = 2 << 20
)

type githubClient struct {
	clientID     string
	clientSecret string
	authorizeURL string
	tokenURL     string
	profileURL   string
	emailsURL    string
	scope        string
	client       *http.Client
}

type oauthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
}

type normalizedProfile struct {
	Provider          string
	ProviderAccountID string
	Email             string
	Name              string
	Username          string
	Image             string
	EmailVerified     bool
	Nonce             string
}

type accountTokens struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Scope        string
	ExpiresAt    int
	IDToken      string
}

type githubProfile struct {
	ID        json.RawMessage `json:"id"`
	Login     string          `json:"login"`
	Name      string          `json:"name"`
	Email     string          `json:"email"`
	AvatarURL string          `json:"avatar_url"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func newGitHubClient(clientID, clientSecret string, client *http.Client) *githubClient {
	if client == nil {
		client = &http.Client{}
	}
	return &githubClient{
		clientID:     strings.TrimSpace(clientID),
		clientSecret: strings.TrimSpace(clientSecret),
		authorizeURL: githubAuthorize,
		tokenURL:     githubToken,
		profileURL:   githubProfileURL,
		emailsURL:    githubEmailList,
		scope:        "read:user user:email",
		client:       client,
	}
}

func (p *githubClient) authorizationURL(state OAuthState, redirectURI string) (string, error) {
	if p.clientID == "" || p.clientSecret == "" {
		return "", errors.New("oauth: GitHub client is not configured")
	}
	endpoint, err := url.Parse(p.authorizeURL)
	if err != nil {
		return "", fmt.Errorf("oauth: parse GitHub authorization URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("client_id", p.clientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("scope", p.scope)
	query.Set("state", state.Value)
	endpoint.RawQuery = query.Encode()
	return endpoint.String(), nil
}

func (p *githubClient) exchange(ctx context.Context, code, redirectURI string) (oauthTokens, error) {
	form := url.Values{
		"client_id":     {p.clientID},
		"client_secret": {p.clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return oauthTokens{}, fmt.Errorf("oauth: create GitHub token request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokens, err := decodeJSONResponse[oauthTokens](p.client, request)
	if err != nil {
		return oauthTokens{}, fmt.Errorf("oauth: exchange GitHub code: %w", err)
	}
	if strings.TrimSpace(tokens.AccessToken) == "" {
		return oauthTokens{}, errors.New("oauth: GitHub token response has no access token")
	}
	return tokens, nil
}

func (p *githubClient) profile(ctx context.Context, tokens oauthTokens) (normalizedProfile, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.profileURL, nil)
	if err != nil {
		return normalizedProfile{}, fmt.Errorf("oauth: create GitHub profile request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "nex-api")
	profile, err := decodeJSONResponse[githubProfile](p.client, request)
	if err != nil {
		return normalizedProfile{}, fmt.Errorf("oauth: fetch GitHub profile: %w", err)
	}
	providerAccountID, err := githubProfileID(profile.ID)
	if err != nil {
		return normalizedProfile{}, err
	}
	email := strings.TrimSpace(profile.Email)
	verified := email != ""
	if email == "" {
		email, verified, err = p.primaryEmail(ctx, tokens.AccessToken)
		if err != nil {
			return normalizedProfile{}, err
		}
	}
	if !verified || email == "" {
		return normalizedProfile{}, errors.New("oauth: GitHub profile has no verified email")
	}
	return normalizedProfile{
		ProviderAccountID: providerAccountID,
		Email:             strings.ToLower(email),
		Name:              strings.TrimSpace(profile.Name),
		Username:          strings.TrimSpace(profile.Login),
		Image:             strings.TrimSpace(profile.AvatarURL),
		EmailVerified:     true,
	}, nil
}

func (p *githubClient) primaryEmail(ctx context.Context, accessToken string) (string, bool, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.emailsURL, nil)
	if err != nil {
		return "", false, fmt.Errorf("oauth: create GitHub email request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "nex-api")
	emails, err := decodeJSONResponse[[]githubEmail](p.client, request)
	if err != nil {
		return "", false, fmt.Errorf("oauth: fetch GitHub emails: %w", err)
	}
	for _, email := range emails {
		if email.Primary && email.Verified && strings.TrimSpace(email.Email) != "" {
			return strings.TrimSpace(email.Email), true, nil
		}
	}
	for _, email := range emails {
		if email.Verified && strings.TrimSpace(email.Email) != "" {
			return strings.TrimSpace(email.Email), true, nil
		}
	}
	return "", false, nil
}

func decodeJSONResponse[T any](client *http.Client, request *http.Request) (T, error) {
	response, err := client.Do(request)
	if err != nil {
		var zero T
		return zero, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, readErr := io.ReadAll(io.LimitReader(response.Body, maxOAuthBodyBytes))
		if readErr != nil {
			var zero T
			return zero, fmt.Errorf("upstream status %d: read error: %w", response.StatusCode, readErr)
		}
		var zero T
		return zero, fmt.Errorf("upstream status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxOAuthBodyBytes))
	var destination T
	if err := decoder.Decode(&destination); err != nil {
		var zero T
		return zero, err
	}
	return destination, nil
}

func githubProfileID(raw json.RawMessage) (string, error) {
	var numeric int64
	if err := json.Unmarshal(raw, &numeric); err == nil && numeric > 0 {
		return strconv.FormatInt(numeric, 10), nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil && strings.TrimSpace(text) != "" {
		return strings.TrimSpace(text), nil
	}
	return "", errors.New("oauth: GitHub profile has no account ID")
}
