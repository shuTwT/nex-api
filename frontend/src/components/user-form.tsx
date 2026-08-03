import { useState, type FormEvent } from "react";
import { Alert, Input, InputNumber, Select } from "antd";
import { api } from "@/lib/api";

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

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    const formData = new FormData(e.currentTarget);
    formData.set("role", role);

    const body: Record<string, string> = {};
    formData.forEach((value, key) => {
      body[key] = String(value);
    });
    if (!isEdit) {
      body.credits = String(parseInt(body.credits, 10) || 1000);
    }

    try {
      const result = isEdit
        ? await api.users_id_route_put({ id: body.id ?? user?.id ?? "" }, body)
        : await api.users_route_post(body);

      if (result.success) {
        onSuccess();
        onClose();
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
    <form id={formId} onSubmit={handleSubmit} className="space-y-4">
      {isEdit && (
        <input type="hidden" name="id" value={user.id} />
      )}

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700" htmlFor="user-email">
          邮箱 <span className="text-red-500">*</span>
        </label>
        <Input
          id="user-email"
          name="email"
          type="email"
          placeholder="user@example.com"
          defaultValue={user?.email}
          required
          disabled={isLoading}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700" htmlFor="user-username">
          用户名 <span className="text-red-500">*</span>
        </label>
        <Input
          id="user-username"
          name="username"
          placeholder="username"
          defaultValue={user?.username}
          required
          disabled={isLoading}
        />
      </div>

      {!isEdit && (
        <div className="space-y-2">
          <label className="text-sm font-medium text-slate-700" htmlFor="user-password">
            密码 <span className="text-red-500">*</span>
          </label>
          <Input
            id="user-password"
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
        <Select
          value={role}
          onChange={setRole}
          disabled={isLoading}
          className="w-full"
          options={[
            { value: "user", label: "普通用户" },
            { value: "admin", label: "管理员" },
          ]}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-slate-700" htmlFor="user-credits">
          积分
        </label>
        <InputNumber
          id="user-credits"
          name="credits"
          placeholder="1000"
          defaultValue={user?.credits || 1000}
          disabled={isLoading}
          className="w-full"
        />
      </div>

      {error && (
        <Alert type="error" message={error} showIcon />
      )}
    </form>
  );
}
