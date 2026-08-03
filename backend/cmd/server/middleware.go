package main

import (
	"net/http"

	"github.com/shuTwT/nex-api/backend/internal/auth"
	"github.com/shuTwT/nex-api/backend/internal/authz"
)

// sessionMiddleware 是"软"会话认证:从会话 cookie 恢复登录态并写入请求
// context。认证失败时不拦截请求,由各业务模块内部的 authz 策略
// (RequireUser/RequireAdmin 等)决定是否拒绝。
func sessionMiddleware(service *auth.Service, next http.Handler) http.Handler {
	if service == nil || next == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(service.SessionCookieName())
		if err == nil && cookie.Value != "" {
			if authContext, authErr := service.Authenticate(r.Context(), cookie.Value); authErr == nil {
				ctx := auth.WithAuthContext(r.Context(), authContext)
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
