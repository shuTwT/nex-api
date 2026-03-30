"use client";

import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { AuditLogForm } from "@/components/audit-log-form";

export interface AuditLog {
  id: string;
  action: string;
  resource: string;
  details: string | null;
  ipAddress: string | null;
  userAgent: string | null;
  level: string;
  status: string;
  metadata: string | null;
}

interface AuditLogFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  auditLog?: AuditLog | null;
  onSuccess: () => void;
}

export function AuditLogFormDialog({
  open,
  onOpenChange,
  auditLog,
  onSuccess,
}: AuditLogFormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md max-h-[85vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>{auditLog ? "编辑审计日志" : "创建审计日志"}</DialogTitle>
        </DialogHeader>
        <div className="flex-1 overflow-y-auto pr-2 -mr-2">
          <AuditLogForm
            auditLog={auditLog || undefined}
            onClose={() => onOpenChange(false)}
            onSuccess={onSuccess}
            formId="audit-log-form"
          />
        </div>
        <DialogFooter className="border-t pt-4 mt-4">
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            className="cursor-pointer"
          >
            取消
          </Button>
          <Button type="submit" form="audit-log-form" className="cursor-pointer">
            {auditLog ? "保存" : "创建"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
