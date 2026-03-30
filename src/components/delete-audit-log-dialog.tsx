"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { deleteAuditLog } from "@/app/actions/audit-logs";

interface DeleteAuditLogDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  auditLogId: string;
  auditLogAction: string;
  onSuccess: () => void;
}

export function DeleteAuditLogDialog({
  open,
  onOpenChange,
  auditLogId,
  auditLogAction,
  onSuccess,
}: DeleteAuditLogDialogProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleDelete() {
    setIsLoading(true);
    setError(null);

    try {
      const result = await deleteAuditLog(auditLogId);
      if (result.success) {
        onSuccess();
        onOpenChange(false);
      } else {
        setError(result.error || "删除失败");
      }
    } catch (err) {
      setError("删除失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>确认删除</DialogTitle>
          <DialogDescription>
            确定要删除审计日志 "{auditLogAction}" 吗？此操作无法撤销。
          </DialogDescription>
        </DialogHeader>
        {error && (
          <div className="p-3 bg-red-50 border border-red-200 rounded-md">
            <p className="text-sm text-red-700">{error}</p>
          </div>
        )}
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isLoading}
            className="cursor-pointer"
          >
            取消
          </Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={isLoading}
            className="cursor-pointer"
          >
            {isLoading ? "删除中..." : "删除"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
