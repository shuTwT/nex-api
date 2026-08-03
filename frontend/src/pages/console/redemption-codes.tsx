import { useEffect, useState, useCallback } from "react";
import { Badge, Button, Card, Checkbox, Input, Modal, Select, Typography } from "antd";
import {
  Plus,
  Trash2,
  Ticket,
  Gift,
  Coins,
  Copy,
  Check,
  Download,
  Search,
} from "lucide-react";
import { api, responseData } from "@/lib/api";
import { RedemptionCodeForm } from "@/components/redemption-code-form";
import { DeleteRedemptionCodeDialog } from "@/components/delete-redemption-code-dialog";
import { Pagination } from "@/components/pagination";
import { toast } from "sonner";

interface RedemptionCode {
  id: string;
  code: string;
  type: string;
  planId: string | null;
  planName: string | null;
  credits: number | null;
  expiresAt: Date | null;
  isUsed: boolean;
  usedBy: string | null;
  usedAt: Date | null;
  createdBy: string;
  batchId: string | null;
  createdAt: Date;
  updatedAt: Date;
}

interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

const typeLabels: Record<string, string> = {
  subscription: "订阅",
  quota: "额度",
};

const typeBadgeColors: Record<string, string> = {
  subscription: "bg-blue-50 text-blue-700 border-blue-200",
  quota: "bg-purple-50 text-purple-700 border-purple-200",
};

