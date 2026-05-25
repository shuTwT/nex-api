"use client";

import { useState } from "react";
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
import { api } from "@/lib/api-client";

interface Token {
  id: string;
  name: string;
  permissions: string;
  expiresAt: string | null;
  isActive: boolean;
}

interface TokenFormProps {
  token?: Token;
  onClose: () => void;
  onSuccess: () => void;
  formId?: string;
}

export function TokenForm({ token, onClose, onSuccess, formId }: TokenFormProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdToken, setCreatedToken] = useState<string | null>(null);
  const [permissions, setPermissions] = useState(token?.permissions || "read");
  const isEdit = !!token;

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setIsLoading(true);
    setError(null);
    setCreatedToken(null);

    const formData = new FormData(e.currentTarget);
    formData.set("permissions", permissions);
    const body = Object.fromEntries(formData.entries());

    try {
      const result = isEdit
        ? await api.put(`/api/tokens/${body.id}`, body)
        : await api.post("/api/tokens", body);

      if (result.success) {
        if (!isEdit && result.data) {
          setCreatedToken(result.data.token);
        } else {
          onSuccess();
          onClose();
        }
      } else {
        setError(result.error || "操作失败");
      }
    } catch (err) {
      setError("操作失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  function handleCopyToken() {
    if (createdToken) {
      navigator.clipboard.writeText(createdToken);
    }
  }

  if (createdToken) {
    return (
      <div className="bg-white rounded-lg shadow-lg max-w-md w-full p-6">
        <div className="text-center space-y-4">
          <div className="h-16 w-16 mx-auto rounded-full bg-green-100 flex items-center justify-center">
            <svg className="h-8 w-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h3 className="text-xl font-semibold text-slate-900">令牌创建成功</h3>
          <p className="text-sm text-slate-600">
            请复制您的令牌，此令牌只会显示一次
          </p>
          <div className="bg-slate-50 rounded-lg p-4">
            <code className="text-sm text-slate-900 break-all">{createdToken}</code>
          </div>
          <div className="flex gap-3">
            <Button
              onClick={handleCopyToken}
              className="flex-1 cursor-pointer"
            >
              复制令牌
            </Button>
            <Button
              variant="outline"
              onClick={() => {
                onSuccess();
                onClose();
              }}
              className="flex-1 cursor-pointer"
            >
              完成
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <form id={formId} onSubmit={handleSubmit} className="space-y-4">
      {isEdit && <input type="hidden" name="id" value={token.id} />}

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">
          令牌名称 <span className="text-red-500">*</span>
        </label>
        <Input
          name="name"
          placeholder="如：生产环境 Token"
          defaultValue={token?.name || ""}
          required
          disabled={isLoading}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">权限</label>
        <Select value={permissions} onValueChange={setPermissions} disabled={isLoading}>
          <SelectTrigger className="w-full">
            <SelectValue placeholder="选择权限" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="read">只读</SelectItem>
              <SelectItem value="read,write">读写</SelectItem>
              <SelectItem value="read,write,delete">读写删除</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">过期时间</label>
        <Input
          name="expiresAt"
          type="datetime-local"
          defaultValue={token?.expiresAt || ""}
          disabled={isLoading}
        />
        <p className="text-xs text-slate-500">留空表示永不过期</p>
      </div>

      {isEdit && (
        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            name="isActive"
            id="isActive"
            value="true"
            defaultChecked={token?.isActive !== false}
            disabled={isLoading}
            className="h-4 w-4 rounded border-slate-300"
          />
          <label htmlFor="isActive" className="text-sm text-slate-700">
            启用此令牌
          </label>
        </div>
      )}

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded-md">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}
    </form>
  );
}
