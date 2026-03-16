import "dotenv/config";
import { PrismaClient } from "../generated/client";
import { PrismaBetterSqlite3 } from "@prisma/adapter-better-sqlite3";

const databaseUrl = process.env.DATABASE_URL;
if (!databaseUrl) {
  throw new Error("DATABASE_URL environment variable is not set");
}

const adapter = new PrismaBetterSqlite3({
  url: databaseUrl,
});

const prisma = new PrismaClient({ adapter });

async function main() {
  console.log("Starting seed...");

  const categories = [
    {
      name: "人工智能",
      description: "AI 相关的 API 接口",
      icon: "brain",
    },
    {
      name: "图像",
      description: "图像识别、处理相关 API",
      icon: "image",
    },
    {
      name: "杂项",
      description: "其他 API",
      icon: "message-square",
    },
    {
      name: "天气地理",
      description: "数据统计、可视化 API",
      icon: "bar-chart",
    },
  ];

  for (const category of categories) {
    const created = await prisma.apiCategory.upsert({
      where: { name: category.name },
      update: {},
      create: category,
    });
    console.log(`Created category: ${created.name}`);
  }

  const aiCategory = await prisma.apiCategory.findUnique({
    where: { name: "人工智能" },
  });

  if (aiCategory) {
    const apis = [
      {
        name: "GPT-4 对话 API",
        description: "OpenAI GPT-4 模型对话接口，支持上下文对话",
        endpoint: "/api/v1/chat/gpt4",
        method: "POST",
        categoryId: aiCategory.id,
        pricing: "0.02积分/次",
        documentation: "https://docs.example.com/gpt4",
        isActive: true,
      },
      {
        name: "Claude 对话 API",
        description: "Anthropic Claude 模型对话接口，支持长文本",
        endpoint: "/api/v1/chat/claude",
        method: "POST",
        categoryId: aiCategory.id,
        pricing: "0.015积分/次",
        documentation: "https://docs.example.com/claude",
        isActive: true,
      },
    ];

    for (const api of apis) {
      const created = await prisma.api.upsert({
        where: { endpoint: api.endpoint },
        update: {},
        create: api,
      });
      console.log(`Created API: ${created.name}`);
    }
  }

  console.log("Seed completed!");
}

main()
  .catch((e) => {
    console.error(e);
    process.exit(1);
  })
  .finally(async () => {
    await prisma.$disconnect();
  });