export default function RedemptionCodesPage() {
  const [codes, setCodes] = useState<RedemptionCode[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [pagination, setPagination] = useState<PaginationInfo | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [searchInput, setSearchInput] = useState("");
  const [typeFilter, setTypeFilter] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [appliedType, setAppliedType] = useState<string>("all");
  const [appliedStatus, setAppliedStatus] = useState<string>("all");
  const [showForm, setShowForm] = useState(false);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [deletingCode, setDeletingCode] = useState<{
    id: string;
    code: string;
  } | null>(null);
  const [copiedCode, setCopiedCode] = useState<string | null>(null);

  const loadCodes = useCallback(async () => {
    setIsLoading(true);
    const query: Record<string, string | number | boolean> = {
      page: currentPage,
      limit: pageSize,
      search: appliedSearch,
    };
    if (appliedType !== "all") query.type = appliedType;
    if (appliedStatus !== "all") query.isUsed = appliedStatus;
    const result = await api.redemption_codes_route_get(query);
    if (result.success) {
      const data = responseData<RedemptionCode[]>(result);
      if (data) setCodes(data);
      if (result.pagination) {
        setPagination(result.pagination);
      }
    }
    setIsLoading(false);
  }, [currentPage, pageSize, appliedSearch, appliedType, appliedStatus]);

  useEffect(() => {
    loadCodes();
  }, [loadCodes]);

  function handleCreateSuccess() {
    toast.success("兑换码生成成功");
    setShowForm(false);
    loadCodes();
  }

  function handleDeleteSuccess() {
    setSelectedIds(new Set());
    loadCodes();
  }

  function handlePageChange(page: number) {
    setCurrentPage(page);
    setSelectedIds(new Set());
  }

  function handlePageSizeChange(size: number) {
    setPageSize(size);
    setCurrentPage(1);
    setSelectedIds(new Set());
  }

  function handleQuery() {
    setAppliedSearch(searchInput);
    setAppliedType(typeFilter);
    setAppliedStatus(statusFilter);
    setCurrentPage(1);
    setSelectedIds(new Set());
  }

  function handleReset() {
    setSearchInput("");
    setTypeFilter("all");
    setStatusFilter("all");
    setAppliedSearch("");
    setAppliedType("all");
    setAppliedStatus("all");
    setCurrentPage(1);
    setSelectedIds(new Set());
  }

  function toggleSelectAll() {
    if (selectedIds.size === codes.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(codes.map((c) => c.id)));
    }
  }

  function toggleSelect(id: string) {
    const next = new Set(selectedIds);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    setSelectedIds(next);
  }

  async function handleBatchDelete() {
    if (selectedIds.size === 0) {
      toast.error("请先选择兑换码");
      return;
    }
    const result = await api.redemption_codes_batch_route_post({ ids: Array.from(selectedIds) });
    if (result.success) {
      toast.success("删除成功");
      handleDeleteSuccess();
    } else {
      toast.error(result.error || "删除失败");
    }
  }

  async function handleExport() {
    const ids = selectedIds.size > 0 ? Array.from(selectedIds) : undefined;
    const query = ids ? { ids: ids.join(",") } : undefined;
    const result = await api.redemption_codes_export_route_get(query);
    if (result.success) {
      const data = responseData<string>(result);
      if (data !== null) {
        const blob = new Blob([data], {
          type: "text/csv;charset=utf-8;",
        });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `兑换码_${new Date().toISOString().slice(0, 10)}.csv`;
        a.click();
        URL.revokeObjectURL(url);
        toast.success("导出成功");
      } else {
        toast.error("导出失败");
      }
    } else {
      toast.error(result.error || "导出失败");
    }
  }

  function handleCopyCode(code: string) {
    navigator.clipboard.writeText(code).then(() => {
      setCopiedCode(code);
      toast.success("兑换码已复制");
      setTimeout(() => setCopiedCode(null), 2000);
    });
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">兑换码管理</h1>
          <p className="text-slate-500 mt-1">生成和管理兑换码</p>
        </div>
        <Button
          className="gap-2 cursor-pointer"
          onClick={() => setShowForm(true)}
        >
          <Plus className="h-4 w-4" />
          生成兑换码
        </Button>
      </div>

      <Card>
        <div className="p-4">
          <div className="flex flex-wrap items-end gap-3">
            <div className="relative flex-1 min-w-[200px]">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <Input
                placeholder="搜索兑换码..."
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                className="pl-10"
              />
            </div>
            <Select value={typeFilter} onChange={setTypeFilter} className="w-[120px]" options={[{ value: "all", label: "全部" }, { value: "subscription", label: "订阅" }, { value: "quota", label: "额度" }]} />
            <Select value={statusFilter} onChange={setStatusFilter} className="w-[130px]" options={[{ value: "all", label: "全部" }, { value: "false", label: "未使用" }, { value: "true", label: "已使用" }]} />
            <Button
              size="small"
              onClick={handleQuery}
              className="cursor-pointer"
            >
              查询
            </Button>
            <Button
              type="default"
              size="small"
              onClick={handleReset}
              className="cursor-pointer"
            >
              重置
            </Button>
          </div>
        </div>
      </Card>

      <Card>
        <div className="p-4">
          <div className="flex items-center justify-between">
            <Typography.Title level={5}>兑换码列表</Typography.Title>
            <div className="flex items-center gap-2">
              {selectedIds.size > 0 && (
                <Button
                  danger
                  size="small"
                  onClick={handleBatchDelete}
                  className="cursor-pointer gap-1"
                >
                  <Trash2 className="h-4 w-4" />
                  删除选中 ({selectedIds.size})
                </Button>
              )}
              <Button
                type="default"
                size="small"
                onClick={handleExport}
                className="cursor-pointer gap-1"
              >
                <Download className="h-4 w-4" />
                导出
              </Button>
            </div>
          </div>
        </div>
        <div className="px-4 pb-4">
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
            </div>
          ) : codes.length === 0 ? (
            <div className="text-center py-12">
              <Ticket className="h-12 w-12 text-slate-300 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-slate-900 mb-2">
                暂无兑换码
              </h3>
              <p className="text-slate-500 mb-4">生成第一批兑换码开始使用</p>
              <Button
                onClick={() => setShowForm(true)}
                className="gap-2 cursor-pointer"
              >
                <Plus className="h-4 w-4" />
                生成兑换码
              </Button>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-slate-200">
                      <th className="py-3 px-2 w-10">
                        <Checkbox
                          checked={
                            codes.length > 0 &&
                            selectedIds.size === codes.length
                          }
                          onChange={toggleSelectAll}
                        />
                      </th>
                      <th className="text-left py-3 px-3 text-sm font-medium text-slate-600">
                        兑换码
                      </th>
                      <th className="text-left py-3 px-3 text-sm font-medium text-slate-600">
                        类型
                      </th>
                      <th className="text-left py-3 px-3 text-sm font-medium text-slate-600">
                        内容
                      </th>
                      <th className="text-left py-3 px-3 text-sm font-medium text-slate-600">
                        过期时间
                      </th>
                      <th className="text-left py-3 px-3 text-sm font-medium text-slate-600">
                        状态
                      </th>
                      <th className="text-left py-3 px-3 text-sm font-medium text-slate-600">
                        创建时间
                      </th>
                      <th className="text-left py-3 px-3 text-sm font-medium text-slate-600">
                        操作
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {codes.map((code) => {
                      const isExpired =
                        code.expiresAt &&
                        new Date(code.expiresAt) < new Date();
                      const isChecked = selectedIds.has(code.id);
                      const isCopied = copiedCode === code.code;

                      return (
                        <tr
                          key={code.id}
                          className={`border-b border-slate-100 hover:bg-slate-50 transition-colors ${
                            code.isUsed || isExpired
                              ? "opacity-60"
                              : ""
                          }`}
                        >
                          <td className="py-3 px-2">
                            <Checkbox
                              checked={isChecked}
                              onChange={() => toggleSelect(code.id)}
                            />
                          </td>
                          <td className="py-3 px-3">
                            <div className="flex items-center gap-2">
                              <span
                                className={`font-mono text-sm ${
                                  code.isUsed
                                    ? "text-slate-400 line-through"
                                    : "text-slate-900"
                                }`}
                              >
                                {code.code}
                              </span>
                              {!code.isUsed && (
                                <Button
                                  type="text"
                                  size="small"
                                  className="h-6 w-6 cursor-pointer"
                                  onClick={() => handleCopyCode(code.code)}
                                >
                                  {isCopied ? (
                                    <Check className="h-3 w-3 text-green-500" />
                                  ) : (
                                    <Copy className="h-3 w-3" />
                                  )}
                                </Button>
                              )}
                            </div>
                          </td>
                          <td className="py-3 px-3">
                            <Badge
                              className={typeBadgeColors[code.type]}
                            >
                              {code.type === "subscription" ? (
                                <Gift className="h-3 w-3 mr-1" />
                              ) : (
                                <Coins className="h-3 w-3 mr-1" />
                              )}
                              {typeLabels[code.type]}
                            </Badge>
                          </td>
                          <td className="py-3 px-3 text-sm text-slate-600">
                            {code.type === "subscription"
                              ? code.planName || "-"
                              : code.credits != null
                                ? `${code.credits.toLocaleString()} 额度`
                                : "-"}
                          </td>
                          <td className="py-3 px-3 text-sm">
                            {code.expiresAt ? (
                              <span
                                className={
                                  isExpired
                                    ? "text-red-500"
                                    : "text-slate-600"
                                }
                              >
                                {new Date(
                                  code.expiresAt,
                                ).toLocaleString("zh-CN")}
                                {isExpired && " (已过期)"}
                              </span>
                            ) : (
                              <span className="text-slate-400">永久</span>
                            )}
                          </td>
                          <td className="py-3 px-3">
                            {code.isUsed ? (
                              <Badge
                                className="bg-green-50 text-green-700 border-green-200"
                              >
                                已使用
                              </Badge>
                            ) : isExpired ? (
                              <Badge
                                className="bg-red-50 text-red-700 border-red-200"
                              >
                                已过期
                              </Badge>
                            ) : (
                              <Badge
                                className="bg-slate-50 text-slate-700 border-slate-200"
                              >
                                未使用
                              </Badge>
                            )}
                          </td>
                          <td className="py-3 px-3 text-sm text-slate-500">
                            {new Date(code.createdAt).toLocaleString(
                              "zh-CN",
                            )}
                          </td>
                          <td className="py-3 px-3">
                            {!code.isUsed && (
                              <Button
                                type="text"
                                size="small"
                                className="h-8 w-8 cursor-pointer text-red-500 hover:text-red-600 hover:bg-red-50"
                                onClick={() =>
                                  setDeletingCode({
                                    id: code.id,
                                    code: code.code,
                                  })
                                }
                              >
                                <Trash2 className="h-4 w-4" />
                              </Button>
                            )}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>

              <Pagination
                currentPage={pagination?.page ?? 1}
                totalPages={pagination?.totalPages ?? 1}
                total={pagination?.total ?? 0}
                pageSize={pagination?.limit ?? pageSize}
                onPageChange={handlePageChange}
                onPageSizeChange={handlePageSizeChange}
              />
            </>
          )}
        </div>
      </Card>

      <Modal open={showForm} title="生成兑换码" onCancel={() => setShowForm(false)} destroyOnHidden footer={[<Button key="cancel" onClick={() => setShowForm(false)}>取消</Button>, <Button key="submit" type="primary" htmlType="submit" form="redemption-code-form">生成</Button>]}>
          <RedemptionCodeForm
            onSuccess={handleCreateSuccess}
            formId="redemption-code-form"
          />
      </Modal>

      {deletingCode && (
        <DeleteRedemptionCodeDialog
          open={!!deletingCode}
          onOpenChange={(open) => {
            if (!open) setDeletingCode(null);
          }}
          codeId={deletingCode.id}
          codeValue={deletingCode.code}
          onSuccess={handleDeleteSuccess}
        />
      )}
    </div>
  );
}
