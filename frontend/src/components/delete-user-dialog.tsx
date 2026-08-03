import { useState } from "react";
import { Alert, Button, Modal, Typography } from "antd";
import { api } from "@/lib/api";
import { AlertCircle } from "lucide-react";

interface DeleteUserDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  userId: string;
  userName: string;
  onSuccess: () => void;
}

export function DeleteUserDialog({
  open,
  onOpenChange,
  userId,
  userName,
  onSuccess,
}: DeleteUserDialogProps) {
  const [isLoading, setIsLoading] = useState(false);

  async function handleDelete() {
    setIsLoading(true);

    try {
      const result = await api.users_id_route_delete({ id: userId });

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
    <Modal
      open={open}
      title={<span className="flex items-center gap-2 text-red-600">
            <AlertCircle className="h-5 w-5" />
            确认删除
          </span>}
      onCancel={() => onOpenChange(false)}
      footer={[
        <Button key="cancel" onClick={() => onOpenChange(false)} disabled={isLoading}>取消</Button>,
        <Button key="delete" danger type="primary" onClick={handleDelete} loading={isLoading}>确认删除</Button>,
      ]}
    >
          <Typography.Paragraph>
            您确定要删除用户 <span className="font-semibold text-slate-900">{userName}</span> 吗？
          </Typography.Paragraph>

        <Alert type="error" showIcon message="此操作无法撤销" description={
          <>
          <p className="text-sm text-red-700">
            删除用户将同时删除该用户的所有数据，包括：
          </p>
          <ul className="mt-2 text-sm text-red-600 list-disc list-inside">
            <li>订阅记录</li>
            <li>API 使用记录</li>
            <li>账单记录</li>
          </ul>
          </>
        } />
    </Modal>
  );
}
