import { useState } from "react";
import { Modal } from "antd";
import { api } from "@/lib/api";
import { toast } from "sonner";

interface DeleteSubscriptionPlanDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  planId: string;
  planTitle: string;
  onSuccess: () => void;
}

export function DeleteSubscriptionPlanDialog({
  open,
  onOpenChange,
  planId,
  planTitle,
  onSuccess,
}: DeleteSubscriptionPlanDialogProps) {
  const [isDeleting, setIsDeleting] = useState(false);

  async function handleDelete() {
    setIsDeleting(true);
    const result = await api.subscription_plans_id_route_delete({ id: planId });

    if (result.success) {
      toast.success("订阅计划已删除");
      onSuccess();
      onOpenChange(false);
    } else {
      toast.error(result.error || "删除订阅计划失败");
    }

    setIsDeleting(false);
  }

  return (
    <Modal open={open} title="删除订阅计划" okText="删除" okButtonProps={{ danger: true, loading: isDeleting }} cancelButtonProps={{ disabled: isDeleting }} onCancel={() => onOpenChange(false)} onOk={handleDelete} destroyOnHidden>
      <p>您确定要删除订阅计划 &quot;{planTitle}&quot; 吗？此操作无法撤销。</p>
    </Modal>
  );
}
