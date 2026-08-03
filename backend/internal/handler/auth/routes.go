// Package auth implements the /api/auth/* endpoints: CSRF token issuance,
// login, current user and logout. Session/cookie and CSRF request adaptation
// lives here and in the middleware package; service/auth only handles
// credentials, sessions and user state.
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shuTwT/nex-api/backend/internal/middleware"
	handlerutils "github.com/shuTwT/nex-api/backend/internal/pkg/utils"
	serviceauth "github.com/shuTwT/nex-api/backend/internal/service/auth"
	"github.com/shuTwT/nex-api/backend/pkg/domain/model"
)

type Handler struct {
	service *serviceauth.Service
	csrf    *middleware.CSRFProtector
	mux     chi.Router
}

type responseEnvelope = model.EnvelopeResp

func NewHandler(service *serviceauth.Service) http.Handler {
	return newHandler(service)
}

func newHandler(service *serviceauth.Service) *Handler {
	handler := &Handler{
		service: service,
		csrf: middleware.NewCSRFProtector(
			service.CSRFTokenCookieName(),
			service.SecureCookies(),
			service.SessionTTL(),
		),
		mux: chi.NewRouter(),
	}
	handler.registerRoutes(handler.mux)
	return handler
}

func RegisterRoutes(r chi.Router, service *serviceauth.Service) error {
	if r == nil {
		return errors.New("auth: route mux is nil")
	}
	handler := newHandler(service)
	r.Group(func(r chi.Router) {
		r.Use(handler.csrf.Middleware)
		handler.registerRoutes(r)
	})
	return nil
}

func (h *Handler) registerRoutes(r chi.Router) {
	r.Get("/api/auth/csrf", h.csrfToken)
	r.Post("/api/auth/login", h.login)
	r.Get("/api/auth/me", h.me)
	r.Post("/api/auth/logout", h.logout)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		if err := writeError(w, http.StatusInternalServerError, "auth_unavailable"); err != nil {
			return
		}
		return
	}
	h.csrf.Middleware(h.mux).ServeHTTP(w, r)
}

func (h *Handler) csrfToken(w http.ResponseWriter, r *http.Request) {
	token, err := h.csrf.EnsureToken(w, r)
	if err != nil {
		if writeErr := writeError(w, http.StatusInternalServerError, "csrf_unavailable"); writeErr != nil {
			return
		}
		return
	}
	if err := writeJSON(w, http.StatusOK, responseEnvelope{Success: true, Data: map[string]string{"token": token}}); err != nil {
		return
	}
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var request model.AuthLoginReq
	if err := handlerutils.DecodeJSON(r, &request); err != nil || strings.TrimSpace(request.Email) == "" || request.Password == "" {
		if writeErr := writeError(w, http.StatusBadRequest, "invalid_request"); writeErr != nil {
			return
		}
		return
	}
	key := loginRateLimitKey(r, request.Email)
	allowed, retryAfter := h.service.AllowLogin(key)
	if !allowed {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", maxAgeSeconds(retryAfter)))
		if writeErr := writeError(w, http.StatusTooManyRequests, "login_rate_limited"); writeErr != nil {
			return
		}
		return
	}
	authContext, err := h.service.Login(r.Context(), request.Email, request.Password, h.tokenFromRequest(r))
	if err != nil {
		if errors.Is(err, serviceauth.ErrInvalidCredentials) {
			if writeErr := writeError(w, http.StatusUnauthorized, "invalid_credentials"); writeErr != nil {
				return
			}
			return
		}
		if writeErr := writeError(w, http.StatusInternalServerError, "login_failed"); writeErr != nil {
			return
		}
		return
	}
	h.service.ResetLogin(key)
	h.setSessionCookie(w, authContext)
	if _, err := h.csrf.EnsureToken(w, r); err != nil {
		if writeErr := writeError(w, http.StatusInternalServerError, "csrf_unavailable"); writeErr != nil {
			return
		}
		return
	}
	if err := writeJSON(w, http.StatusOK, responseEnvelope{Success: true, Data: authContext.User}); err != nil {
		return
	}
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	authContext, err := h.service.Authenticate(r.Context(), h.tokenFromRequest(r))
	if err != nil {
		if writeErr := writeError(w, http.StatusUnauthorized, "unauthorized"); writeErr != nil {
			return
		}
		return
	}
	if err := writeJSON(w, http.StatusOK, responseEnvelope{Success: true, Data: authContext.User}); err != nil {
		return
	}
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Logout(r.Context(), h.tokenFromRequest(r)); err != nil {
		if writeErr := writeError(w, http.StatusInternalServerError, "logout_failed"); writeErr != nil {
			return
		}
		return
	}
	h.clearSessionCookie(w)
	h.csrf.ClearCookie(w)
	if err := writeJSON(w, http.StatusOK, responseEnvelope{Success: true, Data: map[string]string{"message": "logged out"}}); err != nil {
		return
	}
}

func (h *Handler) tokenFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(h.service.SessionCookieName())
	if err != nil {
		return ""
	}
	return cookie.Value
}

func writeJSON(w http.ResponseWriter, status int, value responseEnvelope) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code string) error {
	return writeJSON(w, status, responseEnvelope{Success: false, Error: code})
}

func loginRateLimitKey(r *http.Request, email string) string {
	remoteIP := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		remoteIP = host
	}
	return strings.ToLower(strings.TrimSpace(email)) + "|" + remoteIP
}

func maxAgeSeconds(duration time.Duration) int {
	seconds := int(duration / time.Second)
	if seconds < 1 {
		return 1
	}
	return seconds
}
