package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const CSRFHeaderName = "X-CSRF-Token"

var ErrCSRF = errors.New("auth: csrf validation failed")

var legacyCookieNames = []string{
	"next-auth.session-token",
	"__Secure-next-auth.session-token",
	"next-auth.callback-url",
	"__Secure-next-auth.callback-url",
	"next-auth.csrf-token",
	"__Host-next-auth.csrf-token",
	"next-auth.pkce.code_verifier",
	"__Secure-next-auth.pkce.code_verifier",
	"next-auth.state",
	"__Secure-next-auth.state",
	"next-auth.nonce",
	"__Secure-next-auth.nonce",
	"authjs.session-token",
	"__Secure-authjs.session-token",
}

func (s *Service) EnsureCSRFToken(w http.ResponseWriter, r *http.Request) (string, error) {
	if r == nil {
		return "", errors.New("auth: csrf request is nil")
	}
	if cookie, err := r.Cookie(s.csrfCookieName); err == nil && validCSRFToken(cookie.Value) {
		return cookie.Value, nil
	}
	rawToken := make([]byte, csrfTokenBytes)
	if _, err := rand.Read(rawToken); err != nil {
		return "", fmt.Errorf("generate csrf token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(rawToken)
	s.setCSRFCookie(w, token)
	return token, nil
}

func (s *Service) ValidateCSRF(r *http.Request) error {
	if r == nil {
		return ErrCSRF
	}
	cookie, err := r.Cookie(s.csrfCookieName)
	if err != nil || !validCSRFToken(cookie.Value) {
		return ErrCSRF
	}
	header := r.Header.Get(CSRFHeaderName)
	if !validCSRFToken(header) || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(header)) != 1 {
		return ErrCSRF
	}
	return nil
}

func (s *Service) ClearCSRFCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.csrfCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(1, 0).UTC(),
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (s *Service) RotateSessionCookie(ctx context.Context, w http.ResponseWriter, current AuthContext) (AuthContext, error) {
	rotated, err := s.RotateSession(ctx, current)
	if err != nil {
		return AuthContext{}, err
	}
	s.SetSessionCookie(w, rotated, rotated.token)
	return rotated, nil
}

func (s *Service) SessionCookieName() string { return s.sessionCookieName }

func (s *Service) CSRFTokenCookieName() string { return s.csrfCookieName }

func (s *Service) SetSessionCookie(w http.ResponseWriter, session AuthContext, token string) {
	maxAge := maxAgeSeconds(session.ExpiresAt.Sub(s.clock.Now()))
	http.SetCookie(w, &http.Cookie{
		Name:     s.sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (s *Service) SetSessionCookieForSession(w http.ResponseWriter, session AuthContext) {
	s.SetSessionCookie(w, session, session.token)
}

func (s *Service) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(1, 0).UTC(),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (s *Service) tokenFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(s.sessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func ClearLegacyCookies(w http.ResponseWriter) {
	for _, name := range legacyCookieNames {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(1, 0).UTC(),
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
	}
}

func maxAgeSeconds(duration time.Duration) int {
	seconds := int(duration / time.Second)
	if seconds < 1 {
		return 1
	}
	return seconds
}

func (s *Service) CSRFProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isSafeMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}
		if err := s.ValidateCSRF(r); err != nil {
			if writeErr := writeError(w, http.StatusForbidden, "csrf_failed"); writeErr != nil {
				return
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Service) setCSRFCookie(w http.ResponseWriter, token string) {
	expiresAt := s.clock.Now().UTC().Add(s.sessionTTL)
	http.SetCookie(w, &http.Cookie{
		Name:     s.csrfCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   maxAgeSeconds(expiresAt.Sub(s.clock.Now())),
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func validCSRFToken(token string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil && len(decoded) == csrfTokenBytes
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
