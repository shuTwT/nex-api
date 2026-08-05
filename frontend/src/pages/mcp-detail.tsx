import { useCallback, useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router";
import {
  Activity,
  ArrowLeft,
  BookOpen,
  Check,
  Copy,
  ExternalLink,
  Folder,
  Plug,
  Radio,
  RefreshCw,
  TerminalSquare,
  Users,
  Wrench,
} from "lucide-react";
import { InlineAd } from "@/components/ads";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { api, responseData } from "@/lib/api";
import { AdPosition } from "@/types/ad-position";

interface McpDetail {
  id: string;
  name: string;
  identifier: string;
  category: string;
  description: string;
  documentation?: string;
  type: string;
  pricing: number;
  isFree: boolean;
  isActive: boolean;
  todayCallCount: number;
  userCount: number;
  totalCallCount: number;
}

interface McpTool {
  name: string;
  title?: string;
  description?: string;
  inputSchema?: {
    properties?: Record<string, { type?: string; description?: string }>;
    required?: string[];
  };
}

const TYPE_LABELS: Record<string, string> = {
  stdio: "stdio",
  sse: "SSE",
  streamableHttp: "Streamable HTTP",
};

function CopyButton({ value, label = "复制" }: { value: string; label?: string }) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    await navigator.clipboard?.writeText(value);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 1600);
  };

  return (
    <Button size="sm" variant="ghost" onClick={copy} aria-label={label}>
      {copied ? <Check data-icon="inline-start" /> : <Copy data-icon="inline-start" />}
      {copied ? "已复制" : label}
    </Button>
  );
}

function DetailSkeleton() {
  return (
    <main className="container mx-auto max-w-7xl px-4 py-8 md:px-6 lg:py-12">
      <div className="flex flex-col gap-8">
        <Skeleton className="h-5 w-56" />
        <div className="flex items-center gap-4">
          <Skeleton className="size-16 rounded-xl" />
          <div className="flex flex-col gap-3">
            <Skeleton className="h-8 w-64" />
            <Skeleton className="h-4 w-36" />
          </div>
        </div>
        <div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_320px]">
          <div className="flex flex-col gap-6">
            <Skeleton className="h-40 w-full" />
            <Skeleton className="h-72 w-full" />
          </div>
          <Skeleton className="h-80 w-full" />
        </div>
      </div>
    </main>
  );
}

