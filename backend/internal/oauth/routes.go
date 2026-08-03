package oauth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/auth"
	"github.com/shuTwT/nex-api/backend/internal/config"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
)

type SessionIssuer interface {
	CreateSession(context.Context, auth.User) (auth.AuthContext, error)
	SetSessionCookieForSession(http.ResponseWriter, auth.AuthContext)
}

type Handler struct {
	client   *ent.Client
	sessions SessionIssuer
	baseURL  *url.URL
	state    *StateCodec
	http     *http.Client
	mux      *http.ServeMux
	clock    func() time.Time
}

func New(client *ent.Client, sessions SessionIssuer, cfg config.Config) (*Handler, error) {
	if client == nil {
		return nil, errors.New("oauth: ent client is nil")
	}
	baseURL, err := parseHTTPURL(cfg.AppURL)
	if err != nil {
		return nil, fmt.Errorf("oauth: app URL: %w", err)
	}
	secret := append([]byte(nil), cfg.Auth.SessionSecret...)
	if len(secret) == 0 {
		secret = make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return nil, fmt.Errorf("oauth: generate state secret: %w", err)
		}
	}
	state, err := NewStateCodec(secret)
	if err != nil {
		return nil, err
	}
	handler := &Handler{
		client:   client,
		sessions: sessions,
		baseURL:  baseURL,
		state:    state,
		http:     &http.Client{Timeout: 15 * time.Second},
		mux:      http.NewServeMux(),
		clock:    time.Now,
	}
	handler.registerRoutes()
	return handler, nil
}

func NewHandler(client *ent.Client, sessions SessionIssuer, cfg config.Config) (http.Handler, error) {
	handler, err := New(client, sessions, cfg)
	if err != nil {
		return nil, err
	}
	return handler, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("GET /api/auth/providers", h.providers)
	h.mux.HandleFunc("GET /api/auth/signin/{provider}", h.authorize)
	h.mux.HandleFunc("GET /api/auth/callback/{provider}", h.callback)

	// Keep the legacy endpoint aliases while provider definitions move to system
	// settings. New callers should use /signin/{provider} and /callback/{provider}.
	for _, providerID := range []string{"github", "easy1", "easy1auth"} {
		providerID := providerID
		h.mux.HandleFunc("GET /api/auth/"+providerID, func(w http.ResponseWriter, r *http.Request) {
			h.authorizeProvider(w, r, providerID)
		})
		h.mux.HandleFunc("GET /api/auth/"+providerID+"/callback", func(w http.ResponseWriter, r *http.Request) {
			h.callbackProvider(w, r, providerID)
		})
	}
}

func (h *Handler) authorize(w http.ResponseWriter, r *http.Request) {
	h.authorizeProvider(w, r, r.PathValue("provider"))
}

func (h *Handler) callback(w http.ResponseWriter, r *http.Request) {
	h.callbackProvider(w, r, r.PathValue("provider"))
}

func (h *Handler) authorizeProvider(w http.ResponseWriter, r *http.Request, providerID string) {
	provider, err := h.provider(r.Context(), providerID)
	if err != nil {
		h.writeOAuthError(w, oauthProviderStatus(err), "oauth_provider_unavailable")
		return
	}
	returnURL, err := ResolveReturnURL(firstQuery(r, "callbackUrl", "returnTo"), h.baseURL.String())
	if err != nil {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_return_url")
		return
	}
	state, err := h.state.New(provider.ID, returnURL, provider.usesPKCE())
	if err != nil {
		h.writeOAuthError(w, http.StatusInternalServerError, "oauth_unavailable")
		return
	}
	authorizationURL, err := provider.authorizationURL(state, h.callbackURL(provider.ID))
	if err != nil {
		h.writeOAuthError(w, http.StatusBadGateway, "oauth_provider_unavailable")
		return
	}
	h.state.Write(w, state)
	http.Redirect(w, r, authorizationURL, http.StatusFound)
}

func (h *Handler) callbackProvider(w http.ResponseWriter, r *http.Request, providerID string) {
	providerID = normalizeProviderID(providerID)
	state, err := h.state.Read(r, providerID, r.URL.Query().Get("state"))
	h.state.Clear(w)
	if err != nil {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_oauth_state")
		return
	}
	if providerError := r.URL.Query().Get("error"); providerError != "" {
		h.redirectError(w, r, state.ReturnURL, "oauth_denied")
		return
	}
	provider, err := h.provider(r.Context(), providerID)
	if err != nil {
		h.redirectError(w, r, state.ReturnURL, "oauth_provider_unavailable")
		return
	}
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		h.redirectError(w, r, state.ReturnURL, "missing_oauth_code")
		return
	}
	profile, tokens, err := provider.authenticate(r.Context(), code, h.callbackURL(provider.ID), state)
	if err != nil {
		h.redirectError(w, r, state.ReturnURL, "oauth_failed")
		return
	}
	profile.Provider = provider.ID
	accountUser, err := h.provision(r.Context(), profile, tokens)
	if err != nil || h.sessions == nil {
		h.redirectError(w, r, state.ReturnURL, "oauth_failed")
		return
	}
	session, err := h.sessions.CreateSession(r.Context(), auth.User{
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
	returnURL, err := ResolveReturnURL(rawReturnURL, h.baseURL.String())
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
