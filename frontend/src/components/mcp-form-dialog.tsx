import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  Checkbox,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
} from "antd";
import { api } from "@/lib/api";

export interface McpServiceData {
  id: string;
  name: string;
  identifier: string;
  categoryId: string;
  category: { id: string; name: string } | null;
  description: string | null;
  documentation: string | null;
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
  categories: { id: string; name: string }[];
  categoriesLoading?: boolean;
  onSuccess: () => void;
}
interface McpFormValues {
  name: string;
  identifier: string;
  categoryId: string;
  description?: string;
  documentation?: string;
  type: "stdio" | "sse" | "streamableHttp";
  command?: string;
  endpoint?: string;
  envVars?: string;
  pricing?: number;
  isActive: boolean;
}

function initialValues(service?: McpServiceData | null): McpFormValues {
  return {
    name: service?.name ?? "",
    identifier: service?.identifier ?? "",
    categoryId: service?.categoryId ?? service?.category?.id ?? "",
    description: service?.description ?? "",
    documentation: service?.documentation ?? "",
    type: (service?.type as McpFormValues["type"]) ?? "stdio",
    command: service?.command ?? "",
    endpoint: service?.endpoint ?? "",
    envVars: service?.envVars ?? "",
    pricing: service?.pricing ?? 0,
    isActive: service?.isActive !== false,
  };
}

export function McpFormDialog({
  open,
  onOpenChange,
  service,
  categories,
  categoriesLoading = false,
  onSuccess,
}: McpFormDialogProps) {
  const [form] = Form.useForm<McpFormValues>();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const type = Form.useWatch("type", form) ?? "stdio";
  const isEdit = !!service;

  useEffect(() => {
    if (open) {
      form.setFieldsValue(initialValues(service));
      setError(null);
    }
  }, [form, open, service]);

  async function handleFinish(values: McpFormValues) {
    setIsLoading(true);
    setError(null);
    const body: Record<string, string | number | boolean | null> = {
      name: values.name,
      identifier: values.identifier,
      categoryId: values.categoryId,
      type: values.type,
      pricing: values.pricing ?? 0,
      isActive: values.isActive,
      description: values.description ?? "",
      documentation: values.documentation ?? "",
      envVars: values.envVars || null,
    };
    if (values.type === "stdio") body.command = values.command ?? "";
    else body.endpoint = values.endpoint ?? "";
    try {
      const result = isEdit
        ? await api.mcp_services_id_route_put({ id: service!.id }, body)
        : await api.mcp_services_route_post(body);
      if (result.success) {
        onSuccess();
        onOpenChange(false);
      } else setError(result.error || "操作失败");
    } catch {
      setError("操作失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Modal
      open={open}
      title={isEdit ? "编辑 MCP 服务" : "添加 MCP 服务"}
      onCancel={() => onOpenChange(false)}
      destroyOnHidden
      footer={[
        <Button key="cancel" onClick={() => onOpenChange(false)}>
          取消
        </Button>,
        <Button
          key="submit"
          type="primary"
          htmlType="submit"
          form="mcp-form"
          loading={isLoading}
        >
          {isEdit ? "保存" : "添加"}
        </Button>,
      ]}
    >
      <Form<McpFormValues>
        id="mcp-form"
        form={form}
        layout="vertical"
        onFinish={handleFinish}
        disabled={isLoading}
        initialValues={initialValues(service)}
      >
        <Form.Item
          name="name"
          label="名称"
          rules={[{ required: true, message: "请输入名称" }]}
        >
          <Input placeholder="如：Filesystem MCP" />
        </Form.Item>
        <Form.Item
          name="identifier"
          label="标识"
          extra="以字母开头，只能包含字母、数字和连字符"
          rules={[
            { required: true, message: "请输入标识" },
            { pattern: /^[a-zA-Z][a-zA-Z0-9-]*$/, message: "格式不正确" },
          ]}
        >
          <Input placeholder="如：filesystem（用于路径 /api/v1/mcp/filesystem）" />
        </Form.Item>
        <Form.Item
          name="categoryId"
          label="分类"
          rules={[{ required: true, message: "请选择分类" }]}
        >
          <Select
            placeholder="选择分类"
            loading={categoriesLoading}
            disabled={categoriesLoading || categories.length === 0}
            showSearch={{ optionFilterProp: "label" }}
            options={categories.map((category) => ({
              value: category.id,
              label: category.name,
            }))}
            notFoundContent={categoriesLoading ? "正在加载分类..." : "暂无分类，请先在 HTTP 接口管理中添加"}
          />
        </Form.Item>
        <Form.Item name="description" label="描述">
          <Input.TextArea
            placeholder="简要说明该 MCP 服务的能力和适用场景"
            rows={3}
            maxLength={500}
            showCount
          />
        </Form.Item>
        <Form.Item
          name="documentation"
          label="文档链接"
          rules={[{ type: "url", message: "请输入有效的 URL" }]}
        >
          <Input placeholder="如：https://example.com/docs" />
        </Form.Item>
        <Form.Item
          name="type"
          label="类型"
          rules={[{ required: true, message: "请选择类型" }]}
        >
          <Select
            options={[
              { value: "stdio", label: "stdio（子进程）" },
              { value: "sse", label: "SSE（Server-Sent Events）" },
              { value: "streamableHttp", label: "streamable HTTP" },
            ]}
          />
        </Form.Item>
        {type === "stdio" ? (
          <Form.Item
            name="command"
            label="启动命令"
            rules={[{ required: true, message: "请输入启动命令" }]}
          >
            <Input placeholder="如：npx @modelcontextprotocol/server-filesystem /tmp" />
          </Form.Item>
        ) : (
          <Form.Item
            name="endpoint"
            label="端点 URL"
            rules={[
              { required: true, message: "请输入端点 URL" },
              { type: "url", message: "请输入有效的 URL" },
            ]}
          >
            <Input placeholder="如：https://mcp.example.com/sse" />
          </Form.Item>
        )}
        <Form.Item
          name="envVars"
          label="环境变量"
          extra="JSON 格式，可选。stdio 类型会注入到子进程环境变量，sse/streamableHttp 会作为请求头发送"
        >
          <Input.TextArea
            placeholder={'{"KEY": "value", "API_KEY": "sk-xxx"}'}
            rows={4}
            className="font-mono text-sm"
          />
        </Form.Item>
        <Form.Item name="pricing" label="定价（积分/次）">
          <InputNumber min={0} step={1} className="w-full" placeholder="0" />
        </Form.Item>
        <Form.Item name="isActive" valuePropName="checked">
          <Checkbox>立即启用此服务</Checkbox>
        </Form.Item>
        {error && <Alert type="error" message={error} showIcon />}
      </Form>
    </Modal>
  );
}
