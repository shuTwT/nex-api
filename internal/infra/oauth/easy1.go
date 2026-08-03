package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const Easy1Provider = "easy1auth"

type Easy1Client struct {
	clientID     string
	clientSecret string
	authorizeURL string
	tokenURL     string
	userinfoURL  string
	scope        string
	client       HTTPClient
}

type easy1Profile struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified *bool  `json:"email_verified"`
	Name          string `json:"name"`
	Username      string `json:"preferred_username"`
	Picture       string `json:"picture"`
	Nonce         string `json:"nonce"`
}

func NewEasy1Client(clientID, clientSecret, authorizationURL, tokenURL, userinfoURL, scope string, client HTTPClient) (*Easy1Client, error) {
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" {
		return nil, errors.New("oauth: Easy1Auth client is not configured")
	}
	for name, endpoint := range map[string]string{
		"authorization": authorizationURL,
		"token":         tokenURL,
		"userinfo":      userinfoURL,
	} {
		if _, err := ParseHTTPURL(endpoint); err != nil {
			return nil, fmt.Errorf("oauth: invalid Easy1Auth %s URL: %w", name, err)
		}
	}
	if client == nil {
		client = NewHTTPClient(0)
	}
	if strings.TrimSpace(scope) == "" {
		scope = "openid profile email"
	}
	return &Easy1Client{
		clientID:     strings.TrimSpace(clientID),
		clientSecret: strings.TrimSpace(clientSecret),
		authorizeURL: strings.TrimSpace(authorizationURL),
		tokenURL:     strings.TrimSpace(tokenURL),
		userinfoURL:  strings.TrimSpace(userinfoURL),
		scope:        strings.TrimSpace(scope),
		client:       client,
	}, nil
}

func (p *Easy1Client) AuthorizationURL(state OAuthState, redirectURI string) (string, error) {
	if state.CodeChallenge == "" || state.Nonce == "" {
		return "", errors.New("oauth: Easy1Auth state is missing PKCE or nonce")
	}
	endpoint, err := url.Parse(p.authorizeURL)
	if err != nil {
		return "", fmt.Errorf("oauth: parse Easy1Auth authorization URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("client_id", p.clientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("response_type", "code")
	query.Set("scope", p.scope)
	query.Set("state", state.Value)
	query.Set("code_challenge", state.CodeChallenge)
	query.Set("code_challenge_method", "S256")
	query.Set("nonce", state.Nonce)
	endpoint.RawQuery = query.Encode()
	return endpoint.String(), nil
}

func (p *Easy1Client) Exchange(ctx context.Context, code, redirectURI, verifier string) (OAuthTokens, error) {
	if verifier == "" {
		return OAuthTokens{}, errors.New("oauth: Easy1Auth callback has no PKCE verifier")
	}
	form := url.Values{
		"client_id":     {p.clientID},
		"client_secret": {p.clientSecret},
		"code":          {code},
		"code_verifier": {verifier},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return OAuthTokens{}, fmt.Errorf("oauth: create Easy1Auth token request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokens, err := DecodeJSONResponse[OAuthTokens](p.client, request)
	if err != nil {
		return OAuthTokens{}, fmt.Errorf("oauth: exchange Easy1Auth code: %w", err)
	}
	if strings.TrimSpace(tokens.AccessToken) == "" {
		return OAuthTokens{}, errors.New("oauth: Easy1Auth token response has no access token")
	}
	return tokens, nil
}

func (p *Easy1Client) Profile(ctx context.Context, tokens OAuthTokens, expectedNonce string) (NormalizedProfile, AccountTokens, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userinfoURL, nil)
	if err != nil {
		return NormalizedProfile{}, AccountTokens{}, fmt.Errorf("oauth: create Easy1Auth profile request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	request.Header.Set("Accept", "application/json")
	profile, err := DecodeJSONResponse[easy1Profile](p.client, request)
	if err != nil {
		return NormalizedProfile{}, AccountTokens{}, fmt.Errorf("oauth: fetch Easy1Auth profile: %w", err)
	}
	if strings.TrimSpace(profile.Subject) == "" || strings.TrimSpace(profile.Email) == "" {
		return NormalizedProfile{}, AccountTokens{}, errors.New("oauth: Easy1Auth profile is missing sub or email")
	}
	if profile.EmailVerified != nil && !*profile.EmailVerified {
		return NormalizedProfile{}, AccountTokens{}, errors.New("oauth: Easy1Auth email is not verified")
	}
	if profile.Nonce != "" && !constantTimeEqual(profile.Nonce, expectedNonce) {
		return NormalizedProfile{}, AccountTokens{}, ErrInvalidOAuthState
	}
	if tokens.IDToken != "" {
		if err := ValidateOIDCNonce(tokens.IDToken, expectedNonce); err != nil {
			return NormalizedProfile{}, AccountTokens{}, fmt.Errorf("oauth: validate Easy1Auth nonce: %w", err)
		}
	}
	expiresAt := 0
	if tokens.ExpiresIn > 0 {
		expiresAt = int(time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second).Unix())
	}
	return NormalizedProfile{
			ProviderAccountID: strings.TrimSpace(profile.Subject),
			Email:             strings.ToLower(strings.TrimSpace(profile.Email)),
			Name:              strings.TrimSpace(profile.Name),
			Username:          strings.TrimSpace(profile.Username),
			Image:             strings.TrimSpace(profile.Picture),
			EmailVerified:     profile.EmailVerified == nil || *profile.EmailVerified,
			Nonce:             profile.Nonce,
		}, AccountTokens{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			TokenType:    tokens.TokenType,
			Scope:        tokens.Scope,
			ExpiresAt:    expiresAt,
			IDToken:      tokens.IDToken,
		}, nil
}

func constantTimeEqual(left, right string) bool {
	if left == "" || right == "" {
		return false
	}
	return len(left) == len(right) && subtleCompare([]byte(left), []byte(right)) == 1
}

func subtleCompare(left, right []byte) int {
	if len(left) != len(right) {
		return 0
	}
	var result byte
	for index := range left {
		result |= left[index] ^ right[index]
	}
	if result == 0 {
		return 1
	}
	return 0
}
