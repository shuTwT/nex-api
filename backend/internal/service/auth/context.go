package auth

import (
	"context"
)

type authContextKey struct{}

// WithAuthContext stores the authenticated session context; the middleware
// layer injects it during session restoration.
func WithAuthContext(ctx context.Context, authContext AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey{}, authContext)
}

// AuthFromContext reads the authenticated session context.
func AuthFromContext(ctx context.Context) (AuthContext, bool) {
	if ctx == nil {
		return AuthContext{}, false
	}
	authContext, ok := ctx.Value(authContextKey{}).(AuthContext)
	return authContext, ok
}

// UserFromContext reads the authenticated user, if any.
func UserFromContext(ctx context.Context) (User, bool) {
	authContext, ok := AuthFromContext(ctx)
	if !ok {
		return User{}, false
	}
	return authContext.User, true
}

// RequireAuth is a business-policy helper: it returns the authenticated
// context or ErrUnauthenticated.
func RequireAuth(ctx context.Context) (AuthContext, error) {
	authContext, ok := AuthFromContext(ctx)
	if !ok {
		return AuthContext{}, ErrUnauthenticated
	}
	return authContext, nil
}

// RequireAdmin is a business-policy helper: it returns the authenticated user
// or ErrUnauthenticated/ErrForbidden.
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
