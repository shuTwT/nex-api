"use client";

import { useEffect, useState, useTransition } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
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
  Edit
} from "lucide-react";
import { getTokens, getTokenStats, toggleTokenStatus } from "@/app/actions/tokens";
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
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedStatus, setSelectedStatus] = useState("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [isLoading, setIsLoading] = useState(true);
  const [showTokenForm, setShowTokenForm] = useState(false);
  const [editingToken, setEditingToken] = useState<Token | null>(null);
  const [deletingToken, setDeletingToken] = useState<Token | null>(null);
  const [showToken, setShowToken] = useState<string | null>(null);
  const [copiedToken, setCopiedToken] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  useEffect(() => {
    loadTokens();
    loadStats();
  }, [searchQuery, selectedStatus, currentPage, pageSize]);

  async function loadTokens() {
    setIsLoading(true);
    const result = await getTokens({
      search: searchQuery,
      status: selectedStatus,
      page: currentPage,
      limit: pageSize,
    });
    
    if (result.success && result.data) {
      setTokens(result.data);
      if (result.pagination) {
        setPagination(result.pagination);
      }
    }
    setIsLoading(false);
  }

  async function loadStats() {
    const result = await getTokenStats();
    if (result.success && result.data) {
      setStats(result.data);
    }
  }

  function handleSearchChange(value: string) {
    setSearchQuery(value);
    setCurrentPage(1);
  }

  function handleStatusChange(status: string) {
    setSelectedStatus(status);
    setCurrentPage(1);
  }

  function handlePageChange(page: number) {
    setCurrentPage(page);
  }

  function handlePageSizeChange(size: number) {
    setPageSize(size);
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
      const result = await toggleTokenStatus(id);
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

  return (
    <div className="space-y-6">
      {/* Page Header */}
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

      {/* Stats */}
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

      {/* Security Notice */}
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

      {/* Search & Filter */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <Input
                placeholder="搜索令牌名称..."
                value={searchQuery}
                onChange={(e) => handleSearchChange(e.target.value)}
                className="pl-10"
              />
            </div>
            <div className="flex gap-2">
              <Button
                variant={selectedStatus === "all" ? "default" : "outline"}
                size="sm"
                onClick={() => handleStatusChange("all")}
                className="cursor-pointer"
              >
                全部
              </Button>
              <Button
                variant={selectedStatus === "active" ? "default" : "outline"}
                size="sm"
                onClick={() => handleStatusChange("active")}
                className="cursor-pointer"
              >
                活跃
              </Button>
              <Button
                variant={selectedStatus === "inactive" ? "default" : "outline"}
                size="sm"
                onClick={() => handleStatusChange("inactive")}
                className="cursor-pointer"
              >
                已停用
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Tokens List */}
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

          {/* Pagination */}
          {pagination && (
            <div className="mt-4">
              <Pagination
                currentPage={pagination.page}
                totalPages={pagination.totalPages}
                total={pagination.total}
                pageSize={pagination.limit}
                onPageChange={handlePageChange}
                onPageSizeChange={handlePageSizeChange}
              />
            </div>
          )}
        </>
      )}

      {/* Token Form Dialog */}
      <TokenFormDialog
        open={showTokenForm}
        onOpenChange={setShowTokenForm}
        token={editingToken}
        onSuccess={handleFormSuccess}
      />

      {/* Delete Confirmation Dialog */}
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
