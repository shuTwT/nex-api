"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { api } from "@/lib/api-client";

interface DeleteTokenDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tokenId: string;
  tokenName: string;
  onSuccess: () => void;
}

export function DeleteTokenDialog({
  open,
  onOpenChange,
  tokenId,
  tokenName,
  onSuccess,
}: DeleteTokenDialogProps) {
  const [isLoading, setIsLoading] = useState(false);

  async function handleDelete() {
    setIsLoading(true);
    try {
      const result = await api.delete(`/api/tokens/${tokenId}`);
      if (result.success) {
        onSuccess();
        onOpenChange(false);
      } else {
        alert(result.error || "删除失败");
      }
    } catch (error) {
      alert("删除失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  if (!open) return null;

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-lg max-w-md w-full">
        <div className="p-6">
          <div className="flex items-center gap-3 mb-4">
            <div className="h-10 w-10 rounded-full bg-red-100 flex items-center justify-center">
              <svg className="h-5 w-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <div>
              <h3 className="text-lg font-semibold text-slate-900">删除令牌</h3>
              <p className="text-sm text-slate-500">此操作无法撤销</p>
            </div>
          </div>

          <div className="p-4 bg-red-50 border border-red-200 rounded-lg mb-4">
            <p className="text-sm text-red-700">
              确定要删除令牌 <span className="font-semibold">{tokenName}</span> 吗？
              删除后，使用此令牌的应用将无法访问 API。
            </p>
          </div>

          <div className="flex gap-3">
            <Button
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
              className="flex-1 cursor-pointer"
            >
              取消
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={isLoading}
              className="flex-1 cursor-pointer"
            >
              {isLoading ? "删除中..." : "确认删除"}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
