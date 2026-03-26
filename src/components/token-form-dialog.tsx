"use client";

import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { TokenForm } from "@/components/token-form";

export interface Token {
  id: string;
  name: string;
  permissions: string;
  expiresAt: string | null;
  isActive: boolean;
}

interface TokenFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  token?: Token | null;
  onSuccess: () => void;
}

export function TokenFormDialog({
  open,
  onOpenChange,
  token,
  onSuccess,
}: TokenFormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{token ? "编辑令牌" : "创建令牌"}</DialogTitle>
        </DialogHeader>
        <TokenForm
          token={token || undefined}
          onClose={() => onOpenChange(false)}
          onSuccess={onSuccess}
          formId="token-form"
        />
        <DialogFooter className="border-t pt-4 mt-4">
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            className="cursor-pointer"
          >
            取消
          </Button>
          <Button type="submit" form="token-form" className="cursor-pointer">
            {token ? "保存" : "创建"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
