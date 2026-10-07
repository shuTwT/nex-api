// Package authz 集中处理授权逻辑：请求者是谁（Principal）、API Token 允许做什么
// （Permissions），以及中间件和处理器使用的各类策略门禁。
package authz

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shuTwT/nex-api/internal/service/auth"
)

// 从 auth 包转出的哨兵错误，调用方（中间件、处理器）据此把失败映射为 401/403，
// 无需直接依赖 auth 包。
var (
	ErrUnauthenticated = auth.ErrUnauthenticated
	ErrForbidden       = auth.ErrForbidden
)

// CredentialSource 标识调用方通过何种方式完成认证。中间件用它区分两种凭证，
// 例如 RequireAPIPermission 会拒绝非 API Token 的请求。
type CredentialSource string

const (
	// BrowserSessionCredential 表示通过浏览器 Cookie 会话认证的身份。
	BrowserSessionCredential CredentialSource = "browser_session"
	// APITokenCredential 表示通过 API Token（sk_... Bearer 凭证）认证的身份，
	// 并带有显式的权限集。
	APITokenCredential CredentialSource = "api_token"
)

// Permission 是 API Token 的具名权限范围。Token 只允许三档递增的权限组合：
// read、read+write、read+write+delete。
type Permission string

const (
	PermissionRead   Permission = "read"
	PermissionWrite  Permission = "write"
	PermissionDelete Permission = "delete"
)

// permissionMask 是 Permissions 背后的位掩码表示。
type permissionMask uint8

const (
	readPermission permissionMask = 1 << iota
	writePermission
	deletePermission
)

// Permissions 是不可变的 API Token 权限集合，底层用位掩码存储。
type Permissions struct {
	mask permissionMask
}

// ParsePermissions 把 API Token 上存储的权限字符串（"read"、"read,write" 或
// "read,write,delete"，其余一律拒绝）解析成 Permissions。在 Token 认证时调用。
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

// Allows 判断当前权限集是否包含指定权限。
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

// AllowsMethod 按 HTTP 方法校验 Token 权限：读方法（GET/HEAD/OPTIONS）需要 read，
// 变更方法（POST/PUT/PATCH）需要 write，DELETE 需要 delete。未知方法一律拒绝。
func (p Permissions) AllowsMethod(method string) bool {
	permission, ok := permissionForMethod(method)
	return ok && p.Allows(permission)
}

// String 把权限集序列化回规范的逗号分隔存储格式，例如 "read,write"。
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

// permissionForMethod 把 HTTP 动词映射为所需的 API Token 权限，
// 第二个返回值为 false 表示该方法不在已知范围内。
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

// Principal 是请求的已认证执行者：以哪个用户身份、什么角色、通过何种凭证认证。
// 只有 API Token 凭证才会填充 Permissions（json:"-" 避免泄漏到响应里）。
type Principal struct {
	UserID      string
	Role        string
	Source      CredentialSource
	TokenID     string
	Permissions Permissions `json:"-"`
}

type principalContextKey struct{}

// WithPrincipal 把 Principal 存入请求上下文。认证中间件在认证成功后调用，
// 处理器通过 RequestPrincipal 取回。
func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

// PrincipalFromContext 取回之前用 WithPrincipal 存入的 Principal，不存在则返回 false。
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	if ctx == nil {
		return Principal{}, false
	}
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok
}

// BrowserSessionPolicy 基于浏览器会话的认证上下文构建 Principal；
// 没有会话用户时返回 ErrUnauthenticated，因此它不接受单独的 API Token 凭证。
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

// RequestPrincipal 返回认证中间件注入的 Principal（浏览器会话或 API Token 均可），
// 未注入时回退到会话认证。它是处理器识别调用方的默认入口。
func RequestPrincipal(ctx context.Context) (Principal, error) {
	if principal, ok := PrincipalFromContext(ctx); ok && principal.UserID != "" {
		return principal, nil
	}
	return BrowserSessionPolicy(ctx)
}

// UserPolicy 要求已登录的浏览器会话，是 RequireUser 中间件背后的门禁。
func UserPolicy(ctx context.Context) (Principal, error) {
	return BrowserSessionPolicy(ctx)
}

// AdminPolicy 要求已登录且角色为 "admin" 的浏览器会话，
// 是 RequireAdmin 中间件背后的门禁。
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

// OwnsResource 判断 userID 是否与资源属主 ID 非空且完全一致。
func OwnsResource(userID, resourceOwnerID string) bool {
	return userID != "" && resourceOwnerID != "" && userID == resourceOwnerID
}

// CanAccessResource 判断 principal 能否访问属主为 resourceOwnerID 的资源：
// admin 可以访问任何资源，其他用户只能访问自己的。
func CanAccessResource(principal Principal, resourceOwnerID string) bool {
	return principal.UserID != "" && resourceOwnerID != "" &&
		(principal.Role == "admin" || OwnsResource(principal.UserID, resourceOwnerID))
}

// CheckOwnership 是返回错误形式的归属门禁，供归属校验中间件使用：
// 调用方无身份时返回 ErrUnauthenticated，既非 admin 也非属主时返回 ErrForbidden。
func CheckOwnership(principal Principal, resourceOwnerID string) error {
	if principal.UserID == "" {
		return ErrUnauthenticated
	}
	if !CanAccessResource(principal, resourceOwnerID) {
		return ErrForbidden
	}
	return nil
}

// IsUnauthenticated 判断 err 是否为（或包装了）未认证哨兵错误，用于映射 401 响应。
func IsUnauthenticated(err error) bool {
	return errors.Is(err, ErrUnauthenticated)
}

// IsForbidden 判断 err 是否为（或包装了）禁止访问哨兵错误，用于映射 403 响应。
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}
