package authz

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shuTwT/nex-api/internal/service/auth"
)

var (
	ErrUnauthenticated = auth.ErrUnauthenticated
	ErrForbidden       = auth.ErrForbidden
)

type CredentialSource string

const (
	BrowserSessionCredential CredentialSource = "browser_session"
	APITokenCredential       CredentialSource = "api_token"
)

type Permission string

const (
	PermissionRead   Permission = "read"
	PermissionWrite  Permission = "write"
	PermissionDelete Permission = "delete"
)

type permissionMask uint8

const (
	readPermission permissionMask = 1 << iota
	writePermission
	deletePermission
)

type Permissions struct {
	mask permissionMask
}

func ParsePermissions(raw string) (Permissions, error) {
	switch raw {
	case string(PermissionRead):
		return Permissions{mask: readPermission}, nil
	case string(PermissionRead) + "," + string(PermissionWrite):
		return Permissions{mask: readPermission | writePermission}, nil
	case string(PermissionRead) + "," + string(PermissionWrite) + "," + string(PermissionDelete):
		return Permissions{mask: readPermission | writePermission | deletePermission}, nil
	default:
		return Permissions{}, fmt.Errorf("%w: %q", ErrInvalidPermissions, raw)
	}
}

func (p Permissions) Allows(permission Permission) bool {
	var required permissionMask
	switch permission {
	case PermissionRead:
		required = readPermission
	case PermissionWrite:
		required = writePermission
	case PermissionDelete:
		required = deletePermission
	default:
		return false
	}
	return p.mask&required != 0
}

func (p Permissions) AllowsMethod(method string) bool {
	permission, ok := permissionForMethod(method)
	return ok && p.Allows(permission)
}

func (p Permissions) String() string {
	values := make([]string, 0, 3)
	if p.Allows(PermissionRead) {
		values = append(values, string(PermissionRead))
	}
	if p.Allows(PermissionWrite) {
		values = append(values, string(PermissionWrite))
	}
	if p.Allows(PermissionDelete) {
		values = append(values, string(PermissionDelete))
	}
	return strings.Join(values, ",")
}

func permissionForMethod(method string) (Permission, bool) {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "GET", "HEAD", "OPTIONS":
		return PermissionRead, true
	case "POST", "PUT", "PATCH":
		return PermissionWrite, true
	case "DELETE":
		return PermissionDelete, true
	default:
		return "", false
	}
}

type Principal struct {
	UserID      string
	Role        string
	Source      CredentialSource
	TokenID     string
	Permissions Permissions `json:"-"`
}

type principalContextKey struct{}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	if ctx == nil {
		return Principal{}, false
	}
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok
}

func BrowserSessionPolicy(ctx context.Context) (Principal, error) {
	authContext, ok := auth.AuthFromContext(ctx)
	if !ok || authContext.User.ID == "" {
		return Principal{}, ErrUnauthenticated
	}
	return Principal{
		UserID: authContext.User.ID,
		Role:   authContext.User.Role,
		Source: BrowserSessionCredential,
	}, nil
}

func RequestPrincipal(ctx context.Context) (Principal, error) {
	if principal, ok := PrincipalFromContext(ctx); ok && principal.UserID != "" {
		return principal, nil
	}
	return BrowserSessionPolicy(ctx)
}

func UserPolicy(ctx context.Context) (Principal, error) {
	return BrowserSessionPolicy(ctx)
}

func AdminPolicy(ctx context.Context) (Principal, error) {
	principal, err := BrowserSessionPolicy(ctx)
	if err != nil {
		return Principal{}, err
	}
	if principal.Role != "admin" {
		return Principal{}, ErrForbidden
	}
	return principal, nil
}

func OwnsResource(userID, resourceOwnerID string) bool {
	return userID != "" && resourceOwnerID != "" && userID == resourceOwnerID
}

func CanAccessResource(principal Principal, resourceOwnerID string) bool {
	return principal.UserID != "" && resourceOwnerID != "" &&
		(principal.Role == "admin" || OwnsResource(principal.UserID, resourceOwnerID))
}

func CheckOwnership(principal Principal, resourceOwnerID string) error {
	if principal.UserID == "" {
		return ErrUnauthenticated
	}
	if !CanAccessResource(principal, resourceOwnerID) {
		return ErrForbidden
	}
	return nil
}

func IsUnauthenticated(err error) bool {
	return errors.Is(err, ErrUnauthenticated)
}

func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}
