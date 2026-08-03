package middleware

import (
	"errors"
	"net/http"

	serviceauthz "github.com/shuTwT/nex-api/internal/service/authz"
)

// RequireUser rejects requests without a browser-session principal and
// injects the principal into the request context.
func RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, err := serviceauthz.UserPolicy(r.Context())
		if err != nil {
			writeAuthorizationError(w, r, err)
			return
		}
		serveWithPrincipal(next, w, r, principal)
	})
}

// RequireAdmin rejects requests without an admin principal.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, err := serviceauthz.AdminPolicy(r.Context())
		if err != nil {
			writeAuthorizationError(w, r, err)
			return
		}
		serveWithPrincipal(next, w, r, principal)
	})
}

// RequireOwnership rejects requests whose resource owner differs from the
// request principal (admins may access any resource).
func RequireOwnership(ownerID func(*http.Request) (string, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, err := serviceauthz.RequestPrincipal(r.Context())
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
			if err := serviceauthz.CheckOwnership(principal, resourceOwnerID); err != nil {
				writeAuthorizationError(w, r, err)
				return
			}
			serveWithPrincipal(next, w, r, principal)
		})
	}
}

// RequireAPIToken authenticates a Bearer API token and enforces method-level
// permissions.
func RequireAPIToken(service *serviceauthz.TokenService, next http.Handler) http.Handler {
	return requireAPIToken(service, next, true)
}

// RequireAPITokenOnly authenticates a Bearer API token without method
// permission enforcement.
func RequireAPITokenOnly(service *serviceauthz.TokenService, next http.Handler) http.Handler {
	return requireAPIToken(service, next, false)
}

func requireAPIToken(service *serviceauthz.TokenService, next http.Handler, requireMethodPermission bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, err := service.AuthenticateBearer(r.Context(), r.Header.Get("Authorization"))
		if err != nil {
			writeAuthorizationError(w, r, err)
			return
		}
		if requireMethodPermission && !principal.Permissions.AllowsMethod(r.Method) {
			writeAuthorizationError(w, r, serviceauthz.ErrForbidden)
			return
		}
		serveWithPrincipal(next, w, r, principal)
	})
}

// RequireAPIPermission requires a previously authenticated API-token
// principal (injected by RequireAPIToken) with the given permission.
func RequireAPIPermission(permission serviceauthz.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := serviceauthz.PrincipalFromContext(r.Context())
			if !ok || principal.Source != serviceauthz.APITokenCredential || principal.UserID == "" {
				writeAuthorizationError(w, r, serviceauthz.ErrUnauthenticated)
				return
			}
			if !principal.Permissions.Allows(permission) {
				writeAuthorizationError(w, r, serviceauthz.ErrForbidden)
				return
			}
			serveWithPrincipal(next, w, r, principal)
		})
	}
}

func serveWithPrincipal(next http.Handler, w http.ResponseWriter, r *http.Request, principal serviceauthz.Principal) {
	if next == nil {
		next = http.NotFoundHandler()
	}
	next.ServeHTTP(w, r.WithContext(serviceauthz.WithPrincipal(r.Context(), principal)))
}

func writeAuthorizationError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	message := "internal server error"
	switch {
	case errors.Is(err, serviceauthz.ErrUnauthenticated),
		errors.Is(err, serviceauthz.ErrInvalidBearer),
		errors.Is(err, serviceauthz.ErrInvalidToken),
		errors.Is(err, serviceauthz.ErrInactiveToken),
		errors.Is(err, serviceauthz.ErrExpiredToken),
		errors.Is(err, serviceauthz.ErrInvalidPermissions):
		status = http.StatusUnauthorized
		message = "authentication required"
	case errors.Is(err, serviceauthz.ErrForbidden):
		status = http.StatusForbidden
		message = "access denied"
	}
	writeErrorResponse(w, r, status, message)
}
