# NexApi - 全栈 API 管理与 MCP 服务平台

基于 Next.js 16 的全栈 API 管理系统，支持 HTTP API 管理、MCP 服务管理、API 市场、支付订阅、用量统计等完整功能。

## 功能概览

- **API 管理** — 创建和管理 HTTP API 接口，支持定价策略、调用限制、参数配置
- **MCP 服务管理** — 管理 Stdio / SSE / Streamable HTTP 三种 MCP 服务，统一 Gateway 转发
- **API 市场 & MCP 市场** — 公开展示可用服务，支持类型筛选、搜索、网格/列表视图
- **用户系统** — 注册、登录、OAuth 支持、角色管理（管理员/普通用户）
- **积分与订阅** — 积分充值、兑换码、订阅计划，基于 Redis 的用量统计和限流
- **支付系统** — 支持微信支付、支付宝、模拟支付，包含回调处理和业务通知
- **广告系统** — 横幅广告和内联广告管理，支持上传图片和自定义链接
- **审计日志** — 记录用户操作和 API 调用审计日志
- **管理控制台** — 完整的后台管理面板，涵盖用户/Token/API/MCP/广告/订阅/审计等模块
- **代码沙箱** — 内置代码执行沙箱，支持在线测试 API 调用示例

## 技术栈

| 分类 | 技术 |
|---|---|
| 前端框架 | Next.js 16.1.6 (App Router) |
| 语言 | TypeScript 5.x (strict mode, zero `any`) |
| UI 组件 | Radix UI + shadcn/ui |
| 样式 | Tailwind CSS 4 |
| 认证 | NextAuth.js (Credentials + OAuth) |
| ORM | Prisma 7.5.0 (Driver Adapter) |
| 数据库 | SQLite / MySQL / PostgreSQL |
| 缓存/限流 | Redis (ioredis) |
| 支付 | better-wechatpay + alipay-sdk |
| 图表 | Chart.js + react-chartjs-2 |
| 代码编辑器 | Monaco Editor |
| 运行时 | ESM (ECMAScript Modules) |

> **必须启用 Redis，否则用量统计和限流功能无法使用。**

## 项目结构

