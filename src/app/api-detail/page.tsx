"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { MainLayout } from "@/components/main-layout";
import { 
  Book, 
  Code, 
  ArrowLeft, 
  Link as LinkIcon, 
  Tag, 
  Info, 
  Send, 
  RotateCcw,
  AlertTriangle,
  CheckCircle,
  Copy,
  Users,
  Zap
} from "lucide-react";
import { useState, useEffect, Suspense } from "react";
import { getApiDetail } from "@/app/actions/marketplace";

interface ApiDetail {
  id: string;
  name: string;
  description: string;
  alias: string;
  endpoint: string;
  method: string;
  pricing: number;
  category: string;
  isFree: boolean;
  isActive: boolean;
  todayCallCount: number;
  userCount: number;
  totalCallCount: number;
  createdAt: Date;
  updatedAt: Date;
}

function ApiDetailContent() {
  const searchParams = useSearchParams();
  const apiId = searchParams.get("id");
  const [api, setApi] = useState<ApiDetail | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [activeSection, setActiveSection] = useState("detail");
  const [baseUrl, setBaseUrl] = useState("");

  useEffect(() => {
    setBaseUrl(window.location.origin);
  }, []);

  useEffect(() => {
    if (apiId) {
      loadApiDetail();
    }
  }, [apiId]);

  async function loadApiDetail() {
    if (!apiId) return;
    
    setIsLoading(true);
    const result = await getApiDetail(apiId);
    if (result.success && result.data) {
      setApi(result.data);
    }
    setIsLoading(false);
  }

  useEffect(() => {
    const handleScroll = () => {
      const sections = ["detail", "request", "request-example", "response", "response-example", "error-codes", "examples"];
      const scrollPosition = window.scrollY + 100;

      for (const section of sections) {
        const element = document.getElementById(section);
        if (element) {
          const offsetTop = element.offsetTop;
          const offsetHeight = element.offsetHeight;
          
          if (scrollPosition >= offsetTop && scrollPosition < offsetTop + offsetHeight) {
            setActiveSection(section);
            break;
          }
        }
      }
    };

    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  const scrollToSection = (sectionId: string) => {
    const element = document.getElementById(sectionId);
    if (element) {
      const offsetTop = element.offsetTop - 80;
      window.scrollTo({
        top: offsetTop,
        behavior: "smooth"
      });
    }
  };

  if (isLoading) {
    return (
      <MainLayout>
        <div className="flex items-center justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
        </div>
      </MainLayout>
    );
  }

  if (!api) {
    return (
      <MainLayout>
        <div className="flex flex-col items-center justify-center py-12">
          <AlertTriangle className="h-12 w-12 text-slate-300 mb-4" />
          <h3 className="text-lg font-medium text-slate-900 mb-2">API 不存在</h3>
          <p className="text-slate-500 mb-4">请检查 API ID 是否正确</p>
          <Link href="/api-market">
            <Button>
              <ArrowLeft className="h-4 w-4 mr-2" />
              返回 API 市场
            </Button>
          </Link>
        </div>
      </MainLayout>
    );
  }

  const apiUrl = `${baseUrl}/api/v1/${api.alias}`;

  return (
    <MainLayout>
      {/* Hero Section */}
      <section className="relative bg-gradient-to-br from-blue-50 via-white to-blue-50">
        <div className="absolute inset-0 overflow-hidden">
          <div className="absolute -top-40 -left-40 w-80 h-80 bg-blue-100/50 rounded-full blur-3xl"></div>
          <div className="absolute top-40 right-20 w-96 h-96 bg-blue-50/50 rounded-full blur-3xl"></div>
        </div>
        <div className="container relative px-4 py-16 md:px-6 mx-auto">
          <div className="flex flex-col items-center text-center space-y-6">
            <div>
              <h1 className="text-4xl md:text-5xl font-bold mb-3 text-slate-900">
                {api.name}
              </h1>
              <p className="text-lg text-slate-600">
                {api.description}
              </p>
            </div>

            {/* API Stats */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
              <div className="text-center">
                <div className="text-sm text-slate-500 mb-1">接口状态</div>
                <div className="flex items-center justify-center gap-1">
                  <CheckCircle className="h-4 w-4 text-green-600" />
                  <span className="font-semibold text-slate-700">{api.isActive ? "正常" : "停用"}</span>
                </div>
              </div>
              <div className="text-center">
                <div className="text-sm text-slate-500 mb-1">请求方式</div>
                <Badge className="bg-blue-100 text-blue-700 border-blue-200">{api.method}</Badge>
              </div>
              <div className="text-center">
                <div className="text-sm text-slate-500 mb-1">计费方式</div>
                <div className="flex items-center justify-center gap-1">
                  <span className="font-semibold text-slate-700">
                    {api.isFree ? "免费" : `${api.pricing} 积分/次`}
                  </span>
                </div>
              </div>
              <div className="text-center">
                <div className="text-sm text-slate-500 mb-1">今日调用</div>
                <div className="flex items-center justify-center gap-1">
                  <Zap className="h-4 w-4 text-blue-600" />
                  <span className="font-semibold text-slate-700">{api.todayCallCount.toLocaleString()}</span>
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => scrollToSection("detail")}>
                接口详情
              </Button>
              <Button variant="outline" size="sm" onClick={() => scrollToSection("request-example")}>
                请求示例
              </Button>
              <Button variant="outline" size="sm" onClick={() => scrollToSection("response-example")}>
                响应示例
              </Button>
              <Button variant="outline" size="sm" onClick={() => scrollToSection("error-codes")}>
                错误代码
              </Button>
            </div>
          </div>
        </div>
      </section>

      {/* Main Content */}
      <div className="flex-1 container px-4 py-8 md:px-6 mx-auto">
        <div className="grid grid-cols-1 md:grid-cols-[280px_1fr] gap-6">
          {/* Sidebar - Documentation Navigation */}
          <aside className="hidden md:block">
            <div className="sticky top-20 space-y-4">
              <div className="rounded-lg border bg-white p-4">
                <div className="flex items-center gap-2 mb-4">
                  <Book className="h-5 w-5 text-blue-600" />
                  <h3 className="font-semibold">文档目录</h3>
                </div>
                <ScrollArea className="h-[600px]">
                  <nav className="space-y-1">
                    <div className="space-y-2">
                      <div className="text-sm font-medium text-gray-700 px-3 py-1">接口信息</div>
                      <button
                        onClick={() => scrollToSection("detail")}
                        className={`flex items-center gap-2 text-sm w-full text-left px-3 py-2 rounded-md transition-colors ${
                          activeSection === "detail"
                            ? "text-blue-600 bg-blue-50"
                            : "text-gray-600 hover:bg-gray-50"
                        }`}
                      >
                        <Info className="h-4 w-4" />
                        接口详情
                      </button>
                      <button
                        onClick={() => scrollToSection("request")}
                        className={`flex items-center gap-2 text-sm w-full text-left px-3 py-2 rounded-md transition-colors ${
                          activeSection === "request"
                            ? "text-blue-600 bg-blue-50"
                            : "text-gray-600 hover:bg-gray-50"
                        }`}
                      >
                        <Send className="h-4 w-4" />
                        请求参数
                      </button>
                      <button
                        onClick={() => scrollToSection("request-example")}
                        className={`flex items-center gap-2 text-sm w-full text-left px-3 py-2 rounded-md transition-colors ${
                          activeSection === "request-example"
                            ? "text-blue-600 bg-blue-50"
                            : "text-gray-600 hover:bg-gray-50"
                        }`}
                      >
                        <Send className="h-4 w-4" />
                        请求示例
                      </button>
                    </div>
                    <Separator className="my-2" />
                    <div className="space-y-2">
                      <div className="text-sm font-medium text-gray-700 px-3 py-1">返回信息</div>
                      <button
                        onClick={() => scrollToSection("response")}
                        className={`flex items-center gap-2 text-sm w-full text-left px-3 py-2 rounded-md transition-colors ${
                          activeSection === "response"
                            ? "text-blue-600 bg-blue-50"
                            : "text-gray-600 hover:bg-gray-50"
                        }`}
                      >
                        <RotateCcw className="h-4 w-4" />
                        返回参数
                      </button>
                      <button
                        onClick={() => scrollToSection("response-example")}
                        className={`flex items-center gap-2 text-sm w-full text-left px-3 py-2 rounded-md transition-colors ${
                          activeSection === "response-example"
                            ? "text-blue-600 bg-blue-50"
                            : "text-gray-600 hover:bg-gray-50"
                        }`}
                      >
                        <RotateCcw className="h-4 w-4" />
                        响应示例
                      </button>
                      <button
                        onClick={() => scrollToSection("error-codes")}
                        className={`flex items-center gap-2 text-sm w-full text-left px-3 py-2 rounded-md transition-colors ${
                          activeSection === "error-codes"
                            ? "text-blue-600 bg-blue-50"
                            : "text-gray-600 hover:bg-gray-50"
                        }`}
                      >
                        <AlertTriangle className="h-4 w-4" />
                        错误代码
                      </button>
                    </div>
                    <Separator className="my-2" />
                    <div className="space-y-2">
                      <div className="text-sm font-medium text-gray-700 px-3 py-1">代码示例</div>
                      <button
                        onClick={() => scrollToSection("examples")}
                        className={`flex items-center gap-2 text-sm w-full text-left px-3 py-2 rounded-md transition-colors ${
                          activeSection === "examples"
                            ? "text-blue-600 bg-blue-50"
                            : "text-gray-600 hover:bg-gray-50"
                        }`}
                      >
                        <Code className="h-4 w-4" />
                        代码示例
                      </button>
                    </div>
                  </nav>
                </ScrollArea>
              </div>

              <Link href="/api-market" className="block">
                <Button className="w-full gap-2" variant="outline">
                  <ArrowLeft className="h-4 w-4" />
                  返回 API 市场
                </Button>
              </Link>
            </div>
          </aside>

          {/* Main Content Area */}
          <main className="space-y-6">
            {/* API Info Card */}
            <div id="detail" className="rounded-lg border bg-white">
              <div className="p-6 space-y-6">
                <div>
                  <h2 className="text-2xl font-bold flex items-center gap-2 mb-4">
                    <Info className="h-5 w-5 text-blue-600" />
                    接口详情
                  </h2>
                  <Separator />
                </div>

                {/* API URL */}
                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-sm font-medium">
                    <LinkIcon className="h-4 w-4 text-gray-500" />
                    接口地址
                  </div>
                  <div className="flex gap-2">
                    <code className="flex-1 block bg-slate-900 text-white p-3 rounded-md text-sm font-mono">
                      {apiUrl}
                    </code>
                    <Button size="icon" variant="outline" className="shrink-0">
                      <Copy className="h-4 w-4" />
                    </Button>
                  </div>
                </div>

                {/* API Tags */}
                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-sm font-medium">
                    <Tag className="h-4 w-4 text-purple-600" />
                    接口标签
                  </div>
                  <div className="flex gap-2 flex-wrap">
                    <Badge className="bg-blue-50 text-blue-700 border-blue-200">{api.method}</Badge>
                    <Badge className="bg-green-50 text-green-700 border-green-200">{api.category}</Badge>
                    <Badge className={api.isActive ? "bg-green-50 text-green-700 border-green-200" : "bg-red-50 text-red-700 border-red-200"}>
                      {api.isActive ? "正常" : "停用"}
                    </Badge>
                    <Badge className={api.isFree ? "bg-orange-50 text-orange-700 border-orange-200" : "bg-purple-50 text-purple-700 border-purple-200"}>
                      {api.isFree ? "免费" : "会员专享"}
                    </Badge>
                  </div>
                </div>

                {/* API Description */}
                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-sm font-medium">
                    <Info className="h-4 w-4 text-gray-500" />
                    接口描述
                  </div>
                  <div className="bg-gray-50 p-4 rounded-md">
                    <p className="text-sm text-gray-700">{api.description}</p>
                  </div>
                </div>

                {/* API Stats */}
                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-sm font-medium">
                    <Users className="h-4 w-4 text-gray-500" />
                    使用统计
                  </div>
                  <div className="grid grid-cols-3 gap-4">
                    <div className="bg-gray-50 p-4 rounded-md text-center">
                      <p className="text-2xl font-bold text-slate-900">{api.userCount}</p>
                      <p className="text-xs text-slate-500 mt-1">使用人数</p>
                    </div>
                    <div className="bg-gray-50 p-4 rounded-md text-center">
                      <p className="text-2xl font-bold text-slate-900">{api.todayCallCount.toLocaleString()}</p>
                      <p className="text-xs text-slate-500 mt-1">今日调用</p>
                    </div>
                    <div className="bg-gray-50 p-4 rounded-md text-center">
                      <p className="text-2xl font-bold text-slate-900">{api.totalCallCount.toLocaleString()}</p>
                      <p className="text-xs text-slate-500 mt-1">累计调用</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Request Parameters */}
            <div id="request" className="rounded-lg border bg-white">
              <div className="p-6 space-y-4">
                <div>
                  <h2 className="text-2xl font-bold flex items-center gap-2 mb-4">
                    <Send className="h-5 w-5 text-blue-600" />
                    请求参数
                  </h2>
                  <Separator />
                </div>

                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b bg-gray-50">
                        <th className="text-left p-3 font-medium">参数名</th>
                        <th className="text-left p-3 font-medium">类型</th>
                        <th className="text-left p-3 font-medium">必填</th>
                        <th className="text-left p-3 font-medium">说明</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr className="border-b">
                        <td className="p-3 font-mono text-blue-600">token</td>
                        <td className="p-3">string</td>
                        <td className="p-3"><Badge variant="outline" className="bg-red-50 text-red-700 border-red-200">是</Badge></td>
                        <td className="p-3">您的 API Token</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>

            {/* Request Example */}
            <div id="request-example" className="rounded-lg border bg-white p-6">
              <h2 className="text-2xl font-bold flex items-center gap-2 mb-4">
                <Send className="h-5 w-5 text-blue-600" />
                请求示例
              </h2>
              <Separator className="mb-6" />

              <div className="space-y-4">
                <div className="space-y-2">
                  <h3 className="font-semibold">cURL</h3>
                  <div className="bg-slate-900 text-white p-4 rounded-md overflow-x-auto">
                    <pre className="text-sm font-mono">
{`curl -X ${api.method} "${apiUrl}?token=YOUR_TOKEN" \\
  -H "Content-Type: application/json"`}</pre>
                  </div>
                </div>

                <div className="space-y-2">
                  <h3 className="font-semibold">HTTP</h3>
                  <div className="bg-slate-900 text-white p-4 rounded-md overflow-x-auto">
                    <pre className="text-sm font-mono">
{`${api.method} ${api.endpoint}?token=YOUR_TOKEN HTTP/1.1
Host: ${baseUrl?.replace(/^https?:\/\//, '') || '[YOUR_DOMAIN]'}
Content-Type: application/json`}</pre>
                  </div>
                </div>
              </div>
            </div>

            {/* Response Parameters */}
            <div id="response" className="rounded-lg border bg-white p-6">
              <h2 className="text-2xl font-bold flex items-center gap-2 mb-4">
                <RotateCcw className="h-5 w-5 text-blue-600" />
                返回参数说明
              </h2>
              <Separator className="mb-6" />

              <div className="space-y-6">
                <div className="space-y-2">
                  <h3 className="font-semibold">返回参数</h3>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b bg-gray-50">
                          <th className="text-left p-3 font-medium">字段</th>
                          <th className="text-left p-3 font-medium">类型</th>
                          <th className="text-left p-3 font-medium">说明</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr className="border-b">
                          <td className="p-3 font-mono text-blue-600">success</td>
                          <td className="p-3">boolean</td>
                          <td className="p-3">请求是否成功</td>
                        </tr>
                        <tr className="border-b">
                          <td className="p-3 font-mono text-blue-600">data</td>
                          <td className="p-3">object</td>
                          <td className="p-3">返回数据对象</td>
                        </tr>
                        <tr className="border-b">
                          <td className="p-3 font-mono text-blue-600">error</td>
                          <td className="p-3">string</td>
                          <td className="p-3">错误信息（失败时）</td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </div>
              </div>
            </div>

            {/* Response Example */}
            <div id="response-example" className="rounded-lg border bg-white p-6">
              <h2 className="text-2xl font-bold flex items-center gap-2 mb-4">
                <RotateCcw className="h-5 w-5 text-blue-600" />
                响应示例
              </h2>
              <Separator className="mb-6" />

              <div className="space-y-4">
                <div className="space-y-2">
                  <h3 className="font-semibold">成功响应</h3>
                  <div className="bg-slate-900 text-white p-4 rounded-md overflow-x-auto">
                    <pre className="text-sm font-mono">
{`{
  "success": true,
  "data": {
    // 具体返回数据根据接口而定
  }
}`}</pre>
                  </div>
                </div>

                <div className="space-y-2">
                  <h3 className="font-semibold">失败响应</h3>
                  <div className="bg-slate-900 text-white p-4 rounded-md overflow-x-auto">
                    <pre className="text-sm font-mono">
{`{
  "success": false,
  "error": "错误信息描述"
}`}</pre>
                  </div>
                </div>
              </div>
            </div>

            {/* Error Codes */}
            <div id="error-codes" className="rounded-lg border bg-white p-6">
              <h2 className="text-2xl font-bold flex items-center gap-2 mb-4">
                <AlertTriangle className="h-5 w-5 text-orange-600" />
                错误代码
              </h2>
              <Separator className="mb-6" />

              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b bg-gray-50">
                      <th className="text-left p-3 font-medium">错误代码</th>
                      <th className="text-left p-3 font-medium">说明</th>
                      <th className="text-left p-3 font-medium">解决方案</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="border-b">
                      <td className="p-3 font-mono text-red-600">400</td>
                      <td className="p-3">请求参数错误</td>
                      <td className="p-3">检查请求参数是否正确</td>
                    </tr>
                    <tr className="border-b">
                      <td className="p-3 font-mono text-red-600">401</td>
                      <td className="p-3">未授权访问</td>
                      <td className="p-3">请检查 Token 是否正确</td>
                    </tr>
                    <tr className="border-b">
                      <td className="p-3 font-mono text-red-600">403</td>
                      <td className="p-3">禁止访问</td>
                      <td className="p-3">检查账户权限或余额是否充足</td>
                    </tr>
                    <tr className="border-b">
                      <td className="p-3 font-mono text-red-600">404</td>
                      <td className="p-3">接口不存在</td>
                      <td className="p-3">检查请求的接口地址是否正确</td>
                    </tr>
                    <tr className="border-b">
                      <td className="p-3 font-mono text-red-600">429</td>
                      <td className="p-3">请求过于频繁</td>
                      <td className="p-3">降低请求频率或升级套餐</td>
                    </tr>
                    <tr className="border-b">
                      <td className="p-3 font-mono text-red-600">500</td>
                      <td className="p-3">服务器内部错误</td>
                      <td className="p-3">联系技术支持或稍后重试</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            {/* Code Examples */}
            <div id="examples" className="rounded-lg border bg-white p-6">
              <h2 className="text-2xl font-bold flex items-center gap-2 mb-4">
                <Code className="h-5 w-5 text-blue-600" />
                代码示例
              </h2>
              <Separator className="mb-6" />

              <Tabs defaultValue="curl" className="space-y-4">
                <TabsList>
                  <TabsTrigger value="curl">cURL</TabsTrigger>
                  <TabsTrigger value="javascript">JavaScript</TabsTrigger>
                  <TabsTrigger value="python">Python</TabsTrigger>
                  <TabsTrigger value="java">Java</TabsTrigger>
                  <TabsTrigger value="php">PHP</TabsTrigger>
                </TabsList>

                <TabsContent value="curl" className="space-y-4">
                  <div className="flex items-center justify-between">
                    <h3 className="font-semibold">cURL 请求示例</h3>
                    <Button size="sm" variant="outline">
                      <Copy className="h-4 w-4 mr-2" />
                      复制
                    </Button>
                  </div>
                  <div className="bg-slate-900 text-white p-4 rounded-md overflow-x-auto">
                    <pre className="text-sm font-mono">
{`curl -X ${api.method} "${apiUrl}?token=YOUR_TOKEN" \\
  -H "Content-Type: application/json"`}</pre>
                  </div>
                </TabsContent>

                <TabsContent value="javascript" className="space-y-4">
                  <div className="flex items-center justify-between">
                    <h3 className="font-semibold">JavaScript (Fetch)</h3>
                    <Button size="sm" variant="outline">
                      <Copy className="h-4 w-4 mr-2" />
                      复制
                    </Button>
                  </div>
                  <div className="bg-slate-900 text-white p-4 rounded-md overflow-x-auto">
                    <pre className="text-sm font-mono">
{`const url = '${apiUrl}?token=YOUR_TOKEN';

fetch(url, {
  method: '${api.method}',
  headers: {
    'Content-Type': 'application/json'
  }
})
.then(response => response.json())
.then(data => {
  console.log(data);
})
.catch(error => {
  console.error('Error:', error);
});`}</pre>
                  </div>
                </TabsContent>

                <TabsContent value="python" className="space-y-4">
                  <div className="flex items-center justify-between">
                    <h3 className="font-semibold">Python (requests)</h3>
                    <Button size="sm" variant="outline">
                      <Copy className="h-4 w-4 mr-2" />
                      复制
                    </Button>
                  </div>
                  <div className="bg-slate-900 text-white p-4 rounded-md overflow-x-auto">
                    <pre className="text-sm font-mono">
{`import requests

url = '${apiUrl}'
params = {
    'token': 'YOUR_TOKEN'
}
headers = {
    'Content-Type': 'application/json'
}

response = requests.${api.method.toLowerCase()}(url, params=params, headers=headers)
data = response.json()
print(data)`}</pre>
                  </div>
                </TabsContent>

                <TabsContent value="java" className="space-y-4">
                  <div className="flex items-center justify-between">
                    <h3 className="font-semibold">Java (OkHttp)</h3>
                    <Button size="sm" variant="outline">
                      <Copy className="h-4 w-4 mr-2" />
                      复制
                    </Button>
                  </div>
                  <div className="bg-slate-900 text-white p-4 rounded-md overflow-x-auto">
                    <pre className="text-sm font-mono">
{`OkHttpClient client = new OkHttpClient();

String url = "${apiUrl}?token=YOUR_TOKEN";

Request request = new Request.Builder()
    .url(url)
    .${api.method.toLowerCase()}()
    .addHeader("Content-Type", "application/json")
    .build();

try (Response response = client.newCall(request).execute()) {
    String responseData = response.body().string();
    System.out.println(responseData);
}`}</pre>
                  </div>
                </TabsContent>

                <TabsContent value="php" className="space-y-4">
                  <div className="flex items-center justify-between">
                    <h3 className="font-semibold">PHP (cURL)</h3>
                    <Button size="sm" variant="outline">
                      <Copy className="h-4 w-4 mr-2" />
                      复制
                    </Button>
                  </div>
                  <div className="bg-slate-900 text-white p-4 rounded-md overflow-x-auto">
                    <pre className="text-sm font-mono">
{`<?php
$url = '${apiUrl}?token=YOUR_TOKEN';

$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, $url);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Content-Type: application/json'
]);

$response = curl_exec($ch);
curl_close($ch);

$data = json_decode($response, true);
print_r($data);
?>`}</pre>
                  </div>
                </TabsContent>
              </Tabs>
            </div>
          </main>
        </div>
      </div>
    </MainLayout>
  );
}

export default function ApiDetailPage() {
  return (
    <Suspense fallback={
      <MainLayout>
        <div className="flex items-center justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
        </div>
      </MainLayout>
    }>
      <ApiDetailContent />
    </Suspense>
  );
}
