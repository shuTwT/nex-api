# 后端五层目录一次性重构计划

## 概要

将项目根目录的 `internal` 一次性重构为仅包含以下五个顶层目录：

```text
internal/
├── handler/
├── infra/
├── job/
├── middleware/
└── service/
```

保持所有 HTTP 路由、JSON、状态码、配置键、数据库 Schema、启动命令和 OpenAPI 行为兼容。本次只做架构重构，不修复统计双写、SSRF、幂等等既有行为问题。

依赖方向固定为：

```text
cmd/server → handler | middleware | job | service | infra | ent
handler    → service | middleware
middleware → service | infra
job        → service | infra/schedule
service    → infra | ent
infra      → ent | 标准库及第三方库
ent        → 标准库及第三方库
```

禁止 `service` 反向依赖 `handler`、`middleware`、`job`，禁止 `infra` 依赖其他四层。

## 实施内容

### 1. Infra 层

建立以下基础设施子包：

```text
infra/
├── config/          # 当前 config
├── database/
│   ├── database.go  # 数据库创建、连接和 Ent Client 实例入口
│   └── redis.go     # Redis 客户端实例入口
├── logger/          # slog 创建和上下文日志
├── pay/             # Provider 接口、微信、支付宝、Mock 客户端
├── schedule/        # 纯调度引擎、任务注册表和生命周期
├── storage/         # 本地文件存储
├── httpserver/      # HTTP Server、健康检查和优雅退出
├── oauth/           # GitHub 等外部 OAuth 客户端
├── proxy/           # API/MCP 上游 HTTP、SSE、stdio 适配
└── worker/          # 脚本 Worker、进程池和 IPC
```

- Ent Schema、生成入口和生成代码统一位于后端根目录 `ent/`；`infra/database` 只负责创建数据库连接和 Ent Client 实例。
- `infra/pay` 维护支付客户端所需的基础类型，`service/payment` 直接依赖该包，避免反向依赖和循环引用。
- `infra/schedule` 不包含具体业务任务，只负责注册、启停、触发和关闭。
- 将原 `runtime` 拆分：Server/health 进入 `infra/httpserver`，logger 进入 `infra/logger`，HTTP 错误输出进入 `handler/httpkit`，HTTP 中间件进入 `middleware`。
- OpenAPI 文档由 `cmd/server/swagger.go` 中的 Swaggo 注释生成；再通过标准 `swagger2openapi` 转为 OpenAPI 3，前端使用 `openapi-typescript` 消费该文档。

### 2. Service 层

按业务域建立子包：

```text
service/{accounts,ads,auth,authz,catalog,dashboard,gateway,
marketplace,mcpgateway,membership,oauth,payment,schedule,
settings,stats,system,upload}
```

- Service 可以直接依赖 `infra/database`、`infra/pay`、`infra/storage` 等具体实现，不增加 Repository/Port 抽象。
- 将现有领域包中的业务规则、校验、事务、查询、统计和 DTO 移入对应 Service。
- 为 `dashboard`、`marketplace`、`settings`、`upload` 补齐 Service，Handler 不再直接执行 Ent 查询或事务。
- 将支付最低金额、积分换算、所有权判断等规则从 Handler 移入 `service/payment`。
- `service/gateway` 和 `service/mcpgateway` 负责编排鉴权、计费、统计及审计；HTTP、SSE、stdio 细节分别留在 Handler 或 Infra。
- Service 返回业务错误，不再创建 HTTP 状态码或写响应；HTTP 映射由 Handler 完成。
- `service/auth` 只处理凭证、会话和用户状态；Cookie、Header、CSRF 请求适配迁往 Handler/Middleware。
- `service/schedule` 负责计划任务配置的数据库 CRUD，调度运行时由 `infra/schedule` 提供。

### 3. Handler 与 Middleware

按业务域建立 `handler/<domain>`，每个包只负责：

- 路由注册；
- 请求参数、Path、Query、Header 和 JSON 解码；
- 调用对应 Service；
- 将业务错误映射为兼容的 HTTP 状态和响应；

增加：

```text
handler/httpkit/     # 统一 JSON、分页、错误响应、请求元数据
handler/router/      # 根路由装配，不创建 Service
```

所有运行时 HTTP 路由统一使用 Chi 注册；不混用 `http.ServeMux`。

`middleware` 集中存放：

- 请求 ID、恢复、访问日志、Body 限制、开发 CORS；
- 软会话恢复；
- User/Admin/API Token/Permission/Ownership；
- CSRF 校验和 Session/CSRF Cookie 适配。

认证 Principal、Permission、业务策略保留在 `service/authz`；`middleware` 仅负责 HTTP 上下文注入和拒绝响应。

### 4. Job 与启动装配

`job` 中定义并注册当前两个任务：

- Redis API/MCP 统计同步；
- Pending 支付订单过期。

具体任务调用 `service/stats` 与 `service/payment`，通过 `infra/schedule` 注册。原 `/api/cron/*` HTTP 入口属于 `handler/cron`，不放入 `job`。

`cmd/server` 最终只负责：

1. 加载 Infra；
2. 创建 Service；
3. 注册 Job；
4. 创建 Handler/Router；
5. 组装 Middleware；
6. 启动和关闭 HTTP Server、Scheduler、Worker。

删除迁移完成后的旧 `internal/accounts`、`payment`、`runtime` 等顶层目录，确保 `internal` 顶层只剩五个目标目录。

## 接口与兼容性

- 外部 HTTP 路由、请求/响应 JSON、Cookie、Header、状态码保持不变。
- OpenAPI 文档由可执行的 Swaggo 注释生成；`operationId` 和前端客户端兼容调用保持稳定。
- 数据库 Schema、Atlas migration、Ent 字段和已有数据不变，不新增迁移。
- `.env` 键、配置文件格式、默认值和覆盖顺序不变。
- `go run ./cmd/server`、`go run ./cmd/script-worker`、Makefile 入口不变。
- 内部构造接口统一为 `service.New(...)`、`handler.New(...)`/`RegisterRoutes(...)` 和 `job.RegisterAll(...)`；旧内部 import path 不保留兼容包装。

## 测试与验收

- 迁移现有测试到职责对应的新包，保持原测试语义不变。
- 为新抽出的 `dashboard`、`marketplace`、`settings` Service 补充业务及事务测试；Handler 测试只验证 HTTP 映射。
- 验证登录、CSRF、Session Cookie、API Token、管理员和 Ownership 中间件行为完全兼容。
- 验证支付创建、回调幂等、取消、过期及会员发放流程。
- 验证 API/MCP 网关转发、计费、退款、脚本转换、SSE/stdio 行为。
- 运行 OpenAPI 生成并确认生成结果可重复，执行 contract fixtures 和 OpenAPI lint。
- 执行 `go build ./cmd/server ./cmd/script-worker` 与 `go test -race -shuffle=on -count=1 ./...`。
- 使用 `go list` 检查实际 import 图符合规定方向，并确认 `internal` 顶层恰好只有五个目录；不新增自动化架构测试。
- 检查数据库迁移状态与前端生成客户端，确认无外部契约或 Schema 差异。

## 已确定的约束

- 采用一次性切换，在同一改动中完成全部移动、拆分、import 更新和验证，不提供旧包过渡层。
- 当前工作区已有修改均视为有效基线；迁移其当前内容，不回退或用 HEAD 版本覆盖。
- 本次不顺带修复统计计数覆盖、支付/网关幂等、SSRF 或初始化竞争问题，单独记录后续任务。
- 分层规则写入后端架构文档并由 README 引用，不增加 CI 架构守护测试。
