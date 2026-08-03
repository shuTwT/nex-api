package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	serviceauth "github.com/shuTwT/nex-api/backend/internal/service/auth"
	"github.com/shuTwT/nex-api/backend/internal/service/authz"
)

// CSRFHeaderName is the header that must echo the CSRF cookie value on
// state-changing requests.
const CSRFHeaderName = "X-CSRF-Token"

// ErrCSRF reports a failed CSRF validation.
var ErrCSRF = errors.New("auth: csrf validation failed")

const csrfTokenBytes = 32

// CSRFProtector implements the double-submit CSRF cookie/header protection.
// It is constructed by the handler layer with the cookie configuration owned
// by the auth service.
type CSRFProtector struct {
	cookieName string
	secure     bool
	ttl        time.Duration
	clock      func() time.Time
}

func NewCSRFProtector(cookieName string, secure bool, ttl time.Duration) *CSRFProtector {
	clock := time.Now
	return &CSRFProtector{cookieName: cookieName, secure: secure, ttl: ttl, clock: clock}
}

// EnsureToken returns the existing valid CSRF token or issues a new cookie.
func (p *CSRFProtector) EnsureToken(w http.ResponseWriter, r *http.Request) (string, error) {
	if r == nil {
		return "", errors.New("auth: csrf request is nil")
	}
	if cookie, err := r.Cookie(p.cookieName); err == nil && validCSRFToken(cookie.Value) {
		return cookie.Value, nil
	}
	rawToken := make([]byte, csrfTokenBytes)
	if _, err := rand.Read(rawToken); err != nil {
		return "", fmt.Errorf("generate csrf token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(rawToken)
	p.setCookie(w, token)
	return token, nil
}

// Validate checks the cookie/header pair on a request.
func (p *CSRFProtector) Validate(r *http.Request) error {
	if r == nil {
		return ErrCSRF
	}
	cookie, err := r.Cookie(p.cookieName)
	if err != nil || !validCSRFToken(cookie.Value) {
		return ErrCSRF
	}
	header := r.Header.Get(CSRFHeaderName)
	if !validCSRFToken(header) || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(header)) != 1 {
		return ErrCSRF
	}
	return nil
}

// ClearCookie removes the CSRF cookie (used on logout).
func (p *CSRFProtector) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     p.cookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(1, 0).UTC(),
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   p.secure,
		SameSite: http.SameSiteStrictMode,
	})
}

// Middleware enforces CSRF validation on unsafe methods.
func (p *CSRFProtector) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isSafeMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}
		if err := p.Validate(r); err != nil {
			writeErrorResponse(w, r, http.StatusForbidden, "csrf_failed")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (p *CSRFProtector) setCookie(w http.ResponseWriter, token string) {
	expiresAt := p.clock().UTC().Add(p.ttl)
	http.SetCookie(w, &http.Cookie{
		Name:     p.cookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   maxAgeSeconds(expiresAt.Sub(p.clock())),
		HttpOnly: false,
		Secure:   p.secure,
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

func maxAgeSeconds(duration time.Duration) int {
	seconds := int(duration / time.Second)
	if seconds < 1 {
		return 1
	}
	return seconds
}

// SessionAuth is the "soft" session middleware: it restores the login state
// from the session cookie into the request context and injects the authz
// principal. Authentication failures do not block the request; per-route
// policies decide whether to reject.
func SessionAuth(service *serviceauth.Service, next http.Handler) http.Handler {
	if service == nil || next == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(service.SessionCookieName())
		if err == nil && cookie.Value != "" {
			if authContext, authErr := service.Authenticate(r.Context(), cookie.Value); authErr == nil {
				ctx := serviceauth.WithAuthContext(r.Context(), authContext)
				ctx = authz.WithPrincipal(ctx, authz.Principal{
					UserID: authContext.User.ID,
					Role:   authContext.User.Role,
					Source: authz.BrowserSessionCredential,
				})
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func tokenFromRequest(r *http.Request, cookieName string) string {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
