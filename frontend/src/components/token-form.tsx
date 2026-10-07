import { useState } from "react";
import { Alert, Button, Checkbox, Form, Input, Select } from "antd";
import { api, responseData } from "@/lib/api";

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
  onCreated?: () => void;
  formId?: string;
}
interface TokenFormValues {
  name: string;
  permissions: string;
  expiresAt?: string;
  isActive?: boolean;
}

function toDateTimeLocal(value: string | null | undefined) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const pad = (n: number) => String(n).padStart(2, "0");
  // datetime-local 输入框按本地时区解析无时区标记的值，这里必须输出本地时间，
  // 否则编辑回显会按 UTC 显示、原样保存会导致时间偏移。
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

export function TokenForm({
  token,
  onClose,
  onSuccess,
  onCreated,
  formId,
}: TokenFormProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdToken, setCreatedToken] = useState<string | null>(null);
  const isEdit = !!token;

  async function handleFinish(values: TokenFormValues) {
    setIsLoading(true);
    setError(null);
    setCreatedToken(null);
    const body: Record<string, string | boolean | undefined> = {
      name: values.name,
      permissions: values.permissions,
      // 后端 expiresAt 是 *time.Time（RFC3339），空值必须省略字段而非传 ""；
      // datetime-local 的值无时区标记，按本地时区转为 ISO 后再提交。
      expiresAt: values.expiresAt
        ? new Date(values.expiresAt).toISOString()
        : undefined,
    };
    if (isEdit) body.isActive = values.isActive ?? false;
    try {
      const result = isEdit
        ? await api.tokens_id_route_put({ id: token!.id }, body)
        : await api.tokens_route_post(body);
      const data = responseData<{ token: string }>(result);
      if (result.success) {
        if (!isEdit && data) {
          setCreatedToken(data.token);
          onCreated?.();
        } else {
          onSuccess();
          onClose();
        }
      } else setError(result.error || "操作失败");
    } catch {
      setError("操作失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  if (createdToken)
    return (
      <div className="space-y-4 text-center">
        <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
          <svg
            className="h-8 w-8 text-green-600"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M5 13l4 4L19 7"
            />
          </svg>
        </div>
        <h3 className="text-xl font-semibold text-slate-900">令牌创建成功</h3>
        <p className="text-sm text-slate-600">
          请复制您的令牌，此令牌只会显示一次
        </p>
        <div className="rounded-lg bg-slate-50 p-4">
          <code className="break-all text-sm text-slate-900">
            {createdToken}
          </code>
        </div>
        <div className="flex gap-3">
          <Button
            onClick={() => void navigator.clipboard.writeText(createdToken)}
            className="flex-1"
          >
            复制令牌
          </Button>
          <Button
            onClick={() => {
              onSuccess();
              onClose();
            }}
            className="flex-1"
          >
            完成
          </Button>
        </div>
      </div>
    );

  return (
    <Form<TokenFormValues>
      id={formId}
      layout="vertical"
      onFinish={handleFinish}
      disabled={isLoading}
      initialValues={{
        name: token?.name ?? "",
        permissions: token?.permissions ?? "read",
        expiresAt: toDateTimeLocal(token?.expiresAt),
        isActive: token?.isActive !== false,
      }}
    >
      <Form.Item
        name="name"
        label="令牌名称"
        rules={[{ required: true, message: "请输入令牌名称" }]}
      >
        <Input placeholder="如：生产环境 Token" />
      </Form.Item>
      <Form.Item
        name="permissions"
        label="权限"
        rules={[{ required: true, message: "请选择权限" }]}
      >
        <Select
          options={[
            { value: "read", label: "只读" },
            { value: "read,write", label: "读写" },
            { value: "read,write,delete", label: "读写删除" },
          ]}
        />
      </Form.Item>
      <Form.Item name="expiresAt" label="过期时间" extra="留空表示永不过期">
        <Input type="datetime-local" />
      </Form.Item>
      {isEdit && (
        <Form.Item name="isActive" valuePropName="checked">
          <Checkbox>启用此令牌</Checkbox>
        </Form.Item>
      )}
      {error && <Alert type="error" message={error} showIcon />}
    </Form>
  );
}
