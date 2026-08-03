package oauth

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/systemsetting"
	infraoauth "github.com/shuTwT/nex-api/internal/infra/oauth"
	"golang.org/x/oauth2"
)

const (
	oauthProvidersSettingKey          = "oauthProviders"
	githubOAuthClientIDSettingKey     = "githubOAuthClientId"
	githubOAuthClientSecretSettingKey = "githubOAuthClientSecret"
	oidcProviderSettingKey            = "oidcProvider"
	customOIDCProviderID              = "custom-oidc"
	providerKindOIDC                  = "oidc"
)

var (
	errOAuthProviderNotFound = errors.New("oauth: provider not configured")
	errOAuthProviderInvalid  = errors.New("oauth: provider configuration is invalid")
)

// ProviderConfig is persisted as one entry in SystemSetting.oauthProviders.
// RoleField is deliberately not used: an upstream identity provider must never
// be able to grant a local administrative role.
type ProviderConfig struct {
	Kind             string `json:"kind,omitempty"`
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
	Issuer           string `json:"issuer,omitempty"`
}

// OIDCProviderConfig is the separate, single OIDC relying-party configuration.
// The issuer discovery document supplies the protocol endpoints and signing keys.
type OIDCProviderConfig struct {
	Name         string `json:"name"`
	Issuer       string `json:"issuer"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Scopes       string `json:"scopes"`
}

// ConfiguredProvider is a validated ProviderConfig bound to an HTTP client.
type ConfiguredProvider struct {
	ProviderConfig
	client infraoauth.HTTPClient
}

// Service owns OAuth provider configuration and account provisioning.
type Service struct {
	client *ent.Client
	http   infraoauth.HTTPClient
	clock  func() time.Time
}

func NewService(client *ent.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("oauth: ent client is nil")
	}
	return &Service{client: client, http: infraoauth.NewHTTPClient(15 * time.Second), clock: time.Now}, nil
}

// ConfiguredProviders loads the provider list from system settings.
func (s *Service) ConfiguredProviders(ctx context.Context) ([]ProviderConfig, error) {
	settings, err := s.client.SystemSetting.Query().Where(systemsetting.KeyIn(
		oauthProvidersSettingKey,
		githubOAuthClientIDSettingKey,
		githubOAuthClientSecretSettingKey,
		oidcProviderSettingKey,
	)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("oauth: load providers: %w", err)
	}
	values := make(map[string]string, len(settings))
	for _, setting := range settings {
		values[setting.Key] = setting.Value
	}

	providers := make([]ProviderConfig, 0)
	var legacyGitHub *ProviderConfig
	if rawProviders := strings.TrimSpace(values[oauthProvidersSettingKey]); rawProviders != "" {
		if err := json.Unmarshal([]byte(rawProviders), &providers); err != nil {
			return nil, fmt.Errorf("%w: %v", errOAuthProviderInvalid, err)
		}
	}
	seen := make(map[string]struct{}, len(providers))
	customProviders := make([]ProviderConfig, 0, len(providers)+1)
	for _, provider := range providers {
		provider.ID = NormalizeProviderID(provider.ID)
		if provider.ID == "" {
			return nil, errOAuthProviderInvalid
		}
		if _, ok := seen[provider.ID]; ok {
			return nil, fmt.Errorf("%w: duplicate provider %q", errOAuthProviderInvalid, provider.ID)
		}
		seen[provider.ID] = struct{}{}
		if provider.ID == infraoauth.GitHubProvider {
			legacyGitHub = &provider
			continue
		}
		customProviders = append(customProviders, provider)
	}

	githubClientID := strings.TrimSpace(values[githubOAuthClientIDSettingKey])
	githubClientSecret := strings.TrimSpace(values[githubOAuthClientSecretSettingKey])
	if githubClientID != "" || githubClientSecret != "" {
		customProviders = append(customProviders, ProviderConfig{
			ID:           infraoauth.GitHubProvider,
			Name:         "GitHub",
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
		})
	} else if legacyGitHub != nil {
		customProviders = append(customProviders, *legacyGitHub)
	}
	if rawOIDCProvider := strings.TrimSpace(values[oidcProviderSettingKey]); rawOIDCProvider != "" {
		var oidcProvider OIDCProviderConfig
		if err := json.Unmarshal([]byte(rawOIDCProvider), &oidcProvider); err != nil {
			return nil, fmt.Errorf("%w: %v", errOAuthProviderInvalid, err)
		}
		customProviders = append(customProviders, oidcProvider.providerConfig())
	}
	return customProviders, nil
}

func (config OIDCProviderConfig) providerConfig() ProviderConfig {
	return ProviderConfig{
		Kind:         providerKindOIDC,
		ID:           customOIDCProviderID,
		Name:         strings.TrimSpace(config.Name),
		Issuer:       strings.TrimSpace(config.Issuer),
		ClientID:     strings.TrimSpace(config.ClientID),
		ClientSecret: strings.TrimSpace(config.ClientSecret),
		Scopes:       strings.TrimSpace(config.Scopes),
	}
}

// Providers returns the validated provider list for display.
func (s *Service) Providers(ctx context.Context) ([]ConfiguredProvider, error) {
	providers, err := s.ConfiguredProviders(ctx)
	if err != nil {
		return nil, err
	}
	configured := make([]ConfiguredProvider, 0, len(providers))
	for _, config := range providers {
		provider := ConfiguredProvider{ProviderConfig: config, client: s.http}
		if provider.validate() == nil {
			configured = append(configured, provider)
		}
	}
	return configured, nil
}

// Provider returns a single validated provider by ID.
func (s *Service) Provider(ctx context.Context, rawID string) (*ConfiguredProvider, error) {
	providerID := NormalizeProviderID(rawID)
	if providerID == "" {
		return nil, errOAuthProviderNotFound
	}
	providers, err := s.ConfiguredProviders(ctx)
	if err != nil {
		return nil, err
	}
	for _, config := range providers {
		if config.ID != providerID {
			continue
		}
		provider := &ConfiguredProvider{ProviderConfig: config, client: s.http}
		if err := provider.validate(); err != nil {
			return nil, err
		}
		return provider, nil
	}
	return nil, errOAuthProviderNotFound
}

func (p *ConfiguredProvider) validate() error {
	if !ValidProviderID(p.ID) || strings.TrimSpace(p.ClientID) == "" || strings.TrimSpace(p.ClientSecret) == "" {
		return errOAuthProviderInvalid
	}
	if p.Kind == providerKindOIDC {
		issuer, err := infraoauth.ParseHTTPURL(p.Issuer)
		if err != nil || issuer.Scheme != "https" || issuer.RawQuery != "" || issuer.Fragment != "" {
			return errOAuthProviderInvalid
		}
		return nil
	}
	if p.Kind != "" {
		return errOAuthProviderInvalid
	}
	if p.ID == infraoauth.GitHubProvider {
		return nil
	}
	for _, rawURL := range []string{p.AuthorizationURL, p.TokenURL, p.UserInfoURL} {
		parsed, err := url.Parse(strings.TrimSpace(rawURL))
		if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.User != nil {
			return errOAuthProviderInvalid
		}
	}
	return nil
}

func (p *ConfiguredProvider) DisplayName() string {
	if name := strings.TrimSpace(p.Name); name != "" {
		return name
	}
	return p.ID
}

func (p *ConfiguredProvider) BuildAuthorizationURL(ctx context.Context, state OAuthState, redirectURI string) (string, error) {
	infraState := toInfraState(state)
	if p.ID == infraoauth.GitHubProvider {
		github := p.github()
		return github.AuthorizationURL(infraState, redirectURI)
	}
	if p.Kind == providerKindOIDC {
		provider, err := p.oidcProvider(ctx)
		if err != nil {
			return "", err
		}
		config := oauth2.Config{ClientID: p.ClientID, ClientSecret: p.ClientSecret, RedirectURL: redirectURI, Endpoint: provider.Endpoint(), Scopes: oidcScopes(p.Scopes)}
		return config.AuthCodeURL(state.Value, oauth2.S256ChallengeOption(state.CodeVerifier), oidc.Nonce(state.Nonce)), nil
	}
	endpoint, err := url.Parse(p.AuthorizationURL)
	if err != nil {
		return "", fmt.Errorf("oauth: parse authorization URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("client_id", p.ClientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("response_type", "code")
	query.Set("scope", infraoauth.NormalizedScopes(p.Scopes))
	query.Set("state", state.Value)
	query.Set("code_challenge", state.CodeChallenge)
	query.Set("code_challenge_method", "S256")
	query.Set("nonce", state.Nonce)
	endpoint.RawQuery = query.Encode()
	return endpoint.String(), nil
}

func (p *ConfiguredProvider) Authenticate(ctx context.Context, code, redirectURI string, state OAuthState) (infraoauth.NormalizedProfile, infraoauth.AccountTokens, error) {
	if p.ID == infraoauth.GitHubProvider {
		githubTokens, err := p.github().Exchange(ctx, code, redirectURI)
		if err != nil {
			return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, err
		}
		profile, err := p.github().Profile(ctx, githubTokens)
		return profile, infraoauth.AccountTokensFromOAuth(githubTokens), err
	}
	if p.Kind == providerKindOIDC {
		return p.authenticateOIDC(ctx, code, redirectURI, state)
	}
	form := url.Values{
		"client_id":     {p.ClientID},
		"client_secret": {p.ClientSecret},
		"code":          {code},
		"code_verifier": {state.CodeVerifier},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	}
	request, err := infraoauth.NewRequestWithContext(ctx, "POST", p.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, fmt.Errorf("oauth: create token request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokens, err := infraoauth.DecodeJSONResponse[infraoauth.OAuthTokens](p.client, request)
	if err != nil {
		return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, fmt.Errorf("oauth: exchange code: %w", err)
	}
	if strings.TrimSpace(tokens.AccessToken) == "" {
		return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, errors.New("oauth: token response has no access token")
	}
	profile, err := p.profile(ctx, tokens)
	if err != nil {
		return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, err
	}
	return profile, infraoauth.AccountTokensFromOAuth(tokens), nil
}

func (p *ConfiguredProvider) UsesPKCE() bool {
	return p.ID != infraoauth.GitHubProvider
}

func (p *ConfiguredProvider) github() *infraoauth.GitHubClient {
	client := infraoauth.NewGitHubClient(p.ClientID, p.ClientSecret, p.client)
	if authorizationURL := strings.TrimSpace(p.AuthorizationURL); authorizationURL != "" {
		client.SetAuthorizationURL(authorizationURL)
	}
	if tokenURL := strings.TrimSpace(p.TokenURL); tokenURL != "" {
		client.SetTokenURL(tokenURL)
	}
	if userInfoURL := strings.TrimSpace(p.UserInfoURL); userInfoURL != "" {
		client.SetProfileURL(userInfoURL)
	}
	if scope := infraoauth.NormalizedScopes(p.Scopes); scope != "" {
		client.SetScope(scope)
	}
	return client
}

func (p *ConfiguredProvider) oidcProvider(ctx context.Context) (*oidc.Provider, error) {
	oidcContext, err := p.oidcContext(ctx)
	if err != nil {
		return nil, err
	}
	return oidc.NewProvider(oidcContext, p.Issuer)
}

func (p *ConfiguredProvider) authenticateOIDC(ctx context.Context, code, redirectURI string, state OAuthState) (infraoauth.NormalizedProfile, infraoauth.AccountTokens, error) {
	provider, err := p.oidcProvider(ctx)
	if err != nil {
		return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, fmt.Errorf("oidc: discover provider: %w", err)
	}
	oidcContext, err := p.oidcContext(ctx)
	if err != nil {
		return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, err
	}
	config := oauth2.Config{ClientID: p.ClientID, ClientSecret: p.ClientSecret, RedirectURL: redirectURI, Endpoint: provider.Endpoint(), Scopes: oidcScopes(p.Scopes)}
	token, err := config.Exchange(oidcContext, code, oauth2.VerifierOption(state.CodeVerifier))
	if err != nil {
		return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, fmt.Errorf("oidc: exchange code: %w", err)
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || strings.TrimSpace(rawIDToken) == "" {
		return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, errors.New("oidc: token response has no ID token")
	}
	idToken, err := provider.Verifier(&oidc.Config{ClientID: p.ClientID}).Verify(oidcContext, rawIDToken)
	if err != nil {
		return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, fmt.Errorf("oidc: verify ID token: %w", err)
	}
	if subtle.ConstantTimeCompare([]byte(idToken.Nonce), []byte(state.Nonce)) != 1 {
		return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, infraoauth.ErrInvalidOAuthState
	}
	var claims struct {
		AuthorizedParty string `json:"azp"`
		Email           string `json:"email"`
		EmailVerified   *bool  `json:"email_verified"`
		Name            string `json:"name"`
		Username        string `json:"preferred_username"`
		Picture         string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, fmt.Errorf("oidc: decode ID token claims: %w", err)
	}
	if len(idToken.Audience) > 1 && strings.TrimSpace(claims.AuthorizedParty) != p.ClientID {
		return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, errors.New("oidc: ID token authorized party does not match client")
	}
	if strings.TrimSpace(idToken.Subject) == "" || strings.TrimSpace(claims.Email) == "" || claims.EmailVerified == nil || !*claims.EmailVerified {
		return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, errors.New("oidc: ID token is missing a verified subject or email")
	}
	username := strings.TrimSpace(claims.Username)
	if username == "" {
		username = strings.Split(strings.TrimSpace(claims.Email), "@")[0]
	}
	return infraoauth.NormalizedProfile{ProviderAccountID: idToken.Subject, Email: strings.ToLower(strings.TrimSpace(claims.Email)), Name: strings.TrimSpace(claims.Name), Username: username, Image: strings.TrimSpace(claims.Picture), EmailVerified: true}, infraoauth.AccountTokens{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, TokenType: token.Type(), Scope: scopeFromToken(token), ExpiresAt: int(token.Expiry.Unix()), IDToken: rawIDToken}, nil
}

func oidcScopes(raw string) []string {
	scopes := []string{oidc.ScopeOpenID}
	for _, scope := range strings.Fields(infraoauth.NormalizedScopes(raw)) {
		if scope != oidc.ScopeOpenID {
			scopes = append(scopes, scope)
		}
	}
	if len(scopes) == 1 {
		scopes = append(scopes, "profile", "email")
	}
	return scopes
}

func (p *ConfiguredProvider) oidcContext(ctx context.Context) (context.Context, error) {
	client, ok := p.client.(*http.Client)
	if !ok {
		return nil, errors.New("oidc: HTTP client must be a net/http client")
	}
	return oidc.ClientContext(ctx, client), nil
}

func scopeFromToken(token *oauth2.Token) string {
	if scope, ok := token.Extra("scope").(string); ok {
		return scope
	}
	return ""
}

func (p *ConfiguredProvider) profile(ctx context.Context, tokens infraoauth.OAuthTokens) (infraoauth.NormalizedProfile, error) {
	request, err := infraoauth.NewRequestWithContext(ctx, "GET", p.UserInfoURL, nil)
	if err != nil {
		return infraoauth.NormalizedProfile{}, fmt.Errorf("oauth: create profile request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	request.Header.Set("Accept", "application/json")
	attributes, err := infraoauth.DecodeJSONResponse[map[string]any](p.client, request)
	if err != nil {
		return infraoauth.NormalizedProfile{}, fmt.Errorf("oauth: fetch profile: %w", err)
	}
	accountID := infraoauth.StringAttribute(attributes, p.UserIDField, "sub", "id")
	email := strings.ToLower(infraoauth.StringAttribute(attributes, p.EmailField, "email"))
	username := infraoauth.StringAttribute(attributes, p.UsernameField, "preferred_username", "username", "login", "name")
	if accountID == "" || email == "" {
		return infraoauth.NormalizedProfile{}, errors.New("oauth: profile is missing account ID or email")
	}
	if verified, exists := infraoauth.BoolAttribute(attributes, "email_verified", "emailVerified"); exists && !verified {
		return infraoauth.NormalizedProfile{}, errors.New("oauth: profile email is not verified")
	}
	if username == "" {
		username = strings.Split(email, "@")[0]
	}
	return infraoauth.NormalizedProfile{
		ProviderAccountID: accountID,
		Email:             email,
		Name:              infraoauth.StringAttribute(attributes, "name", "display_name", "displayName"),
		Username:          username,
		Image:             infraoauth.StringAttribute(attributes, "picture", "avatar_url", "avatarUrl"),
		EmailVerified:     true,
	}, nil
}

// NormalizeProviderID normalizes a provider identifier for lookup.
func NormalizeProviderID(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

// ValidProviderID validates a provider identifier shape.
func ValidProviderID(value string) bool {
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

// ProviderUnavailable reports whether a provider does not exist or has an
// invalid configuration. HTTP handlers choose the outward status code.
func ProviderUnavailable(err error) bool {
	return errors.Is(err, errOAuthProviderNotFound) || errors.Is(err, errOAuthProviderInvalid)
}
