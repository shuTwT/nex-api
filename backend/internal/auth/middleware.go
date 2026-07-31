package auth

import (
	"context"
	"net/http"
)

type authContextKey struct{}

func WithAuthContext(ctx context.Context, authContext AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey{}, authContext)
}

func AuthFromContext(ctx context.Context) (AuthContext, bool) {
	if ctx == nil {
		return AuthContext{}, false
	}
	authContext, ok := ctx.Value(authContextKey{}).(AuthContext)
	return authContext, ok
}

func UserFromContext(ctx context.Context) (User, bool) {
	authContext, ok := AuthFromContext(ctx)
	if !ok {
		return User{}, false
	}
	return authContext.User, true
}

func (s *Service) CutoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ClearLegacyCookies(w)
		next.ServeHTTP(w, r)
	})
}

func (s *Service) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ClearLegacyCookies(w)
		authContext, err := s.Authenticate(r.Context(), s.tokenFromRequest(r))
		if err != nil {
			if writeErr := writeError(w, http.StatusUnauthorized, "unauthorized"); writeErr != nil {
				return
			}
			return
		}
		next.ServeHTTP(w, r.WithContext(WithAuthContext(r.Context(), authContext)))
	})
}

func (s *Service) RequireUser(next http.Handler) http.Handler {
	return s.AuthMiddleware(next)
}

func (s *Service) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return s.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authContext, ok := AuthFromContext(r.Context())
			if !ok {
				if writeErr := writeError(w, http.StatusUnauthorized, "unauthorized"); writeErr != nil {
					return
				}
				return
			}
			if role == "" || authContext.User.Role != role {
				if writeErr := writeError(w, http.StatusForbidden, "forbidden"); writeErr != nil {
					return
				}
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}

func (s *Service) RequireAdmin(next http.Handler) http.Handler {
	return s.RequireRole("admin")(next)
}

func RequireAuth(ctx context.Context) (AuthContext, error) {
	authContext, ok := AuthFromContext(ctx)
	if !ok {
		return AuthContext{}, ErrUnauthenticated
	}
	return authContext, nil
}

func RequireAdmin(ctx context.Context) (User, error) {
	authContext, err := RequireAuth(ctx)
	if err != nil {
		return User{}, err
	}
	if authContext.User.Role != "admin" {
		return User{}, ErrForbidden
	}
	return authContext.User, nil
}
