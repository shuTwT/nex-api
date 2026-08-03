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
import { api } from "@/lib/api";

interface DeleteAdvertisementDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  advertisementId: string;
  advertisementTitle: string;
  onSuccess: () => void;
}

export function DeleteAdvertisementDialog({
  open,
  onOpenChange,
  advertisementId,
  advertisementTitle,
  onSuccess,
}: DeleteAdvertisementDialogProps) {
  const [isLoading, setIsLoading] = useState(false);

  async function handleDelete() {
    setIsLoading(true);
    try {
      const result = await api.advertisements_id_route_delete({ id: advertisementId });
      if (result.success) {
        onSuccess();
        onOpenChange(false);
      }
    } catch {
      // dialog stays open on failure
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>删除广告</DialogTitle>
          <DialogDescription>
            确定要删除广告 <span className="font-medium text-slate-900">{advertisementTitle}</span> 吗？此操作无法撤销。
          </DialogDescription>
        </DialogHeader>
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
