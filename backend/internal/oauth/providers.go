package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/systemsetting"
	"github.com/shuTwT/nex-api/backend/internal/runtime"
)

const oauthProvidersSettingKey = "oauthProviders"

var (
	errOAuthProviderNotFound = errors.New("oauth: provider not configured")
	errOAuthProviderInvalid  = errors.New("oauth: provider configuration is invalid")
)

// providerConfig is persisted as one entry in SystemSetting.oauthProviders.
// RoleField is deliberately not used: an upstream identity provider must never
// be able to grant a local administrative role.
type providerConfig struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ClientID         string `json:"clientId"`
	ClientSecret     string `json:"clientSecret"`
	AuthorizationURL string `json:"authorizationUrl"`
	TokenURL         string `json:"tokenUrl"`
	UserInfoURL      string `json:"userInfoUrl"`
	Scopes           string `json:"scopes"`
	UserIDField      string `json:"userIdField"`
	EmailField       string `json:"emailField"`
	UsernameField    string `json:"usernameField"`
	RoleField        string `json:"roleField"`
}

type publicProvider struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type configuredProvider struct {
	providerConfig
	client *http.Client
}

func (h *Handler) providers(w http.ResponseWriter, r *http.Request) {
	providers, err := h.configuredProviders(r.Context())
	if err != nil {
		h.writeOAuthError(w, http.StatusServiceUnavailable, "oauth_unavailable")
		return
	}
	available := make([]publicProvider, 0, len(providers))
	for _, provider := range providers {
		if provider.validate() == nil {
			available = append(available, publicProvider{ID: provider.ID, Name: provider.displayName()})
		}
	}
	_ = runtime.WriteData(w, http.StatusOK, available)
}

func (h *Handler) provider(ctx context.Context, rawID string) (*configuredProvider, error) {
	providerID := normalizeProviderID(rawID)
	if providerID == "" {
		return nil, errOAuthProviderNotFound
	}
	providers, err := h.configuredProviders(ctx)
	if err != nil {
		return nil, err
	}
	for _, config := range providers {
		if config.ID != providerID {
			continue
		}
		if err := config.validate(); err != nil {
			return nil, err
		}
		return &configuredProvider{providerConfig: config, client: h.http}, nil
	}
	return nil, errOAuthProviderNotFound
}

func (h *Handler) configuredProviders(ctx context.Context) ([]providerConfig, error) {
	setting, err := h.client.SystemSetting.Query().Where(systemsetting.Key(oauthProvidersSettingKey)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("oauth: load providers: %w", err)
	}
	if strings.TrimSpace(setting.Value) == "" {
		return nil, nil
	}
	var providers []providerConfig
	if err := json.Unmarshal([]byte(setting.Value), &providers); err != nil {
		return nil, fmt.Errorf("%w: %v", errOAuthProviderInvalid, err)
	}
	seen := make(map[string]struct{}, len(providers))
	for index := range providers {
		providers[index].ID = normalizeProviderID(providers[index].ID)
		if providers[index].ID == "" {
			return nil, errOAuthProviderInvalid
		}
		if _, ok := seen[providers[index].ID]; ok {
			return nil, fmt.Errorf("%w: duplicate provider %q", errOAuthProviderInvalid, providers[index].ID)
		}
		seen[providers[index].ID] = struct{}{}
	}
	return providers, nil
}

func (p providerConfig) validate() error {
	if !validProviderID(p.ID) || strings.TrimSpace(p.ClientID) == "" || strings.TrimSpace(p.ClientSecret) == "" {
		return errOAuthProviderInvalid
	}
	for _, rawURL := range []string{p.AuthorizationURL, p.TokenURL, p.UserInfoURL} {
		parsed, err := url.Parse(strings.TrimSpace(rawURL))
		if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.User != nil {
			return errOAuthProviderInvalid
		}
	}
	return nil
}

func (p providerConfig) displayName() string {
	if name := strings.TrimSpace(p.Name); name != "" {
		return name
	}
	return p.ID
}

