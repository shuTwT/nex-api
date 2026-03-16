"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { createApi, updateApi } from "@/app/actions/apis";
import type { Api } from "@/app/console/api-management/page";

interface ApiFormProps {
  api?: Api;
  categories: { id: string; name: string }[];
  onClose: () => void;
  onSuccess: () => void;
  formId?: string;
}

export function ApiForm({ api, categories, onClose, onSuccess, formId }: ApiFormProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [method, setMethod] = useState(api?.method || "GET");
  const [categoryId, setCategoryId] = useState(api?.category?.id || "");
  const isEdit = !!api;

  async function handleSubmit(formData: FormData) {
    setIsLoading(true);
    setError(null);

    formData.set("method", method);
    formData.set("categoryId", categoryId);

    try {
      const result = isEdit ? await updateApi(formData) : await createApi(formData);

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
      {isEdit && <input type="hidden" name="id" value={api.id} />}

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">
          接口名称 <span className="text-red-500">*</span>
        </label>
        <Input
          name="name"
          placeholder="如：GPT-4 对话 API"
          defaultValue={api?.name || ""}
          required
          disabled={isLoading}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">
          别名 <span className="text-red-500">*</span>
        </label>
        <Input
          name="alias"
          placeholder="如：gpt4Chat（用于 API 路径，只能包含字母和数字，且不能以数字开头）"
          defaultValue={api?.alias || ""}
          required
          disabled={isLoading}
          pattern="^[a-zA-Z][a-zA-Z0-9]*$"
        />
        <p className="text-xs text-slate-500">
          别名用于生成 API 访问路径，如 /api/gpt4Chat
        </p>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">
          接口描述 <span className="text-red-500">*</span>
        </label>
        <Input
          name="description"
          placeholder="如：OpenAI GPT-4 模型对话接口"
          defaultValue={api?.description || ""}
          required
          disabled={isLoading}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">
          上游端点 <span className="text-red-500">*</span>
        </label>
        <Input
          name="endpoint"
          placeholder="如：https://api.openai.com/v1/chat/completions"
          defaultValue={api?.endpoint || ""}
          required
          disabled={isLoading}
        />
        <p className="text-xs text-slate-500">
          上游渠道的实际接口地址
        </p>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium text-slate-700">
            请求方法 <span className="text-red-500">*</span>
          </label>
          <Select value={method} onValueChange={setMethod} disabled={isLoading}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="选择方法" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="GET">GET</SelectItem>
                <SelectItem value="POST">POST</SelectItem>
                <SelectItem value="PUT">PUT</SelectItem>
                <SelectItem value="DELETE">DELETE</SelectItem>
                <SelectItem value="PATCH">PATCH</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium text-slate-700">
            分类 <span className="text-red-500">*</span>
          </label>
          <Select value={categoryId} onValueChange={setCategoryId} disabled={isLoading}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="选择分类" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                {categories.map((cat) => (
                  <SelectItem key={cat.id} value={cat.id}>
                    {cat.name}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">定价</label>
        <Input
          name="pricing"
          placeholder="如：0.02积分/次 或 免费"
          defaultValue={api?.pricing || ""}
          disabled={isLoading}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">文档链接</label>
        <Input
          name="documentation"
          type="url"
          placeholder="https://docs.example.com/api"
          defaultValue={api?.documentation || ""}
          disabled={isLoading}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">预处理脚本</label>
        <textarea
          name="preScript"
          placeholder="JavaScript 代码，用于在请求上游前处理请求参数。可使用变量：request（请求对象）、params（请求参数）"
          defaultValue={api?.preScript || ""}
          disabled={isLoading}
          rows={4}
          className="w-full px-3 py-2 border border-slate-200 rounded-md text-sm font-mono resize-none"
        />
        <p className="text-xs text-slate-500">
          在转发请求到上游前执行的脚本，可用于修改请求参数
        </p>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">后处理脚本</label>
        <textarea
          name="postScript"
          placeholder="JavaScript 代码，用于在收到上游响应后处理响应数据。可使用变量：response（响应对象）、data（响应数据）"
          defaultValue={api?.postScript || ""}
          disabled={isLoading}
          rows={4}
          className="w-full px-3 py-2 border border-slate-200 rounded-md text-sm font-mono resize-none"
        />
        <p className="text-xs text-slate-500">
          在收到上游响应后执行的脚本，可用于转换响应格式
        </p>
      </div>

      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          name="isActive"
          id="isActive"
          value="true"
          defaultChecked={api?.isActive !== false}
          disabled={isLoading}
          className="h-4 w-4 rounded border-slate-300"
        />
        <label htmlFor="isActive" className="text-sm text-slate-700">
          立即启用此接口
        </label>
      </div>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded-md">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}
    </form>
  );
}
