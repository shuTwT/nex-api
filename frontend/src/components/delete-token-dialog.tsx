import { useState } from "react";
import { Alert, Modal } from "antd";
import { api } from "@/lib/api";

interface DeleteTokenDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tokenId: string;
  tokenName: string;
  onSuccess: () => void;
}

export function DeleteTokenDialog({ open, onOpenChange, tokenId, tokenName, onSuccess }: DeleteTokenDialogProps) {
  const [isLoading, setIsLoading] = useState(false);
  async function handleDelete() {
    setIsLoading(true);
    try {
      const result = await api.tokens_id_route_delete({ id: tokenId });
      if (result.success) { onSuccess(); onOpenChange(false); } else alert(result.error || "删除失败");
    } catch { alert("删除失败，请重试"); } finally { setIsLoading(false); }
  }
  return (
    <Modal open={open} title="删除令牌" okText="确认删除" okButtonProps={{ danger: true, loading: isLoading }} cancelButtonProps={{ disabled: isLoading }} onCancel={() => onOpenChange(false)} onOk={handleDelete} destroyOnHidden>
      <p className="mb-4 text-slate-500">此操作无法撤销</p>
      <Alert type="error" showIcon message={<>确定要删除令牌 <strong>{tokenName}</strong> 吗？删除后，使用此令牌的应用将无法访问 API。</>} />
    </Modal>
  );
}
