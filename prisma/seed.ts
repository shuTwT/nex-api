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
      name: "天气地理",
      description: "天气、地理信息相关 API",
      icon: "cloud",
    },
    {
      name: "工具",
      description: "实用工具类 API",
      icon: "wrench",
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

  const imageCategory = await prisma.apiCategory.findUnique({
    where: { name: "图像" },
  });

  const weatherCategory = await prisma.apiCategory.findUnique({
    where: { name: "天气地理" },
  });

  const toolCategory = await prisma.apiCategory.findUnique({
    where: { name: "工具" },
  });

  if (aiCategory) {
    const apis = [
      {
        name: "GPT-4 对话",
        alias: "gpt4-chat",
        description: "OpenAI GPT-4 模型对话接口，支持上下文对话",
        endpoint: "/api/v1/chat/gpt4",
        method: "POST",
        categoryId: aiCategory.id,
        pricing: 20,
        documentation: "https://docs.example.com/gpt4",
        isActive: true,
      },
      {
        name: "Claude 对话",
        alias: "claude-chat",
        description: "Anthropic Claude 模型对话接口，支持长文本",
        endpoint: "/api/v1/chat/claude",
        method: "POST",
        categoryId: aiCategory.id,
        pricing: 15,
        documentation: "https://docs.example.com/claude",
        isActive: true,
      },
    ];

    for (const api of apis) {
      const created = await prisma.api.upsert({
        where: { 
          alias: api.alias,
        },
        update: {
          endpoint: api.endpoint,
        },
        create: api,
      });
      console.log(`Created API: ${created.name}`);
    }
  }

  if (imageCategory) {
    const apis = [
      {
        name: "图像识别",
        alias: "image-recognition",
        description: "通用图像识别接口，支持物体、场景识别",
        endpoint: "http://localhost:3000/api/v1/image/recognize",
        method: "POST",
        categoryId: imageCategory.id,
        pricing: 5,
        documentation: "https://docs.example.com/image-recognition",
        isActive: true,
      },
      {
        name: "OCR 文字识别",
        alias: "ocr",
        description: "图片文字识别提取，支持多语言",
        endpoint: "http://localhost:3000/api/v1/image/ocr",
        method: "POST",
        categoryId: imageCategory.id,
        pricing: 3,
        documentation: "https://docs.example.com/ocr",
        isActive: true,
      },
    ];

    for (const api of apis) {
      const created = await prisma.api.upsert({
        where: { 
          alias: api.alias,
         },
        update: {
          endpoint: api.endpoint,
        },
        create: api,
      });
      console.log(`Created API: ${created.name}`);
    }
  }

  if (weatherCategory) {
    const apis = [
      {
        name: "天气查询",
        alias: "weather",
        description: "实时天气查询，支持城市名称或经纬度",
        endpoint: "http://localhost:3000/api/v1/weather/query",
        method: "GET",
        categoryId: weatherCategory.id,
        pricing: 1,
        documentation: "https://docs.example.com/weather",
        isActive: true,
      },
      {
        name: "IP 地址查询",
        alias: "ip-query",
        description: "IP 地址定位查询，返回地理位置信息",
        endpoint: "http://localhost:3000/api/v1/ip/query",
        method: "GET",
        categoryId: weatherCategory.id,
        pricing: 0,
        documentation: "https://docs.example.com/ip-query",
        isActive: true,
      },
    ];

    for (const api of apis) {
      const created = await prisma.api.upsert({
        where: { 
          alias: api.alias,
         },
        update: {
          endpoint: api.endpoint,
        },
        create: api,
      });
      console.log(`Created API: ${created.name}`);
    }
  }

  if (toolCategory) {
    const apis = [
      {
        name: "二维码生成",
        alias: "qrcode-gen",
        description: "生成二维码图片，支持自定义尺寸和颜色",
        endpoint: "http://localhost:3000/api/v1/qrcode/generate",
        method: "POST",
        categoryId: toolCategory.id,
        pricing: 0,
        documentation: "https://docs.example.com/qrcode",
        isActive: true,
      },
      {
        name: "短链接生成",
        alias: "shorturl",
        description: "将长链接转换为短链接",
        endpoint: "http://localhost:3000/api/v1/shorturl/create",
        method: "POST",
        categoryId: toolCategory.id,
        pricing: 0,
        documentation: "https://docs.example.com/shorturl",
        isActive: true,
      },
    ];

    for (const api of apis) {
      const created = await prisma.api.upsert({
        where: { 
          alias: api.alias,
        },
        update: {
          endpoint: api.endpoint,
        },
        create: api,
      });
      console.log(`Created API: ${created.name}`);
    }
  }

  const subscriptionPlans = [
    {
      title: "免费版",
      price: 0,
      totalCredits: 1000,
      sortOrder: 1,
      validityDuration: 1,
      validityUnit: "month",
      creditResetCycle: "month",
      isActive: true,
    },
    {
      title: "专业版",
      price: 29.9,
      totalCredits: 3000,
      sortOrder: 2,
      validityDuration: 1,
      validityUnit: "month",
      creditResetCycle: "month",
      isActive: true,
    },
  ];

  for (const plan of subscriptionPlans) {
    const created = await prisma.subscriptionPlan.upsert({
      where: { title: plan.title },
      update: {},
      create: plan,
    });
    console.log(`Created subscription plan: ${created.title}`);
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
