# 后端架构：五层目录

`backend/internal` 只包含五个顶层目录，依赖方向固定：

```text
internal/
├── handler/     HTTP 适配层：路由注册、请求解码、错误映射、OpenAPI ServerInterface
├── infra/       基础设施：config、database、logger、httpserver、pay、schedule、
│                storage、oauth、proxy、worker
├── job/         内置计划任务注册（统计同步、支付过期）
├── middleware/  HTTP 中间件：请求 ID、恢复、访问日志、Body 限制、CORS、
│                会话恢复、User/Admin/API Token/Permission/Ownership、CSRF
└── service/     业务规则、校验、事务、查询、统计与 DTO
```

依赖方向：

```text
cmd/server → handler | middleware | job | service | infra | ent
handler    → service | middleware
middleware → service | infra
job        → service | infra/schedule
service    → infra | ent
infra      → ent | 标准库及第三方库
ent        → 标准库及第三方库
```

约束：

- 禁止 `service` 反向依赖 `handler`、`middleware`、`job`；
- 禁止 `infra` 依赖其他四层；
- `handler` 只做 HTTP 适配，不执行 Ent 查询或事务（业务查询在对应 `service`）；
- `service` 返回业务错误，不创建 HTTP 状态码或写响应；HTTP 映射由 `handler` 完成；
- `infra/schedule` 是纯调度引擎，不包含业务任务；具体任务在 `job` 注册，业务逻辑在 `service`。
- `backend/ent` 保存 Ent Schema、生成入口及全部生成代码；`infra/database` 负责创建 SQL/Ent 与 Redis 客户端实例。

分层规则以本文件为准，不通过 CI 架构守护测试强制（见 `docs/backend-five-layer-refactor-plan.md`）。
