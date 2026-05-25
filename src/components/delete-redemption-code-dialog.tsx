"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { api } from "@/lib/api-client";
import { toast } from "sonner";

interface DeleteRedemptionCodeDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  codeId?: string;
  codeValue?: string;
  batchId?: string;
  onSuccess: () => void;
}

export function DeleteRedemptionCodeDialog({
  open,
  onOpenChange,
  codeId,
  codeValue,
  batchId,
  onSuccess,
}: DeleteRedemptionCodeDialogProps) {
  const [isDeleting, setIsDeleting] = useState(false);

  const isBatchDelete = !!batchId;

  async function handleDelete() {
    setIsDeleting(true);
    const result = isBatchDelete
      ? await api.delete("/api/redemption-codes/batch", { batchId: batchId! })
      : await api.delete(`/api/redemption-codes/${codeId}`);

    if (result.success) {
      toast.success(result.message || "兑换码已删除");
      onSuccess();
      onOpenChange(false);
    } else {
      toast.error(result.error || "删除兑换码失败");
    }

    setIsDeleting(false);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {isBatchDelete ? "批量删除兑换码" : "删除兑换码"}
          </DialogTitle>
          <DialogDescription>
            {isBatchDelete
              ? "您确定要删除该批次所有未使用的兑换码吗？此操作无法撤销。"
              : `您确定要删除兑换码 "${codeValue}" 吗？此操作无法撤销。`}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isDeleting}
            className="cursor-pointer"
          >
            取消
          </Button>
          <Button
            type="button"
            variant="destructive"
            onClick={handleDelete}
            disabled={isDeleting}
            className="cursor-pointer"
          >
            {isDeleting ? "删除中..." : "删除"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
