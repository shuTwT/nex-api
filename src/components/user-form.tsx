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
import { api } from "@/lib/api-client";

interface User {
  id: string;
  email: string;
  username: string;
  role: string;
  credits: number;
}

interface UserFormProps {
  user?: User;
  onClose: () => void;
  onSuccess: () => void;
  formId?: string;
}

export function UserForm({ user, onClose, onSuccess, formId }: UserFormProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [role, setRole] = useState(user?.role || "user");

  const isEdit = !!user;

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    const formData = new FormData(e.currentTarget);
    formData.set("role", role);

    const body = Object.fromEntries(formData.entries());
    if (!isEdit) {
      body.credits = String(parseInt(body.credits as string) || 1000);
    }

    try {
      const result = isEdit
        ? await api.put(`/api/users/${body.id}`, body)
        : await api.post("/api/users", body);

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
      {isEdit && (
        <input type="hidden" name="id" value={user.id} />
      )}

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">
          邮箱 <span className="text-red-500">*</span>
        </label>
        <Input
          name="email"
          type="email"
          placeholder="user@example.com"
          defaultValue={user?.email}
          required
          disabled={isLoading}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">
          用户名 <span className="text-red-500">*</span>
        </label>
        <Input
          name="username"
          placeholder="username"
          defaultValue={user?.username}
          required
          disabled={isLoading}
        />
      </div>

      {!isEdit && (
        <div className="space-y-2">
          <label className="text-sm font-medium text-slate-700">
            密码 <span className="text-red-500">*</span>
          </label>
          <Input
            name="password"
            type="password"
            placeholder="至少8个字符"
            required
            disabled={isLoading}
          />
        </div>
      )}

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">
          角色 <span className="text-red-500">*</span>
        </label>
        <Select value={role} onValueChange={setRole} disabled={isLoading}>
          <SelectTrigger className="w-full">
            <SelectValue placeholder="选择角色" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="user">普通用户</SelectItem>
              <SelectItem value="admin">管理员</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700">
          积分
        </label>
        <Input
          name="credits"
          type="number"
          placeholder="1000"
          defaultValue={user?.credits || 1000}
          disabled={isLoading}
        />
      </div>

      {error && (
        <div className="p-3 text-sm text-red-600 bg-red-50 border border-red-200 rounded-md">
          {error}
        </div>
      )}
    </form>
  );
}
