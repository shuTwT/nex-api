"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { api } from "@/lib/api-client";

interface AuditLog {
  id: string;
  action: string;
  resource: string;
  details: string | null;
  ipAddress: string | null;
  userAgent: string | null;
  level: string;
  status: string;
  metadata: string | null;
}

interface AuditLogFormProps {
  auditLog?: AuditLog;
  onClose: () => void;
  onSuccess: () => void;
  formId?: string;
}

export function AuditLogForm({ auditLog, onClose, onSuccess, formId }: AuditLogFormProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [action, setAction] = useState(auditLog?.action || "");
  const [resource, setResource] = useState(auditLog?.resource || "");
  const [details, setDetails] = useState(auditLog?.details || "");
  const [ipAddress, setIpAddress] = useState(auditLog?.ipAddress || "");
  const [userAgent, setUserAgent] = useState(auditLog?.userAgent || "");
  const [level, setLevel] = useState(auditLog?.level || "info");
  const [status, setStatus] = useState(auditLog?.status || "success");
  const [metadata, setMetadata] = useState(auditLog?.metadata || "");
  const isEdit = !!auditLog;

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    const body = { action, resource, details, ipAddress, userAgent, level, status, metadata };

    try {
      const result = isEdit
        ? await api.put(`/api/audit-logs/${auditLog!.id}`, body)
        : await api.post("/api/audit-logs", body);

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
    <form id={formId} onSubmit={handleSubmit} className="space-y-4">
      {isEdit && <input type="hidden" name="id" value={auditLog.id} />}

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">
          操作 <span className="text-red-500">*</span>
        </label>
        <Input
          name="action"
          placeholder="如：用户登录"
          value={action}
          onChange={(e) => setAction(e.target.value)}
          required
          disabled={isLoading}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">
          资源 <span className="text-red-500">*</span>
        </label>
        <Input
          name="resource"
          placeholder="如：用户系统"
          value={resource}
          onChange={(e) => setResource(e.target.value)}
          required
          disabled={isLoading}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">详情</label>
        <Textarea
          name="details"
          placeholder="操作详情描述"
          value={details}
          onChange={(e) => setDetails(e.target.value)}
          rows={3}
          disabled={isLoading}
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium text-slate-700">IP 地址</label>
          <Input
            name="ipAddress"
            placeholder="如：192.168.1.100"
            value={ipAddress}
            onChange={(e) => setIpAddress(e.target.value)}
            disabled={isLoading}
          />
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium text-slate-700">级别</label>
          <Select value={level} onValueChange={setLevel} disabled={isLoading}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="选择级别" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="info">信息</SelectItem>
                <SelectItem value="warning">警告</SelectItem>
                <SelectItem value="error">错误</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">状态</label>
        <Select value={status} onValueChange={setStatus} disabled={isLoading}>
          <SelectTrigger className="w-full">
            <SelectValue placeholder="选择状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="success">成功</SelectItem>
              <SelectItem value="error">失败</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">User Agent</label>
        <Input
          name="userAgent"
          placeholder="浏览器 User Agent"
          value={userAgent}
          onChange={(e) => setUserAgent(e.target.value)}
          disabled={isLoading}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">元数据（JSON）</label>
        <Textarea
          name="metadata"
          placeholder="额外的元数据，JSON 格式"
          value={metadata}
          onChange={(e) => setMetadata(e.target.value)}
          rows={3}
          disabled={isLoading}
        />
      </div>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded-md">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}
    </form>
  );
}
