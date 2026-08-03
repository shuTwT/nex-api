# Nex API

Nex API 是一个由 Go 后端和 Vite 前端组成的 API 管理与 MCP 服务平台。Go 服务提供认证、API/MCP 网关、支付、用量统计与管理 API；Vite 应用提供公开站点和管理控制台。

## 项目结构

```text
backend/   Go HTTP 服务、Ent 数据模型、迁移和 OpenAPI 生成器
frontend/  React + Vite 单页应用
```

## 本地开发

需要 Go 1.26+、Node.js 22+ 和 npm。

```bash
make doctor
cp backend/.env.example backend/.env
npm --prefix frontend ci
```

在一个终端启动后端。配置文件只是环境变量模板，启动前需要由 shell、进程管理器或容器导出：

```bash
cd backend
set -a; . ./.env; set +a
go run ./cmd/server
```

在另一个终端启动前端：

```bash
npm --prefix frontend run dev
```

Vite 在 `http://localhost:3000` 运行，并将 `/api` 请求代理到 Go 后端的 `http://localhost:8080`。生产部署时使用 `VITE_API_BASE_URL` 指向公开 API 地址，并将该前端来源加入 `SERVER_CORS_ORIGINS`。

## 构建与验证

```bash
make generate  # 从独立 HTTP 契约重新生成 OpenAPI 与 Vite 客户端
make test      # Go 测试和 Vite 类型检查
make build     # 构建 Go 二进制与 Vite 静态资源
```

HTTP 契约位于 `backend/test/contract/manifest.json`。它是独立于实现文件的 API 清单，用于生成 `backend/openapi/openapi.yaml`、Go 路由类型和 `frontend/src/api/generated/` 客户端。

## 配置

后端配置模板在 `backend/.env.example`。支付与 OAuth 提供商配置由管理员在系统设置中维护并保存到数据库；不要把这些凭据写入前端变量或提交到仓库。

## License

MIT
