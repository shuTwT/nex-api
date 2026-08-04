import { useState } from "react";
import { Alert, Form, Input, InputNumber, Select } from "antd";
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
interface UserFormValues {
  email: string;
  username: string;
  password?: string;
  role: string;
  credits?: number;
}

export function UserForm({ user, onClose, onSuccess, formId }: UserFormProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const isEdit = !!user;

  async function handleFinish(values: UserFormValues) {
    setIsLoading(true);
    setError(null);
    const body: Record<string, string> = {
      email: values.email,
      username: values.username,
      role: values.role,
      credits: String(values.credits ?? 1000),
    };
    if (!isEdit) body.password = values.password ?? "";

    try {
      const result = isEdit
        ? await api.users_id_route_put({ id: user!.id }, body)
        : await api.users_route_post(body);
      if (result.success) {
        onSuccess();
        onClose();
      } else setError(result.error || "操作失败");
    } catch {
      setError("操作失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Form<UserFormValues>
      id={formId}
      layout="vertical"
      onFinish={handleFinish}
      disabled={isLoading}
      initialValues={{
        email: user?.email ?? "",
        username: user?.username ?? "",
        role: user?.role ?? "user",
        credits: user?.credits ?? 1000,
      }}
    >
      <Form.Item
        name="email"
        label="邮箱"
        rules={[
          { required: true, message: "请输入邮箱" },
          { type: "email", message: "请输入有效的邮箱地址" },
        ]}
      >
        <Input placeholder="user@example.com" />
      </Form.Item>
      <Form.Item
        name="username"
        label="用户名"
        rules={[{ required: true, message: "请输入用户名" }]}
      >
        <Input placeholder="username" />
      </Form.Item>
      {!isEdit && (
        <Form.Item
          name="password"
          label="密码"
          rules={[
            { required: true, message: "请输入密码" },
            { min: 8, message: "密码至少 8 个字符" },
          ]}
        >
          <Input.Password placeholder="至少8个字符" />
        </Form.Item>
      )}
      <Form.Item
        name="role"
        label="角色"
        rules={[{ required: true, message: "请选择角色" }]}
      >
        <Select
          options={[
            { value: "user", label: "普通用户" },
            { value: "admin", label: "管理员" },
          ]}
        />
      </Form.Item>
      <Form.Item name="credits" label="积分">
        <InputNumber min={0} className="w-full" placeholder="1000" />
      </Form.Item>
      {error && <Alert type="error" message={error} showIcon />}
    </Form>
  );
}
