# One API - 全栈 SSR 项目

一个基于 Next.js 16 的全栈 API 管理系统。

## 概览

![首页](./images/screenshot-index.png)

![API 市场](./images/screenshot-store.png)

![定价](./images/screenshot-price.png)



## 技术栈

- **前端框架**: Next.js 16.1.6 (App Router)
- **语言**: TypeScript 5.x
- **UI 组件**: Radix UI + shadcn/ui
- **样式**: Tailwind CSS 4
- **ORM**: Prisma 7.5.0
- **数据库**: SQLite (better-sqlite3)、mysql8、postgresql
- **运行时**: ESM (ECMAScript Modules)

**必须启用 redis 功能，否则无法统计用量和限流**

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

## 快速开始

### 1. 安装依赖

```bash
pnpm install
```

### 2. 配置环境变量

复制一份`.env.example`并重命名为 `.env` 文件：

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

访问 <http://localhost:3000>

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

## License

MIT
