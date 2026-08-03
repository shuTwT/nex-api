import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { ApiForm } from "@/components/api-form";

export interface Api {
  id: string;
  name: string;
  alias: string;
  description: string;
  endpoint: string;
  method: string;
  category: {
    id: string;
    name: string;
  };
  pricing: string;
  documentation: string | null;
  preScript: string | null;
  postScript: string | null;
  isActive: boolean;
  callCount: number;
  createdAt: string;
}

interface Category {
  id: string;
  name: string;
}

interface ApiFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  api?: Api | null;
  categories: Category[];
  onSuccess: () => void;
}

export function ApiFormDialog({
  open,
  onOpenChange,
  api,
  categories,
  onSuccess,
}: ApiFormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[85vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>{api ? "编辑接口" : "添加接口"}</DialogTitle>
        </DialogHeader>
        <div className="flex-1 overflow-y-auto pr-2 -mr-2">
          <ApiForm
            apiItem={api || undefined}
            categories={categories}
            onClose={() => onOpenChange(false)}
            onSuccess={onSuccess}
            formId="api-form"
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
          <Button type="submit" form="api-form" className="cursor-pointer">
            {api ? "保存" : "添加"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
