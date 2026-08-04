import { useState } from "react";
import { Alert, Checkbox, Form, Input, InputNumber, Select, Tabs } from "antd";
import { MonacoEditor } from "@/components/monaco-editor";
import { api } from "@/lib/api";

interface Api {
  id: string;
  name: string;
  alias: string;
  description: string;
  endpoint: string;
  method: string;
  category: { id: string; name: string };
  pricing: string;
  documentation: string | null;
  preScript: string | null;
  postScript: string | null;
  isActive: boolean;
  callCount: number;
  createdAt: string;
}

interface ApiFormProps {
  apiItem?: Api;
  categories: { id: string; name: string }[];
  onClose: () => void;
  onSuccess: () => void;
  formId?: string;
}

interface ApiFormValues {
  name: string;
  alias: string;
  description: string;
  endpoint: string;
  method: string;
  categoryId: string;
  pricing?: number;
  documentation?: string;
  isActive: boolean;
  preScript?: string;
  postScript?: string;
}

export function ApiForm({
  apiItem,
  categories,
  onClose,
  onSuccess,
  formId,
}: ApiFormProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const isEdit = !!apiItem;

  const preScriptPlaceholder = `function preScript(headers, query, body) {
  headers['X-Custom-Header'] = 'value';
  query['param'] = 'value';
  body = { ...body, extra: 'data' };
}
  `;
  const postScriptPlaceholder = `function postScript(response, responseHeaders) {
  return response.data;
}
  `;

  async function handleFinish(values: ApiFormValues) {
    setIsLoading(true);
    setError(null);

    const body = {
      ...values,
      pricing: Number(values.pricing ?? 0),
      documentation: values.documentation ?? "",
      preScript: values.preScript ?? "",
      postScript: values.postScript ?? "",
    };

    try {
      const result = isEdit
        ? await api.apis_id_route_put({ id: apiItem!.id }, body)
        : await api.apis_route_post(body);

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
    <Form<ApiFormValues>
      id={formId}
      layout="vertical"
      onFinish={handleFinish}
      disabled={isLoading}
      className="space-y-4"
      initialValues={{
        name: apiItem?.name ?? "",
        alias: apiItem?.alias ?? "",
        description: apiItem?.description ?? "",
        endpoint: apiItem?.endpoint ?? "",
        method: apiItem?.method ?? "GET",
        categoryId: apiItem?.category?.id ?? "",
        pricing: Number(apiItem?.pricing ?? 0),
        documentation: apiItem?.documentation ?? "",
        isActive: apiItem?.isActive !== false,
        preScript: apiItem?.preScript ?? "",
        postScript: apiItem?.postScript ?? "",
      }}
    >
      <Tabs
        defaultActiveKey="basic"
        items={[
          {
            key: "basic",
            label: "基本信息",
            children: (
              <div className="mt-4 space-y-4">
                <Form.Item
                  name="name"
                  label="接口名称"
                  rules={[{ required: true, message: "请输入接口名称" }]}
                >
                  <Input placeholder="如：GPT-4 对话 API" />
                </Form.Item>
                <Form.Item
                  name="alias"
                  label="别名"
                  extra="别名用于生成 API 访问路径，如 /api/gpt4Chat"
                  rules={[
                    { required: true, message: "请输入别名" },
                    {
                      pattern: /^[a-zA-Z][a-zA-Z0-9]*$/,
                      message: "只能包含字母和数字，且不能以数字开头",
                    },
                  ]}
                >
                  <Input placeholder="如：gpt4Chat" />
                </Form.Item>
                <Form.Item
                  name="description"
                  label="接口描述"
                  rules={[{ required: true, message: "请输入接口描述" }]}
                >
                  <Input placeholder="如：OpenAI GPT-4 模型对话接口" />
                </Form.Item>
                <Form.Item
                  name="endpoint"
                  label="上游端点"
                  extra="上游渠道的实际接口地址"
                  rules={[
                    { required: true, message: "请输入上游端点" },
                    { type: "url", message: "请输入有效的 URL" },
                  ]}
                >
                  <Input placeholder="如：https://api.openai.com/v1/chat/completions" />
                </Form.Item>
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                  <Form.Item
                    name="method"
                    label="请求方法"
                    rules={[{ required: true, message: "请选择请求方法" }]}
                  >
                    <Select
                      options={["GET", "POST", "PUT", "DELETE", "PATCH"].map(
                        (value) => ({ value, label: value }),
                      )}
                    />
                  </Form.Item>
                  <Form.Item
                    name="categoryId"
                    label="分类"
                    rules={[{ required: true, message: "请选择分类" }]}
                  >
                    <Select
                      placeholder="选择分类"
                      options={categories.map((category) => ({
                        value: category.id,
                        label: category.name,
                      }))}
                    />
                  </Form.Item>
                </div>
                <Form.Item
                  name="pricing"
                  label="定价（积分/次）"
                  extra="每次调用此接口消耗的积分数量，0 表示免费"
                >
                  <InputNumber
                    min={0}
                    step={1}
                    className="w-full"
                    placeholder="如：10"
                  />
                </Form.Item>
                <Form.Item
                  name="documentation"
                  label="文档链接"
                  rules={[{ type: "url", message: "请输入有效的 URL" }]}
                >
                  <Input placeholder="https://docs.example.com/api" />
                </Form.Item>
                <Form.Item name="isActive" valuePropName="checked">
                  <Checkbox>立即启用此接口</Checkbox>
                </Form.Item>
              </div>
            ),
          },
          {
            key: "pre-script",
            label: "预处理脚本",
            children: (
              <div className="mt-4 space-y-4">
                <Form.Item
                  name="preScript"
                  label="预处理脚本"
                  extra={
                    <>
                      <span>
                        在转发请求到上游前执行的脚本，可用于修改请求头、查询参数和请求体。
                      </span>
                      <br />
                      <span>
                        可用变量：<code>headers</code>（请求头）、
                        <code>query</code>（查询参数）、<code>body</code>
                        （请求体）
                      </span>
                    </>
                  }
                >
                  <MonacoEditor
                    language="javascript"
                    height="400px"
                    placeholder={preScriptPlaceholder}
                    disabled={isLoading}
                  />
                </Form.Item>
              </div>
            ),
          },
          {
            key: "post-script",
            label: "后处理脚本",
            children: (
              <div className="mt-4 space-y-4">
                <Form.Item
                  name="postScript"
                  label="后处理脚本"
                  extra={
                    <>
                      <span>
                        在收到上游响应后执行的脚本，可用于转换响应格式或过滤敏感数据。
                      </span>
                      <br />
                      <span>
                        可用变量：<code>responseBody</code>（响应体）、
                        <code>responseHeaders</code>（响应头）
                      </span>
                    </>
                  }
                >
                  <MonacoEditor
                    language="javascript"
                    height="400px"
                    placeholder={postScriptPlaceholder}
                    disabled={isLoading}
                  />
                </Form.Item>
              </div>
            ),
          },
        ]}
      />
      {error && <Alert type="error" message={error} showIcon />}
    </Form>
  );
}
