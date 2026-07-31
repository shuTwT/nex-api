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
	github   *githubClient
	easy1    *easy1Client
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
		secret = append(secret, cfg.Auth.NextAuthSecret...)
	}
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
	clientHTTP := &http.Client{Timeout: 15 * time.Second}
	github := newGitHubClient(cfg.Auth.GitHubClientID, cfg.Auth.GitHubClientSecret, clientHTTP)
	var easy1 *easy1Client
	if strings.TrimSpace(cfg.Auth.SSOClientID) != "" || strings.TrimSpace(cfg.Auth.SSOClientSecret) != "" || strings.TrimSpace(cfg.Auth.SSOAuthorizationURL) != "" {
		easy1, err = newEasy1Client(
			cfg.Auth.SSOClientID,
			cfg.Auth.SSOClientSecret,
			cfg.Auth.SSOAuthorizationURL,
			cfg.Auth.SSOTokenURL,
			cfg.Auth.SSOUserInfoURL,
			cfg.Auth.SSOScope,
			clientHTTP,
		)
		if err != nil {
			return nil, err
		}
	}
	handler := &Handler{
		client:   client,
		sessions: sessions,
		baseURL:  baseURL,
		state:    state,
		github:   github,
		easy1:    easy1,
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
	for _, provider := range []string{githubProvider, easy1Provider} {
		h.mux.HandleFunc("GET /api/auth/signin/"+provider, h.authorize(provider))
		h.mux.HandleFunc("GET /api/auth/callback/"+provider, h.callback(provider))
		h.mux.HandleFunc("GET /api/auth/"+provider, h.authorize(provider))
		h.mux.HandleFunc("GET /api/auth/"+provider+"/callback", h.callback(provider))
	}
	h.mux.HandleFunc("GET /api/auth/signin/easy1", h.authorize(easy1Provider))
	h.mux.HandleFunc("GET /api/auth/callback/easy1", h.callback(easy1Provider))
}

func (h *Handler) authorize(provider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		returnURL, err := ResolveReturnURL(firstQuery(r, "callbackUrl", "returnTo"), h.baseURL.String())
		if err != nil {
			h.writeOAuthError(w, http.StatusBadRequest, "invalid_return_url")
			return
		}
		state, err := h.state.New(provider, returnURL, provider == easy1Provider)
		if err != nil {
			h.writeOAuthError(w, http.StatusInternalServerError, "oauth_unavailable")
			return
		}
		redirectURI := h.callbackURL(provider)
		var authorizationURL string
		switch provider {
		case githubProvider:
			authorizationURL, err = h.github.authorizationURL(state, redirectURI)
		case easy1Provider:
			if h.easy1 == nil {
				h.writeOAuthError(w, http.StatusNotFound, "oauth_provider_unavailable")
				return
			}
			authorizationURL, err = h.easy1.authorizationURL(state, redirectURI)
		}
		if err != nil {
			h.writeOAuthError(w, http.StatusBadGateway, "oauth_provider_unavailable")
			return
		}
		h.state.Write(w, state)
		http.Redirect(w, r, authorizationURL, http.StatusFound)
	}
}

func (h *Handler) callback(provider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := h.state.Read(r, provider, r.URL.Query().Get("state"))
		h.state.Clear(w)
		if err != nil {
			h.writeOAuthError(w, http.StatusBadRequest, "invalid_oauth_state")
			return
		}
		if providerError := r.URL.Query().Get("error"); providerError != "" {
			h.redirectError(w, r, state.ReturnURL, "oauth_denied")
			return
		}
		code := strings.TrimSpace(r.URL.Query().Get("code"))
		if code == "" {
			h.redirectError(w, r, state.ReturnURL, "missing_oauth_code")
			return
		}
		redirectURI := h.callbackURL(provider)
		var profile normalizedProfile
		var tokens accountTokens
		switch provider {
		case githubProvider:
			githubTokens, exchangeErr := h.github.exchange(r.Context(), code, redirectURI)
			if exchangeErr == nil {
				profile, exchangeErr = h.github.profile(r.Context(), githubTokens)
				tokens = accountTokensFromOAuth(githubTokens)
			}
			err = exchangeErr
		case easy1Provider:
			easy1Tokens, exchangeErr := h.easy1.exchange(r.Context(), code, redirectURI, state.CodeVerifier)
			if exchangeErr == nil {
				profile, tokens, exchangeErr = h.easy1.profile(r.Context(), easy1Tokens, state.Nonce)
			}
			err = exchangeErr
		}
		if err != nil {
			h.redirectError(w, r, state.ReturnURL, "oauth_failed")
			return
		}
		profile.Provider = provider
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
