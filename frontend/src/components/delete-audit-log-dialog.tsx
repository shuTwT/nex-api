import { useState } from "react";
import { Alert, Button, Modal, Typography } from "antd";
import { api } from "@/lib/api";

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
      const result = await api.audit_logs_id_route_delete({ id: auditLogId });
      if (result.success) {
        onSuccess();
        onOpenChange(false);
      } else {
        setError(result.error || "删除失败");
      }
    } catch {
      setError("删除失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Modal
      open={open}
      title="确认删除"
      onCancel={() => onOpenChange(false)}
      footer={[
        <Button key="cancel" onClick={() => onOpenChange(false)} disabled={isLoading}>取消</Button>,
        <Button key="delete" danger type="primary" onClick={handleDelete} loading={isLoading}>删除</Button>,
      ]}
    >
          <Typography.Paragraph>
            确定要删除审计日志 &ldquo;{auditLogAction}&rdquo; 吗？此操作无法撤销。
          </Typography.Paragraph>
        {error && (
          <Alert type="error" message={error} showIcon />
        )}
    </Modal>
  );
}
