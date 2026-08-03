// Package oauth implements the /api/auth/* OAuth endpoints (providers list,
// sign-in redirect and callback). Protocol details live in infra/oauth;
// provider configuration and account provisioning live in service/oauth.
package oauth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	serviceauth "github.com/shuTwT/nex-api/internal/service/auth"
	serviceoauth "github.com/shuTwT/nex-api/internal/service/oauth"
)

// SessionIssuer creates sessions and writes the session cookie after a
// successful OAuth callback.
type SessionIssuer interface {
	CreateSession(context.Context, serviceauth.User) (serviceauth.AuthContext, error)
	SetSessionCookieForSession(http.ResponseWriter, serviceauth.AuthContext)
}

type Config struct {
	AppURL        string
	SessionSecret []byte
}

type Handler struct {
	service  *serviceoauth.Service
	sessions SessionIssuer
	baseURL  *url.URL
	state    *serviceoauth.StateManager
	mux      chi.Router
}

func New(service *serviceoauth.Service, sessions SessionIssuer, value any) (*Handler, error) {
	return newHandler(chi.NewRouter(), service, sessions, value)
}

// RegisterRoutes installs OAuth endpoints directly onto the application's router.
func RegisterRoutes(mux chi.Router, service *serviceoauth.Service, sessions SessionIssuer, value any) error {
	if mux == nil {
		return errors.New("oauth: mux is nil")
	}
	_, err := newHandler(mux, service, sessions, value)
	return err
}

func newHandler(mux chi.Router, service *serviceoauth.Service, sessions SessionIssuer, value any) (*Handler, error) {
	cfg, err := normalizeConfig(value)
	if err != nil {
		return nil, err
	}
	if service == nil {
		return nil, errors.New("oauth: service is nil")
	}
	baseURL, err := url.Parse(cfg.AppURL)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("oauth: app URL: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("oauth: app URL: %w", err)
	}
	secret := append([]byte(nil), cfg.SessionSecret...)
	if len(secret) == 0 {
		secret = make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return nil, fmt.Errorf("oauth: generate state secret: %w", err)
		}
	}
	state, err := serviceoauth.NewStateManager(secret)
	if err != nil {
		return nil, err
	}
	handler := &Handler{
		service:  service,
		sessions: sessions,
		baseURL:  baseURL,
		state:    state,
		mux:      mux,
	}
	handler.registerRoutes()
	return handler, nil
}

func NewHandler(service *serviceoauth.Service, sessions SessionIssuer, cfg any) (http.Handler, error) {
	handler, err := New(service, sessions, cfg)
	if err != nil {
		return nil, err
	}
	return handler, nil
}

func normalizeConfig(value any) (Config, error) {
	if cfg, ok := value.(Config); ok {
		return cfg, nil
	}
	reflected := reflect.ValueOf(value)
	if reflected.Kind() == reflect.Pointer {
		reflected = reflected.Elem()
	}
	if reflected.Kind() != reflect.Struct {
		return Config{}, errors.New("oauth: invalid configuration")
	}
	appURL := reflected.FieldByName("AppURL")
	auth := reflected.FieldByName("Auth")
	if !appURL.IsValid() || appURL.Kind() != reflect.String || !auth.IsValid() {
		return Config{}, errors.New("oauth: invalid configuration")
	}
	if auth.Kind() == reflect.Pointer {
		auth = auth.Elem()
	}
	secret := auth.FieldByName("SessionSecret")
	if !secret.IsValid() {
		return Config{}, errors.New("oauth: invalid configuration")
	}
	var sessionSecret []byte
	switch secret.Kind() {
	case reflect.String:
		sessionSecret = []byte(secret.String())
	case reflect.Slice:
		if secret.Type().Elem().Kind() != reflect.Uint8 {
			return Config{}, errors.New("oauth: invalid configuration")
		}
		sessionSecret = append([]byte(nil), secret.Bytes()...)
	default:
		return Config{}, errors.New("oauth: invalid configuration")
	}
	return Config{AppURL: appURL.String(), SessionSecret: sessionSecret}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) registerRoutes() {
	h.mux.Get("/api/auth/providers", h.providers)
	h.mux.Get("/api/auth/signin/{provider}", h.authorize)
	h.mux.Get("/api/auth/callback/{provider}", h.callback)

	// Keep the legacy endpoint aliases while provider definitions move to system
	// settings. New callers should use /signin/{provider} and /callback/{provider}.
	for _, providerID := range []string{"github", "easy1", "easy1auth"} {
		providerID := providerID
		h.mux.Get("/api/auth/"+providerID, func(w http.ResponseWriter, r *http.Request) {
			h.authorizeProvider(w, r, providerID)
		})
		h.mux.Get("/api/auth/"+providerID+"/callback", func(w http.ResponseWriter, r *http.Request) {
			h.callbackProvider(w, r, providerID)
		})
	}
}

func (h *Handler) authorize(w http.ResponseWriter, r *http.Request) {
	h.authorizeProvider(w, r, chi.URLParam(r, "provider"))
}

