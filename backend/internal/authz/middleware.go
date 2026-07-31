package authz

import (
	"errors"
	"net/http"

	appRuntime "github.com/shuTwT/nex-api/backend/internal/runtime"
)

func RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, err := UserPolicy(r.Context())
		if err != nil {
			writeAuthorizationError(w, r, err)
			return
		}
		serveWithPrincipal(next, w, r, principal)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, err := AdminPolicy(r.Context())
		if err != nil {
			writeAuthorizationError(w, r, err)
			return
		}
		serveWithPrincipal(next, w, r, principal)
	})
}

func RequireOwnership(ownerID func(*http.Request) (string, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, err := RequestPrincipal(r.Context())
			if err != nil {
				writeAuthorizationError(w, r, err)
				return
			}
			if ownerID == nil {
				writeAuthorizationError(w, r, errors.New("authz: resource owner resolver is nil"))
				return
			}
			resourceOwnerID, err := ownerID(r)
			if err != nil {
				writeAuthorizationError(w, r, err)
				return
			}
			if err := CheckOwnership(principal, resourceOwnerID); err != nil {
				writeAuthorizationError(w, r, err)
				return
			}
			serveWithPrincipal(next, w, r, principal)
		})
	}
}

func (s *TokenService) RequireAPIToken(next http.Handler) http.Handler {
	return s.requireAPIToken(next, true)
}

func (s *TokenService) RequireAPITokenOnly(next http.Handler) http.Handler {
	return s.requireAPIToken(next, false)
}

func (s *TokenService) requireAPIToken(next http.Handler, requireMethodPermission bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, err := s.AuthenticateBearer(r.Context(), r.Header.Get("Authorization"))
		if err != nil {
			writeAuthorizationError(w, r, err)
			return
		}
		if requireMethodPermission && !principal.Permissions.AllowsMethod(r.Method) {
			writeAuthorizationError(w, r, ErrForbidden)
			return
		}
		serveWithPrincipal(next, w, r, principal)
	})
}

func (s *TokenService) RequireAPIPermission(permission Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := PrincipalFromContext(r.Context())
			if !ok || principal.Source != APITokenCredential || principal.UserID == "" {
				writeAuthorizationError(w, r, ErrUnauthenticated)
				return
			}
			if !principal.Permissions.Allows(permission) {
				writeAuthorizationError(w, r, ErrForbidden)
				return
			}
			serveWithPrincipal(next, w, r, principal)
		})
	}
}

func serveWithPrincipal(next http.Handler, w http.ResponseWriter, r *http.Request, principal Principal) {
	if next == nil {
		next = http.NotFoundHandler()
	}
	next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), principal)))
}

func writeAuthorizationError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	message := "internal server error"
	switch {
	case errors.Is(err, ErrUnauthenticated),
		errors.Is(err, ErrInvalidBearer),
		errors.Is(err, ErrInvalidToken),
		errors.Is(err, ErrInactiveToken),
		errors.Is(err, ErrExpiredToken),
		errors.Is(err, ErrInvalidPermissions):
		status = http.StatusUnauthorized
		code = "unauthorized"
		message = "authentication required"
	case errors.Is(err, ErrForbidden):
		status = http.StatusForbidden
		code = "forbidden"
		message = "access denied"
	}
	apiError := appRuntime.NewAPIError(status, code, message, nil)
	if writeErr := appRuntime.WriteError(w, r, apiError); writeErr != nil {
		return
	}
}
