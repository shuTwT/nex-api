"use client";

import { useEffect, useState, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Plus,
  Trash2,
  Ticket,
  Gift,
  Coins,
  Copy,
  Check,
  Download,
} from "lucide-react";
import { api } from "@/lib/api-client";
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
  const [pageSize, setPageSize] = useState(20);
  const [showForm, setShowForm] = useState(false);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [deletingCode, setDeletingCode] = useState<{
    id: string;
    code: string;
  } | null>(null);
  const [copiedCode, setCopiedCode] = useState<string | null>(null);

  const loadCodes = useCallback(async () => {
    setIsLoading(true);
    const result = await api.paginated("/api/redemption-codes", {
      page: currentPage,
      limit: pageSize,
    });
    if (result.success && result.data) {
      setCodes(result.data);
      if (result.pagination) {
        setPagination(result.pagination);
      }
    }
    setIsLoading(false);
  }, [currentPage, pageSize]);

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
    const result = await api.post("/api/redemption-codes/batch", { ids: Array.from(selectedIds) });
    if (result.success) {
      toast.success(result.message || "删除成功");
      handleDeleteSuccess();
    } else {
      toast.error(result.error || "删除失败");
    }
  }

  async function handleExport() {
    const ids =
      selectedIds.size > 0 ? Array.from(selectedIds) : undefined;
    const result = await api.get("/api/redemption-codes/export", ids ? { ids: ids.join(",") } : undefined);
    if (result.success && result.data) {
      const blob = new Blob([result.data], {
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
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg">兑换码列表</CardTitle>
            <div className="flex items-center gap-2">
              {selectedIds.size > 0 && (
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={handleBatchDelete}
                  className="cursor-pointer gap-1"
                >
                  <Trash2 className="h-4 w-4" />
                  删除选中 ({selectedIds.size})
                </Button>
              )}
              <Button
                variant="outline"
                size="sm"
                onClick={handleExport}
                className="cursor-pointer gap-1"
              >
                <Download className="h-4 w-4" />
                导出
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
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
                        <input
                          type="checkbox"
                          checked={
                            codes.length > 0 &&
                            selectedIds.size === codes.length
                          }
                          onChange={toggleSelectAll}
                          className="h-4 w-4 rounded border-slate-300 cursor-pointer"
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
                            <input
                              type="checkbox"
                              checked={isChecked}
                              onChange={() => toggleSelect(code.id)}
                              className="h-4 w-4 rounded border-slate-300 cursor-pointer"
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
                                  variant="ghost"
                                  size="icon"
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
                              variant="outline"
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
                                  code.expiresAt
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
                                variant="outline"
                                className="bg-green-50 text-green-700 border-green-200"
                              >
                                已使用
                              </Badge>
                            ) : isExpired ? (
                              <Badge
                                variant="outline"
                                className="bg-red-50 text-red-700 border-red-200"
                              >
                                已过期
                              </Badge>
                            ) : (
                              <Badge
                                variant="outline"
                                className="bg-slate-50 text-slate-700 border-slate-200"
                              >
                                未使用
                              </Badge>
                            )}
                          </td>
                          <td className="py-3 px-3 text-sm text-slate-500">
                            {new Date(code.createdAt).toLocaleString(
                              "zh-CN"
                            )}
                          </td>
                          <td className="py-3 px-3">
                            {!code.isUsed && (
                              <Button
                                variant="ghost"
                                size="icon"
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

              {pagination && (
                <Pagination
                  currentPage={pagination.page}
                  totalPages={pagination.totalPages}
                  total={pagination.total}
                  pageSize={pagination.limit}
                  onPageChange={handlePageChange}
                  onPageSizeChange={handlePageSizeChange}
                />
              )}
            </>
          )}
        </CardContent>
      </Card>

      <Dialog open={showForm} onOpenChange={setShowForm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>生成兑换码</DialogTitle>
          </DialogHeader>
          <RedemptionCodeForm
            onSuccess={handleCreateSuccess}
            formId="redemption-code-form"
          />
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setShowForm(false)}
              className="cursor-pointer"
            >
              取消
            </Button>
            <Button
              type="submit"
              form="redemption-code-form"
              className="cursor-pointer"
            >
              生成
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

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
