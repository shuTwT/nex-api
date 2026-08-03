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
import { api } from "@/lib/api";

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
      const result = await api.tokens_id_route_delete({ id: tokenId });
      if (result.success) {
        onSuccess();
        onOpenChange(false);
      } else {
        alert(result.error || "删除失败");
      }
    } catch {
      alert("删除失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="text-lg font-semibold text-slate-900">删除令牌</DialogTitle>
          <DialogDescription className="text-sm text-slate-500">此操作无法撤销</DialogDescription>
        </DialogHeader>
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg mb-4">
          <p className="text-sm text-red-700">
            确定要删除令牌 <span className="font-semibold">{tokenName}</span> 吗？
            删除后，使用此令牌的应用将无法访问 API。
          </p>
        </div>
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
            {isLoading ? "删除中..." : "确认删除"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
