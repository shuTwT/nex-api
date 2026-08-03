import { useState } from "react";
import { Modal } from "antd";
import { api } from "@/lib/api";
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
      ? await api.redemption_codes_batch_route_delete({ batchId: batchId! })
      : await api.redemption_codes_id_route_delete({ id: codeId ?? "" });

    if (result.success) {
      toast.success("兑换码已删除");
      onSuccess();
      onOpenChange(false);
    } else {
      toast.error(result.error || "删除兑换码失败");
    }

    setIsDeleting(false);
  }

  return (
    <Modal open={open} title={isBatchDelete ? "批量删除兑换码" : "删除兑换码"} okText="删除" okButtonProps={{ danger: true, loading: isDeleting }} cancelButtonProps={{ disabled: isDeleting }} onCancel={() => onOpenChange(false)} onOk={handleDelete} destroyOnHidden>
      <p>{isBatchDelete ? "您确定要删除该批次所有未使用的兑换码吗？此操作无法撤销。" : `您确定要删除兑换码 "${codeValue}" 吗？此操作无法撤销。`}</p>
    </Modal>
  );
}
