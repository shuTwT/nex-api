import { useEffect, useState, useCallback, useTransition } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Key,
  Plus,
  Search,
  Copy,
  Trash2,
  Eye,
  EyeOff,
  Shield,
  AlertCircle,
  CheckCircle,
  Edit,
} from "lucide-react";
import { api, responseData } from "@/lib/api";
import { TokenFormDialog } from "@/components/token-form-dialog";
import { DeleteTokenDialog } from "@/components/delete-token-dialog";
import { Pagination } from "@/components/pagination";

interface Token {
  id: string;
  name: string;
  token: string;
  permissions: string;
  lastUsedAt: string | null;
  expiresAt: string | null;
  isActive: boolean;
  createdAt: string;
}

interface TokenStats {
  totalTokens: number;
  activeTokens: number;
  inactiveTokens: number;
  expiredTokens: number;
}

interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export default function TokensPage() {
  const [tokens, setTokens] = useState<Token[]>([]);
  const [stats, setStats] = useState<TokenStats | null>(null);
  const [pagination, setPagination] = useState<PaginationInfo | null>(null);
  const [searchInput, setSearchInput] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [appliedStatus, setAppliedStatus] = useState<string>("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [isLoading, setIsLoading] = useState(true);
  const [showTokenForm, setShowTokenForm] = useState(false);
  const [editingToken, setEditingToken] = useState<Token | null>(null);
  const [deletingToken, setDeletingToken] = useState<Token | null>(null);
  const [showToken, setShowToken] = useState<string | null>(null);
  const [copiedToken, setCopiedToken] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  const loadTokens = useCallback(async () => {
    setIsLoading(true);
    const query: Record<string, string | number | boolean> = {
      search: appliedSearch,
      status: appliedStatus,
      page: currentPage,
      limit: pageSize,
    };
    const result = await api.tokens_route_get(query);

    if (result.success) {
      const data = responseData<Token[]>(result);
      if (data) setTokens(data);
      if (result.pagination) {
        setPagination(result.pagination);
      }
    }
    setIsLoading(false);
  }, [currentPage, pageSize, appliedSearch, appliedStatus]);

  useEffect(() => {
    loadTokens();
  }, [loadTokens]);

  async function loadStats() {
    const result = await api.tokens_stats_route_get();
    const data = responseData<TokenStats>(result);
    if (data) setStats(data);
  }

  useEffect(() => {
    loadStats();
  }, []);

  function handlePageChange(page: number) {
    setCurrentPage(page);
  }

  function handlePageSizeChange(size: number) {
    setPageSize(size);
    setCurrentPage(1);
  }

  function handleQuery() {
    setAppliedSearch(searchInput);
    setAppliedStatus(statusFilter);
    setCurrentPage(1);
  }

  function handleReset() {
    setSearchInput("");
    setStatusFilter("all");
    setAppliedSearch("");
    setAppliedStatus("all");
    setCurrentPage(1);
  }

  function handleAddToken() {
    setEditingToken(null);
    setShowTokenForm(true);
  }

  function handleEditToken(token: Token) {
    setEditingToken(token);
    setShowTokenForm(true);
  }

  function handleDeleteToken(token: Token) {
    setDeletingToken(token);
  }

  function handleFormSuccess() {
    startTransition(() => {
      loadTokens();
      loadStats();
    });
  }

  async function handleToggleStatus(id: string) {
    startTransition(async () => {
      const result = await api.tokens_id_toggle_route_put({ id });
      if (result.success) {
        loadTokens();
        loadStats();
      } else {
        alert(result.error || "切换状态失败");
      }
    });
  }

  const handleCopyToken = (tokenId: string, token: string) => {
    navigator.clipboard.writeText(token);
    setCopiedToken(tokenId);
    setTimeout(() => setCopiedToken(null), 2000);
  };

  const toggleShowToken = (tokenId: string) => {
    setShowToken(showToken === tokenId ? null : tokenId);
  };

  const maskToken = (token: string) => {
    return token.substring(0, 10) + "•".repeat(20) + token.substring(token.length - 10);
  };

  const isTokenExpired = (expiresAt: string | null) => {
    if (!expiresAt) return false;
    return new Date(expiresAt) < new Date();
  };

  const statsCards = [
    {
      title: "活跃令牌",
      value: stats?.activeTokens || 0,
      icon: CheckCircle,
      color: "green",
    },
    {
      title: "已停用",
      value: stats?.inactiveTokens || 0,
      icon: AlertCircle,
      color: "orange",
    },
    {
      title: "已过期",
      value: stats?.expiredTokens || 0,
      icon: AlertCircle,
      color: "red",
    },
    {
      title: "总令牌数",
      value: stats?.totalTokens || 0,
      icon: Key,
      color: "blue",
    },
  ];

  void isPending;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">令牌管理</h1>
          <p className="text-slate-500 mt-1">管理您的 API 访问令牌</p>
        </div>
        <Button className="gap-2 cursor-pointer" onClick={handleAddToken}>
          <Plus className="h-4 w-4" />
          创建令牌
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        {statsCards.map((stat) => {
          const Icon = stat.icon;
          const colorClasses = {
            green: "bg-green-50 text-green-600",
            orange: "bg-orange-50 text-orange-600",
            red: "bg-red-50 text-red-600",
            blue: "bg-blue-50 text-blue-600",
          };

          return (
            <Card key={stat.title} className="hover:shadow-md transition-shadow cursor-pointer">
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-slate-500">{stat.title}</p>
                    <p className="text-2xl font-bold text-slate-900 mt-1">{stat.value}</p>
                  </div>
                  <div className={`h-10 w-10 rounded-lg flex items-center justify-center ${colorClasses[stat.color as keyof typeof colorClasses]}`}>
                    <Icon className="h-5 w-5" />
                  </div>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      <Card className="border-orange-200 bg-orange-50">
        <CardContent className="p-4">
          <div className="flex items-start gap-3">
            <Shield className="h-5 w-5 text-orange-600 flex-shrink-0 mt-0.5" />
            <div>
              <h3 className="text-sm font-medium text-orange-900">安全提示</h3>
              <p className="text-sm text-orange-700 mt-1">
                请妥善保管您的 API 令牌，不要在公开场合分享。令牌创建后只会显示一次，请及时复制保存。
                如果令牌泄露，请立即删除并创建新令牌。
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-4">
          <div className="flex flex-wrap items-end gap-3">
            <div className="relative flex-1 min-w-[200px]">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <Input
                placeholder="搜索令牌名称..."
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                className="pl-10"
              />
            </div>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-[120px]">
                <SelectValue placeholder="状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">全部</SelectItem>
                  <SelectItem value="active">活跃</SelectItem>
                  <SelectItem value="inactive">已停用</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
            <Button
              size="sm"
              onClick={handleQuery}
              className="cursor-pointer"
            >
              查询
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleReset}
              className="cursor-pointer"
            >
              重置
            </Button>
          </div>
        </CardContent>
      </Card>

      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
        </div>
      ) : tokens.length === 0 ? (
        <Card>
          <CardContent className="p-12">
            <div className="text-center">
              <Key className="h-12 w-12 text-slate-300 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-slate-900 mb-2">没有找到令牌</h3>
              <p className="text-slate-500 mb-4">创建您的第一个 API 令牌开始使用</p>
              <Button className="gap-2 cursor-pointer" onClick={handleAddToken}>
                <Plus className="h-4 w-4" />
                创建令牌
              </Button>
            </div>
          </CardContent>
        </Card>
      ) : (
        <>
          <div className="space-y-4">
            {tokens.map((token) => {
              const expired = isTokenExpired(token.expiresAt);

              return (
                <Card key={token.id} className="hover:shadow-md transition-shadow">
                  <CardContent className="p-6">
                    <div className="flex items-start justify-between">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-3 mb-3">
                          <div className="h-10 w-10 rounded-lg bg-blue-50 flex items-center justify-center">
                            <Key className="h-5 w-5 text-blue-600" />
                          </div>
                          <div>
                            <h3 className="text-lg font-semibold text-slate-900">{token.name}</h3>
                            <div className="flex items-center gap-2 mt-1">
                              <Badge
                                variant="outline"
                                className={`cursor-pointer ${
                                  !token.isActive
                                    ? "bg-gray-50 text-gray-700 border-gray-200"
                                    : expired
                                    ? "bg-red-50 text-red-700 border-red-200"
                                    : "bg-green-50 text-green-700 border-green-200"
                                }`}
                                onClick={() => handleToggleStatus(token.id)}
                              >
                                {!token.isActive ? "已停用" : expired ? "已过期" : "活跃"}
                              </Badge>
                              <Badge variant="outline" className="bg-slate-50 text-slate-600 border-slate-200">
                                {token.permissions === "read" ? "只读" :
                                 token.permissions === "read,write" ? "读写" :
                                 token.permissions === "read,write,delete" ? "读写删除" : token.permissions}
                              </Badge>
                            </div>
                          </div>
                        </div>

                        <div className="bg-slate-50 rounded-lg p-3 mb-3">
                          <div className="flex items-center justify-between">
                            <code className="text-sm text-slate-700 font-mono flex-1 overflow-x-auto">
                              {showToken === token.id ? token.token : maskToken(token.token)}
                            </code>
                            <div className="flex items-center gap-2 ml-3">
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => toggleShowToken(token.id)}
                                className="h-8 w-8 p-0 cursor-pointer"
                              >
                                {showToken === token.id ? (
                                  <EyeOff className="h-4 w-4" />
                                ) : (
                                  <Eye className="h-4 w-4" />
                                )}
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleCopyToken(token.id, token.token)}
                                className="h-8 w-8 p-0 cursor-pointer"
                              >
                                {copiedToken === token.id ? (
                                  <CheckCircle className="h-4 w-4 text-green-600" />
                                ) : (
                                  <Copy className="h-4 w-4" />
                                )}
                              </Button>
                            </div>
                          </div>
                        </div>

                        <div className="grid grid-cols-3 gap-4 text-sm">
                          <div>
                            <p className="text-slate-500">创建时间</p>
                            <p className="text-slate-900 font-medium mt-1">
                              {new Date(token.createdAt).toLocaleString("zh-CN")}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-500">最后使用</p>
                            <p className="text-slate-900 font-medium mt-1">
                              {token.lastUsedAt ? new Date(token.lastUsedAt).toLocaleString("zh-CN") : "从未使用"}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-500">过期时间</p>
                            <p className={`font-medium mt-1 ${expired ? "text-red-600" : "text-slate-900"}`}>
                              {token.expiresAt ? new Date(token.expiresAt).toLocaleString("zh-CN") : "永不过期"}
                            </p>
                          </div>
                        </div>
                      </div>

                      <div className="ml-4 flex gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleEditToken(token)}
                          className="cursor-pointer"
                          title="编辑"
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDeleteToken(token)}
                          className="text-red-600 hover:text-red-700 hover:bg-red-50 cursor-pointer"
                          title="删除"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              );
            })}
          </div>

          <div className="mt-4">
            <Pagination
              currentPage={pagination?.page ?? 1}
              totalPages={pagination?.totalPages ?? 1}
              total={pagination?.total ?? 0}
              pageSize={pagination?.limit ?? pageSize}
              onPageChange={handlePageChange}
              onPageSizeChange={handlePageSizeChange}
            />
          </div>
        </>
      )}

      <TokenFormDialog
        open={showTokenForm}
        onOpenChange={setShowTokenForm}
        token={editingToken}
        onSuccess={handleFormSuccess}
      />

      {deletingToken && (
        <DeleteTokenDialog
          open={!!deletingToken}
          onOpenChange={(open) => !open && setDeletingToken(null)}
          tokenId={deletingToken.id}
          tokenName={deletingToken.name}
          onSuccess={handleFormSuccess}
        />
      )}
    </div>
  );
}
