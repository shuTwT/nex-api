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
