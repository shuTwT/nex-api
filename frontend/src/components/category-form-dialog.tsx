import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { CategoryForm } from "@/components/category-form";

export interface Category {
  id: string;
  name: string;
  description: string;
  icon: string | null;
  apiCount: number;
}

interface CategoryFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  category?: Category | null;
  onSuccess: () => void;
}

export function CategoryFormDialog({
  open,
  onOpenChange,
  category,
  onSuccess,
}: CategoryFormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>{category ? "编辑分类" : "添加分类"}</DialogTitle>
        </DialogHeader>
        <div className="flex-1 overflow-y-auto">
          <CategoryForm
            category={category || undefined}
            onClose={() => onOpenChange(false)}
            onSuccess={onSuccess}
            formId="category-form"
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
          <Button type="submit" form="category-form" className="cursor-pointer">
            {category ? "保存" : "添加"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