export default function McpDetailPage() {
  const { id: mcpId } = useParams<{ id: string }>();
  const [service, setService] = useState<McpDetail | null>(null);
  const [tools, setTools] = useState<McpTool[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingTools, setIsLoadingTools] = useState(false);
  const [hasRequestedTools, setHasRequestedTools] = useState(false);
  const [toolsError, setToolsError] = useState("");
  const [baseUrl, setBaseUrl] = useState("");

  useEffect(() => setBaseUrl(window.location.origin), []);

  const loadTools = useCallback(async () => {
    if (!mcpId) return;
    setHasRequestedTools(true);
    setIsLoadingTools(true);
    setToolsError("");
    const result = await api.marketplace_mcp_services_id_tools_route_get({ id: mcpId });
    const data = responseData<McpTool[]>(result);
    if (data) {
      setTools(data);
    } else {
      setToolsError(result.error || "暂时无法连接服务端点，请稍后重试。");
    }
    setIsLoadingTools(false);
  }, [mcpId]);

  const loadService = useCallback(async () => {
    if (!mcpId) {
      setIsLoading(false);
      return;
    }
    setIsLoading(true);
    const result = await api.marketplace_mcp_services_id_route_get({ id: mcpId });
    const data = responseData<McpDetail>(result);
    setService(data);
    setIsLoading(false);
  }, [mcpId]);

  useEffect(() => {
    void loadService();
  }, [loadService]);

  const accessUrl = service ? `${baseUrl}/api/v1/mcp/${service.identifier}` : "";
  const commands = useMemo(() => {
    if (!service) return {};
    return {
      claude: `claude mcp add ${service.identifier} --transport http ${accessUrl}`,
      codex: `codex mcp add ${service.identifier} --url ${accessUrl}`,
      cursor: `在 Cursor 设置中添加远程 MCP：${accessUrl}`,
      vscode: `code --add-mcp '{"name":"${service.identifier}","url":"${accessUrl}"}'`,
    };
  }, [accessUrl, service]);

  if (isLoading) return <DetailSkeleton />;

  if (!service) {
    return (
      <main className="container mx-auto flex max-w-4xl flex-col items-center gap-4 px-4 py-24 text-center">
        <Plug className="size-12 text-muted-foreground" />
        <div className="flex flex-col gap-1">
          <h1 className="text-xl font-semibold">MCP 服务不存在</h1>
          <p className="text-muted-foreground">请检查服务链接是否正确。</p>
        </div>
        <Button variant="outline" asChild>
          <Link to="/mcp-market"><ArrowLeft data-icon="inline-start" />返回 MCP 市场</Link>
        </Button>
      </main>
    );
  }

  return (
    <main className="container mx-auto max-w-7xl px-4 py-8 md:px-6 lg:py-12">
      <div className="flex flex-col gap-8">
        <nav aria-label="面包屑" className="flex items-center gap-2 text-sm text-muted-foreground">
          <Link to="/mcp-market" className="transition-colors hover:text-foreground">MCP 市场</Link>
          <span aria-hidden="true">/</span>
          <span>{service.category || "未分类"}</span>
          <span aria-hidden="true">/</span>
          <span className="truncate text-foreground">{service.name}</span>
        </nav>

        <header className="flex flex-col gap-5 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex min-w-0 items-center gap-4">
            <div className="flex size-16 shrink-0 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
              <Plug className="size-8" />
            </div>
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-2">
                <h1 className="text-3xl font-semibold tracking-tight md:text-4xl">{service.name}</h1>
                {service.isActive && <Badge variant="secondary">已验证</Badge>}
              </div>
              <p className="mt-1 font-mono text-sm text-muted-foreground">@{service.identifier}</p>
            </div>
          </div>
          {service.documentation && (
            <Button variant="outline" asChild>
              <a href={service.documentation} target="_blank" rel="noreferrer">
                <BookOpen data-icon="inline-start" />文档<ExternalLink data-icon="inline-end" />
              </a>
            </Button>
          )}
        </header>

        <div className="grid items-start gap-10 lg:grid-cols-[minmax(0,1fr)_320px]">
          <div className="flex min-w-0 flex-col gap-10">
            <section aria-labelledby="about-title" className="flex flex-col gap-3">
              <h2 id="about-title" className="text-2xl font-semibold">关于 {service.name}</h2>
              <p className="max-w-4xl text-base leading-7 text-muted-foreground">
                {service.description || "该 MCP 服务暂未提供描述。"}
              </p>
            </section>

            <section aria-labelledby="connection-title" className="flex flex-col gap-4">
              <div>
                <h2 id="connection-title" className="text-2xl font-semibold">连接信息</h2>
                <p className="mt-1 text-sm text-muted-foreground">使用 API Token 访问 NexApi 提供的统一 MCP 网关。</p>
              </div>
              <Card size="sm">
                <CardContent className="flex items-center gap-3">
                  <code className="min-w-0 flex-1 break-all font-mono text-sm">{accessUrl}</code>
                  <CopyButton value={accessUrl} label="复制地址" />
                </CardContent>
              </Card>
            </section>

            <section aria-labelledby="access-title" className="flex flex-col gap-4">
              <h2 id="access-title" className="text-2xl font-semibold">接入方式</h2>
              <Tabs defaultValue="claude">
                <TabsList className="max-w-full overflow-x-auto">
                  <TabsTrigger value="claude"><TerminalSquare data-icon="inline-start" />Claude Code</TabsTrigger>
                  <TabsTrigger value="codex"><TerminalSquare data-icon="inline-start" />Codex</TabsTrigger>
                  <TabsTrigger value="cursor"><TerminalSquare data-icon="inline-start" />Cursor</TabsTrigger>
                  <TabsTrigger value="vscode"><TerminalSquare data-icon="inline-start" />VS Code</TabsTrigger>
                </TabsList>
                {Object.entries(commands).map(([client, command]) => (
                  <TabsContent key={client} value={client}>
                    <Card size="sm">
                      <CardContent className="flex items-center gap-3">
                        <code className="min-w-0 flex-1 break-all font-mono text-sm">{command}</code>
                        <CopyButton value={command} label="复制命令" />
                      </CardContent>
                    </Card>
                  </TabsContent>
                ))}
              </Tabs>
            </section>

            <section aria-labelledby="tools-title" className="flex flex-col gap-4">
              <div className="flex items-center justify-between gap-4">
                <div>
                  <h2 id="tools-title" className="text-2xl font-semibold">工具</h2>
                  <p className="mt-1 text-sm text-muted-foreground">由 MCP 服务实时返回的可用能力。</p>
                </div>
                <Button variant="outline" size="sm" onClick={() => void loadTools()} disabled={isLoadingTools}>
                  <RefreshCw data-icon="inline-start" />{isLoadingTools ? "获取中" : "获取工具"}
                </Button>
              </div>

              {isLoadingTools && tools.length === 0 && (
                <div className="flex flex-col gap-3">
                  <Skeleton className="h-24 w-full" />
                  <Skeleton className="h-24 w-full" />
                </div>
              )}

              {toolsError && (
                <Alert variant="destructive">
                  <Wrench />
                  <AlertTitle>工具获取失败</AlertTitle>
                  <AlertDescription>{toolsError}</AlertDescription>
                </Alert>
              )}

              {!isLoadingTools && !toolsError && tools.length === 0 && (
                <Card className="border-dashed">
                  <CardContent className="flex flex-col items-center gap-2 py-10 text-center">
                    <Wrench className="size-8 text-muted-foreground" />
                    <p className="font-medium">{hasRequestedTools ? "该服务暂未声明工具" : "尚未获取工具"}</p>
                    <p className="text-sm text-muted-foreground">
                      {hasRequestedTools ? "服务连接正常，但 tools/list 返回了空列表。" : "点击上方“获取工具”，实时连接服务并读取工具列表。"}
                    </p>
                  </CardContent>
                </Card>
              )}

              {tools.length > 0 && (
                <div className="grid gap-3">
                  {tools.map((tool) => {
                    const properties = Object.entries(tool.inputSchema?.properties || {});
                    const required = new Set(tool.inputSchema?.required || []);
                    return (
                      <Card key={tool.name} size="sm">
                        <CardHeader>
                          <CardTitle className="flex items-center gap-2">
                            <Wrench className="size-4 text-muted-foreground" />
                            <code>{tool.title || tool.name}</code>
                          </CardTitle>
                          <CardDescription>{tool.description || "该工具暂未提供说明。"}</CardDescription>
                          <CardAction><Badge variant="outline">{properties.length} 个参数</Badge></CardAction>
                        </CardHeader>
                        {properties.length > 0 && (
                          <CardContent className="flex flex-col gap-2">
                            <Separator />
                            <dl className="grid gap-2 pt-2 sm:grid-cols-2">
                              {properties.map(([name, property]) => (
                                <div key={name} className="min-w-0 rounded-lg bg-muted/50 px-3 py-2">
                                  <dt className="flex items-center gap-1 font-mono text-xs font-medium">
                                    {name}{required.has(name) && <span className="text-destructive">*</span>}
                                    {property.type && <Badge variant="secondary">{property.type}</Badge>}
                                  </dt>
                                  {property.description && <dd className="mt-1 text-xs text-muted-foreground">{property.description}</dd>}
                                </div>
                              ))}
                            </dl>
                          </CardContent>
                        )}
                      </Card>
                    );
                  })}
                </div>
              )}
            </section>
          </div>

          <aside className="flex flex-col gap-6 lg:sticky lg:top-24">
            <section aria-labelledby="sponsor-title" className="flex flex-col gap-3">
              <h2 id="sponsor-title" className="text-sm font-medium text-muted-foreground">推荐内容</h2>
              <InlineAd size="lg" position={AdPosition.SIDEBAR_TOP} />
            </section>

            <section aria-labelledby="basic-title" className="flex flex-col gap-3">
              <h2 id="basic-title" className="text-sm font-medium text-muted-foreground">基本信息</h2>
              <Card>
                <CardContent className="flex flex-col gap-4">
                  <div className="flex items-center gap-3">
                    <Radio className="size-5 text-muted-foreground" />
                    <div><div className="text-xs text-muted-foreground">传输方式</div><div className="font-medium">{TYPE_LABELS[service.type] || service.type}</div></div>
                  </div>
                  <Separator />
                  <div className="flex items-center gap-3">
                    <Folder className="size-5 text-muted-foreground" />
                    <div><div className="text-xs text-muted-foreground">分类</div><div className="font-medium">{service.category || "未分类"}</div></div>
                  </div>
                  <Separator />
                  <div className="flex items-center gap-3">
                    <Users className="size-5 text-muted-foreground" />
                    <div><div className="text-xs text-muted-foreground">使用人数</div><div className="font-medium">{service.userCount.toLocaleString()}</div></div>
                  </div>
                  <Separator />
                  <div className="flex items-center gap-3">
                    <Activity className="size-5 text-muted-foreground" />
                    <div><div className="text-xs text-muted-foreground">累计调用</div><div className="font-medium">{service.totalCallCount.toLocaleString()}</div></div>
                  </div>
                  <Separator />
                  <div className="flex items-center justify-between gap-3">
                    <span className="text-sm text-muted-foreground">计费</span>
                    <Badge variant={service.isFree ? "secondary" : "outline"}>{service.isFree ? "免费" : `${service.pricing} 积分/次`}</Badge>
                  </div>
                </CardContent>
              </Card>
            </section>
          </aside>
        </div>
      </div>
    </main>
  );
}
