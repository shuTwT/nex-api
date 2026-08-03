import { Button, Modal } from "antd";
import { CategoryForm } from "@/components/category-form";

export interface Category { id: string; name: string; description: string; icon: string | null; apiCount: number; }
interface CategoryFormDialogProps { open: boolean; onOpenChange: (open: boolean) => void; category?: Category | null; onSuccess: () => void; }

export function CategoryFormDialog({ open, onOpenChange, category, onSuccess }: CategoryFormDialogProps) {
  return <Modal open={open} title={category ? "编辑分类" : "添加分类"} onCancel={() => onOpenChange(false)} destroyOnHidden footer={[<Button key="cancel" onClick={() => onOpenChange(false)}>取消</Button>, <Button key="submit" type="primary" htmlType="submit" form="category-form">{category ? "保存" : "添加"}</Button>]}>
    <CategoryForm category={category || undefined} onClose={() => onOpenChange(false)} onSuccess={onSuccess} formId="category-form" />
  </Modal>;
}
