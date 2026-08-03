import { useState } from "react";
import { Button, Modal, Typography } from "antd";
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
    <Modal
      open={open}
      title="删除广告"
      onCancel={() => onOpenChange(false)}
      footer={[
        <Button key="cancel" onClick={() => onOpenChange(false)} disabled={isLoading}>取消</Button>,
        <Button key="delete" danger type="primary" onClick={handleDelete} loading={isLoading}>删除</Button>,
      ]}
    >
          <Typography.Paragraph>
            确定要删除广告 <span className="font-medium text-slate-900">{advertisementTitle}</span> 吗？此操作无法撤销。
          </Typography.Paragraph>
    </Modal>
  );
}
