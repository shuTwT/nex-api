"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { createCategory, updateCategory } from "@/app/actions/categories";

interface CategoryFormProps {
  category?: {
    id: string;
    name: string;
    description: string;
    icon: string | null;
  };
  onClose: () => void;
  onSuccess: () => void;
  formId?: string;
}

export function CategoryForm({ category, onClose, onSuccess, formId }: CategoryFormProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const isEdit = !!category;

  async function handleSubmit(formData: FormData) {
    setIsLoading(true);
    setError(null);

    try {
      const result = isEdit ? await updateCategory(formData) : await createCategory(formData);

      if (result.success) {
        onSuccess();
        onClose();
      } else {
        setError(result.error || "操作失败");
      }
    } catch (err) {
      setError("操作失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <form id={formId} action={handleSubmit} className="space-y-4">
      {isEdit && <input type="hidden" name="id" value={category.id} />}

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">
          分类名称 <span className="text-red-500">*</span>
        </label>
        <Input
          name="name"
          placeholder="如：人工智能"
          defaultValue={category?.name || ""}
          required
          disabled={isLoading}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">分类描述</label>
        <Input
          name="description"
          placeholder="如：AI 相关 API 接口"
          defaultValue={category?.description || ""}
          disabled={isLoading}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">图标类名</label>
        <Input
          name="icon"
          placeholder="如：Zap, Settings, Users"
          defaultValue={category?.icon || ""}
          disabled={isLoading}
        />
        <p className="text-xs text-slate-500">填写 Lucide React 图标组件名称</p>
      </div>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded-md">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}
    </form>
  );
}
