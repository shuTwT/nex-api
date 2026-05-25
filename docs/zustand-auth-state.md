# Zustand 用户状态管理系统

## 概述

本项目使用 Zustand 实现了完整的用户状态管理系统，包含用户信息、认证状态、持久化存储等功能。

## 核心特性

### 1. **类型安全**
- 完整的 TypeScript 类型定义
- 不可变的状态更新
- 类型推导的选择器

### 2. **持久化存储**
- 使用 localStorage 自动持久化
- 页面刷新后自动恢复状态
- 可配置的存储策略

### 3. **原子化更新**
- 精确的状态更新操作
- 避免不必要的重渲染
- 选择器优化性能

### 4. **订阅机制**
- 状态变更自动通知
- 组件自动重渲染
- 细粒度的订阅控制

## 文件结构

```
src/
├── types/
│   └── auth.ts              # 类型定义
├── stores/
│   └── auth-store.ts        # Zustand store
├── hooks/
│   └── use-auth.ts          # 认证 hook
└── components/
    ├── main-layout.tsx      # 主布局（使用状态）
    └── auth-provider.tsx    # 认证提供者
```

## 使用方法

### 1. **在组件中使用状态**

```typescript
import { useAuth } from "@/hooks/use-auth";

function MyComponent() {
  const { 
    user,              // 当前用户信息
    isAuthenticated,   // 是否已登录
    isAdmin,          // 是否是管理员
    credits,          // 用户积分
    login,            // 登录方法
    logout,           // 登出方法
    updateUser,       // 更新用户信息
    updateCredits,    // 更新积分
  } = useAuth();

  if (!isAuthenticated) {
    return <div>请先登录</div>;
  }

  return (
    <div>
      <p>欢迎，{user?.username}</p>
      <p>积分：{credits}</p>
      <button onClick={logout}>退出登录</button>
    </div>
  );
}
```

### 2. **使用选择器优化性能**

```typescript
import { useAuthStore, selectUserCredits, selectIsAdmin } from "@/stores/auth-store";

function CreditsDisplay() {
  // 只订阅积分变化，不会因为其他状态变化而重渲染
  const credits = useAuthStore(selectUserCredits);
  
  return <div>积分：{credits}</div>;
}

function AdminPanel() {
  // 只订阅管理员状态
  const isAdmin = useAuthStore(selectIsAdmin);
  
  if (!isAdmin) return null;
  
  return <div>管理员面板</div>;
}
```

### 3. **登录流程**

```typescript
import { useAuth } from "@/hooks/use-auth";

function LoginForm() {
  const { login } = useAuth();

  const handleLogin = async (email: string, password: string) => {
    const response = await fetch("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });

    const data = await response.json();
    
    if (data.success) {
      // 登录成功后，状态会自动更新并持久化
      login(data.data.user, data.data.token);
      // 页面会自动跳转或更新 UI
    }
  };

  return <form onSubmit={handleLogin}>...</form>;
}
```

### 4. **更新用户信息**

```typescript
import { useAuth } from "@/hooks/use-auth";

function UserProfile() {
  const { user, updateUser, updateCredits } = useAuth();

  const handleUpdateUsername = (newUsername: string) => {
    // 原子化更新，只更新 username 字段
    updateUser({ username: newUsername });
  };

  const handleAddCredits = (amount: number) => {
    // 更新积分
    const newCredits = (user?.credits ?? 0) + amount;
    updateCredits(newCredits);
  };

  return (
    <div>
      <p>用户名：{user?.username}</p>
      <button onClick={() => handleUpdateUsername("新用户名")}>
        更新用户名
      </button>
      <button onClick={() => handleAddCredits(100)}>
        增加 100 积分
      </button>
    </div>
  );
}
```

## API 参考

### AuthStore 状态

| 属性 | 类型 | 说明 |
|------|------|------|
| `user` | `User \| null` | 当前用户信息 |
| `token` | `string \| null` | 认证令牌 |
| `isAuthenticated` | `boolean` | 是否已认证 |
| `isLoading` | `boolean` | 是否正在加载 |

### AuthStore 方法

