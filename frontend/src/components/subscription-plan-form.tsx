import { useState } from "react";
import { Alert, Card, Checkbox, Form, Input, InputNumber, Select } from "antd";
import { api } from "@/lib/api";

interface SubscriptionPlan {
  id?: string;
  title: string;
  price: number;
  totalCredits: number;
  sortOrder: number;
  validityDuration: number;
  validityUnit: string;
  creditResetCycle: string;
  isActive: boolean;
}
interface SubscriptionPlanFormProps {
  plan?: SubscriptionPlan;
  onSuccess: () => void;
  formId: string;
}
interface SubscriptionPlanValues {
  title: string;
  price: number;
  totalCredits: number;
  sortOrder?: number;
  validityDuration: number;
  validityUnit: string;
  creditResetCycle: string;
  isActive: boolean;
}

export function SubscriptionPlanForm({
  plan,
  onSuccess,
  formId,
}: SubscriptionPlanFormProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const isEdit = !!plan;

  async function handleFinish(values: SubscriptionPlanValues) {
    setIsLoading(true);
    setError(null);
    const body = { ...values, sortOrder: values.sortOrder ?? 0 };
    try {
      const result = isEdit
        ? await api.subscription_plans_id_route_put({ id: plan!.id! }, body)
        : await api.subscription_plans_route_post(body);
      if (result.success) onSuccess();
      else setError(result.error || "操作失败");
    } catch {
      setError("操作失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Form<SubscriptionPlanValues>
      id={formId}
      layout="vertical"
      onFinish={handleFinish}
      disabled={isLoading}
      initialValues={{
        title: plan?.title ?? "",
        price: plan?.price ?? 0,
        totalCredits: plan?.totalCredits ?? 0,
        sortOrder: plan?.sortOrder ?? 0,
        validityDuration: plan?.validityDuration ?? 30,
        validityUnit: plan?.validityUnit ?? "day",
        creditResetCycle: plan?.creditResetCycle ?? "month",
        isActive: plan?.isActive ?? true,
      }}
    >
      <Card styles={{ body: { padding: 24 } }}>
        <div className="grid gap-4 md:grid-cols-2">
          <Form.Item
            name="title"
            label="套餐标题"
            rules={[{ required: true, message: "请输入套餐标题" }]}
          >
            <Input placeholder="例如：基础版、专业版" />
          </Form.Item>
          <Form.Item
            name="price"
            label="价格"
            rules={[{ required: true, message: "请输入价格" }]}
          >
            <InputNumber
              min={0}
              step={0.01}
              className="w-full"
              placeholder="例如：99.99"
            />
          </Form.Item>
          <Form.Item
            name="totalCredits"
            label="总额度"
            rules={[{ required: true, message: "请输入总额度" }]}
          >
            <InputNumber min={0} className="w-full" placeholder="例如：1000" />
          </Form.Item>
          <Form.Item name="sortOrder" label="排序">
            <InputNumber
              min={0}
              className="w-full"
              placeholder="数字越小越靠前"
            />
          </Form.Item>
          <Form.Item
            name="validityDuration"
            label="有效期时长"
            rules={[{ required: true, message: "请输入有效期时长" }]}
          >
            <InputNumber min={1} className="w-full" placeholder="例如：30" />
          </Form.Item>
          <Form.Item
            name="validityUnit"
            label="有效期单位"
            rules={[{ required: true, message: "请选择有效期单位" }]}
          >
            <Select
              options={[
                { value: "day", label: "天" },
                { value: "week", label: "周" },
                { value: "month", label: "月" },
                { value: "year", label: "年" },
              ]}
            />
          </Form.Item>
        </div>
        <Form.Item
          name="creditResetCycle"
          label="额度重置周期"
          rules={[{ required: true, message: "请选择额度重置周期" }]}
        >
          <Select
            options={[
              { value: "day", label: "每天" },
              { value: "week", label: "每周" },
              { value: "month", label: "每月" },
              { value: "year", label: "每年" },
            ]}
          />
        </Form.Item>
        <Form.Item name="isActive" valuePropName="checked">
          <Checkbox>启用</Checkbox>
        </Form.Item>
        {error && <Alert type="error" message={error} showIcon />}
      </Card>
    </Form>
  );
}