func (p *configuredProvider) authorizationURL(state OAuthState, redirectURI string) (string, error) {
	if p.ID == githubProvider {
		github := p.github()
		return github.authorizationURL(state, redirectURI)
	}
	if p.isEasy1() {
		easy1, err := p.easy1()
		if err != nil {
			return "", err
		}
		return easy1.authorizationURL(state, redirectURI)
	}
	endpoint, err := url.Parse(p.AuthorizationURL)
	if err != nil {
		return "", fmt.Errorf("oauth: parse authorization URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("client_id", p.ClientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("response_type", "code")
	query.Set("scope", normalizedScopes(p.Scopes))
	query.Set("state", state.Value)
	query.Set("code_challenge", state.CodeChallenge)
	query.Set("code_challenge_method", "S256")
	query.Set("nonce", state.Nonce)
	endpoint.RawQuery = query.Encode()
	return endpoint.String(), nil
}

func (p *configuredProvider) authenticate(ctx context.Context, code, redirectURI string, state OAuthState) (normalizedProfile, accountTokens, error) {
	if p.ID == githubProvider {
		githubTokens, err := p.github().exchange(ctx, code, redirectURI)
		if err != nil {
			return normalizedProfile{}, accountTokens{}, err
		}
		profile, err := p.github().profile(ctx, githubTokens)
		return profile, accountTokensFromOAuth(githubTokens), err
	}
	if p.isEasy1() {
		easy1, err := p.easy1()
		if err != nil {
			return normalizedProfile{}, accountTokens{}, err
		}
		easy1Tokens, err := easy1.exchange(ctx, code, redirectURI, state.CodeVerifier)
		if err != nil {
			return normalizedProfile{}, accountTokens{}, err
		}
		return easy1.profile(ctx, easy1Tokens, state.Nonce)
	}
	form := url.Values{
		"client_id":     {p.ClientID},
		"client_secret": {p.ClientSecret},
		"code":          {code},
		"code_verifier": {state.CodeVerifier},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return normalizedProfile{}, accountTokens{}, fmt.Errorf("oauth: create token request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokens, err := decodeJSONResponse[oauthTokens](p.client, request)
	if err != nil {
		return normalizedProfile{}, accountTokens{}, fmt.Errorf("oauth: exchange code: %w", err)
	}
	if strings.TrimSpace(tokens.AccessToken) == "" {
		return normalizedProfile{}, accountTokens{}, errors.New("oauth: token response has no access token")
	}
	profile, err := p.profile(ctx, tokens)
	if err != nil {
		return normalizedProfile{}, accountTokens{}, err
	}
	return profile, accountTokensFromOAuth(tokens), nil
}

func (p *configuredProvider) usesPKCE() bool {
	return p.ID != githubProvider
}

func (p *configuredProvider) isEasy1() bool {
	return p.ID == easy1Provider || p.ID == "easy1"
}

func (p *configuredProvider) github() *githubClient {
	client := newGitHubClient(p.ClientID, p.ClientSecret, p.client)
	client.authorizeURL = p.AuthorizationURL
	client.tokenURL = p.TokenURL
	client.profileURL = p.UserInfoURL
	if scope := normalizedScopes(p.Scopes); scope != "" {
		client.scope = scope
	}
	return client
}

func (p *configuredProvider) easy1() (*easy1Client, error) {
	return newEasy1Client(
		p.ClientID,
		p.ClientSecret,
		p.AuthorizationURL,
		p.TokenURL,
		p.UserInfoURL,
		normalizedScopes(p.Scopes),
		p.client,
	)
}

func (p *configuredProvider) profile(ctx context.Context, tokens oauthTokens) (normalizedProfile, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.UserInfoURL, nil)
	if err != nil {
		return normalizedProfile{}, fmt.Errorf("oauth: create profile request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	request.Header.Set("Accept", "application/json")
	attributes, err := decodeJSONResponse[map[string]any](p.client, request)
	if err != nil {
		return normalizedProfile{}, fmt.Errorf("oauth: fetch profile: %w", err)
	}
	accountID := stringAttribute(attributes, p.UserIDField, "sub", "id")
	email := strings.ToLower(stringAttribute(attributes, p.EmailField, "email"))
	username := stringAttribute(attributes, p.UsernameField, "preferred_username", "username", "login", "name")
	if accountID == "" || email == "" {
		return normalizedProfile{}, errors.New("oauth: profile is missing account ID or email")
	}
	if verified, exists := boolAttribute(attributes, "email_verified", "emailVerified"); exists && !verified {
		return normalizedProfile{}, errors.New("oauth: profile email is not verified")
	}
	if username == "" {
		username = strings.Split(email, "@")[0]
	}
	return normalizedProfile{
		ProviderAccountID: accountID,
		Email:             email,
		Name:              stringAttribute(attributes, "name", "display_name", "displayName"),
		Username:          username,
		Image:             stringAttribute(attributes, "picture", "avatar_url", "avatarUrl"),
		EmailVerified:     true,
	}, nil
}

func normalizeProviderID(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func validProviderID(value string) bool {
	if len(value) == 0 || len(value) > 64 {
		return false
	}
	for index, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || (index > 0 && (character == '-' || character == '_')) {
			continue
		}
		return false
	}
	return true
}

func normalizedScopes(raw string) string {
	parts := strings.FieldsFunc(raw, func(character rune) bool {
		return character == ',' || unicode.IsSpace(character)
	})
	return strings.Join(parts, " ")
}

func stringAttribute(attributes map[string]any, preferred string, fallbacks ...string) string {
	for _, field := range append([]string{preferred}, fallbacks...) {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if value, ok := nestedAttribute(attributes, field); ok {
			switch typed := value.(type) {
			case string:
				if value := strings.TrimSpace(typed); value != "" {
					return value
				}
			case json.Number:
				return typed.String()
			case float64:
				return strconv.FormatFloat(typed, 'f', -1, 64)
			}
		}
	}
	return ""
}

func boolAttribute(attributes map[string]any, fields ...string) (bool, bool) {
	for _, field := range fields {
		value, ok := nestedAttribute(attributes, field)
		if !ok {
			continue
		}
		if typed, ok := value.(bool); ok {
			return typed, true
		}
	}
	return false, false
}

func nestedAttribute(attributes map[string]any, path string) (any, bool) {
	var current any = attributes
	for _, part := range strings.Split(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func oauthProviderStatus(err error) int {
	if errors.Is(err, errOAuthProviderNotFound) || errors.Is(err, errOAuthProviderInvalid) {
		return http.StatusNotFound
	}
	return http.StatusServiceUnavailable
}
