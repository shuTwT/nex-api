"use client";

import { useState, useRef } from "react";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { MonacoEditor } from "@/components/monaco-editor";
import { api as apiClient } from "@/lib/api-client";
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
  const [name, setName] = useState(api?.name || "");
  const [alias, setAlias] = useState(api?.alias || "");
  const [description, setDescription] = useState(api?.description || "");
  const [endpoint, setEndpoint] = useState(api?.endpoint || "");
  const [method, setMethod] = useState(api?.method || "GET");
  const [categoryId, setCategoryId] = useState(api?.category?.id || "");
  const [pricing, setPricing] = useState(api?.pricing || "0");
  const [documentation, setDocumentation] = useState(api?.documentation || "");
  const [isActive, setIsActive] = useState(api?.isActive !== false);
  const [preScript, setPreScript] = useState(api?.preScript || "");
  const [postScript, setPostScript] = useState(api?.postScript || "");
  const isEdit = !!api;

  const formRef = useRef<HTMLFormElement>(null);
  const preScriptPlaceholder = `function preScript(headers, query, body) {
  // 示例：
  // headers['X-Custom-Header'] = 'value';
  // query['param'] = 'value';
  // body = { ...body, extra: 'data' };
}
  `
  const postScriptPlaceholder = `function postScript(response, responseHeaders) {
  return response.data;
}
  `

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    const body = {
      name, alias, description, endpoint, method, categoryId, pricing: parseInt(pricing) || 0,
      documentation, isActive, preScript, postScript,
    };

    try {
      const result = isEdit
        ? await apiClient.put(`/api/apis/${api!.id}`, body)
        : await apiClient.post("/api/apis", body);

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
    <form id={formId} ref={formRef} onSubmit={handleSubmit} className="space-y-4">
      {isEdit && <input type="hidden" name="id" value={api.id} />}

      <Tabs defaultValue="basic" className="w-full">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="basic">基本信息</TabsTrigger>
          <TabsTrigger value="pre-script">预处理脚本</TabsTrigger>
          <TabsTrigger value="post-script">后处理脚本</TabsTrigger>
        </TabsList>

        <TabsContent value="basic" className="space-y-4 mt-4">
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-700">
              接口名称 <span className="text-red-500">*</span>
            </label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="如：GPT-4 对话 API"
              required
              disabled={isLoading}
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-700">
              别名 <span className="text-red-500">*</span>
            </label>
            <Input
              value={alias}
              onChange={(e) => setAlias(e.target.value)}
              placeholder="如：gpt4Chat（用于 API 路径，只能包含字母和数字，且不能以数字开头）"
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
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="如：OpenAI GPT-4 模型对话接口"
              required
              disabled={isLoading}
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-700">
              上游端点 <span className="text-red-500">*</span>
            </label>
            <Input
              value={endpoint}
              onChange={(e) => setEndpoint(e.target.value)}
              placeholder="如：https://api.openai.com/v1/chat/completions"
              required
              disabled={isLoading}
            />
            <p className="text-xs text-slate-500">
              上游渠道的实际接口地址
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
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
            <label className="text-sm font-medium text-slate-700">定价（积分/次）</label>
            <Input
              value={pricing}
              onChange={(e) => setPricing(e.target.value)}
              type="number"
              min="0"
              step="1"
              placeholder="如：10（表示每次调用消耗10积分）"
              disabled={isLoading}
            />
            <p className="text-xs text-slate-500">
              每次调用此接口消耗的积分数量，0 表示免费
            </p>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-700">文档链接</label>
            <Input
              value={documentation}
              onChange={(e) => setDocumentation(e.target.value)}
              type="url"
              placeholder="https://docs.example.com/api"
              disabled={isLoading}
            />
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              name="isActive"
              id="isActive"
              checked={isActive}
              onChange={(e) => setIsActive(e.target.checked)}
              disabled={isLoading}
              className="h-4 w-4 rounded border"
            />
            <label htmlFor="isActive" className="text-sm text-slate-700">
              立即启用此接口
            </label>
          </div>
        </TabsContent>

        <TabsContent value="pre-script" className="space-y-4 mt-4">
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-700">预处理脚本</label>
            <p className="text-xs text-slate-500 mb-2">
              在转发请求到上游前执行的脚本，可用于修改请求头、查询参数和请求体
            </p>
            <p className="text-xs text-slate-500 mb-2">
              可用变量：<code className="bg-slate-100 px-1 py-0.5 rounded">headers</code>（请求头）、<code className="bg-slate-100 px-1 py-0.5 rounded">query</code>（查询参数）、<code className="bg-slate-100 px-1 py-0.5 rounded">body</code>（请求体）
            </p>
            <MonacoEditor
              value={preScript}
              onChange={setPreScript}
              language="javascript"
              height="400px"
              placeholder={preScriptPlaceholder}
              disabled={isLoading}
            />
          </div>
        </TabsContent>

        <TabsContent value="post-script" className="space-y-4 mt-4">
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-700">后处理脚本</label>
            <p className="text-xs text-slate-500 mb-2">
              在收到上游响应后执行的脚本，可用于转换响应格式或过滤敏感数据
            </p>
            <p className="text-xs text-slate-500 mb-2">
              可用变量：<code className="bg-slate-100 px-1 py-0.5 rounded">responseBody</code>（响应体）、<code className="bg-slate-100 px-1 py-0.5 rounded">responseHeaders</code>（响应头）
            </p>
            <MonacoEditor
              value={postScript}
              onChange={setPostScript}
              language="javascript"
              height="400px"
              placeholder={postScriptPlaceholder}
              disabled={isLoading}
            />
          </div>
        </TabsContent>
      </Tabs>

      {error && (
        <div className="pmt-3 bg-red-50 border border-red-200 rounded-md">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}
    </form>
  );
}
