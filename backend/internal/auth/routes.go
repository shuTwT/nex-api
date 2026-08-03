package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Handler struct {
	service *Service
	mux     *http.ServeMux
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type responseEnvelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewHandler(service *Service) http.Handler {
	handler := &Handler{service: service, mux: http.NewServeMux()}
	handler.registerRoutes()
	return handler
}

func (s *Service) Handler() http.Handler { return NewHandler(s) }

func RegisterRoutes(mux *http.ServeMux, service *Service) error {
	if mux == nil {
		return errors.New("auth: route mux is nil")
	}
	mux.Handle("/api/auth/", NewHandler(service))
	return nil
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("GET /api/auth/csrf", h.csrf)
	h.mux.HandleFunc("POST /api/auth/login", h.login)
	h.mux.HandleFunc("GET /api/auth/me", h.me)
	h.mux.HandleFunc("POST /api/auth/logout", h.logout)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		if err := writeError(w, http.StatusInternalServerError, "auth_unavailable"); err != nil {
			return
		}
		return
	}
	h.service.CutoverMiddleware(h.service.CSRFProtection(h.mux)).ServeHTTP(w, r)
}

func (h *Handler) csrf(w http.ResponseWriter, r *http.Request) {
	token, err := h.service.EnsureCSRFToken(w, r)
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
	var request loginRequest
	if err := decodeJSON(r, &request); err != nil || strings.TrimSpace(request.Email) == "" || request.Password == "" {
		if writeErr := writeError(w, http.StatusBadRequest, "invalid_request"); writeErr != nil {
			return
		}
		return
	}
	key := loginRateLimitKey(r, request.Email)
	allowed, retryAfter := h.service.rateLimiter.allow(key)
	if !allowed {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", maxAgeSeconds(retryAfter)))
		if writeErr := writeError(w, http.StatusTooManyRequests, "login_rate_limited"); writeErr != nil {
			return
		}
		return
	}
	authContext, err := h.service.Login(r.Context(), request.Email, request.Password, h.service.tokenFromRequest(r))
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
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
	h.service.rateLimiter.reset(key)
	h.service.SetSessionCookie(w, authContext, authContext.token)
	if _, err := h.service.EnsureCSRFToken(w, r); err != nil {
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
	authContext, err := h.service.Authenticate(r.Context(), h.service.tokenFromRequest(r))
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
	if err := h.service.Logout(r.Context(), h.service.tokenFromRequest(r)); err != nil {
		if writeErr := writeError(w, http.StatusInternalServerError, "logout_failed"); writeErr != nil {
			return
		}
		return
	}
	h.service.ClearSessionCookie(w)
	h.service.ClearCSRFCookie(w)
	if err := writeJSON(w, http.StatusOK, responseEnvelope{Success: true, Data: map[string]string{"message": "logged out"}}); err != nil {
		return
	}
}

func decodeJSON(r *http.Request, destination any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("auth: multiple json values")
		}
		return err
	}
	return nil
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

type loginRateLimiter struct {
	mu      sync.Mutex
	entries map[string]loginAttempt
	limit   int
	window  time.Duration
	clock   Clock
}

type loginAttempt struct {
	startedAt time.Time
	count     int
}

func WithTokenGenerator(generator func(int) ([]byte, error)) Option {
	return func(options *serviceOptions) {
		if generator != nil {
			options.tokenGenerator = generator
		}
	}
}

func newLoginRateLimiter(limit int, window time.Duration, clock Clock) *loginRateLimiter {
	return &loginRateLimiter{entries: make(map[string]loginAttempt), limit: limit, window: window, clock: clock}
}

func (l *loginRateLimiter) allow(key string) (bool, time.Duration) {
	now := l.clock.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, exists := l.entries[key]
	if !exists || !now.Before(entry.startedAt.Add(l.window)) {
		l.entries[key] = loginAttempt{startedAt: now, count: 1}
		return true, 0
	}
	if entry.count >= l.limit {
		return false, entry.startedAt.Add(l.window).Sub(now)
	}
	entry.count++
	l.entries[key] = entry
	return true, 0
}

func (l *loginRateLimiter) reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}
