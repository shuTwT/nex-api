# One API - 全栈 SSR 项目

一个基于 Next.js 16 的全栈 API 管理系统，使用 Prisma ORM 和 SQLite 数据库。

## 技术栈

- **前端框架**: Next.js 16.1.6 (App Router)
- **语言**: TypeScript 5.x
- **UI 组件**: Radix UI + shadcn/ui
- **样式**: Tailwind CSS 4
- **ORM**: Prisma 7.5.0
- **数据库**: SQLite (better-sqlite3)
- **运行时**: ESM (ECMAScript Modules)

## 项目结构

```
one-api/
├── prisma/
│   ├── schema.prisma          # Prisma 数据模型定义
│   ├── seed.ts                # 数据库种子文件
│   └── migrations/            # 数据库迁移文件
├── generated/                 # Prisma Client 生成目录
├── src/
│   ├── app/                   # Next.js App Router
│   │   ├── api/              # API 路由
│   │   │   ├── categories/   # 分类 API
│   │   │   └── apis/         # API 列表接口
│   │   ├── api-detail/       # API 详情页
│   │   ├── api-market/       # API 市场页
│   │   ├── pricing/          # 定价页
│   │   └── page.tsx          # 首页
│   ├── components/           # React 组件
│   │   └── ui/              # shadcn/ui 组件
│   └── lib/                  # 工具库
│       ├── prisma.ts        # Prisma Client 实例
│       └── utils.ts         # 工具函数
├── prisma.config.ts          # Prisma v7 配置文件
├── .env                      # 环境变量
└── package.json

```

## 数据模型

### User (用户)
- 用户基本信息、角色、积分
- 关联订阅和 API 使用记录

### ApiCategory (API 分类)
- 分类名称、描述、图标
- 关联多个 API

### Api (API 接口)
- API 基本信息、端点、方法
- 定价、文档链接
- 关联参数和响应模型

### ApiParameter (API 参数)
- 参数名称、类型、是否必需
- 默认值和描述

### ApiResponse (API 响应)
- 响应字段名称、类型、描述

### ApiUsage (API 使用记录)
- 用户、API、消耗积分、状态
- 创建时间

### Subscription (订阅)
- 用户订阅计划、积分、价格
- 开始和结束时间

## 快速开始

### 1. 安装依赖

```bash
npm install
```

### 2. 配置环境变量

创建 `.env` 文件：

```env
DATABASE_URL="file:./dev.db"
```

### 3. 初始化数据库

```bash
# 生成 Prisma Client
npm run db:generate

# 运行数据库迁移
npm run db:migrate

# 填充种子数据
npm run db:seed
```

### 4. 启动开发服务器

```bash
npm run dev
```

访问 http://localhost:3000

## 可用脚本

- `npm run dev` - 启动开发服务器
- `npm run build` - 构建生产版本（包含 Prisma Client 生成）
- `npm run start` - 启动生产服务器
- `npm run lint` - 运行 ESLint
- `npm run db:generate` - 生成 Prisma Client
- `npm run db:push` - 推送 schema 变更到数据库
- `npm run db:migrate` - 创建并运行迁移
- `npm run db:studio` - 打开 Prisma Studio
- `npm run db:seed` - 运行种子文件

## API 路由

### GET /api/categories
获取所有 API 分类及其关联的 API

**响应示例:**
```json
{
  "success": true,
  "data": [
    {
      "id": "xxx",
      "name": "人工智能",
      "description": "AI 相关的 API 接口",
      "icon": "brain",
      "apis": [...]
    }
  ]
}
```

### GET /api/apis
获取所有 API 及其关联的分类、参数和响应

**响应示例:**
```json
{
  "success": true,
  "data": [
    {
      "id": "xxx",
      "name": "GPT-4 对话 API",
      "description": "OpenAI GPT-4 模型对话接口",
      "endpoint": "/api/v1/chat/gpt4",
      "method": "POST",
      "category": {...},
      "parameters": [...],
      "responses": [...]
    }
  ]
}
```

## Prisma v7 特性

本项目使用 Prisma v7，具有以下特性：

- **ESM Only**: 项目使用 ES Modules
- **Driver Adapter**: 使用 `@prisma/adapter-better-sqlite3`
- **配置分离**: `prisma.config.ts` 替代 schema 中的环境变量
- **显式输出**: Prisma Client 生成到 `./generated` 目录

## 开发注意事项

1. **修改数据模型后**:
   ```bash
   npm run db:migrate
   ```

2. **重新生成 Prisma Client**:
   ```bash
   npm run db:generate
   ```

3. **查看数据库**:
   ```bash
   npm run db:studio
   ```

4. **环境变量加载**:
   - Prisma v7 需要手动加载环境变量
   - 在 `prisma.config.ts` 和 `src/lib/prisma.ts` 中使用 `import "dotenv/config"`

## 页面路由

- `/` - 首页
- `/api-market` - API 市场
- `/api-detail` - API 详情
- `/pricing` - 定价方案

## License

MIT
