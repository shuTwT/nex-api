package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/backend/ent"
	"github.com/shuTwT/nex-api/backend/ent/systemsetting"
	infraoauth "github.com/shuTwT/nex-api/backend/internal/infra/oauth"
)

const oauthProvidersSettingKey = "oauthProviders"

var (
	errOAuthProviderNotFound = errors.New("oauth: provider not configured")
	errOAuthProviderInvalid  = errors.New("oauth: provider configuration is invalid")
)

// ProviderConfig is persisted as one entry in SystemSetting.oauthProviders.
// RoleField is deliberately not used: an upstream identity provider must never
// be able to grant a local administrative role.
type ProviderConfig struct {
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
	setting, err := s.client.SystemSetting.Query().Where(systemsetting.Key(oauthProvidersSettingKey)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("oauth: load providers: %w", err)
	}
	if strings.TrimSpace(setting.Value) == "" {
		return nil, nil
	}
	var providers []ProviderConfig
	if err := json.Unmarshal([]byte(setting.Value), &providers); err != nil {
		return nil, fmt.Errorf("%w: %v", errOAuthProviderInvalid, err)
	}
	seen := make(map[string]struct{}, len(providers))
	for index := range providers {
		providers[index].ID = NormalizeProviderID(providers[index].ID)
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

func (p *ConfiguredProvider) BuildAuthorizationURL(state OAuthState, redirectURI string) (string, error) {
	infraState := toInfraState(state)
	if p.ID == infraoauth.GitHubProvider {
		github := p.github()
		return github.AuthorizationURL(infraState, redirectURI)
	}
	if p.isEasy1() {
		easy1, err := p.easy1()
		if err != nil {
			return "", err
		}
		return easy1.AuthorizationURL(infraState, redirectURI)
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
	infraState := toInfraState(state)
	if p.ID == infraoauth.GitHubProvider {
		githubTokens, err := p.github().Exchange(ctx, code, redirectURI)
		if err != nil {
			return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, err
		}
		profile, err := p.github().Profile(ctx, githubTokens)
		return profile, infraoauth.AccountTokensFromOAuth(githubTokens), err
	}
	if p.isEasy1() {
		easy1, err := p.easy1()
		if err != nil {
			return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, err
		}
		easy1Tokens, err := easy1.Exchange(ctx, code, redirectURI, state.CodeVerifier)
		if err != nil {
			return infraoauth.NormalizedProfile{}, infraoauth.AccountTokens{}, err
		}
		return easy1.Profile(ctx, easy1Tokens, infraState.Nonce)
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

func (p *ConfiguredProvider) isEasy1() bool {
	return p.ID == infraoauth.Easy1Provider || p.ID == "easy1"
}

func (p *ConfiguredProvider) github() *infraoauth.GitHubClient {
	client := infraoauth.NewGitHubClient(p.ClientID, p.ClientSecret, p.client)
	client.SetAuthorizationURL(p.AuthorizationURL)
	client.SetTokenURL(p.TokenURL)
	client.SetProfileURL(p.UserInfoURL)
	if scope := infraoauth.NormalizedScopes(p.Scopes); scope != "" {
		client.SetScope(scope)
	}
	return client
}

func (p *ConfiguredProvider) easy1() (*infraoauth.Easy1Client, error) {
	return infraoauth.NewEasy1Client(
		p.ClientID,
		p.ClientSecret,
		p.AuthorizationURL,
		p.TokenURL,
		p.UserInfoURL,
		infraoauth.NormalizedScopes(p.Scopes),
		p.client,
	)
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
