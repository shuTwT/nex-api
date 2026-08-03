package auth

import (
	"context"
	"net/http"
	"time"

	serviceauth "github.com/shuTwT/nex-api/backend/internal/service/auth"
)

// setSessionCookie writes the session cookie for the given auth context.
func (h *Handler) setSessionCookie(w http.ResponseWriter, session serviceauth.AuthContext) {
	maxAge := maxAgeSeconds(session.ExpiresAt.Sub(time.Now()))
	http.SetCookie(w, &http.Cookie{
		Name:     h.service.SessionCookieName(),
		Value:    session.Token(),
		Path:     "/",
		Expires:  session.ExpiresAt,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.service.SecureCookies(),
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.service.SessionCookieName(),
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(1, 0).UTC(),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.service.SecureCookies(),
		SameSite: http.SameSiteStrictMode,
	})
}

// SessionCookieWriter adapts the auth service for OAuth flows: it creates
// sessions and writes the session cookie, satisfying the session issuer
// interface expected by handler/oauth.
type SessionCookieWriter struct {
	service *serviceauth.Service
}

func NewSessionCookieWriter(service *serviceauth.Service) *SessionCookieWriter {
	return &SessionCookieWriter{service: service}
}

func (w *SessionCookieWriter) CreateSession(ctx context.Context, user serviceauth.User) (serviceauth.AuthContext, error) {
	return w.service.CreateSession(ctx, user)
}

func (w *SessionCookieWriter) SetSessionCookieForSession(rw http.ResponseWriter, session serviceauth.AuthContext) {
	maxAge := maxAgeSeconds(session.ExpiresAt.Sub(time.Now()))
	http.SetCookie(rw, &http.Cookie{
		Name:     w.service.SessionCookieName(),
		Value:    session.Token(),
		Path:     "/",
		Expires:  session.ExpiresAt,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   w.service.SecureCookies(),
		SameSite: http.SameSiteStrictMode,
	})
}
