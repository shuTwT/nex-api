import { useState, type FormEvent } from "react";
import { Button, Checkbox, Input, Modal, Select } from "antd";
import { api } from "@/lib/api";

export interface McpServiceData { id: string; name: string; identifier: string; type: string; command: string | null; endpoint: string | null; envVars: string | null; pricing: number; isActive: boolean; callCount: number; createdAt: string; }
interface McpFormDialogProps { open: boolean; onOpenChange: (open: boolean) => void; service?: McpServiceData | null; onSuccess: () => void; }

export function McpFormDialog({ open, onOpenChange, service, onSuccess }: McpFormDialogProps) {
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

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); setIsLoading(true); setError(null);
    const body: Record<string, string | number | boolean | null> = { name, identifier, type, pricing: parseInt(pricing, 10) || 0, isActive: isActiveVal, envVars: envVars || null };
    if (type === "stdio") body.command = command; else body.endpoint = endpoint;
    try {
      const result = isEdit ? await api.mcp_services_id_route_put({ id: service!.id }, body) : await api.mcp_services_route_post(body);
      if (result.success) { onSuccess(); onOpenChange(false); } else setError(result.error || "操作失败");
    } catch { setError("操作失败，请重试"); } finally { setIsLoading(false); }
  }

  return <Modal open={open} title={isEdit ? "编辑 MCP 服务" : "添加 MCP 服务"} onCancel={() => onOpenChange(false)} destroyOnHidden footer={[<Button key="cancel" onClick={() => onOpenChange(false)}>取消</Button>, <Button key="submit" type="primary" htmlType="submit" form="mcp-form" loading={isLoading}>{isEdit ? "保存" : "添加"}</Button>]}>
    <form id="mcp-form" onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2"><label className="text-sm font-medium text-slate-700">名称 <span className="text-red-500">*</span></label><Input value={name} onChange={e => setName(e.target.value)} placeholder="如：Filesystem MCP" required disabled={isLoading} /></div>
      <div className="space-y-2"><label className="text-sm font-medium text-slate-700">标识 <span className="text-red-500">*</span></label><Input value={identifier} onChange={e => setIdentifier(e.target.value)} placeholder="如：filesystem（用于路径 /api/v1/mcp/filesystem）" required disabled={isLoading} pattern="^[a-zA-Z][a-zA-Z0-9-]*$" /><p className="text-xs text-slate-500">以字母开头，只能包含字母、数字和连字符</p></div>
      <div className="space-y-2"><label className="text-sm font-medium text-slate-700">类型 <span className="text-red-500">*</span></label><Select value={type} onChange={setType} disabled={isLoading} options={[{ value: "stdio", label: "stdio（子进程）" }, { value: "sse", label: "SSE（Server-Sent Events）" }, { value: "streamableHttp", label: "streamable HTTP" }]} /></div>
      {type === "stdio" ? <div className="space-y-2"><label className="text-sm font-medium text-slate-700">启动命令 <span className="text-red-500">*</span></label><Input value={command} onChange={e => setCommand(e.target.value)} placeholder="如：npx @modelcontextprotocol/server-filesystem /tmp" required disabled={isLoading} /></div> : <div className="space-y-2"><label className="text-sm font-medium text-slate-700">端点 URL <span className="text-red-500">*</span></label><Input value={endpoint} onChange={e => setEndpoint(e.target.value)} placeholder="如：https://mcp.example.com/sse" required disabled={isLoading} /></div>}
      <div className="space-y-2"><label className="text-sm font-medium text-slate-700">环境变量</label><Input.TextArea value={envVars} onChange={e => setEnvVars(e.target.value)} placeholder='{"KEY": "value", "API_KEY": "sk-xxx"}' rows={4} disabled={isLoading} className="font-mono text-sm" /><p className="text-xs text-slate-500">JSON 格式，可选。stdio 类型会注入到子进程环境变量，sse/streamableHttp 会作为请求头发送</p></div>
      <div className="space-y-2"><label className="text-sm font-medium text-slate-700">定价（积分/次）</label><Input value={pricing} onChange={e => setPricing(e.target.value)} type="number" min="0" step="1" placeholder="0" disabled={isLoading} /></div>
      <Checkbox checked={isActiveVal} onChange={e => setIsActiveVal(e.target.checked)} disabled={isLoading}>立即启用此服务</Checkbox>
      {error && <p className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{error}</p>}
    </form>
  </Modal>;
}
