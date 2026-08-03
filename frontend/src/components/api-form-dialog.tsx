import { Button, Modal } from "antd";
import { ApiForm } from "@/components/api-form";

export interface Api {
  id: string; name: string; alias: string; description: string; endpoint: string; method: string;
  category: { id: string; name: string }; pricing: string; documentation: string | null;
  preScript: string | null; postScript: string | null; isActive: boolean; callCount: number; createdAt: string;
}
interface Category { id: string; name: string; }
interface ApiFormDialogProps { open: boolean; onOpenChange: (open: boolean) => void; api?: Api | null; categories: Category[]; onSuccess: () => void; }

export function ApiFormDialog({ open, onOpenChange, api, categories, onSuccess }: ApiFormDialogProps) {
  return <Modal open={open} title={api ? "编辑接口" : "添加接口"} width={800} style={{ top: 32 }} onCancel={() => onOpenChange(false)} destroyOnHidden footer={[<Button key="cancel" onClick={() => onOpenChange(false)}>取消</Button>, <Button key="submit" type="primary" htmlType="submit" form="api-form">{api ? "保存" : "添加"}</Button>]}>
    <div className="max-h-[70vh] overflow-y-auto pr-2"><ApiForm apiItem={api || undefined} categories={categories} onClose={() => onOpenChange(false)} onSuccess={onSuccess} formId="api-form" /></div>
  </Modal>;
}
