# 前端认证状态

新版 Vite 前端使用 React Context 管理认证状态，而不是在浏览器中持久化访问令牌。

- [`frontend/src/providers/auth.tsx`](../frontend/src/providers/auth.tsx) 负责读取会话、登录、登出与 CSRF Token。
- [`frontend/src/hooks/use-auth.ts`](../frontend/src/hooks/use-auth.ts) 为页面和组件提供 `user`、`isAuthenticated`、`isLoading`、`login`、`logout` 和 `refreshUser`。
- Go 后端通过 HttpOnly 会话 Cookie 认证；前端请求始终使用 `credentials: "include"`。

应用根节点必须被 `AuthProvider` 包裹。页面加载时会调用 `/api/auth/me` 恢复会话；用户名和权限变化后调用 `refreshUser()` 即可同步状态。
