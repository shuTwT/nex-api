import "dotenv/config";
import { Prisma, PrismaClient } from "../generated/client";
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
    const apis: Prisma.ApiCreateManyInput[] = [
      
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
    const apis: Prisma.ApiCreateManyInput[] = [
      
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
    const apis: Prisma.ApiCreateManyInput[] = [
      {
        name: "天气查询",
        alias: "weather",
        description: "实时天气查询，支持城市名称或经纬度",
        endpoint: "http://api-collection:8005/api/v1/weather/query",
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
        endpoint: "http://api-collection:8005/api/v1/ip/query",
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
        endpoint: "http://api-collection:8005/api/v1/qrcode",
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
        endpoint: "http://api-collection:8005/api/v1/shorturl/create",
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