| 方法 | 参数 | 说明 |
|------|------|------|
| `login` | `(user: User, token: string)` | 登录并设置用户信息 |
| `logout` | - | 登出并清除所有状态 |
| `updateUser` | `(userData: Partial<User>)` | 更新用户信息（部分更新） |
| `updateCredits` | `(credits: number)` | 更新用户积分 |
| `setLoading` | `(loading: boolean)` | 设置加载状态 |
| `initializeAuth` | - | 初始化认证状态 |

### useAuth Hook

返回对象包含所有状态和方法，以及额外的便捷属性：

| 属性 | 类型 | 说明 |
|------|------|------|
| `isAdmin` | `boolean` | 是否是管理员 |
| `credits` | `number` | 用户积分（默认 0） |

### 选择器函数

```typescript
// 选择器函数，用于优化性能
selectUser(state)          // 选择用户信息
selectIsAuthenticated(state) // 选择认证状态
selectToken(state)         // 选择令牌
selectIsLoading(state)     // 选择加载状态
selectUserCredits(state)   // 选择用户积分
selectUserRole(state)      // 选择用户角色
selectIsAdmin(state)       // 选择是否是管理员
```

## 持久化策略

- **存储位置**：localStorage
- **存储键名**：`nex-api-auth`
- **存储内容**：user、token、isAuthenticated
- **自动恢复**：页面加载时自动从 localStorage 恢复状态
- **安全清理**：登出时自动清除 localStorage

## 最佳实践

### 1. **使用选择器避免不必要的重渲染**

```typescript
// ❌ 不好：订阅整个 store
const store = useAuthStore();

// ✅ 好：只订阅需要的状态
const user = useAuthStore(selectUser);
const credits = useAuthStore(selectUserCredits);
```

### 2. **使用 useAuth hook 简化代码**

```typescript
// ❌ 不好：直接使用 store
const user = useAuthStore(state => state.user);
const login = useAuthStore(state => state.login);

// ✅ 好：使用 hook
const { user, login } = useAuth();
```

### 3. **原子化更新状态**

```typescript
// ❌ 不好：手动合并状态
const user = useAuthStore(state => state.user);
const newUser = { ...user, username: "新名字" };

// ✅ 好：使用原子化更新方法
updateUser({ username: "新名字" });
```

### 4. **处理敏感信息**

```typescript
// ❌ 不好：在控制台打印敏感信息
console.log(user);

// ✅ 好：避免打印敏感信息
console.log({ id: user.id, username: user.username });
```

## 安全考虑

1. **Token 存储**：Token 存储在 localStorage，生产环境建议使用 HttpOnly Cookie
2. **敏感信息**：不要在 localStorage 存储敏感信息（如密码）
3. **XSS 防护**：React 自动转义，但要注意 dangerouslySetInnerHTML
4. **登出清理**：登出时会清除所有状态和 localStorage

## 调试技巧

### 1. **查看当前状态**

```typescript
// 在浏览器控制台
localStorage.getItem('nex-api-auth')
```

### 2. **手动清除状态**

```typescript
// 在浏览器控制台
localStorage.removeItem('nex-api-auth')
location.reload()
```

### 3. **订阅状态变化**

```typescript
// 在组件中
useEffect(() => {
  const unsubscribe = useAuthStore.subscribe(
    (state) => console.log('State changed:', state)
  );
  return unsubscribe;
}, []);
```

## 扩展功能

### 1. **添加新的用户属性**

```typescript
// types/auth.ts
export interface User {
  id: string;
  email: string;
  username: string;
  role: "user" | "admin";
  credits: number;
  avatar?: string;  // 新增头像字段
  phone?: string;   // 新增手机号字段
  createdAt?: string;
  updatedAt?: string;
}
```

### 2. **添加新的状态**

```typescript
// stores/auth-store.ts
interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  lastLoginTime: number | null;  // 新增最后登录时间
}
```

### 3. **添加新的方法**

```typescript
// stores/auth-store.ts
updateAvatar: (avatar: string) => {
  const currentUser = get().user;
  if (currentUser) {
    set({ user: { ...currentUser, avatar } });
  }
}
```

## 总结

这个状态管理系统提供了：
- ✅ 完整的类型安全
- ✅ 自动持久化
- ✅ 原子化更新
- ✅ 性能优化
- ✅ 易于使用的 API
- ✅ 可扩展的架构

通过 Zustand 的简洁 API 和 React 的最佳实践，实现了高效、可靠的用户状态管理。
