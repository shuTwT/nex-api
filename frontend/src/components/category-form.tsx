import { useState } from "react";
import { Alert, Form, Input } from "antd";
import { api } from "@/lib/api";

interface CategoryFormProps {
  category?: {
    id: string;
    name: string;
    description: string;
    icon: string | null;
  };
  onClose: () => void;
  onSuccess: () => void;
  formId?: string;
}
interface CategoryFormValues {
  name: string;
  description?: string;
  icon?: string;
}

export function CategoryForm({
  category,
  onClose,
  onSuccess,
  formId,
}: CategoryFormProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const isEdit = !!category;
  async function handleFinish(values: CategoryFormValues) {
    setIsLoading(true);
    setError(null);
    const body = {
      name: values.name,
      description: values.description ?? "",
      icon: values.icon ?? "",
    };
    try {
      const result = isEdit
        ? await api.categories_id_route_put({ id: category!.id }, body)
        : await api.categories_route_post(body);
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
    <Form<CategoryFormValues>
      id={formId}
      layout="vertical"
      onFinish={handleFinish}
      disabled={isLoading}
      initialValues={{
        name: category?.name ?? "",
        description: category?.description ?? "",
        icon: category?.icon ?? "",
      }}
    >
      <Form.Item
        name="name"
        label="分类名称"
        rules={[{ required: true, message: "请输入分类名称" }]}
      >
        <Input placeholder="如：人工智能" />
      </Form.Item>
      <Form.Item name="description" label="分类描述">
        <Input placeholder="如：AI 相关 API 接口" />
      </Form.Item>
      <Form.Item
        name="icon"
        label="图标类名"
        extra="填写 Lucide React 图标组件名称"
      >
        <Input placeholder="如：Zap, Settings, Users" />
      </Form.Item>
      {error && <Alert type="error" message={error} showIcon />}
    </Form>
  );
}
