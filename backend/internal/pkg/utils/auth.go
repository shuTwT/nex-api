package utils

import (
	"errors"
	"net/http"

	"github.com/shuTwT/nex-api/backend/internal/handler/httpkit"
	serviceauthz "github.com/shuTwT/nex-api/backend/internal/service/authz"
)

// RequireAdmin validates that the request carries an authenticated admin session.
// It writes the standard API error response and returns false when authorization fails.
func RequireAdmin(w http.ResponseWriter, r *http.Request) bool {
	principal, err := serviceauthz.AdminPolicy(r.Context())
	if err == nil && principal.UserID != "" {
		return true
	}

	status := http.StatusUnauthorized
	code, message := "unauthorized", "authentication required"
	if errors.Is(err, serviceauthz.ErrForbidden) {
		status, code, message = http.StatusForbidden, "forbidden", "access denied"
	}
	WriteError(w, r, httpkit.NewAPIError(status, code, message, err))
	return false
}
