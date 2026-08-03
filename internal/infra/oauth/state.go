package oauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	StateCookieName = "nex_oauth_state"
	stateTTL        = 10 * time.Minute
)

var (
	ErrInvalidOAuthState = errors.New("oauth: invalid state")
	ErrUnsafeReturnURL   = errors.New("oauth: unsafe return URL")
)

type OAuthState struct {
	Provider      string
	Value         string
	ReturnURL     string
	CodeVerifier  string
	CodeChallenge string
	Nonce         string
	IssuedAt      time.Time
}

type stateCookie struct {
	Provider     string `json:"provider"`
	Value        string `json:"value"`
	ReturnURL    string `json:"return_url"`
	CodeVerifier string `json:"code_verifier,omitempty"`
	Nonce        string `json:"nonce,omitempty"`
	IssuedAt     int64  `json:"issued_at"`
}

type StateCodec struct {
	secret     []byte
	cookieName string
	ttl        time.Duration
}

type StateCookie struct {
	Name, Value, Path string
	Expires           time.Time
	MaxAge            int
	HttpOnly, Secure  bool
}

func (c *StateCodec) CookieName() string {
	if c == nil {
		return StateCookieName
	}
	return c.cookieName
}

func (c *StateCodec) Encode(state OAuthState) (StateCookie, error) {
	if c == nil {
		return StateCookie{}, ErrInvalidOAuthState
	}
	payload, err := json.Marshal(stateCookie{Provider: state.Provider, Value: state.Value, ReturnURL: state.ReturnURL, CodeVerifier: state.CodeVerifier, Nonce: state.Nonce, IssuedAt: state.IssuedAt.Unix()})
	if err != nil {
		return StateCookie{}, err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return StateCookie{Name: c.cookieName, Value: encoded + "." + c.sign(encoded), Path: "/", Expires: state.IssuedAt.Add(c.ttl), MaxAge: int(c.ttl / time.Second), HttpOnly: true, Secure: true}, nil
}

func (c *StateCodec) Decode(cookieValue, provider, expectedValue string) (OAuthState, error) {
	if c == nil {
		return OAuthState{}, ErrInvalidOAuthState
	}
	parts := strings.Split(cookieValue, ".")
	if len(parts) != 2 || subtle.ConstantTimeCompare([]byte(parts[1]), []byte(c.sign(parts[0]))) != 1 {
		return OAuthState{}, ErrInvalidOAuthState
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return OAuthState{}, fmt.Errorf("oauth: decode state: %w", ErrInvalidOAuthState)
	}
	var stored stateCookie
	if err := json.Unmarshal(payload, &stored); err != nil {
		return OAuthState{}, fmt.Errorf("oauth: parse state: %w", ErrInvalidOAuthState)
	}
	issuedAt := time.Unix(stored.IssuedAt, 0).UTC()
	now := time.Now().UTC()
	if stored.Provider != provider || stored.Value == "" || stored.Value != expectedValue || stored.IssuedAt <= 0 || now.Before(issuedAt) || now.Sub(issuedAt) > c.ttl {
		return OAuthState{}, ErrInvalidOAuthState
	}
	state := OAuthState{Provider: stored.Provider, Value: stored.Value, ReturnURL: stored.ReturnURL, CodeVerifier: stored.CodeVerifier, Nonce: stored.Nonce, IssuedAt: issuedAt}
	if stored.CodeVerifier != "" {
		hash := sha256.Sum256([]byte(stored.CodeVerifier))
		state.CodeChallenge = base64.RawURLEncoding.EncodeToString(hash[:])
	}
	return state, nil
}

func NewStateCodec(secret []byte) (*StateCodec, error) {
	if len(secret) == 0 {
		return nil, errors.New("oauth: state secret is empty")
	}
	return &StateCodec{
		secret:     append([]byte(nil), secret...),
		cookieName: StateCookieName,
		ttl:        stateTTL,
	}, nil
}

func (c *StateCodec) New(provider, returnURL string, withPKCE bool) (OAuthState, error) {
	if c == nil || len(c.secret) == 0 {
		return OAuthState{}, errors.New("oauth: state codec is not configured")
	}
	provider = strings.TrimSpace(provider)
	if provider == "" || returnURL == "" {
		return OAuthState{}, fmt.Errorf("oauth: state inputs: %w", ErrInvalidOAuthState)
	}
	stateValue, err := RandomToken(32)
	if err != nil {
		return OAuthState{}, fmt.Errorf("oauth: generate state: %w", err)
	}
	nonce, err := RandomToken(32)
	if err != nil {
		return OAuthState{}, fmt.Errorf("oauth: generate nonce: %w", err)
	}
	state := OAuthState{
		Provider:  provider,
		Value:     stateValue,
		ReturnURL: returnURL,
		Nonce:     nonce,
		IssuedAt:  time.Now().UTC(),
	}
	if withPKCE {
		state.CodeVerifier, err = RandomToken(32)
		if err != nil {
			return OAuthState{}, fmt.Errorf("oauth: generate PKCE verifier: %w", err)
		}
		hash := sha256.Sum256([]byte(state.CodeVerifier))
		state.CodeChallenge = base64.RawURLEncoding.EncodeToString(hash[:])
	}
	return state, nil
}

func (c *StateCodec) Write(w http.ResponseWriter, state OAuthState) {
	if c == nil || w == nil {
		return
	}
	cookie, err := c.Encode(state)
	if err != nil {
		return
	}
	http.SetCookie(w, &http.Cookie{Name: cookie.Name, Value: cookie.Value, Path: cookie.Path, Expires: cookie.Expires, MaxAge: cookie.MaxAge, HttpOnly: cookie.HttpOnly, Secure: cookie.Secure, SameSite: http.SameSiteLaxMode})
}

func (c *StateCodec) Read(r *http.Request, provider, expectedValue string) (OAuthState, error) {
	if c == nil || r == nil {
		return OAuthState{}, ErrInvalidOAuthState
	}
	cookie, err := r.Cookie(c.cookieName)
	if err != nil {
		return OAuthState{}, fmt.Errorf("oauth: state cookie: %w", ErrInvalidOAuthState)
	}
	return c.Decode(cookie.Value, provider, expectedValue)
}

func (c *StateCodec) Clear(w http.ResponseWriter) {
	if c == nil || w == nil {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     c.cookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(1, 0).UTC(),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (c *StateCodec) sign(value string) string {
	digest := hmac.New(sha256.New, c.secret)
	_, _ = digest.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(digest.Sum(nil))
}

func ResolveReturnURL(raw, base string) (string, error) {
	baseURL, err := ParseHTTPURL(base)
	if err != nil {
		return "", fmt.Errorf("oauth: parse base URL: %w", err)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return baseURL.String(), nil
	}
	if strings.HasPrefix(raw, "//") || strings.ContainsAny(raw, "\r\n") {
		return "", ErrUnsafeReturnURL
	}
	returnURL, err := url.Parse(raw)
	if err != nil || returnURL.User != nil || returnURL.Opaque != "" {
		return "", ErrUnsafeReturnURL
	}
	if !returnURL.IsAbs() {
		if returnURL.Host != "" || returnURL.Scheme != "" || !strings.HasPrefix(returnURL.Path, "/") {
			return "", ErrUnsafeReturnURL
		}
		returnURL = baseURL.ResolveReference(returnURL)
	}
	if !sameOrigin(baseURL, returnURL) {
		return "", ErrUnsafeReturnURL
	}
	return returnURL.String(), nil
}

func ValidateOIDCNonce(idToken, expectedNonce string) error {
	if strings.TrimSpace(idToken) == "" || strings.TrimSpace(expectedNonce) == "" {
		return ErrInvalidOAuthState
	}
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return fmt.Errorf("oauth: invalid ID token: %w", ErrInvalidOAuthState)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("oauth: decode ID token: %w", ErrInvalidOAuthState)
	}
	var claims struct {
		Nonce string `json:"nonce"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Nonce == "" || subtle.ConstantTimeCompare([]byte(claims.Nonce), []byte(expectedNonce)) != 1 {
		return ErrInvalidOAuthState
	}
	return nil
}

func RandomToken(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func ParseHTTPURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.User != nil || !parsed.IsAbs() || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, ErrUnsafeReturnURL
	}
	return parsed, nil
}

func sameOrigin(left, right *url.URL) bool {
	return strings.EqualFold(left.Scheme, right.Scheme) &&
		strings.EqualFold(left.Hostname(), right.Hostname()) &&
		left.Port() == right.Port()
}