func (h *Handler) callback(w http.ResponseWriter, r *http.Request) {
	h.callbackProvider(w, r, chi.URLParam(r, "provider"))
}

func (h *Handler) authorizeProvider(w http.ResponseWriter, r *http.Request, providerID string) {
	provider, err := h.service.Provider(r.Context(), providerID)
	if err != nil {
		status := http.StatusServiceUnavailable
		if serviceoauth.ProviderUnavailable(err) {
			status = http.StatusNotFound
		}
		h.writeOAuthError(w, status, "oauth_provider_unavailable")
		return
	}
	returnURL, err := serviceoauth.ResolveReturnURL(firstQuery(r, "callbackUrl", "returnTo"), h.baseURL.String())
	if err != nil {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_return_url")
		return
	}
	state, cookie, err := h.state.New(provider.ID, returnURL, provider.UsesPKCE())
	if err != nil {
		h.writeOAuthError(w, http.StatusInternalServerError, "oauth_unavailable")
		return
	}
	authorizationURL, err := provider.BuildAuthorizationURL(state, h.callbackURL(provider.ID))
	if err != nil {
		h.writeOAuthError(w, http.StatusBadGateway, "oauth_provider_unavailable")
		return
	}
	h.writeStateCookie(w, cookie)
	http.Redirect(w, r, authorizationURL, http.StatusFound)
}

func (h *Handler) callbackProvider(w http.ResponseWriter, r *http.Request, providerID string) {
	providerID = serviceoauth.NormalizeProviderID(providerID)
	cookie, _ := r.Cookie(h.state.CookieName())
	cookieValue := ""
	if cookie != nil {
		cookieValue = cookie.Value
	}
	state, err := h.state.Read(cookieValue, providerID, r.URL.Query().Get("state"))
	h.clearStateCookie(w)
	if err != nil {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_oauth_state")
		return
	}
	if providerError := r.URL.Query().Get("error"); providerError != "" {
		h.redirectError(w, r, state.ReturnURL, "oauth_denied")
		return
	}
	provider, err := h.service.Provider(r.Context(), providerID)
	if err != nil {
		h.redirectError(w, r, state.ReturnURL, "oauth_provider_unavailable")
		return
	}
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		h.redirectError(w, r, state.ReturnURL, "missing_oauth_code")
		return
	}
	profile, tokens, err := provider.Authenticate(r.Context(), code, h.callbackURL(provider.ID), state)
	if err != nil {
		h.redirectError(w, r, state.ReturnURL, "oauth_failed")
		return
	}
	profile.Provider = provider.ID
	accountUser, err := h.service.Provision(r.Context(), profile, tokens)
	if err != nil || h.sessions == nil {
		h.redirectError(w, r, state.ReturnURL, "oauth_failed")
		return
	}
	session, err := h.sessions.CreateSession(r.Context(), serviceauth.User{
		ID:       accountUser.ID,
		Email:    accountUser.Email,
		Username: accountUser.Username,
		Role:     accountUser.Role,
		Credits:  accountUser.Credits,
	})
	if err != nil {
		h.redirectError(w, r, state.ReturnURL, "oauth_failed")
		return
	}
	h.sessions.SetSessionCookieForSession(w, session)
	http.Redirect(w, r, state.ReturnURL, http.StatusFound)
}

func (h *Handler) callbackURL(provider string) string {
	callbackPath := "/api/auth/callback/" + provider
	basePath := strings.TrimRight(h.baseURL.Path, "/")
	return strings.TrimRight(h.baseURL.Scheme+"://"+h.baseURL.Host+basePath, "/") + callbackPath
}

func (h *Handler) redirectError(w http.ResponseWriter, r *http.Request, rawReturnURL, code string) {
	returnURL, err := serviceoauth.ResolveReturnURL(rawReturnURL, h.baseURL.String())
	if err != nil {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_return_url")
		return
	}
	parsed, err := url.Parse(returnURL)
	if err != nil {
		h.writeOAuthError(w, http.StatusInternalServerError, "oauth_failed")
		return
	}
	query := parsed.Query()
	query.Set("error", code)
	parsed.RawQuery = query.Encode()
	http.Redirect(w, r, parsed.String(), http.StatusFound)
}

func (h *Handler) writeStateCookie(w http.ResponseWriter, cookie serviceoauth.StateCookie) {
	http.SetCookie(w, &http.Cookie{Name: cookie.Name, Value: cookie.Value, Path: cookie.Path, Expires: cookie.Expires, MaxAge: cookie.MaxAge, HttpOnly: cookie.HttpOnly, Secure: cookie.Secure, SameSite: http.SameSiteLaxMode})
}
func (h *Handler) clearStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: h.state.CookieName(), Value: "", Path: "/", Expires: time.Unix(1, 0).UTC(), MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
}

func (h *Handler) writeOAuthError(w http.ResponseWriter, status int, code string) {
	http.Error(w, code, status)
}

func firstQuery(r *http.Request, names ...string) string {
	for _, name := range names {
		if value := r.URL.Query().Get(name); value != "" {
			return value
		}
	}
	return ""
}
