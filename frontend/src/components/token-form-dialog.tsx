import { Button, Modal } from "antd";
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

export function TokenFormDialog({ open, onOpenChange, token, onSuccess }: TokenFormDialogProps) {
  return (
    <Modal
      open={open}
      title={token ? "编辑令牌" : "创建令牌"}
      onCancel={() => onOpenChange(false)}
      footer={[
        <Button key="cancel" onClick={() => onOpenChange(false)}>取消</Button>,
        <Button key="submit" type="primary" htmlType="submit" form="token-form">
          {token ? "保存" : "创建"}
        </Button>,
      ]}
      destroyOnHidden
    >
      <TokenForm token={token || undefined} onClose={() => onOpenChange(false)} onSuccess={onSuccess} formId="token-form" />
    </Modal>
  );
}
