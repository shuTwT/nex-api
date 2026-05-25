"use client";

import { useState } from "react";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { api as apiClient } from "@/lib/api-client";

export interface McpServiceData {
  id: string;
  name: string;
  identifier: string;
  type: string;
  command: string | null;
  endpoint: string | null;
  envVars: string | null;
  pricing: number;
  isActive: boolean;
  callCount: number;
  createdAt: string;
}

interface McpFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  service?: McpServiceData | null;
  onSuccess: () => void;
}

export function McpFormDialog({
  open,
  onOpenChange,
  service,
  onSuccess,
}: McpFormDialogProps) {
  const isEdit = !!service;
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [name, setName] = useState(service?.name || "");
  const [identifier, setIdentifier] = useState(service?.identifier || "");
  const [type, setType] = useState(service?.type || "stdio");
  const [command, setCommand] = useState(service?.command || "");
  const [endpoint, setEndpoint] = useState(service?.endpoint || "");
  const [envVars, setEnvVars] = useState(service?.envVars || "");
  const [pricing, setPricing] = useState(String(service?.pricing ?? 0));
  const [isActiveVal, setIsActiveVal] = useState(service?.isActive !== false);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    const body: Record<string, unknown> = {
      name,
      identifier,
      type,
      pricing: parseInt(pricing) || 0,
      isActive: isActiveVal,
    };

    if (type === "stdio") {
      body.command = command;
    } else {
      body.endpoint = endpoint;
    }
    body.envVars = envVars || null;

    try {
      const result = isEdit
        ? await apiClient.put(`/api/mcp-services/${service!.id}`, body)
        : await apiClient.post("/api/mcp-services", body);

      if (result.success) {
        onSuccess();
        onOpenChange(false);
      } else {
        setError(result.error || "操作失败");
      }
    } catch {
      setError("操作失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg max-h-[85vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>{isEdit ? "编辑 MCP 服务" : "添加 MCP 服务"}</DialogTitle>
        </DialogHeader>
        <form
          id="mcp-form"
          onSubmit={handleSubmit}
          className="flex-1 overflow-y-auto space-y-4"
        >
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-700">
              名称 <span className="text-red-500">*</span>
            </label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="如：Filesystem MCP"
              required
              disabled={isLoading}
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-700">
              标识 <span className="text-red-500">*</span>
            </label>
            <Input
              value={identifier}
              onChange={(e) => setIdentifier(e.target.value)}
              placeholder="如：filesystem（用于路径 /api/v1/mcp/filesystem）"
              required
              disabled={isLoading}
              pattern="^[a-zA-Z][a-zA-Z0-9-]*$"
            />
            <p className="text-xs text-slate-500">
              以字母开头，只能包含字母、数字和连字符
            </p>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-700">
              类型 <span className="text-red-500">*</span>
            </label>
            <Select value={type} onValueChange={setType} disabled={isLoading}>
              <SelectTrigger className="w-full">
                <SelectValue placeholder="选择类型" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="stdio">stdio（子进程）</SelectItem>
                  <SelectItem value="sse">SSE（Server-Sent Events）</SelectItem>
                  <SelectItem value="streamableHttp">streamable HTTP</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>

          {type === "stdio" && (
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">
                启动命令 <span className="text-red-500">*</span>
              </label>
              <Input
                value={command}
                onChange={(e) => setCommand(e.target.value)}
                placeholder="如：npx @modelcontextprotocol/server-filesystem /tmp"
                required
                disabled={isLoading}
              />
            </div>
          )}

          {type !== "stdio" && (
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">
                端点 URL <span className="text-red-500">*</span>
              </label>
              <Input
                value={endpoint}
                onChange={(e) => setEndpoint(e.target.value)}
                placeholder="如：https://mcp.example.com/sse"
                required
                disabled={isLoading}
              />
            </div>
          )}

          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-700">环境变量</label>
            <Textarea
              value={envVars}
              onChange={(e) => setEnvVars(e.target.value)}
              placeholder='{"KEY": "value", "API_KEY": "sk-xxx"}'
              rows={4}
              disabled={isLoading}
              className="font-mono text-sm"
            />
            <p className="text-xs text-slate-500">
              JSON 格式，可选。stdio 类型会注入到子进程环境变量，sse/streamableHttp 会作为请求头发送
            </p>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-700">定价（积分/次）</label>
            <Input
              value={pricing}
              onChange={(e) => setPricing(e.target.value)}
              type="number"
              min="0"
              step="1"
              placeholder="0"
              disabled={isLoading}
            />
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="mcp-isActive"
              checked={isActiveVal}
              onChange={(e) => setIsActiveVal(e.target.checked)}
              disabled={isLoading}
              className="h-4 w-4 rounded border"
            />
            <label htmlFor="mcp-isActive" className="text-sm text-slate-700">
              立即启用此服务
            </label>
          </div>

          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-700">{error}</p>
            </div>
          )}
        </form>
        <DialogFooter className="border-t pt-4 mt-4">
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            className="cursor-pointer"
          >
            取消
          </Button>
          <Button type="submit" form="mcp-form" disabled={isLoading} className="cursor-pointer">
            {isEdit ? "保存" : "添加"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