```
nex-api/
├── prisma/
│   ├── schema.prisma              # 数据模型（User, Api, McpService, Payment, etc.）
│   ├── seed.ts                    # 数据库种子
│   └── migrations/                # 迁移文件
├── generated/                     # Prisma Client 生成目录
├── src/
│   ├── app/                       # Next.js App Router
│   │   ├── api/                   # API 路由
│   │   │   ├── apis/              # HTTP API 管理 CRUD
│   │   │   ├── mcp-services/      # MCP 服务管理 CRUD
│   │   │   ├── categories/        # API 分类管理
│   │   │   ├── tokens/            # API Token 管理
│   │   │   ├── users/             # 用户管理
│   │   │   ├── subscription-plans/# 订阅计划管理
│   │   │   ├── redemption-codes/  # 兑换码管理
│   │   │   ├── audit-logs/        # 审计日志
│   │   │   ├── advertisements/    # 广告管理
│   │   │   ├── payment/           # 支付接口 + 业务回调
│   │   │   ├── marketplace/       # API/MCP 市场公开接口
│   │   │   ├── dashboard/         # 仪表盘统计
│   │   │   ├── usage/             # 用户用量查询
│   │   │   ├── stats/             # 全局统计
│   │   │   ├── auth/              # 认证相关
│   │   │   ├── membership/        # 会员订阅
│   │   │   ├── personal/          # 个人中心
│   │   │   ├── recharge/          # 积分充值
│   │   │   ├── upload/            # 文件上传
│   │   │   ├── system/            # 系统初始化
│   │   │   ├── system-settings/   # 系统设置
│   │   │   ├── cron/              # 定时任务
│   │   │   └── v1/                # API Gateway + MCP Gateway
│   │   │       ├── mcp/           # MCP 服务 Gateway (用户端)
│   │   │       └── ...            # HTTP API Gateway (用户端)
│   │   ├── console/               # 管理控制台
│   │   │   ├── page.tsx           # 仪表盘
│   │   │   ├── api-management/    # API 管理
│   │   │   ├── mcp-services/      # MCP 服务管理
│   │   │   ├── users/             # 用户管理
│   │   │   ├── tokens/            # Token 管理
│   │   │   ├── advertisements/    # 广告管理
│   │   │   ├── audit-logs/        # 审计日志
│   │   │   ├── subscription-plans/# 订阅计划
│   │   │   ├── redemption-codes/  # 兑换码
│   │   │   ├── settings/          # 系统设置
│   │   │   ├── usage/             # 用量统计
│   │   │   ├── personal/          # 个人中心
│   │   │   └── membership/        # 会员管理
│   │   ├── api-detail/            # API 详情页（公开）
│   │   ├── api-market/            # API 市场页（公开）
│   │   ├── mcp-market/            # MCP 市场页（公开）
│   │   ├── pricing/               # 定价页
│   │   ├── payment/               # 支付页面
│   │   ├── auth/                  # 登录/注册
│   │   ├── initialize/            # 系统初始化
│   │   └── page.tsx               # 首页
│   ├── components/                # React 组件
│   │   ├── ui/                    # shadcn/ui 基础组件
│   │   ├── ads/                   # 广告组件
│   │   ├── main-layout.tsx        # 公共布局
│   │   ├── console-layout.tsx     # 控制台布局
│   │   ├── pagination.tsx         # 分页组件
│   │   ├── usage-trend-chart.tsx  # 用量趋势图
│   │   ├── monaco-editor.tsx      # 代码编辑器
│   │   └── *.tsx                  # 各业务组件
│   └── lib/                       # 工具库
│       ├── prisma.ts              # Prisma Client 实例
│       ├── redis.ts               # Redis 客户端 + 用量/限流
│       ├── api-auth.ts            # API 认证 + 响应工具
│       ├── api-client.ts          # 前端 API 客户端
│       ├── config.ts              # 系统配置管理
│       ├── audit-log.ts           # 审计日志工具
│       ├── request-stats.ts       # 请求统计
│       ├── sandbox.ts             # 代码沙箱
│       ├── auth/                  # NextAuth 配置
│       ├── payment/               # 支付模块（类型/回调/实现）
│       └── validations/           # Zod 验证
├── prisma.config.ts               # Prisma v7 配置
├── eslint.config.mjs              # ESLint 配置
├── .env                           # 环境变量
└── package.json
```

## 数据模型

| 模型 | 说明 |
|---|---|
| `User` | 用户（积分、角色、API Key） |
| `Api` | HTTP API 接口（名称、端点、参数、定价） |
| `McpService` | MCP 服务（Stdio/SSE/Streamable HTTP） |
| `ApiToken` | 用户 API 访问令牌 |
| `Subscription` | 用户订阅记录 |
| `SubscriptionPlan` | 订阅计划 |
| `Payment` | 支付记录 |
| `RedemptionCode` | 积分兑换码 |
| `Advertisement` | 广告 |
| `AuditLog` | 审计日志 |
| `SystemSetting` | 系统设置键值对 |
| `ApiCategory` | API 分类 |

## 快速开始

### 1. 安装依赖

```bash
pnpm install
```

### 2. 配置环境变量

复制 `.env.example` 为 `.env`，配置数据库和 Redis 连接信息。

### 3. 初始化数据库

```bash
pnpm run db:generate      # 生成 Prisma Client
pnpm run db:migrate       # 运行迁移
pnpm run db:seed          # 填充种子数据
```

### 4. 启动

```bash
pnpm run dev              # 开发模式 → http://localhost:3000
pnpm run build && pnpm run start  # 生产构建
```

## Prisma v7 特性

- **ESM Only**: ES Modules 运行时
- **Driver Adapter**: 使用 `@prisma/adapter-better-sqlite3`（可按需切换到 MySQL/PostgreSQL）
- **配置分离**: `prisma.config.ts` 替代 schema 中的环境变量
- **显式输出**: Prisma Client 生成到 `./generated` 目录

## TypeScript 规范

- **strict mode** 启用，禁止 `any` 类型
- 所有 API 响应和参数使用泛型类型（`apiSuccess<T>`、`apiPaginated<T>`）
- Prisma 生成类型用于数据层，前端使用 `PaymentInfo` 等业务类型
- ESLint 0 error，`@typescript-eslint/no-explicit-any` 严格开启

## License

MIT
